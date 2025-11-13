package audit

import (
	"encoding/json"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/sirupsen/logrus"
)

// EventType represents different audit event types
type EventType string

const (
	// Authentication events
	EventLogin              EventType = "auth.login"
	EventLoginFailed        EventType = "auth.login_failed"
	EventLogout             EventType = "auth.logout"
	EventPasswordChange     EventType = "auth.password_change"
	EventPasswordReset      EventType = "auth.password_reset"
	EventAPIKeyUsed         EventType = "auth.api_key_used"
	EventAPIKeyInvalid      EventType = "auth.api_key_invalid"

	// User management events
	EventUserCreated        EventType = "user.created"
	EventUserDeleted        EventType = "user.deleted"
	EventUserModified       EventType = "user.modified"

	// Campaign events
	EventCampaignCreated    EventType = "campaign.created"
	EventCampaignLaunched   EventType = "campaign.launched"
	EventCampaignDeleted    EventType = "campaign.deleted"
	EventCampaignCompleted  EventType = "campaign.completed"

	// Configuration events
	EventConfigChanged      EventType = "config.changed"

	// Security events
	EventPermissionDenied   EventType = "permission.denied"
	EventRateLimitExceeded  EventType = "rate_limit.exceeded"
	EventSuspiciousActivity EventType = "security.suspicious"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   EventType              `json:"event_type"`
	UserID      int64                  `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	IPAddress   string                 `json:"ip_address"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// Log records an audit event
func Log(event AuditEvent) {
	event.Timestamp = time.Now().UTC()

	// Marshal to JSON for structured logging
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Errorf("Failed to marshal audit event: %v", err)
		return
	}

	// Log to structured logger with audit flag
	log.Logger.WithFields(logrus.Fields{
		"audit":      true,
		"event_type": event.EventType,
		"user_id":    event.UserID,
		"success":    event.Success,
	}).Info(string(eventJSON))

	// Optional: Send to SIEM/log aggregation service
	// sendToSIEM(eventJSON)
}

// LogLogin records a login attempt
func LogLogin(username, ip, userAgent string, success bool, reason string) {
	Log(AuditEvent{
		EventType: EventLogin,
		Username:  username,
		IPAddress: ip,
		UserAgent: userAgent,
		Success:   success,
		Message:   reason,
	})
}

// LogLoginFailed records a failed login attempt
func LogLoginFailed(username, ip, userAgent, reason string) {
	Log(AuditEvent{
		EventType: EventLoginFailed,
		Username:  username,
		IPAddress: ip,
		UserAgent: userAgent,
		Success:   false,
		Message:   reason,
	})
}

// LogPasswordChange records a password change
func LogPasswordChange(userID int64, username, ip string, success bool) {
	Log(AuditEvent{
		EventType: EventPasswordChange,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   success,
		Message:   "Password changed",
	})
}

// LogPasswordReset records a password reset
func LogPasswordReset(userID int64, username, ip string, success bool) {
	Log(AuditEvent{
		EventType: EventPasswordReset,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   success,
		Message:   "Password reset",
	})
}

// LogAPIKeyUsed records API key usage
func LogAPIKeyUsed(userID int64, username, ip, endpoint string) {
	Log(AuditEvent{
		EventType: EventAPIKeyUsed,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   true,
		Message:   "API key authenticated",
		Details: map[string]interface{}{
			"endpoint": endpoint,
		},
	})
}

// LogAPIKeyInvalid records an invalid API key attempt
func LogAPIKeyInvalid(ip, endpoint string) {
	Log(AuditEvent{
		EventType: EventAPIKeyInvalid,
		IPAddress: ip,
		Success:   false,
		Message:   "Invalid API key attempted",
		Details: map[string]interface{}{
			"endpoint": endpoint,
		},
	})
}

// LogUserCreated records user creation
func LogUserCreated(adminUserID int64, adminUsername, newUsername, ip string) {
	Log(AuditEvent{
		EventType: EventUserCreated,
		UserID:    adminUserID,
		Username:  adminUsername,
		IPAddress: ip,
		Success:   true,
		Message:   "User created",
		Details: map[string]interface{}{
			"new_username": newUsername,
		},
	})
}

// LogUserDeleted records user deletion
func LogUserDeleted(adminUserID int64, adminUsername, deletedUsername, ip string) {
	Log(AuditEvent{
		EventType: EventUserDeleted,
		UserID:    adminUserID,
		Username:  adminUsername,
		IPAddress: ip,
		Success:   true,
		Message:   "User deleted",
		Details: map[string]interface{}{
			"deleted_username": deletedUsername,
		},
	})
}

// LogCampaignCreated records campaign creation
func LogCampaignCreated(userID int64, username string, campaignID int64, campaignName, ip string) {
	Log(AuditEvent{
		EventType: EventCampaignCreated,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   true,
		Message:   "Campaign created",
		Details: map[string]interface{}{
			"campaign_id":   campaignID,
			"campaign_name": campaignName,
		},
	})
}

// LogCampaignLaunched records campaign launch
func LogCampaignLaunched(userID int64, username string, campaignID int64, campaignName, ip string) {
	Log(AuditEvent{
		EventType: EventCampaignLaunched,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   true,
		Message:   "Campaign launched",
		Details: map[string]interface{}{
			"campaign_id":   campaignID,
			"campaign_name": campaignName,
		},
	})
}

// LogCampaignDeleted records campaign deletion
func LogCampaignDeleted(userID int64, username string, campaignID int64, ip string) {
	Log(AuditEvent{
		EventType: EventCampaignDeleted,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   true,
		Message:   "Campaign deleted",
		Details: map[string]interface{}{
			"campaign_id": campaignID,
		},
	})
}

// LogPermissionDenied records a permission denial
func LogPermissionDenied(userID int64, username, ip, resource, action string) {
	Log(AuditEvent{
		EventType: EventPermissionDenied,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   false,
		Message:   "Permission denied",
		Details: map[string]interface{}{
			"resource": resource,
			"action":   action,
		},
	})
}

// LogRateLimitExceeded records rate limit violations
func LogRateLimitExceeded(ip, endpoint string) {
	Log(AuditEvent{
		EventType: EventRateLimitExceeded,
		IPAddress: ip,
		Success:   false,
		Message:   "Rate limit exceeded",
		Details: map[string]interface{}{
			"endpoint": endpoint,
		},
	})
}

// LogSuspiciousActivity records suspicious behavior
func LogSuspiciousActivity(userID int64, username, ip, activity string, details map[string]interface{}) {
	Log(AuditEvent{
		EventType: EventSuspiciousActivity,
		UserID:    userID,
		Username:  username,
		IPAddress: ip,
		Success:   false,
		Message:   activity,
		Details:   details,
	})
}
