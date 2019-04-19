package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/gophish/gophish/config"
	"github.com/gophish/gophish/models"
)

func (s *ControllersSuite) getFirstCampaign() models.Campaign {
	campaigns, err := models.GetCampaigns(1)
	s.Nil(err)
	return campaigns[0]
}

func (s *ControllersSuite) getFirstEmailRequest() models.EmailRequest {
	campaign := s.getFirstCampaign()
	req := models.EmailRequest{
		TemplateId:    campaign.TemplateId,
		Template:      campaign.Template,
		PageId:        campaign.PageId,
		Page:          campaign.Page,
		URL:           "http://localhost.localdomain",
		UserId:        1,
		BaseRecipient: campaign.Results[0].BaseRecipient,
		SMTP:          campaign.SMTP,
		FromAddress:   campaign.SMTP.FromAddress,
	}
	err := models.PostEmailRequest(&req)
	s.Nil(err)
	return req
}

func (s *ControllersSuite) openEmail(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/track?%s=%s", s.phishServer.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	expected, err := ioutil.ReadFile("static/images/pixel.png")
	s.Nil(err)
	s.Equal(bytes.Compare(body, expected), 0)
}

func (s *ControllersSuite) reportedEmail(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/report?%s=%s", s.phishServer.URL, models.RecipientParameter, rid))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNoContent)
}

func (s *ControllersSuite) reportEmail404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/report?%s=%s", s.phishServer.URL, models.RecipientParameter, rid))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) openEmail404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/track?%s=%s", s.phishServer.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) clickLink(rid string, expectedHTML string) {
	resp, err := http.Get(fmt.Sprintf("%s/?%s=%s", s.phishServer.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	log.Printf("%s\n\n\n", body)
	s.Equal(bytes.Compare(body, []byte(expectedHTML)), 0)
}

func (s *ControllersSuite) clickLink404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/?%s=%s", s.phishServer.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) transparencyRequest(r models.Result, rid, path string) {
	resp, err := http.Get(fmt.Sprintf("%s%s?%s=%s", s.phishServer.URL, path, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Equal(resp.StatusCode, http.StatusOK)
	tr := &TransparencyResponse{}
	err = json.NewDecoder(resp.Body).Decode(tr)
	s.Nil(err)
	s.Equal(tr.ContactAddress, s.config.ContactAddress)
	s.Equal(tr.SendDate, r.SendDate)
	s.Equal(tr.Server, config.ServerName)
}

func (s *ControllersSuite) TestOpenedPhishingEmail() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.StatusSending)

	s.openEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	s.Equal(result.Status, models.EventOpened)
	s.Equal(lastEvent.Message, models.EventOpened)
	s.Equal(result.ModifiedDate, lastEvent.Time)
}

func (s *ControllersSuite) TestReportedPhishingEmail() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.StatusSending)

	s.reportedEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	s.Equal(result.Reported, true)
	s.Equal(lastEvent.Message, models.EventReported)
	s.Equal(result.ModifiedDate, lastEvent.Time)
}

func (s *ControllersSuite) TestClickedPhishingLinkAfterOpen() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.StatusSending)

	s.openEmail(result.RId)
	s.clickLink(result.RId, campaign.Page.HTML)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	s.Equal(result.Status, models.EventClicked)
	s.Equal(lastEvent.Message, models.EventClicked)
	s.Equal(result.ModifiedDate, lastEvent.Time)
}

func (s *ControllersSuite) TestNoRecipientID() {
	resp, err := http.Get(fmt.Sprintf("%s/track", s.phishServer.URL))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)

	resp, err = http.Get(s.phishServer.URL)
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) TestInvalidRecipientID() {
	rid := "XXXXXXXXXX"
	s.openEmail404(rid)
	s.clickLink404(rid)
	s.reportEmail404(rid)
}

func (s *ControllersSuite) TestCompletedCampaignClick() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.StatusSending)
	s.openEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EventOpened)

	models.CompleteCampaign(campaign.Id, 1)
	s.openEmail404(result.RId)
	s.clickLink404(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EventOpened)
}

func (s *ControllersSuite) TestRobotsHandler() {
	expected := []byte("User-agent: *\nDisallow: /\n")
	resp, err := http.Get(fmt.Sprintf("%s/robots.txt", s.phishServer.URL))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusOK)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	s.Equal(bytes.Compare(body, expected), 0)
}

func (s *ControllersSuite) TestInvalidPreviewID() {
	bogusRId := fmt.Sprintf("%sbogus", models.PreviewPrefix)
	s.openEmail404(bogusRId)
	s.clickLink404(bogusRId)
	s.reportEmail404(bogusRId)
}

func (s *ControllersSuite) TestPreviewTrack() {
	req := s.getFirstEmailRequest()
	s.openEmail(req.RId)
}

func (s *ControllersSuite) TestPreviewClick() {
	req := s.getFirstEmailRequest()
	s.clickLink(req.RId, req.Page.HTML)
}

func (s *ControllersSuite) TestInvalidTransparencyRequest() {
	bogusRId := fmt.Sprintf("bogus%s", TransparencySuffix)
	s.openEmail404(bogusRId)
	s.clickLink404(bogusRId)
	s.reportEmail404(bogusRId)
}

func (s *ControllersSuite) TestTransparencyRequest() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	rid := fmt.Sprintf("%s%s", result.RId, TransparencySuffix)
	s.transparencyRequest(result, rid, "/")
	s.transparencyRequest(result, rid, "/track")
	s.transparencyRequest(result, rid, "/report")

	// And check with the URL encoded version of a +
	rid = fmt.Sprintf("%s%s", result.RId, "%2b")
	s.transparencyRequest(result, rid, "/")
	s.transparencyRequest(result, rid, "/track")
	s.transparencyRequest(result, rid, "/report")
}

func (s *ControllersSuite) TestRedirectTemplating() {
	p := models.Page{
		Name:        "Redirect Page",
		HTML:        "<html>Test</html>",
		UserId:      1,
		RedirectURL: "http://example.com/{{.RId}}",
	}
	err := models.PostPage(&p)
	s.Nil(err)
	smtp, _ := models.GetSMTP(1, 1)
	template, _ := models.GetTemplate(1, 1)
	group, _ := models.GetGroup(1, 1)

	campaign := models.Campaign{Name: "Redirect campaign"}
	campaign.UserId = 1
	campaign.Template = template
	campaign.Page = p
	campaign.SMTP = smtp
	campaign.Groups = []models.Group{group}
	err = models.PostCampaign(&campaign, campaign.UserId)
	s.Nil(err)

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	result := campaign.Results[0]
	resp, err := client.PostForm(fmt.Sprintf("%s/?%s=%s", s.phishServer.URL, models.RecipientParameter, result.RId), url.Values{"username": {"test"}, "password": {"test"}})
	s.Nil(err)
	defer resp.Body.Close()
	s.Equal(http.StatusFound, resp.StatusCode)
	expectedURL := fmt.Sprintf("http://example.com/%s", result.RId)
	got, err := resp.Location()
	s.Nil(err)
	s.Equal(expectedURL, got.String())
}
