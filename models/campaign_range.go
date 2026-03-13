package models

import (
	"errors"
	"time"
)

const (
	CampaignStatsModeSnapshot = "snapshot"
	CampaignStatsModeRange    = "range"
)

var ErrCampaignStatsInvalidMode = errors.New("invalid campaign stats mode")
var ErrCampaignStatsEndRequired = errors.New("end date required")
var ErrCampaignStatsStartRequired = errors.New("start date required for range mode")
var ErrCampaignStatsInvalidRange = errors.New("start date must be before end date")

// CampaignRangeStats contains both the current dashboard rollup and the
// event-derived stats for a requested historical window.
type CampaignRangeStats struct {
	CampaignID int64         `json:"campaign_id"`
	Mode       string        `json:"mode"`
	StartDate  *time.Time    `json:"start_date,omitempty"`
	EndDate    *time.Time    `json:"end_date,omitempty"`
	Actual     CampaignStats `json:"actual"`
	Dashboard  CampaignStats `json:"dashboard"`
}

type campaignEventState struct {
	Sent      *time.Time
	Opened    *time.Time
	Clicked   *time.Time
	Submitted *time.Time
	Reported  *time.Time
	Error     *time.Time
}

// GetCampaignRangeStats returns the current dashboard rollup and the
// event-derived stats for a requested historical window.
func GetCampaignRangeStats(id int64, uid int64, start, end *time.Time, mode string) (CampaignRangeStats, error) {
	stats := CampaignRangeStats{}
	mode = normalizeCampaignStatsMode(mode)
	if mode == "" {
		return stats, ErrCampaignStatsInvalidMode
	}
	if end == nil {
		if mode == CampaignStatsModeSnapshot {
			now := time.Now().UTC()
			end = &now
		} else {
			return stats, ErrCampaignStatsEndRequired
		}
	}
	if mode == CampaignStatsModeRange && start == nil {
		return stats, ErrCampaignStatsStartRequired
	}

	startUTC := cloneUTCTime(start)
	endUTC := cloneUTCTime(end)
	if startUTC != nil && endUTC != nil && startUTC.After(*endUTC) {
		return stats, ErrCampaignStatsInvalidRange
	}

	cr, err := GetCampaignResults(id, uid)
	if err != nil {
		return stats, err
	}

	stats.CampaignID = id
	stats.Mode = mode
	stats.StartDate = startUTC
	stats.EndDate = endUTC
	stats.Dashboard = summarizeCampaignResults(cr.Results)
	stats.Actual = summarizeCampaignEvents(cr.Results, cr.Events, startUTC, endUTC, mode)
	return stats, nil
}

func normalizeCampaignStatsMode(mode string) string {
	switch mode {
	case "", CampaignStatsModeSnapshot:
		return CampaignStatsModeSnapshot
	case CampaignStatsModeRange:
		return CampaignStatsModeRange
	default:
		return ""
	}
}

func cloneUTCTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	cloned := t.UTC()
	return &cloned
}

func summarizeCampaignResults(results []Result) CampaignStats {
	stats := CampaignStats{
		Total: int64(len(results)),
	}
	for _, result := range results {
		switch result.Status {
		case EventDataSubmit:
			stats.SubmittedData++
			stats.ClickedLink++
			stats.OpenedEmail++
			stats.EmailsSent++
		case EventClicked:
			stats.ClickedLink++
			stats.OpenedEmail++
			stats.EmailsSent++
		case EventOpened:
			stats.OpenedEmail++
			stats.EmailsSent++
		case EventSent:
			stats.EmailsSent++
		case Error:
			stats.Error++
		}
		if result.Reported {
			stats.EmailReported++
		}
	}
	return stats
}

func summarizeCampaignEvents(results []Result, events []Event, start, end *time.Time, mode string) CampaignStats {
	stats := CampaignStats{
		Total: int64(len(results)),
	}
	if end == nil {
		return stats
	}

	states := map[string]*campaignEventState{}
	for _, result := range results {
		states[result.Email] = &campaignEventState{}
	}
	for _, event := range events {
		if event.Email == "" {
			continue
		}
		state, ok := states[event.Email]
		if !ok {
			continue
		}
		switch event.Message {
		case EventSent:
			setEventTime(&state.Sent, event.Time)
		case EventOpened:
			setEventTime(&state.Opened, event.Time)
		case EventClicked:
			setEventTime(&state.Opened, event.Time)
			setEventTime(&state.Clicked, event.Time)
		case EventDataSubmit:
			setEventTime(&state.Opened, event.Time)
			setEventTime(&state.Clicked, event.Time)
			setEventTime(&state.Submitted, event.Time)
		case EventReported:
			setEventTime(&state.Reported, event.Time)
		case EventSendingError:
			setEventTime(&state.Error, event.Time)
		}
	}

	for _, state := range states {
		if stageMatchesWindow(state.Sent, start, end, mode) {
			stats.EmailsSent++
		}
		if stageMatchesWindow(state.Opened, start, end, mode) {
			stats.OpenedEmail++
		}
		if stageMatchesWindow(state.Clicked, start, end, mode) {
			stats.ClickedLink++
		}
		if stageMatchesWindow(state.Submitted, start, end, mode) {
			stats.SubmittedData++
		}
		if stageMatchesWindow(state.Reported, start, end, mode) {
			stats.EmailReported++
		}
		if stageMatchesWindow(state.Error, start, end, mode) {
			stats.Error++
		}
	}
	return stats
}

func setEventTime(target **time.Time, candidate time.Time) {
	if candidate.IsZero() {
		return
	}
	if *target == nil || candidate.Before(**target) {
		eventTime := candidate.UTC()
		*target = &eventTime
	}
}

func stageMatchesWindow(stageTime, start, end *time.Time, mode string) bool {
	if stageTime == nil || end == nil {
		return false
	}
	switch mode {
	case CampaignStatsModeRange:
		if start == nil {
			return false
		}
		return !stageTime.Before(*start) && !stageTime.After(*end)
	default:
		return !stageTime.After(*end)
	}
}
