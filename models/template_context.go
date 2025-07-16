package models

import (
	"bytes"
	"fmt"
	"net/mail"
	"net/url"
	"path"
	"strings"
	"text/template"
)

// TemplateContext is an interface that allows both campaigns and email
// requests to have a PhishingTemplateContext generated for them.
type TemplateContext interface {
	getFromAddress() string
	getBaseURL() string
}

// PhishingTemplateContext is the context that is sent to any template, such
// as the email or landing page content.
type PhishingTemplateContext struct {
	From        string
	URL         string
	Tracker     string
	TrackingURL string
	RId         string
	BaseURL     string
	BaseRecipient
}

// NewPhishingTemplateContext returns a populated PhishingTemplateContext,
// parsing the correct fields from the provided TemplateContext and recipient.
func NewPhishingTemplateContext(ctx TemplateContext, r BaseRecipient, rid string) (PhishingTemplateContext, error) {
	f, err := mail.ParseAddress(ctx.getFromAddress())
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	fn := f.Name
	if fn == "" {
		fn = f.Address
	}
	templateURL, err := ExecuteTemplate(ctx.getBaseURL(), r)
	if err != nil {
		return PhishingTemplateContext{}, err
	}

	// For the base URL, we'll reset the the path and the query
	// This will create a URL in the form of http://example.com
	baseURL, err := url.Parse(templateURL)
	if err != nil {
		return PhishingTemplateContext{}, err
	}
	baseURL.Path = ""
	baseURL.RawQuery = ""

	phishURL, _ := url.Parse(templateURL)
	q := phishURL.Query()
	q.Set(RecipientParameter, rid)
	phishURL.RawQuery = q.Encode()

	trackingURL, _ := url.Parse(templateURL)
	trackingURL.Path = path.Join(trackingURL.Path, "/track")
	trackingURL.RawQuery = q.Encode()

	return PhishingTemplateContext{
		BaseRecipient: r,
		BaseURL:       baseURL.String(),
		URL:           phishURL.String(),
		TrackingURL:   trackingURL.String(),
		Tracker:       "<img alt='' style='display: none' src='" + trackingURL.String() + "'/>",
		From:          fn,
		RId:           rid,
	}, nil
}

// ExecuteTemplate creates a templated string based on the provided
// template body and data.
func ExecuteTemplate(text string, data interface{}) (string, error) {
	buff := bytes.Buffer{}
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return buff.String(), err
	}
	err = tmpl.Execute(&buff, data)
	return buff.String(), err
}

// ValidationContext is used for validating templates and pages
type ValidationContext struct {
	FromAddress string
	BaseURL     string
}

func (vc ValidationContext) getFromAddress() string {
	return vc.FromAddress
}

func (vc ValidationContext) getBaseURL() string {
	return vc.BaseURL
}

// validateTemplateContent checks if text contains problematic template syntax
// and provides specific error messages with line numbers
func validateTemplateContent(text string) error {
	// Check for common problematic patterns: {{ followed by = character
	if strings.Contains(text, "{{") && strings.Contains(text, "=") {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			// Look for lines containing both {{ and = which are likely problematic
			if strings.Contains(line, "{{") && strings.Contains(line, "=") {
				// Additional check: ensure it's not a valid Go template assignment
				if !strings.Contains(line, ":=") && !strings.Contains(line, "eq") {
					return fmt.Errorf("template syntax error on line %d: '%s' - this appears to be non-Go template syntax (CSS, JSON, or other framework). Consider escaping {{}} braces in imported HTML", i+1, strings.TrimSpace(line))
				}
			}
		}
	}
	return nil
}

// ValidateTemplate ensures that the provided text in the page or template
// uses the supported template variables correctly.
func ValidateTemplate(text string) error {
	// Pre-validate for common template syntax issues
	if err := validateTemplateContent(text); err != nil {
		return err
	}

	vc := ValidationContext{
		FromAddress: "foo@bar.com",
		BaseURL:     "http://example.com",
	}
	td := Result{
		BaseRecipient: BaseRecipient{
			Email:     "foo@bar.com",
			FirstName: "Foo",
			LastName:  "Bar",
			Position:  "Test",
		},
		RId: "123456",
	}
	ptx, err := NewPhishingTemplateContext(vc, td.BaseRecipient, td.RId)
	if err != nil {
		return err
	}
	_, err = ExecuteTemplate(text, ptx)
	if err != nil {
		// Enhance error message to mention template syntax issues in imported HTML
		if strings.Contains(err.Error(), "bad character") || strings.Contains(err.Error(), "unexpected") {
			return fmt.Errorf("template validation failed: %v - if this is imported HTML, it may contain invalid template syntax that needs escaping", err)
		}
		return err
	}
	return nil
}
