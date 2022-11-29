package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/textproto"
	"testing"
	"time"

	"github.com/gophish/gophish/config"

	"github.com/gophish/gomail"
	"github.com/jordan-wright/email"
	"gopkg.in/check.v1"
)

func (s *ModelsSuite) emailFromFirstMailLog(campaign Campaign, ch *check.C) *email.Email {
	result := campaign.Results[0]
	m := &MailLog{}
	err := db.Where("r_id=? AND campaign_id=?", result.RId, campaign.Id).
		Find(m).Error
	ch.Assert(err, check.Equals, nil)

	msg := gomail.NewMessage()
	err = m.Generate(msg)
	ch.Assert(err, check.Equals, nil)

	msgBuff := &bytes.Buffer{}
	_, err = msg.WriteTo(msgBuff)
	ch.Assert(err, check.Equals, nil)

	got, err := email.NewEmailFromReader(msgBuff)
	ch.Assert(err, check.Equals, nil)
	return got
}

func (s *ModelsSuite) TestGetQueuedMailLogs(ch *check.C) {
	campaign := s.createCampaign(ch)
	// By default, for campaigns with no launch date, the maillogs are set as
	// being processed. We need to unlock them first.
	ms, err := GetMailLogsByCampaign(campaign.Id)
	ch.Assert(err, check.Equals, nil)
	err = LockMailLogs(ms, false)
	ch.Assert(err, check.Equals, nil)
	ms, err = GetQueuedMailLogs(campaign.LaunchDate)
	ch.Assert(err, check.Equals, nil)
	got := make(map[string]*MailLog)
	for _, m := range ms {
		got[m.RId] = m
	}
	for _, r := range campaign.Results {
		if m, ok := got[r.RId]; ok {
			ch.Assert(m.RId, check.Equals, r.RId)
			ch.Assert(m.CampaignId, check.Equals, campaign.Id)
			ch.Assert(m.SendDate, check.Equals, campaign.LaunchDate)
			ch.Assert(m.UserId, check.Equals, campaign.UserId)
			ch.Assert(m.SendAttempt, check.Equals, 0)
		} else {
			ch.Fatalf("Result not found in maillogs: %s", r.RId)
		}
	}
}

func (s *ModelsSuite) TestMailLogBackoff(ch *check.C) {
	campaign := s.createCampaign(ch)
	result := campaign.Results[0]
	m := &MailLog{}
	err := db.Where("r_id=? AND campaign_id=?", result.RId, campaign.Id).
		Find(m).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(m.SendAttempt, check.Equals, 0)
	ch.Assert(m.SendDate, check.Equals, campaign.LaunchDate)

	expectedError := &textproto.Error{
		Code: 500,
		Msg:  "Recipient not found",
	}
	for i := m.SendAttempt; i < MaxSendAttempts; i++ {
		err = m.Lock()
		ch.Assert(err, check.Equals, nil)
		ch.Assert(m.Processing, check.Equals, true)

		expectedDuration := math.Pow(2, float64(m.SendAttempt+1))
		expectedSendDate := m.SendDate.Add(time.Minute * time.Duration(expectedDuration))
		err = m.Backoff(expectedError)
		ch.Assert(err, check.Equals, nil)
		ch.Assert(m.SendDate, check.Equals, expectedSendDate)
		ch.Assert(m.Processing, check.Equals, false)
		result, err := GetResult(m.RId)
		ch.Assert(err, check.Equals, nil)
		ch.Assert(result.SendDate, check.Equals, expectedSendDate)
		ch.Assert(result.Status, check.Equals, StatusRetry)
	}
	// Get our updated campaign and check for the added event
	campaign, err = GetCampaign(campaign.Id, int64(1))
	ch.Assert(err, check.Equals, nil)

	// We expect MaxSendAttempts + the initial campaign created event
	ch.Assert(len(campaign.Events), check.Equals, MaxSendAttempts+1)

	// Check that we receive our error after meeting the maximum send attempts
	err = m.Backoff(expectedError)
	ch.Assert(err, check.Equals, ErrMaxSendAttempts)
}

func (s *ModelsSuite) TestMailLogError(ch *check.C) {
	campaign := s.createCampaign(ch)
	result := campaign.Results[0]
	m := &MailLog{}
	err := db.Where("r_id=? AND campaign_id=?", result.RId, campaign.Id).
		Find(m).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(m.RId, check.Equals, result.RId)

	expectedError := &textproto.Error{
		Code: 500,
		Msg:  "Recipient not found",
	}
	err = m.Error(expectedError)
	ch.Assert(err, check.Equals, nil)

	// Get our result and make sure the status is set correctly
	result, err = GetResult(result.RId)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(result.Status, check.Equals, Error)

	// Get our updated campaign and check for the added event
	campaign, err = GetCampaign(campaign.Id, int64(1))
	ch.Assert(err, check.Equals, nil)

	expectedEventLength := 2
	ch.Assert(len(campaign.Events), check.Equals, expectedEventLength)

	gotEvent := campaign.Events[1]
	es := EventError{Error: expectedError.Error()}
	ej, _ := json.Marshal(es)
	expectedEvent := Event{
		Id:         gotEvent.Id,
		Email:      result.Email,
		Message:    EventSendingError,
		CampaignId: campaign.Id,
		Details:    string(ej),
		Time:       gotEvent.Time,
	}
	ch.Assert(gotEvent, check.DeepEquals, expectedEvent)

	ms, err := GetMailLogsByCampaign(campaign.Id)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(ms), check.Equals, len(campaign.Results)-1)
}

func (s *ModelsSuite) TestMailLogSuccess(ch *check.C) {
	campaign := s.createCampaign(ch)
	result := campaign.Results[0]
	m := &MailLog{}
	err := db.Where("r_id=? AND campaign_id=?", result.RId, campaign.Id).
		Find(m).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(m.RId, check.Equals, result.RId)

	err = m.Success()
	ch.Assert(err, check.Equals, nil)

	// Get our result and make sure the status is set correctly
	result, err = GetResult(result.RId)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(result.Status, check.Equals, EventSent)

	// Get our updated campaign and check for the added event
	campaign, err = GetCampaign(campaign.Id, int64(1))
	ch.Assert(err, check.Equals, nil)

	expectedEventLength := 2
	ch.Assert(len(campaign.Events), check.Equals, expectedEventLength)

	gotEvent := campaign.Events[1]
	expectedEvent := Event{
		Id:         gotEvent.Id,
		Email:      result.Email,
		Message:    EventSent,
		CampaignId: campaign.Id,
		Time:       gotEvent.Time,
	}
	ch.Assert(gotEvent, check.DeepEquals, expectedEvent)
	ch.Assert(result.SendDate, check.Equals, gotEvent.Time)

	ms, err := GetMailLogsByCampaign(campaign.Id)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(len(ms), check.Equals, len(campaign.Results)-1)
}

func (s *ModelsSuite) TestGenerateMailLog(ch *check.C) {
	campaign := Campaign{
		Id:     1,
		UserId: 1,
	}
	result := Result{
		RId: "abc1234",
	}
	err := GenerateMailLog(&campaign, &result, campaign.LaunchDate)
	ch.Assert(err, check.Equals, nil)

	m := MailLog{}
	err = db.Where("r_id=?", result.RId).Find(&m).Error
	ch.Assert(err, check.Equals, nil)
	ch.Assert(m.RId, check.Equals, result.RId)
	ch.Assert(m.CampaignId, check.Equals, campaign.Id)
	ch.Assert(m.SendDate, check.Equals, campaign.LaunchDate)
	ch.Assert(m.UserId, check.Equals, campaign.UserId)
	ch.Assert(m.SendAttempt, check.Equals, 0)
	ch.Assert(m.Processing, check.Equals, false)
}

func (s *ModelsSuite) TestMailLogGetSmtpFrom(ch *check.C) {
	template := Template{
		Name:           "OverrideSmtpFrom",
		UserId:         1,
		Text:           "dummytext",
		HTML:           "Dummyhtml",
		Subject:        "Dummysubject",
		EnvelopeSender: "spoofing@example.com",
	}
	ch.Assert(PostTemplate(&template), check.Equals, nil)
	campaign := s.createCampaignDependencies(ch)
	campaign.Template = template

	ch.Assert(PostCampaign(&campaign, campaign.UserId), check.Equals, nil)
	result := campaign.Results[0]

	m := &MailLog{}
	err := db.Where("r_id=? AND campaign_id=?", result.RId, campaign.Id).
		Find(m).Error
	ch.Assert(err, check.Equals, nil)

	msg := gomail.NewMessage()
	err = m.Generate(msg)
	ch.Assert(err, check.Equals, nil)

	msgBuff := &bytes.Buffer{}
	_, err = msg.WriteTo(msgBuff)
	ch.Assert(err, check.Equals, nil)

	got, err := email.NewEmailFromReader(msgBuff)
	ch.Assert(err, check.Equals, nil)
	ch.Assert(got.From, check.Equals, "spoofing@example.com")
}

func (s *ModelsSuite) TestMailLogGenerate(ch *check.C) {
	campaign := s.createCampaign(ch)
	result := campaign.Results[0]
	expected := &email.Email{
		From:    "test@test.com", // Default smtp.FromAddress
		Subject: fmt.Sprintf("%s - Subject", result.RId),
		Text:    []byte(fmt.Sprintf("%s - Text", result.RId)),
		HTML:    []byte(fmt.Sprintf("%s - HTML", result.RId)),
	}
	got := s.emailFromFirstMailLog(campaign, ch)
	ch.Assert(got.From, check.Equals, expected.From)
	ch.Assert(got.Subject, check.Equals, expected.Subject)
	ch.Assert(string(got.Text), check.Equals, string(expected.Text))
	ch.Assert(string(got.HTML), check.Equals, string(expected.HTML))
}

func (s *ModelsSuite) TestMailLogGenerateTransparencyHeaders(ch *check.C) {
	s.config.ContactAddress = "test@test.com"
	expectedHeaders := map[string]string{
		"X-Mailer":          config.ServerName,
		"X-Gophish-Contact": s.config.ContactAddress,
	}
	campaign := s.createCampaign(ch)
	got := s.emailFromFirstMailLog(campaign, ch)
	for k, v := range expectedHeaders {
		ch.Assert(got.Headers.Get(k), check.Equals, v)
	}
}

func (s *ModelsSuite) TestMailLogGenerateOverrideTransparencyHeaders(ch *check.C) {
	expectedHeaders := map[string]string{
		"X-Mailer":          "",
		"X-Gophish-Contact": "",
	}
	smtp := SMTP{
		Name:        "Test SMTP",
		Host:        "1.1.1.1:25",
		FromAddress: "foo@example.com",
		UserId:      1,
		Headers: []Header{
			Header{Key: "X-Gophish-Contact", Value: ""},
			Header{Key: "X-Mailer", Value: ""},
		},
	}
	ch.Assert(PostSMTP(&smtp), check.Equals, nil)
	campaign := s.createCampaignDependencies(ch)
	campaign.SMTP = smtp

	ch.Assert(PostCampaign(&campaign, campaign.UserId), check.Equals, nil)
	got := s.emailFromFirstMailLog(campaign, ch)
	for k, v := range expectedHeaders {
		ch.Assert(got.Headers.Get(k), check.Equals, v)
	}
}

func (s *ModelsSuite) TestUnlockAllMailLogs(ch *check.C) {
	campaign := s.createCampaign(ch)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	ch.Assert(err, check.Equals, nil)
	for _, m := range ms {
		ch.Assert(m.Processing, check.Equals, true)
	}
	err = UnlockAllMailLogs()
	ch.Assert(err, check.Equals, nil)
	ms, err = GetMailLogsByCampaign(campaign.Id)
	ch.Assert(err, check.Equals, nil)
	for _, m := range ms {
		ch.Assert(m.Processing, check.Equals, false)
	}
}

func (s *ModelsSuite) TestURLTemplateRendering(ch *check.C) {
	template := Template{
		Name:    "URLTemplate",
		UserId:  1,
		Text:    "{{.URL}}",
		HTML:    "{{.URL}}",
		Subject: "{{.URL}}",
	}
	ch.Assert(PostTemplate(&template), check.Equals, nil)
	campaign := s.createCampaignDependencies(ch)
	campaign.URL = "http://127.0.0.1/{{.Email}}/"
	campaign.Template = template

	ch.Assert(PostCampaign(&campaign, campaign.UserId), check.Equals, nil)
	result := campaign.Results[0]
	expectedURL := fmt.Sprintf("http://127.0.0.1/%s/?%s=%s", result.Email, RecipientParameter, result.RId)

	got := s.emailFromFirstMailLog(campaign, ch)
	ch.Assert(got.Subject, check.Equals, expectedURL)
	ch.Assert(string(got.Text), check.Equals, expectedURL)
	ch.Assert(string(got.HTML), check.Equals, expectedURL)
}

func (s *ModelsSuite) TestMailLogGenerateEmptySubject(ch *check.C) {

	// in place of using createCampaign, we replicate its small code body
	// here internally as we want to specify an empty subject to createCampaignDependencies
	// campaign := s.createCampaign(ch)
	campaign := s.createCampaignDependencies(ch, "") // specify empty subject
	// Setup and "launch" our campaign
	ch.Assert(PostCampaign(&campaign, campaign.UserId), check.Equals, nil)
	result := campaign.Results[0]

	expected := &email.Email{
		Subject: "",
		Text:    []byte(fmt.Sprintf("%s - Text", result.RId)),
		HTML:    []byte(fmt.Sprintf("%s - HTML", result.RId)),
	}
	got := s.emailFromFirstMailLog(campaign, ch)
	ch.Assert(got.Subject, check.Equals, expected.Subject)
}

func (s *ModelsSuite) TestShouldEmbedAttachment(ch *check.C) {

	// Supported file extensions
	ch.Assert(shouldEmbedAttachment(".png"), check.Equals, true)
	ch.Assert(shouldEmbedAttachment(".jpg"), check.Equals, true)
	ch.Assert(shouldEmbedAttachment(".jpeg"), check.Equals, true)
	ch.Assert(shouldEmbedAttachment(".gif"), check.Equals, true)

	// Some other file extensions
	ch.Assert(shouldEmbedAttachment(".docx"), check.Equals, false)
	ch.Assert(shouldEmbedAttachment(".txt"), check.Equals, false)
	ch.Assert(shouldEmbedAttachment(".jar"), check.Equals, false)
	ch.Assert(shouldEmbedAttachment(".exe"), check.Equals, false)

	// Invalid input
	ch.Assert(shouldEmbedAttachment(""), check.Equals, false)
	ch.Assert(shouldEmbedAttachment("png"), check.Equals, false)
}

func (s *ModelsSuite) TestEmbedAttachment(ch *check.C) {
	campaign := s.createCampaignDependencies(ch)
	campaign.Template.Attachments = []Attachment{
		{
			Name:    "test.png",
			Type:    "image/png",
			Content: "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEUAAACnej3aAAAAAXRSTlMAQObYZgAAAApJREFUCNdjYAAAAAIAAeIhvDMAAAAASUVORK5CYII=",
		},
		{
			Name:    "test.txt",
			Type:    "text/plain",
			Content: "VGVzdCB0ZXh0IGZpbGU=",
		},
	}
	PutTemplate(&campaign.Template)
	ch.Assert(PostCampaign(&campaign, campaign.UserId), check.Equals, nil)
	got := s.emailFromFirstMailLog(campaign, ch)

	// The email package simply ignores attachments where the Content-Disposition header is set
	// to inline, so the best we can do without replacing the whole thing is to check that only
	// the text file was added as an attachment.
	ch.Assert(got.Attachments, check.HasLen, 1)
	ch.Assert(got.Attachments[0].Filename, check.Equals, "test.txt")
}

func BenchmarkMailLogGenerate100(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 100)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	if err != nil {
		b.Fatalf("error getting maillogs for campaign: %v", err)
	}
	ms[0].CacheCampaign(&campaign)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := gomail.NewMessage()
		ms[0].Generate(msg)
	}
	tearDownBenchmark(b)
}

func BenchmarkMailLogGenerate1000(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 1000)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	if err != nil {
		b.Fatalf("error getting maillogs for campaign: %v", err)
	}
	ms[0].CacheCampaign(&campaign)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := gomail.NewMessage()
		ms[0].Generate(msg)
	}
	tearDownBenchmark(b)
}

func BenchmarkMailLogGenerate5000(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 5000)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	if err != nil {
		b.Fatalf("error getting maillogs for campaign: %v", err)
	}
	ms[0].CacheCampaign(&campaign)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := gomail.NewMessage()
		ms[0].Generate(msg)
	}
	tearDownBenchmark(b)
}

func BenchmarkMailLogGenerate10000(b *testing.B) {
	setupBenchmark(b)
	campaign := setupCampaign(b, 10000)
	ms, err := GetMailLogsByCampaign(campaign.Id)
	if err != nil {
		b.Fatalf("error getting maillogs for campaign: %v", err)
	}
	ms[0].CacheCampaign(&campaign)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := gomail.NewMessage()
		ms[0].Generate(msg)
	}
	tearDownBenchmark(b)
}
