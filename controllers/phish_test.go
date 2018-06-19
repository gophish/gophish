package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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
	resp, err := http.Get(fmt.Sprintf("%s/track?%s=%s", ps.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	expected, err := ioutil.ReadFile("static/images/pixel.png")
	s.Nil(err)
	s.Equal(bytes.Compare(body, expected), 0)
}

func (s *ControllersSuite) reportedEmail(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/report?%s=%s", ps.URL, models.RecipientParameter, rid))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNoContent)
}

func (s *ControllersSuite) reportEmail404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/report?%s=%s", ps.URL, models.RecipientParameter, rid))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) openEmail404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/track?%s=%s", ps.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) clickLink(rid string, expectedHTML string) {
	resp, err := http.Get(fmt.Sprintf("%s/?%s=%s", ps.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	s.Nil(err)
	log.Printf("%s\n\n\n", body)
	s.Equal(bytes.Compare(body, []byte(expectedHTML)), 0)
}

func (s *ControllersSuite) clickLink404(rid string) {
	resp, err := http.Get(fmt.Sprintf("%s/?%s=%s", ps.URL, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)
}

func (s *ControllersSuite) transparencyRequest(r models.Result, rid, path string) {
	resp, err := http.Get(fmt.Sprintf("%s%s?%s=%s", ps.URL, path, models.RecipientParameter, rid))
	s.Nil(err)
	defer resp.Body.Close()
	s.Equal(resp.StatusCode, http.StatusOK)
	tr := &TransparencyResponse{}
	err = json.NewDecoder(resp.Body).Decode(tr)
	s.Nil(err)
	s.Equal(tr.ContactAddress, config.Conf.ContactAddress)
	s.Equal(tr.SendDate, r.SendDate)
	s.Equal(tr.Server, config.ServerName)
}

func (s *ControllersSuite) TestOpenedPhishingEmail() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.STATUS_SENDING)

	s.openEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	s.Equal(result.Status, models.EVENT_OPENED)
	s.Equal(lastEvent.Message, models.EVENT_OPENED)
	s.Equal(result.ModifiedDate, lastEvent.Time)
}

func (s *ControllersSuite) TestReportedPhishingEmail() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.STATUS_SENDING)

	s.reportedEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	s.Equal(result.Reported, true)
	s.Equal(lastEvent.Message, models.EVENT_REPORTED)
	s.Equal(result.ModifiedDate, lastEvent.Time)
}

func (s *ControllersSuite) TestClickedPhishingLinkAfterOpen() {
	campaign := s.getFirstCampaign()
	result := campaign.Results[0]
	s.Equal(result.Status, models.STATUS_SENDING)

	s.openEmail(result.RId)
	s.clickLink(result.RId, campaign.Page.HTML)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	lastEvent := campaign.Events[len(campaign.Events)-1]
	s.Equal(result.Status, models.EVENT_CLICKED)
	s.Equal(lastEvent.Message, models.EVENT_CLICKED)
	s.Equal(result.ModifiedDate, lastEvent.Time)
}

func (s *ControllersSuite) TestNoRecipientID() {
	resp, err := http.Get(fmt.Sprintf("%s/track", ps.URL))
	s.Nil(err)
	s.Equal(resp.StatusCode, http.StatusNotFound)

	resp, err = http.Get(ps.URL)
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
	s.Equal(result.Status, models.STATUS_SENDING)
	s.openEmail(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EVENT_OPENED)

	models.CompleteCampaign(campaign.Id, 1)
	s.openEmail404(result.RId)
	s.clickLink404(result.RId)

	campaign = s.getFirstCampaign()
	result = campaign.Results[0]
	s.Equal(result.Status, models.EVENT_OPENED)
}

func (s *ControllersSuite) TestRobotsHandler() {
	expected := []byte("User-agent: *\nDisallow: /\n")
	resp, err := http.Get(fmt.Sprintf("%s/robots.txt", ps.URL))
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
