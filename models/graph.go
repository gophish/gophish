package models

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/mailer"
	"golang.org/x/oauth2"
)

// TokenCache holds the cached access token and its expiry
type TokenCache struct {
	AccessToken string
	ExpiresAt   time.Time
	mu          sync.RWMutex
}

var (
	// Default token endpoint URL
	defaultTokenEndpoint = "https://login.microsoftonline.com/%s/oauth2/v2.0/token"
)

// GraphAPISender implements the mailer.Sender interface for Microsoft Graph API
type GraphAPISender struct {
	client            *http.Client
	graphBaseURL      string
	clientID          string
	clientSecret      string
	providerTenantID  string
	fromAddress       string
	userID            int64
}

// GraphAPI contains the attributes needed to handle sending campaign emails via Microsoft Graph API
type GraphAPI struct {
	Id            int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId        int64     `json:"-" gorm:"column:user_id"`
	Name          string    `json:"name"`
	FromAddress   string    `json:"from_address"`
	Headers       []Header  `json:"headers"`
	ModifiedDate  time.Time `json:"modified_date"`
	InterfaceType string    `json:"interface_type" gorm:"column:interface_type"`
	// Graph API credentials - populated from app_registration and provider_tenant
	ClientID          string        `json:"client_id,omitempty" gorm:"-"`
	ClientSecret      string        `json:"client_secret,omitempty" gorm:"-"`
	ProviderTenant    *ProviderTenant `json:"-" gorm:"-"` // Provider tenant from context
}

// TableName specifies the database tablename for Gorm to use
func (g GraphAPI) TableName() string {
	return "sending_profiles"
}

// Validate ensures that Graph API configs are valid
func (g *GraphAPI) Validate() error {
	log.Infof("Validating GraphAPI config - User ID: %d", g.UserId)

	if g.FromAddress == "" {
		return ErrFromAddressNotSpecified
	}

	if g.ClientID == "" {
		return errors.New("client_id not specified")
	}

	if g.ClientSecret == "" {
		return errors.New("client_secret not specified")
	}

	// Require provider tenant from context
	if g.ProviderTenant == nil {
		return errors.New("provider tenant context is required")
	}

	// Validate the from address
	if _, err := mail.ParseAddress(g.FromAddress); err != nil {
		return fmt.Errorf("invalid from_address: %v", err)
	}

	log.Infof("Getting OAuth token for user %d and provider tenant %s", g.UserId, g.ProviderTenant.ProviderTenantID)
	// Get the OAuth token to validate we have access
	token, err := GetOAuthTokenByUserAndProviderTenant(g.UserId, g.ProviderTenant.ProviderTenantID)
	if err != nil {
		log.Errorf("Failed to get OAuth token for user %d: %v", g.UserId, err)
		return fmt.Errorf("failed to get OAuth token: %v", err)
	}
	if token.AuthorizationCode == "" {
		log.Errorf("No authorization code found for user %d", g.UserId)
		return errors.New("no authorization code found for Graph API")
	}

	log.Infof("Successfully validated GraphAPI config for user %d", g.UserId)
	return nil
}

// GetDialer returns a Dialer for the GraphAPI
func (g *GraphAPI) GetDialer() (mailer.Dialer, error) {
	if err := g.Validate(); err != nil {
		return nil, err
	}

	d := &GraphAPIDialer{
		clientID:          g.ClientID,
		clientSecret:      g.ClientSecret,
		providerTenantID:  g.ProviderTenant.ProviderTenantID,
		fromAddress:       g.FromAddress,
		userID:            g.UserId,
	}

	return d, nil
}

// GraphAPIDialer implements the mailer.Dialer interface for Microsoft Graph API
type GraphAPIDialer struct {
	clientID          string
	clientSecret      string
	providerTenantID  string
	fromAddress       string
	userID            int64
}

// Dial creates a new GraphAPISender
func (d *GraphAPIDialer) Dial() (mailer.Sender, error) {
	return &GraphAPISender{
		client:            &http.Client{},
		graphBaseURL:      "https://graph.microsoft.com/v1.0",
		clientID:          d.clientID,
		clientSecret:      d.clientSecret,
		providerTenantID:  d.providerTenantID,
		fromAddress:       d.fromAddress,
		userID:            d.userID,
	}, nil
}

// Send implements the Sender interface for GraphAPISender
func (s *GraphAPISender) Send(from string, to []string, msg io.WriterTo) error {
	// Get the existing OAuth token for this user and provider tenant
	token, err := GetOAuthTokenByUserAndProviderTenant(s.userID, s.providerTenantID)
	if err != nil {
		return fmt.Errorf("error getting token for user %d: %v", s.userID, err)
	}

	// Use the access token from the stored token
	accessToken := token.AccessTokenEncrypted

	var buf bytes.Buffer
	if _, err := msg.WriteTo(&buf); err != nil {
		return fmt.Errorf("error reading message: %v", err)
	}

	// Parse the email message
	email, err := mail.ReadMessage(strings.NewReader(buf.String()))
	if err != nil {
		return fmt.Errorf("error parsing email: %v", err)
	}

	// Extract subject
	subject := email.Header.Get("Subject")
	if subject == "" {
		subject = "No Subject"
	}

	// Get the message body
	body, err := ioutil.ReadAll(email.Body)
	if err != nil {
		return fmt.Errorf("error reading body: %v", err)
	}

	// Parse the multipart message to get HTML content
	htmlContent := string(body)
	contentType := email.Header.Get("Content-Type")
	
	// Handle multipart messages
	if strings.Contains(contentType, "multipart/") {
		mediaType, params, err := mime.ParseMediaType(contentType)
		if err == nil && strings.HasPrefix(mediaType, "multipart/") {
			mr := multipart.NewReader(bytes.NewReader(body), params["boundary"])
			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					continue
				}
				
				// Look for HTML part
				partContentType := p.Header.Get("Content-Type")
				if strings.Contains(partContentType, "text/html") {
					content, err := ioutil.ReadAll(p)
					if err == nil {
						htmlContent = string(content)
						break
					}
				}
			}
		}
	} else if strings.Contains(contentType, "text/html") {
		// Handle single-part HTML messages
		htmlContent = string(body)
	}

	// Ensure HTML content is properly preserved
	if !strings.Contains(htmlContent, "<!DOCTYPE") && !strings.Contains(htmlContent, "<html") {
		htmlContent = fmt.Sprintf("<!DOCTYPE html><html><body>%s</body></html>", htmlContent)
	}

	// Convert the email message to Graph API format
	graphMessage := struct {
		Message struct {
			Subject      string   `json:"subject"`
			Body        struct {
				ContentType string `json:"contentType"`
				Content    string `json:"content"`
			} `json:"body"`
			ToRecipients []struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			} `json:"toRecipients"`
			From struct {
				EmailAddress struct {
					Address string `json:"address"`
				} `json:"emailAddress"`
			} `json:"from"`
		} `json:"message"`
		SaveToSentItems bool `json:"saveToSentItems"`
	}{}

	// Prepare the message
	graphMessage.Message.Subject = subject
	graphMessage.Message.Body.ContentType = "HTML"
	graphMessage.Message.Body.Content = htmlContent
	graphMessage.Message.From.EmailAddress.Address = s.fromAddress
	graphMessage.SaveToSentItems = false

	for _, recipient := range to {
		graphMessage.Message.ToRecipients = append(graphMessage.Message.ToRecipients, struct {
			EmailAddress struct {
				Address string `json:"address"`
			} `json:"emailAddress"`
		}{
			EmailAddress: struct {
				Address string `json:"address"`
			}{
				Address: recipient,
			},
		})
	}

	// Send the message via Graph API
	jsonData, err := json.Marshal(graphMessage)
	if err != nil {
		return fmt.Errorf("error marshaling message: %v", err)
	}

	// For application permissions, we need to use a specific user's email address
	userEmailEncoded := url.QueryEscape(s.fromAddress)
	
	// Use the /users/{email}/sendMail endpoint with the specific user
	sendMailURL := fmt.Sprintf("%s/users/%s/sendMail", s.graphBaseURL, userEmailEncoded)
	
	log.Infof("Sending mail using Graph API URL: %s", sendMailURL)
	log.Infof("Using from address: %s", s.fromAddress)
	log.Infof("Message subject: %s", subject)

	req, err := http.NewRequest("POST", sendMailURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	// If we get a 401 Unauthorized, the token might be expired. Try to refresh and retry once.
	if resp.StatusCode == http.StatusUnauthorized {
		log.Infof("Got 401 Unauthorized, attempting to refresh token for user %d", s.userID)
		
		// Get a new token
		newAccessToken, _, err := getNewAccessToken(s.clientID, s.clientSecret, s.providerTenantID, s.userID)
		if err != nil {
			return fmt.Errorf("error refreshing token: %v", err)
		}

		// Retry the request with the new token
		req, err = http.NewRequest("POST", sendMailURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("error creating request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+newAccessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err = s.client.Do(req)
		if err != nil {
			return fmt.Errorf("error sending request with refreshed token: %v", err)
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return errors.New("rate limit exceeded")
	}

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("error from Graph API: %d %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	return nil
}

// Close implements the Sender interface
func (s *GraphAPISender) Close() error {
	return nil
}

// Reset implements the Sender interface
func (s *GraphAPISender) Reset() error {
	return nil
}

// getNewAccessToken gets a new access token from Microsoft identity platform
func getNewAccessToken(clientID, clientSecret, providerTenantID string, userID int64) (string, int, error) {
	if userID == 0 {
		return "", 0, fmt.Errorf("invalid user ID: user ID cannot be 0")
	}

	// First, get the oauth token for this tenant
	token, err := GetOAuthTokenByUserAndProviderTenant(userID, providerTenantID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get OAuth token for user %d: %v", userID, err)
	}

	if token == nil {
		return "", 0, fmt.Errorf("no OAuth token found for user %d", userID)
	}

	// Get the app registration for this provider tenant
	appReg, err := GetAppRegistrationByProviderTenant(providerTenantID)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get app registration: %v", err)
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", token.RefreshTokenEncrypted)
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("redirect_uri", appReg.RedirectURI)

	tokenURL := fmt.Sprintf(defaultTokenEndpoint, providerTenantID)
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return "", 0, fmt.Errorf("error requesting token: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("error response from token endpoint: %d %s - %s", resp.StatusCode, resp.Status, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		Scope       string `json:"scope"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("error decoding token response: %v", err)
	}

	// Update the token in the database
	token.AccessTokenEncrypted = result.AccessToken
	token.ExpiresAt = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	if err := token.Update(); err != nil {
		log.Errorf("Failed to update token in database: %v", err)
		return "", 0, fmt.Errorf("failed to update token: %v", err)
	}

	log.Infof("Successfully refreshed token. Type: %s, Expires in: %d seconds", 
		result.TokenType, result.ExpiresIn)
	
	if result.Scope != "" {
		log.Infof("Token scopes: %s", result.Scope)
	}

	return result.AccessToken, result.ExpiresIn, nil
}

// GetGraphClientForUser returns a Graph API client for a specific user
func GetGraphClientForUser(ctx context.Context, userID int64) (*GraphClient, error) {
	// Get and refresh token if needed
	token, err := GetAndRefreshTokenIfNeeded(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %v", err)
	}

	// Get default app registration for endpoint
	defaultAppReg, err := GetDefaultAppRegistration()
	if err != nil {
		return nil, fmt.Errorf("failed to get default app registration: %v", err)
	}

	// Get provider tenant info
	providerTenant, err := GetProviderTenant(defaultAppReg.ProviderTenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider tenant: %v", err)
	}

	// Create HTTP client with token
	httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	return &GraphClient{
		client:         httpClient,
		providerTenant: providerTenant,
	}, nil
}

// SendMailAsUser sends an email using the Graph API on behalf of a user
func SendMailAsUser(ctx context.Context, userID int64, message *GraphMailMessage) error {
	client, err := GetGraphClientForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get Graph client: %v", err)
	}

	return client.SendMail(ctx, message)
}

// GraphClient represents a Microsoft Graph API client
type GraphClient struct {
	client         *http.Client
	providerTenant *ProviderTenant
}

// GraphMailMessage represents an email message for the Graph API
type GraphMailMessage struct {
	Message struct {
		Subject      string `json:"subject"`
		Body         struct {
			ContentType string `json:"contentType"`
			Content     string `json:"content"`
		} `json:"body"`
		ToRecipients []struct {
			EmailAddress struct {
				Address string `json:"address"`
			} `json:"emailAddress"`
		} `json:"toRecipients"`
	} `json:"message"`
	SaveToSentItems bool `json:"saveToSentItems"`
}

// SendMail sends an email using the Graph API
func (c *GraphClient) SendMail(ctx context.Context, message *GraphMailMessage) error {
	// Convert message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://graph.microsoft.com/v1.0/me/sendMail", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send mail: %s - %s", resp.Status, string(body))
	}

	return nil
} 