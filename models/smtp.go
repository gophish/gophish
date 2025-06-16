package models

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gophish/gomail"
	"github.com/gophish/gophish/dialer"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"github.com/jinzhu/gorm"
)

// Dialer is a wrapper around a standard gomail.Dialer in order
// to implement the mailer.Dialer interface. This allows us to better
// separate the mailer package as opposed to forcing a connection
// between mailer and gomail.
type Dialer struct {
	*gomail.Dialer
}

// Dial wraps the gomail dialer's Dial command
func (d *Dialer) Dial() (mailer.Sender, error) {
	return d.Dialer.Dial()
}

// SMTP contains the attributes needed to handle the sending of campaign emails
type SMTP struct {
	Id               int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId           int64     `json:"-" gorm:"column:user_id"`
	Interface        string    `json:"interface_type" gorm:"column:interface_type"`
	Name             string    `json:"name"`
	Host             string    `json:"host"`
	Username         string    `json:"username,omitempty"`
	Password         string    `json:"password,omitempty"`
	FromAddress      string    `json:"from_address"`
	IgnoreCertErrors bool      `json:"ignore_cert_errors"`
	Headers          []Header  `json:"headers"`
	ModifiedDate     time.Time `json:"modified_date"`
	// Graph API fields - reference to app_registration
	AppRegistrationID string    `json:"app_registration_id" gorm:"column:app_registration_id"`
	// Temporary fields for Graph API registration - not stored in database
	ClientID         string `json:"client_id,omitempty" gorm:"-"`
	ClientSecret     string `json:"client_secret,omitempty" gorm:"-"`
	TenantID         string `json:"tenant_id,omitempty" gorm:"-"`
	ProviderTenant   *ProviderTenant `json:"-" gorm:"-"` // For passing provider tenant from context
}

// Header contains the fields and methods for a sending profile to have
// custom headers
type Header struct {
	Id     int64  `json:"-"`
	SMTPId int64  `json:"-"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}

// ErrFromAddressNotSpecified is thrown when there is no "From" address
// specified in the SMTP configuration
var ErrFromAddressNotSpecified = errors.New("No From Address specified")

// ErrInvalidFromAddress is thrown when the SMTP From field in the sending
// profiles containes a value that is not an email address
var ErrInvalidFromAddress = errors.New("Invalid SMTP From address because it is not an email address")

// ErrHostNotSpecified is thrown when there is no Host specified
// in the SMTP configuration
var ErrHostNotSpecified = errors.New("No SMTP Host specified")

// ErrInvalidHost indicates that the SMTP server string is invalid
var ErrInvalidHost = errors.New("Invalid SMTP server address")

// TableName specifies the database tablename for Gorm to use
func (s SMTP) TableName() string {
	return "smtp"
}

// GetProviderTenantByType returns a provider tenant for the given tenant ID and provider type
func GetProviderTenantByType(tenantID string, providerType ProviderType) (*ProviderTenant, error) {
	var providerTenant ProviderTenant
	err := db.Where("tenant_id = ? AND provider_type = ?", tenantID, providerType).First(&providerTenant).Error
	if err != nil {
		return nil, err
	}
	return &providerTenant, nil
}

// Validate ensures that SMTP configs/connections are valid
func (s *SMTP) Validate() error {
	// Debug logging
	log.Infof("Validating SMTP profile - Interface: %s, Host: %s, FromAddress: %s, User ID: %d", 
		s.Interface, s.Host, s.FromAddress, s.UserId)

	// Ensure we have a valid user ID
	if s.UserId == 0 {
		// If we have an ID, try to get the user ID from the database
		if s.Id != 0 {
			var existingSmtp SMTP
			err := db.Where("id = ?", s.Id).First(&existingSmtp).Error
			if err != nil {
				log.Errorf("Failed to get existing SMTP profile: %v", err)
				return fmt.Errorf("failed to get user ID: SMTP profile not found")
			}
			s.UserId = existingSmtp.UserId
			log.Infof("Retrieved user ID %d from existing SMTP profile", s.UserId)
		} else {
			return fmt.Errorf("user ID is required for validation")
		}
	}

	// Always check for From Address
	if s.FromAddress == "" {
		return ErrFromAddressNotSpecified
	}

	// For Graph API interface, validate app registration
	if s.Interface == "GRAPH" {
		log.Infof("Validating Graph API interface for user %d", s.UserId)
		
		// If we have a provider tenant from context, use it
		if s.ProviderTenant != nil {
			log.Infof("Using provider tenant from context: %s (%s) for user %d", 
				s.ProviderTenant.DisplayName, s.ProviderTenant.ID, s.UserId)

			// If we have an app registration ID, validate through that
			if s.AppRegistrationID != "" {
				// Get and validate the app registration
				appReg, err := GetAppRegistration(s.AppRegistrationID)
				if err != nil {
					return fmt.Errorf("Invalid App Registration: %v", err)
				}
				
				// Create a GraphAPI instance with the app registration details
				g := GraphAPI{
					FromAddress:    s.FromAddress,
					ClientID:       appReg.ClientID,
					ClientSecret:   appReg.ClientSecretEncrypted,
					InterfaceType: s.Interface,
					UserId:        s.UserId,
					ProviderTenant: s.ProviderTenant,  // Pass provider tenant from context
				}
				
				return g.Validate()
			}

			// For new profiles without app registration
			if s.ClientID == "" || s.ClientSecret == "" {
				return errors.New("client_id and client_secret are required for Graph API")
			}
			
			// Create a new app registration with the tenant ID from the provider tenant
			appReg := &AppRegistration{
				ProviderTenantID:      s.ProviderTenant.ID,
				ClientID:              s.ClientID,
				ClientSecretEncrypted: s.ClientSecret,
				RedirectURI:           "https://localhost:3333",
			}
			
			// Set default Graph API scopes
			appReg.SetScopes([]string{
				"https://graph.microsoft.com/Mail.Send",
				"https://graph.microsoft.com/Mail.Send.Shared",
				"https://graph.microsoft.com/User.Read",
			})
			
			// Create the app registration
			err := appReg.Create()
			if err != nil {
				return fmt.Errorf("failed to create app registration: %v", err)
			}
			
			// Set the app registration ID in the SMTP profile
			s.AppRegistrationID = appReg.ID

			// Create a GraphAPI instance for validation
			g := GraphAPI{
				FromAddress:    s.FromAddress,
				ClientID:       s.ClientID,
				ClientSecret:   s.ClientSecret,
				InterfaceType: s.Interface,
				UserId:        s.UserId,
				ProviderTenant: s.ProviderTenant,  // Pass provider tenant from context
			}
			
			return g.Validate()
		} else {
			return fmt.Errorf("provider tenant context is required for Graph API validation")
		}
	}

	// SMTP-specific validation
	if s.Host == "" {
		return ErrHostNotSpecified
	}
	
	if !validateFromAddress(s.FromAddress) {
		return ErrInvalidFromAddress
	}
	
	_, err := mail.ParseAddress(s.FromAddress)
	if err != nil {
		return err
	}
	
	// Make sure addr is in host:port format
	hp := strings.Split(s.Host, ":")
	if len(hp) > 2 {
		return ErrInvalidHost
	} else if len(hp) < 2 {
		hp = append(hp, "25")
	}
	_, err = strconv.Atoi(hp[1])
	if err != nil {
		return ErrInvalidHost
	}
	return err
}

// validateFromAddress validates
func validateFromAddress(email string) bool {
	r, _ := regexp.Compile("^([a-zA-Z0-9_\\-\\.]+)@([a-zA-Z0-9_\\-\\.]+)\\.([a-zA-Z]{2,18})$")
	return r.MatchString(email)
}

// GetDialer returns a dialer that implements the Dialer interface
func (s *SMTP) GetDialer() (mailer.Dialer, error) {
	// Ensure we have a valid user ID
	if s.UserId == 0 {
		// If we have an ID, try to get the user ID from the database
		if s.Id != 0 {
			var existingSmtp SMTP
			err := db.Where("id = ?", s.Id).First(&existingSmtp).Error
			if err != nil {
				log.Errorf("Failed to get existing SMTP profile: %v", err)
				return nil, fmt.Errorf("failed to get user ID: SMTP profile not found")
			}
			s.UserId = existingSmtp.UserId
			log.Infof("Retrieved user ID %d from existing SMTP profile", s.UserId)
		} else {
			return nil, fmt.Errorf("user ID is required for dialer")
		}
	}

	// For Graph API interface, use GraphAPI's GetDialer method
	if s.Interface == "GRAPH" {
		log.Infof("Creating dialer for Graph API sending profile - User ID: %d", s.UserId)
		
		// If we have temporary credentials (for test emails), use the existing app registration
		if s.ClientID != "" && s.ClientSecret != "" {
			log.Infof("Using provided client credentials for Graph API test - User ID: %d", s.UserId)
			
			// Check if we have a provider tenant from context
			if s.ProviderTenant == nil {
				return nil, fmt.Errorf("provider tenant context is required for Graph API dialer")
			}

			// If we have an existing app registration, use it
			if s.AppRegistrationID != "" {
				_, err := GetAppRegistration(s.AppRegistrationID)
				if err != nil {
					return nil, fmt.Errorf("failed to get app registration: %v", err)
				}

				g := GraphAPI{
					FromAddress:    s.FromAddress,
					ClientID:       s.ClientID,        // Use test credentials
					ClientSecret:   s.ClientSecret,    // Use test credentials
					InterfaceType: s.Interface,
					UserId:        s.UserId,
					ProviderTenant: s.ProviderTenant,
				}
				log.Infof("Using existing app registration for test with User ID: %d", g.UserId)
				return g.GetDialer()
			}

			// For test emails without app registration, use temporary credentials
			g := GraphAPI{
				FromAddress:    s.FromAddress,
				ClientID:       s.ClientID,
				ClientSecret:   s.ClientSecret,
				InterfaceType: s.Interface,
				UserId:        s.UserId,
				ProviderTenant: s.ProviderTenant,
			}
			log.Infof("Using temporary credentials for test with User ID: %d", g.UserId)
			return g.GetDialer()
		}

		// Otherwise, get the app registration
		appReg, err := GetAppRegistration(s.AppRegistrationID)
		if err != nil {
			return nil, fmt.Errorf("failed to get app registration: %v", err)
		}

		g := GraphAPI{
			FromAddress:    s.FromAddress,
			ClientID:       appReg.ClientID,
			ClientSecret:   appReg.ClientSecretEncrypted,
			InterfaceType: s.Interface,
			UserId:        s.UserId,
			ProviderTenant: s.ProviderTenant,
		}
		log.Infof("Created GraphAPI instance with User ID: %d", g.UserId)
		return g.GetDialer()
	}

	// For standard SMTP interface
	hp := strings.Split(s.Host, ":")
	if len(hp) < 2 {
		hp = append(hp, "25")
	}
	host := hp[0]
	// Any issues should have been caught in validation, but we'll
	// double check here.
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		log.Error(err)
		return nil, err
	}
	dialer := dialer.Dialer()
	d := gomail.NewWithDialer(dialer, host, port, s.Username, s.Password)
	d.TLSConfig = &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: s.IgnoreCertErrors,
	}
	hostname, err := os.Hostname()
	if err != nil {
		log.Error(err)
		hostname = "localhost"
	}
	d.LocalName = hostname
	return &Dialer{d}, err
}

// GetSMTPs returns the SMTPs owned by the given user.
func GetSMTPs(uid int64) ([]SMTP, error) {
	ss := []SMTP{}
	err := db.Where("user_id=?", uid).Find(&ss).Error
	if err != nil {
		log.Error(err)
		return ss, err
	}
	
	// Get the user to find their tenant
	user, err := GetUser(uid)
	if err != nil {
		log.Error(err)
		return ss, err
	}

	// Get the Microsoft provider tenant for this user's tenant
	var providerTenant ProviderTenant
	err = db.Where("tenant_id = ? AND provider_type = ?", user.TenantID, ProviderTypeAzure).First(&providerTenant).Error
	if err != nil {
		log.Error(err)
		// Don't return error as user might not have Microsoft provider tenant yet
	}

	for i := range ss {
		// Get headers
		err = db.Where("smtp_id=?", ss[i].Id).Find(&ss[i].Headers).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			return ss, err
		}
		
		// If this is a Graph API profile, get the credentials from app registration
		if ss[i].Interface == "GRAPH" && ss[i].AppRegistrationID != "" {
			log.Infof("Processing Graph API SMTP profile %d with app_registration_id: %s", ss[i].Id, ss[i].AppRegistrationID)
			
			appReg, err := GetAppRegistration(ss[i].AppRegistrationID)
			if err != nil {
				log.Errorf("Failed to get app registration for SMTP profile %d: %v", ss[i].Id, err)
				continue
			}
			log.Infof("Successfully retrieved app registration with client_id: %s", appReg.ClientID)
			
			// Populate Graph API fields
			ss[i].ClientID = appReg.ClientID
			ss[i].ClientSecret = appReg.ClientSecretEncrypted
			ss[i].TenantID = ss[i].TenantID
			
			log.Infof("Successfully populated Graph API credentials for SMTP profile %d - ClientID: %s", 
				ss[i].Id, ss[i].ClientID)
		}
	}
	return ss, nil
}

// GetSMTP returns the SMTP, if it exists, specified by the given id and user_id.
func GetSMTP(id int64, uid int64) (SMTP, error) {
	s := SMTP{}
	err := db.Where("user_id=? and id=?", uid, id).Find(&s).Error
	if err != nil {
		log.Error(err)
		return s, err
	}
	err = db.Where("smtp_id=?", s.Id).Find(&s.Headers).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return s, err
	}
	
	// Get the user to find their tenant
	user, err := GetUser(uid)
	if err != nil {
		log.Error(err)
		return s, err
	}

	// Get the Microsoft provider tenant for this user's tenant
	var providerTenant ProviderTenant
	err = db.Where("tenant_id = ? AND provider_type = ?", user.TenantID, ProviderTypeAzure).First(&providerTenant).Error
	if err != nil {
		log.Error(err)
		return s, err
	}
	
	// If this is a Graph API profile, get the credentials from app registration
	if s.Interface == "GRAPH" && s.AppRegistrationID != "" {
		log.Infof("Getting app registration details for SMTP profile %d with app_registration_id: %s", s.Id, s.AppRegistrationID)
		
		appReg, err := GetAppRegistration(s.AppRegistrationID)
		if err != nil {
			log.Errorf("Failed to get app registration for SMTP profile %d: %v", s.Id, err)
			return s, err
		}
		
		// Populate Graph API fields
		s.ClientID = appReg.ClientID
		s.ClientSecret = appReg.ClientSecretEncrypted
		
		log.Infof("Successfully populated Graph API credentials for SMTP profile %d", s.Id)
	}
	return s, err
}

// GetSMTPByName returns the SMTP, if it exists, specified by the given name and user_id.
func GetSMTPByName(n string, uid int64) (SMTP, error) {
	s := SMTP{}
	err := db.Where("user_id=? and name=?", uid, n).Find(&s).Error
	if err != nil {
		log.Error(err)
		return s, err
	}
	err = db.Where("smtp_id=?", s.Id).Find(&s.Headers).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
	}
	
	// If this is a Graph API profile, get the credentials from app registration
	if s.Interface == "GRAPH" && s.AppRegistrationID != "" {
		log.Infof("Getting app registration details for SMTP profile %d with app_registration_id: %s", s.Id, s.AppRegistrationID)
		
		appReg, err := GetAppRegistration(s.AppRegistrationID)
		if err != nil {
			log.Errorf("Failed to get app registration for SMTP profile %d: %v", s.Id, err)
			return s, err
		}
		
		// Populate Graph API fields
		s.ClientID = appReg.ClientID
		s.ClientSecret = appReg.ClientSecretEncrypted
		
		log.Infof("Successfully populated Graph API credentials for SMTP profile %d", s.Id)
	}
	return s, err
}

// PostSMTP creates a new SMTP in the database.
func PostSMTP(s *SMTP) error {
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	// Insert into the DB
	err = db.Save(s).Error
	if err != nil {
		log.Error(err)
	}
	// Save custom headers
	for i := range s.Headers {
		s.Headers[i].SMTPId = s.Id
		err := db.Save(&s.Headers[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}
	return err
}

// PutSMTP edits an existing SMTP in the database.
// Per the PUT Method RFC, it presumes all data for a SMTP is provided.
func PutSMTP(s *SMTP) error {
	// For existing profiles, if the interface is GRAPH, we'll skip validation
	// but still update all relevant fields
	if s.Id != 0 && s.Interface == "GRAPH" {
		// First get the existing record to ensure we're not changing critical fields
		existing := SMTP{}
		err := db.Where("id=?", s.Id).First(&existing).Error
		if err != nil {
			log.Error(err)
			return err
		}
		
		// Ensure we're not changing user_id
		s.UserId = existing.UserId
		
		// Update all fields except user_id (which must remain the same)
		updateFields := map[string]interface{}{
			"name": s.Name,
			"interface_type": s.Interface,
			"from_address": s.FromAddress,
			"modified_date": time.Now().UTC(),
		}
		
		// Handle app registration updates
		if existing.AppRegistrationID != "" {
			// Get the existing app registration
			appReg, err := GetAppRegistration(existing.AppRegistrationID)
			if err != nil {
				log.Errorf("Failed to get app registration for SMTP profile %d: %v", s.Id, err)
				return fmt.Errorf("failed to get app registration: %v", err)
			}
			
			// If we have new client credentials, update the app registration
			if s.ClientID != "" && s.ClientSecret != "" {
				appReg.ClientID = s.ClientID
				appReg.ClientSecretEncrypted = s.ClientSecret
				
				// Save the app registration
				if err := appReg.Update(); err != nil {
					log.Errorf("Failed to update app registration for SMTP profile %d: %v", s.Id, err)
					return fmt.Errorf("failed to update app registration: %v", err)
				}
			}
			
			// Keep the existing app registration ID
			updateFields["app_registration_id"] = existing.AppRegistrationID
		} else if s.ClientID != "" && s.ClientSecret != "" {
			// Create new app registration if we have credentials but no existing one
			appReg := &AppRegistration{
				ProviderTenantID: s.ProviderTenant.ID,
				ClientID: s.ClientID,
				ClientSecretEncrypted: s.ClientSecret,
				RedirectURI: "https://localhost:3333",
			}
			
			// Set default Graph API scopes
			appReg.SetScopes([]string{
				"https://graph.microsoft.com/Mail.Send",
				"https://graph.microsoft.com/Mail.Send.Shared",
				"https://graph.microsoft.com/User.Read",
			})
			
			// Create the app registration
			if err := appReg.Create(); err != nil {
				log.Errorf("Failed to create app registration for SMTP profile %d: %v", s.Id, err)
				return fmt.Errorf("failed to create app registration: %v", err)
			}
			
			updateFields["app_registration_id"] = appReg.ID
		}
		
		// Update in the database
		err = db.Model(&SMTP{}).Where("id=?", s.Id).Updates(updateFields).Error
		if err != nil {
			log.Error(err)
			return err
		}
		
		// Delete all custom headers, and replace with new ones
		err = db.Where("smtp_id=?", s.Id).Delete(&Header{}).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Error(err)
			return err
		}
		
		// Save custom headers
		for i := range s.Headers {
			s.Headers[i].SMTPId = s.Id
			err := db.Save(&s.Headers[i]).Error
			if err != nil {
				log.Error(err)
				return err
			}
		}
		
		return nil
	}
	
	// For new profiles or SMTP profiles, proceed with validation
	err := s.Validate()
	if err != nil {
		log.Error(err)
		return err
	}
	
	err = db.Where("id=?", s.Id).Save(s).Error
	if err != nil {
		log.Error(err)
	}
	
	// Delete all custom headers, and replace with new ones
	err = db.Where("smtp_id=?", s.Id).Delete(&Header{}).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error(err)
		return err
	}
	
	// Save custom headers
	for i := range s.Headers {
		s.Headers[i].SMTPId = s.Id
		err := db.Save(&s.Headers[i]).Error
		if err != nil {
			log.Error(err)
			return err
		}
	}
	
	return err
}

// DeleteSMTP deletes an existing SMTP in the database.
// An error is returned if a SMTP with the given user id and SMTP id is not found.
func DeleteSMTP(id int64, uid int64) error {
	// First get the SMTP profile to check if it's a Graph API profile
	smtp, err := GetSMTP(id, uid)
	if err != nil {
		log.Error(err)
		return err
	}

	// If this is a Graph API profile, delete the associated app registration
	if smtp.Interface == "GRAPH" && smtp.AppRegistrationID != "" {
		log.Infof("Deleting app registration for Graph API SMTP profile %d", id)

		// Get the app registration
		appReg, err := GetAppRegistration(smtp.AppRegistrationID)
		if err != nil {
			log.Errorf("Failed to get app registration for SMTP profile %d: %v", id, err)
			// Continue with deletion even if app registration not found
		} else {
			// Delete the app registration
			if err := appReg.Delete(); err != nil {
				log.Errorf("Failed to delete app registration %s: %v", appReg.ID, err)
				return fmt.Errorf("failed to delete app registration: %v", err)
			}
			log.Infof("Successfully deleted app registration %s", appReg.ID)
		}
	}

	// Delete all custom headers
	err = db.Where("smtp_id=?", id).Delete(&Header{}).Error
	if err != nil {
		log.Error(err)
		return err
	}

	// Delete the SMTP profile
	err = db.Where("user_id=?", uid).Delete(SMTP{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
