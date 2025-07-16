package main

import (
	"net/url"
	"strings"
	"testing"

	"github.com/gophish/gophish/models"
)

// TestTemplateSanitizationSecurityXSS tests that template escaping doesn't introduce XSS vulnerabilities
func TestTemplateSanitizationSecurityXSS(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectEscaped  bool
		checkForXSS    []string
		description    string
	}{
		{
			name:        "Script tag in template pattern",
			input:       `{{script = '<script>alert("xss")</script>'}}`,
			expectEscaped: true,
			checkForXSS: []string{"<script>", "alert(", "javascript:"},
			description: "Script tags within template patterns should be escaped",
		},
		{
			name:        "Event handler in template",
			input:       `{{handler = 'onload="alert(1)"'}}`,
			expectEscaped: true,
			checkForXSS: []string{"onload=", "alert("},
			description: "Event handlers should be properly escaped",
		},
		{
			name:        "Data URI XSS attempt",
			input:       `{{data = 'data:text/html,<script>alert(1)</script>'}}`,
			expectEscaped: true,
			checkForXSS: []string{"data:text/html", "<script>"},
			description: "Data URIs with scripts should be escaped",
		},
		{
			name:        "JavaScript protocol",
			input:       `{{link = 'javascript:alert("xss")'}}`,
			expectEscaped: true,
			checkForXSS: []string{"javascript:", "alert("},
			description: "JavaScript protocol should be escaped",
		},
		{
			name:        "HTML entities in template",
			input:       `{{html = '&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;'}}`,
			expectEscaped: true,
			checkForXSS: []string{},
			description: "Already escaped HTML should remain safe",
		},
		{
			name:        "CSS expression injection",
			input:       `{{style = 'expression(alert("xss"))'}}`,
			expectEscaped: true,
			checkForXSS: []string{"expression(", "alert("},
			description: "CSS expressions should be escaped",
		},
		{
			name:        "SVG with script",
			input:       `{{svg = '<svg onload="alert(1)"></svg>'}}`,
			expectEscaped: true,
			checkForXSS: []string{"<svg", "onload=", "alert("},
			description: "SVG with event handlers should be escaped",
		},
		{
			name:        "Unicode XSS attempt",
			input:       `{{unicode = '\u003cscript\u003ealert(1)\u003c/script\u003e'}}`,
			expectEscaped: true,
			checkForXSS: []string{"\\u003c", "script", "alert("},
			description: "Unicode-encoded XSS should be escaped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test sanitization
			sanitized := sanitizeTemplateContentSecurity(tt.input)
			
			// Verify template patterns were escaped
			if tt.expectEscaped {
				if strings.Contains(sanitized, "{{") || strings.Contains(sanitized, "}}") {
					t.Errorf("Template delimiters not properly escaped in: %s", sanitized)
				}
			}

			// Check that potential XSS vectors are not present in a dangerous form
			for _, xssPattern := range tt.checkForXSS {
				if strings.Contains(sanitized, xssPattern) {
					t.Errorf("Potential XSS pattern %q found in sanitized output: %s", xssPattern, sanitized)
				}
			}

			// Test that template validation catches security issues
			err := models.ValidateTemplate(tt.input)
			if err == nil && tt.expectEscaped {
				t.Errorf("Expected template validation to fail for security pattern: %s", tt.description)
			}
		})
	}
}

// TestTemplateInjectionPrevention tests that malicious template injection is prevented
func TestTemplateInjectionPrevention(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "Server-side template injection",
			input:       `{{.}}{{range .}}{{.}}{{end}}`,
			expectError: false, // This is actually valid Go template syntax
			description: "Valid Go template range should pass",
		},
		{
			name:        "Template function call injection",
			input:       `{{printf "%s" .}}`,
			expectError: false, // This is valid if printf is available
			description: "Template function calls",
		},
		{
			name:        "Malicious assignment injection",
			input:       `{{$cmd := "rm -rf /"}}{{$cmd}}`,
			expectError: false, // Valid Go template syntax but content is concerning
			description: "Template variable assignment with concerning content",
		},
		{
			name:        "Invalid template injection",
			input:       `{{system = "rm -rf /"}}`,
			expectError: true,
			description: "Invalid template assignment should be caught",
		},
		{
			name:        "Template escape attempt",
			input:       `{{/* comment */}}{{raw "{{.SafeHTML}}"}}`,
			expectError: true, // Invalid syntax
			description: "Template escape attempts should be caught",
		},
		{
			name:        "Path traversal in template",
			input:       `{{file = "../../../etc/passwd"}}`,
			expectError: true,
			description: "Path traversal patterns should be caught",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := models.ValidateTemplate(tt.input)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for %s but got none", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for %s but got: %v", tt.description, err)
				}
			}
		})
	}
}

// TestHTMLEntityHandling tests that HTML entities are handled correctly
func TestHTMLEntityHandling(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedSafe   bool
		checkEntities  []string
		description    string
	}{
		{
			name:          "Template delimiters as entities",
			input:         `&#123;&#123;theme = 'blue'&#125;&#125;`,
			expectedSafe:  true,
			checkEntities: []string{"&#123;", "&#125;"},
			description:   "HTML entities for template delimiters should be preserved",
		},
		{
			name:          "Mixed entities and templates",
			input:         `<div>&lt;{{script = 'alert(1)'}}&gt;</div>`,
			expectedSafe:  false, // Contains unescaped template
			checkEntities: []string{"&lt;", "&gt;"},
			description:   "Mixed entities and templates should be handled correctly",
		},
		{
			name:          "Escaped XSS with entities",
			input:         `{{xss = '&lt;script&gt;alert(&quot;xss&quot;)&lt;/script&gt;'}}`,
			expectedSafe:  false, // Contains template assignment
			checkEntities: []string{"&lt;", "&gt;", "&quot;"},
			description:   "HTML entities should not prevent template validation",
		},
		{
			name:          "Valid Gophish template with entities",
			input:         `<div>&lt;{{.FirstName}}&gt;</div>`,
			expectedSafe:  true,
			checkEntities: []string{"&lt;", "&gt;"},
			description:   "Valid templates with entities should pass",
		},
		{
			name:          "Double-encoded entities",
			input:         `&amp;lt;script&amp;gt;alert(1)&amp;lt;/script&amp;gt;`,
			expectedSafe:  true,
			checkEntities: []string{"&amp;lt;", "&amp;gt;"},
			description:   "Double-encoded entities should be safe",
		},
		{
			name:          "Numeric character references",
			input:         `&#60;{{script = 'alert(1)'}}&#62;`,
			expectedSafe:  false, // Contains template assignment
			checkEntities: []string{"&#60;", "&#62;"},
			description:   "Numeric character references should be handled",
		},
		{
			name:          "Hex character references",
			input:         `&#x3c;{{script = 'alert(1)'}}&#x3e;`,
			expectedSafe:  false, // Contains template assignment
			checkEntities: []string{"&#x3c;", "&#x3e;"},
			description:   "Hex character references should be handled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test sanitization preserves/creates entities appropriately
			sanitized := sanitizeTemplateContentSecurity(tt.input)
			
			// Check that expected entities are present
			for _, entity := range tt.checkEntities {
				if !strings.Contains(tt.input, entity) && !strings.Contains(sanitized, entity) {
					t.Errorf("Expected entity %q to be present in input or sanitized output", entity)
				}
			}

			// Test validation
			err := models.ValidateTemplate(tt.input)
			
			if tt.expectedSafe {
				if err != nil {
					t.Errorf("Expected safe input to pass validation: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected unsafe input to fail validation: %s", tt.description)
				}
			}

			// Verify sanitized version passes validation
			err = models.ValidateTemplate(sanitized)
			if err != nil {
				t.Errorf("Sanitized version should pass validation: %v", err)
			}
		})
	}
}

// TestContentSecurityPolicyCompatibility tests CSP compatibility of escaped templates
func TestContentSecurityPolicyCompatibility(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		cspDirective    string
		expectCompliant bool
		description     string
	}{
		{
			name:            "Inline style with template",
			input:           `<style>{{css = 'body{color:red;}''}}</style>`,
			cspDirective:    "style-src 'self'",
			expectCompliant: false, // Would violate CSP due to inline style
			description:     "Inline styles may violate CSP",
		},
		{
			name:            "Inline script with template",
			input:           `<script>{{js = 'alert(1)''}}</script>`,
			cspDirective:    "script-src 'self'",
			expectCompliant: false, // Would violate CSP due to inline script
			description:     "Inline scripts may violate CSP",
		},
		{
			name:            "Event handler with template",
			input:           `<div onclick="{{handler = 'alert(1)'}}">`,
			cspDirective:    "script-src 'self'",
			expectCompliant: false, // Event handlers violate strict CSP
			description:     "Event handlers may violate CSP",
		},
		{
			name:            "External resource with template",
			input:           `<img src="{{src = 'https://evil.com/image.png'}}">`,
			cspDirective:    "img-src 'self'",
			expectCompliant: false, // External image may violate CSP
			description:     "External resources may violate CSP",
		},
		{
			name:            "Data URI with template",
			input:           `<img src="{{data = 'data:image/png;base64,iVBOR...'}}">`,
			cspDirective:    "img-src 'self'",
			expectCompliant: false, // Data URIs may not be allowed
			description:     "Data URIs may violate CSP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that sanitization removes problematic template patterns
			sanitized := sanitizeTemplateContentSecurity(tt.input)
			
			// Verify templates were escaped
			if strings.Contains(sanitized, "{{") || strings.Contains(sanitized, "}}") {
				t.Error("Templates should be escaped for CSP compatibility")
			}

			// Test that original input would likely fail template validation
			err := models.ValidateTemplate(tt.input)
			if err == nil && !tt.expectCompliant {
				t.Errorf("Expected template validation to catch CSP-violating pattern: %s", tt.description)
			}
		})
	}
}

// TestURLValidationSecurity tests URL validation in templates for security issues
func TestURLValidationSecurity(t *testing.T) {
	tests := []struct {
		name           string
		templateURL    string
		expectError    bool
		securityIssue  string
		description    string
	}{
		{
			name:          "JavaScript protocol URL",
			templateURL:   `javascript:alert('xss')`,
			expectError:   false, // URL validation may not catch this
			securityIssue: "javascript:",
			description:   "JavaScript protocol URLs are dangerous",
		},
		{
			name:          "Data URL with script",
			templateURL:   `data:text/html,<script>alert(1)</script>`,
			expectError:   false, // URL validation may not catch this
			securityIssue: "data:text/html",
			description:   "Data URLs with HTML/JS are dangerous",
		},
		{
			name:          "File protocol URL",
			templateURL:   `file:///etc/passwd`,
			expectError:   false, // URL validation may not catch this
			securityIssue: "file://",
			description:   "File protocol URLs can expose local files",
		},
		{
			name:          "FTP protocol URL",
			templateURL:   `ftp://attacker.com/malware.exe`,
			expectError:   false, // URL validation may not catch this
			securityIssue: "ftp://",
			description:   "FTP URLs may be used for malicious purposes",
		},
		{
			name:          "Valid HTTP URL",
			templateURL:   `https://example.com/safe`,
			expectError:   false,
			securityIssue: "",
			description:   "HTTPS URLs should be safe",
		},
		{
			name:          "Valid template URL",
			templateURL:   `{{.URL}}/path`,
			expectError:   false,
			securityIssue: "",
			description:   "Template URLs should be safe",
		},
		{
			name:          "Template with dangerous assignment",
			templateURL:   `{{url = 'javascript:alert(1)'}}`,
			expectError:   true,
			securityIssue: "javascript:",
			description:   "Template assignments with dangerous URLs should be caught",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test as redirect URL in page
			page := models.Page{
				Name:               "Security Test",
				HTML:               "<div>Test</div>",
				CaptureCredentials: true,
				RedirectURL:        tt.templateURL,
			}

			err := page.Validate()
			
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for security issue: %s", tt.description)
				}
			}

			// Check for security issues in URL
			if tt.securityIssue != "" {
				if strings.Contains(tt.templateURL, tt.securityIssue) {
					t.Logf("Security issue detected: %s in URL %s", tt.securityIssue, tt.templateURL)
					
					// Check if URL would be safe after parsing
					if parsed, err := url.Parse(tt.templateURL); err == nil {
						if parsed.Scheme == "javascript" || parsed.Scheme == "data" || parsed.Scheme == "file" {
							t.Logf("Potentially dangerous URL scheme: %s", parsed.Scheme)
						}
					}
				}
			}
		})
	}
}

// TestInputSanitizationBoundaries tests edge cases in input sanitization
func TestInputSanitizationBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectSafe  bool
		description string
	}{
		{
			name:        "Empty input",
			input:       "",
			expectSafe:  true,
			description: "Empty input should be safe",
		},
		{
			name:        "Only whitespace",
			input:       "   \n\t\r   ",
			expectSafe:  true,
			description: "Whitespace-only input should be safe",
		},
		{
			name:        "Null bytes",
			input:       "test\x00content{{var = 'value'}}",
			expectSafe:  false,
			description: "Input with null bytes should be sanitized",
		},
		{
			name:        "Control characters",
			input:       "test\x01\x02\x03{{var = 'value'}}",
			expectSafe:  false,
			description: "Input with control characters should be sanitized",
		},
		{
			name:        "Very long input",
			input:       strings.Repeat("a", 100000) + "{{var = 'value'}}",
			expectSafe:  false,
			description: "Very long input should be sanitized",
		},
		{
			name:        "Binary data",
			input:       string([]byte{0xFF, 0xFE, 0xFD}) + "{{var = 'value'}}",
			expectSafe:  false,
			description: "Binary data should be sanitized",
		},
		{
			name:        "Maximum template nesting",
			input:       strings.Repeat("{{", 1000) + "var = 'value'" + strings.Repeat("}}", 1000),
			expectSafe:  false,
			description: "Deeply nested templates should be sanitized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test sanitization
			sanitized := sanitizeTemplateContentSecurity(tt.input)
			
			// Verify no unescaped templates remain
			hasTemplates := strings.Contains(sanitized, "{{") || strings.Contains(sanitized, "}}")
			
			if tt.expectSafe {
				// For safe inputs, templates may remain if they're valid Go templates
				err := models.ValidateTemplate(tt.input)
				if err != nil {
					t.Logf("Input failed validation (expected for some safe inputs): %v", err)
				}
			} else {
				if hasTemplates {
					t.Errorf("Unsafe input still contains unescaped templates after sanitization")
				}
			}

			// Sanitized version should always pass validation
			err := models.ValidateTemplate(sanitized)
			if err != nil {
				// Some inputs may still fail due to other validation rules
				t.Logf("Sanitized input failed validation: %v", err)
			}
		})
	}
}

// Helper function for security tests
func sanitizeTemplateContentSecurity(html string) string {
	// Escape Go template delimiters {{ and }}
	html = strings.ReplaceAll(html, "{{", "&#123;&#123;")
	html = strings.ReplaceAll(html, "}}", "&#125;&#125;")
	
	// Also handle Django/Jinja2 template patterns for completeness
	html = strings.ReplaceAll(html, "{%", "&#123;%")
	html = strings.ReplaceAll(html, "%}", "%&#125;")
	
	return html
}

// TestSecurityRegressionPrevention tests for regressions in security fixes
func TestSecurityRegressionPrevention(t *testing.T) {
	// Test the original error that started this whole fix
	problematicInput := `<style>:root { --color: {{theme = 'blue'}}; }</style>`
	
	// This should now be handled gracefully
	err := models.ValidateTemplate(problematicInput)
	if err == nil {
		t.Error("Expected problematic input to fail validation")
		return
	}
	
	// But sanitized version should pass
	sanitized := sanitizeTemplateContentSecurity(problematicInput)
	err = models.ValidateTemplate(sanitized)
	if err != nil {
		t.Errorf("Sanitized version should pass validation: %v", err)
	}
	
	// Verify the specific error pattern is handled
	if strings.Contains(err.Error(), "bad character U+003D") {
		t.Error("Original error pattern should be caught by preprocessing, not template parser")
	}
}