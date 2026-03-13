package models

import (
	"time"

	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) TestGetCampaignRangeStatsSnapshot(c *check.C) {
	campaign := s.createCampaign(c)

	err := db.Where("campaign_id=?", campaign.Id).Delete(&Event{}).Error
	c.Assert(err, check.Equals, nil)

	baseTime := time.Date(2026, time.January, 10, 9, 0, 0, 0, time.UTC)
	results := campaign.Results

	results[0].Status = EventOpened
	results[0].ModifiedDate = baseTime.Add(1 * time.Hour)
	c.Assert(db.Save(&results[0]).Error, check.Equals, nil)

	results[1].Status = EventClicked
	results[1].ModifiedDate = baseTime.Add(2 * time.Hour)
	c.Assert(db.Save(&results[1]).Error, check.Equals, nil)

	results[2].Status = EventDataSubmit
	results[2].ModifiedDate = baseTime.Add(3 * time.Hour)
	results[2].Reported = true
	c.Assert(db.Save(&results[2]).Error, check.Equals, nil)

	results[3].Status = Error
	results[3].ModifiedDate = baseTime.Add(30 * time.Minute)
	c.Assert(db.Save(&results[3]).Error, check.Equals, nil)

	events := []Event{
		{CampaignId: campaign.Id, Email: results[0].Email, Time: baseTime, Message: EventSent},
		{CampaignId: campaign.Id, Email: results[0].Email, Time: baseTime.Add(1 * time.Hour), Message: EventOpened},
		{CampaignId: campaign.Id, Email: results[1].Email, Time: baseTime, Message: EventSent},
		{CampaignId: campaign.Id, Email: results[1].Email, Time: baseTime.Add(2 * time.Hour), Message: EventClicked},
		{CampaignId: campaign.Id, Email: results[2].Email, Time: baseTime, Message: EventSent},
		{CampaignId: campaign.Id, Email: results[2].Email, Time: baseTime.Add(3 * time.Hour), Message: EventDataSubmit},
		{CampaignId: campaign.Id, Email: results[2].Email, Time: baseTime.Add(4 * time.Hour), Message: EventReported},
		{CampaignId: campaign.Id, Email: results[3].Email, Time: baseTime.Add(30 * time.Minute), Message: EventSendingError},
	}
	for i := range events {
		c.Assert(db.Save(&events[i]).Error, check.Equals, nil)
	}

	end := baseTime.Add(150 * time.Minute)
	stats, err := GetCampaignRangeStats(campaign.Id, campaign.UserId, nil, &end, CampaignStatsModeSnapshot)
	c.Assert(err, check.Equals, nil)

	c.Assert(stats.Actual.Total, check.Equals, int64(4))
	c.Assert(stats.Actual.EmailsSent, check.Equals, int64(3))
	c.Assert(stats.Actual.OpenedEmail, check.Equals, int64(2))
	c.Assert(stats.Actual.ClickedLink, check.Equals, int64(1))
	c.Assert(stats.Actual.SubmittedData, check.Equals, int64(0))
	c.Assert(stats.Actual.EmailReported, check.Equals, int64(0))
	c.Assert(stats.Actual.Error, check.Equals, int64(1))

	c.Assert(stats.Dashboard.Total, check.Equals, int64(4))
	c.Assert(stats.Dashboard.EmailsSent, check.Equals, int64(3))
	c.Assert(stats.Dashboard.OpenedEmail, check.Equals, int64(3))
	c.Assert(stats.Dashboard.ClickedLink, check.Equals, int64(2))
	c.Assert(stats.Dashboard.SubmittedData, check.Equals, int64(1))
	c.Assert(stats.Dashboard.EmailReported, check.Equals, int64(1))
	c.Assert(stats.Dashboard.Error, check.Equals, int64(1))
}

func (s *ModelsSuite) TestGetCampaignRangeStatsRange(c *check.C) {
	campaign := s.createCampaign(c)

	err := db.Where("campaign_id=?", campaign.Id).Delete(&Event{}).Error
	c.Assert(err, check.Equals, nil)

	baseTime := time.Date(2026, time.January, 11, 9, 0, 0, 0, time.UTC)
	results := campaign.Results

	events := []Event{
		{CampaignId: campaign.Id, Email: results[0].Email, Time: baseTime, Message: EventSent},
		{CampaignId: campaign.Id, Email: results[0].Email, Time: baseTime.Add(1 * time.Hour), Message: EventOpened},
		{CampaignId: campaign.Id, Email: results[1].Email, Time: baseTime, Message: EventSent},
		{CampaignId: campaign.Id, Email: results[1].Email, Time: baseTime.Add(2 * time.Hour), Message: EventClicked},
		{CampaignId: campaign.Id, Email: results[2].Email, Time: baseTime, Message: EventSent},
		{CampaignId: campaign.Id, Email: results[2].Email, Time: baseTime.Add(3 * time.Hour), Message: EventDataSubmit},
		{CampaignId: campaign.Id, Email: results[2].Email, Time: baseTime.Add(4 * time.Hour), Message: EventReported},
		{CampaignId: campaign.Id, Email: results[3].Email, Time: baseTime.Add(30 * time.Minute), Message: EventSendingError},
	}
	for i := range events {
		c.Assert(db.Save(&events[i]).Error, check.Equals, nil)
	}

	start := baseTime.Add(90 * time.Minute)
	end := baseTime.Add(270 * time.Minute)
	stats, err := GetCampaignRangeStats(campaign.Id, campaign.UserId, &start, &end, CampaignStatsModeRange)
	c.Assert(err, check.Equals, nil)

	c.Assert(stats.Actual.Total, check.Equals, int64(4))
	c.Assert(stats.Actual.EmailsSent, check.Equals, int64(0))
	c.Assert(stats.Actual.OpenedEmail, check.Equals, int64(2))
	c.Assert(stats.Actual.ClickedLink, check.Equals, int64(2))
	c.Assert(stats.Actual.SubmittedData, check.Equals, int64(1))
	c.Assert(stats.Actual.EmailReported, check.Equals, int64(1))
	c.Assert(stats.Actual.Error, check.Equals, int64(0))
}
