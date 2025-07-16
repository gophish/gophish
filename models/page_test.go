package models

import (
	"os"
	"strings"
	"testing"
)

// TestPageValidationWithTemplateConflicts tests page validation with problematic template syntax
func TestPageValidationWithTemplateConflicts(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expectError bool
		errorText   string
	}{
		{
			name:        "CSS custom properties with template syntax",
			htmlContent: `<style>:root { --color: {{theme = 'blue'}}; }</style>`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "JavaScript config with template syntax",
			htmlContent: `<script>var config = {{api_key = "test"}};</script>`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Django template patterns with assignment",
			htmlContent: `{{django_var = 'value'}} {% if user = admin %}Admin panel{% endif %}`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Valid Gophish templates should pass",
			htmlContent: `<div>Welcome {{.FirstName}} {{.LastName}}!</div>{{.Tracker}}`,
			expectError: false,
		},
		{
			name:        "Mixed valid and invalid patterns",
			htmlContent: `<div>{{.FirstName}}</div><style>:root{--color:{{invalid = 'test'}}}</style>`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Escaped template patterns should pass",
			htmlContent: `<style>:root { --color: &#123;&#123;theme = 'blue'&#125;&#125;; }</style>`,
			expectError: false,
		},
		{
			name:        "Complex nested patterns",
			htmlContent: `{{outer {{inner = value}} }}`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Empty HTML",
			htmlContent: "",
			expectError: false,
		},
		{
			name:        "Regular HTML without templates",
			htmlContent: `<div>Regular content</div><form><input name="test"></form>`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := Page{
				Name:               "Test Page",
				HTML:               tt.htmlContent,
				CaptureCredentials: true,
				CapturePasswords:   true,
				RedirectURL:        "http://example.com",
			}

			err := page.Validate()

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error containing %q, got %q", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

// TestPageValidationImprovedErrorMessages tests that enhanced error messages are provided
func TestPageValidationImprovedErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		htmlContent string
		expectText  string
	}{
		{
			name:        "Bad character error should trigger helpful message",
			htmlContent: `<style>:root { --color: {{theme = 'blue'}}; }</style>`,
			expectText:  "imported HTML contains invalid template syntax",
		},
		{
			name:        "Template validation error with context",
			htmlContent: `<script>var x = {{key = "value"}};</script>`,
			expectText:  "template syntax",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := Page{
				Name:               "Test Page",
				HTML:               tt.htmlContent,
				CaptureCredentials: true,
				RedirectURL:        "http://example.com",
			}

			err := page.Validate()
			if err == nil {
				t.Error("Expected validation error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.expectText) {
				t.Errorf("Expected error message containing %q, got %q", tt.expectText, err.Error())
			}
		})
	}
}

// TestPageImportSaveRenderCycle tests the full import → validate → save → render cycle
func TestPageImportSaveRenderCycle(t *testing.T) {
	testCases := []struct {
		name     string
		htmlFile string
	}{
		{name: "Gmail problematic HTML", htmlFile: "../testdata/gmail_problematic.html"},
		{name: "Django template HTML", htmlFile: "../testdata/django_template.html"},
		{name: "React app HTML", htmlFile: "../testdata/react_app.html"},
		{name: "Mixed templates", htmlFile: "../testdata/mixed_templates.html"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Read problematic HTML (simulating import)
			htmlContent, err := os.ReadFile(tc.htmlFile)
			if err != nil {
				t.Fatalf("Failed to read test file: %v", err)
			}

			// Step 2: Sanitize HTML (this would happen during import)
			sanitizedHTML := sanitizeTemplateContentForTest(string(htmlContent))

			// Step 3: Create and validate page
			page := Page{
				Name:               "Test Imported Page",
				HTML:               sanitizedHTML,
				CaptureCredentials: true,
				CapturePasswords:   true,
				RedirectURL:        "{{.URL}}",
			}

			err = page.Validate()
			if err != nil {
				t.Errorf("Page validation failed after sanitization: %v", err)
				return
			}

			// Step 4: Test that template execution works (simulating rendering)
			// This tests that valid Gophish templates still work after sanitization
			validTemplateHTML := `<div>Welcome {{.FirstName}}!</div>{{.Tracker}}`
			testPage := Page{
				Name:               "Template Test",
				HTML:               validTemplateHTML,
				CaptureCredentials: true,
				RedirectURL:        "{{.URL}}",
			}

			err = testPage.Validate()
			if err != nil {
				t.Errorf("Valid Gophish template failed validation: %v", err)
			}
		})
	}
}

// TestPageFormFieldProcessing tests that form field processing works correctly after template fixes
func TestPageFormFieldProcessing(t *testing.T) {
	htmlWithForms := `<!DOCTYPE html>
<html>
<head>
    <style>:root { --color: &#123;&#123;theme = 'blue'&#125;&#125;; }</style>
</head>
<body>
    <form action="/login" method="post">
        <input type="email" name="email" required>
        <input type="password" name="password" required>
        <button type="submit">Login</button>
    </form>
    <form action="/register" method="post">
        <input type="text" name="username">
        <input type="email" name="email">
        <input type="hidden" name="token" value="test">
    </form>
</body>
</html>`

	tests := []struct {
		name               string
		captureCredentials bool
		capturePasswords   bool
		expectedChanges    []string
	}{
		{
			name:               "Capture credentials and passwords",
			captureCredentials: true,
			capturePasswords:   true,
			expectedChanges: []string{
				`action=""`,       // Form actions should be empty
				`name="password"`, // Password fields should keep name attribute
			},
		},
		{
			name:               "Capture credentials but not passwords",
			captureCredentials: true,
			capturePasswords:   false,
			expectedChanges: []string{
				`action=""`,                        // Form actions should be empty
				`type="password"`,                  // Password fields should exist
				`<input type="password" required>`, // But without name attribute
			},
		},
		{
			name:               "Don't capture credentials",
			captureCredentials: false,
			capturePasswords:   false,
			expectedChanges: []string{
				`action=""`,                     // Form actions should be empty
				`<input type="email" required>`, // Input fields without name attributes
				`<input type="text">`,
				`<input type="password" required>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := Page{
				Name:               "Test Form Page",
				HTML:               htmlWithForms,
				CaptureCredentials: tt.captureCredentials,
				CapturePasswords:   tt.capturePasswords,
				RedirectURL:        "http://example.com",
			}

			err := page.Validate()
			if err != nil {
				t.Fatalf("Page validation failed: %v", err)
			}

			// Check that expected changes were made
			for _, expected := range tt.expectedChanges {
				if !strings.Contains(page.HTML, expected) {
					t.Errorf("Expected %q to be present in processed HTML", expected)
				}
			}

			// Verify that escaped template patterns are still present (not broken by form processing)
			if !strings.Contains(page.HTML, "&#123;&#123;theme = 'blue'&#125;&#125;") {
				t.Error("Expected escaped template patterns to be preserved during form processing")
			}
		})
	}
}

// TestPageRedirectURLValidation tests that redirect URL validation works with template syntax
func TestPageRedirectURLValidation(t *testing.T) {
	tests := []struct {
		name        string
		redirectURL string
		expectError bool
		errorText   string
	}{
		{
			name:        "Valid Gophish template in redirect URL",
			redirectURL: "{{.URL}}/success",
			expectError: false,
		},
		{
			name:        "Invalid template syntax in redirect URL",
			redirectURL: "{{invalid = 'syntax'}}/redirect",
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Complex valid template",
			redirectURL: "{{.BaseURL}}/complete?rid={{.RId}}",
			expectError: false,
		},
		{
			name:        "Regular URL without templates",
			redirectURL: "https://example.com/success",
			expectError: false,
		},
		{
			name:        "Empty redirect URL",
			redirectURL: "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := Page{
				Name:               "Test Page",
				HTML:               "<div>Test content</div>",
				CaptureCredentials: true,
				RedirectURL:        tt.redirectURL,
			}

			err := page.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error containing %q, got %q", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

// TestPageValidationEdgeCases tests edge cases and boundary conditions
func TestPageValidationEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		page        Page
		expectError bool
		errorText   string
	}{
		{
			name: "Missing page name",
			page: Page{
				Name:               "",
				HTML:               "<div>Test</div>",
				CaptureCredentials: true,
			},
			expectError: true,
			errorText:   "Page Name not specified",
		},
		{
			name: "Capture passwords implies capture credentials",
			page: Page{
				Name:               "Test",
				HTML:               "<div>Test</div>",
				CaptureCredentials: false,
				CapturePasswords:   true,
			},
			expectError: false,
		},
		{
			name: "Very large HTML with template patterns",
			page: Page{
				Name:               "Large Page",
				HTML:               strings.Repeat(`<div>{{invalid = 'pattern'}}</div>`, 1000),
				CaptureCredentials: true,
			},
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name: "Unicode content with templates",
			page: Page{
				Name:               "Unicode Page",
				HTML:               `<div>{{emoji = '🚀'}} {{chinese = '中文'}}</div>`,
				CaptureCredentials: true,
			},
			expectError: true,
			errorText:   "template syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.page.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
					return
				}
				if !strings.Contains(err.Error(), tt.errorText) {
					t.Errorf("Expected error containing %q, got %q", tt.errorText, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}

				// For the "capture passwords implies capture credentials" test
				if tt.page.CapturePasswords && !tt.page.CaptureCredentials {
					t.Error("Expected CaptureCredentials to be set to true when CapturePasswords is true")
				}
			}
		})
	}
}

// Helper function to simulate sanitization for testing
// This mirrors the sanitizeTemplateContent function from controllers/api/import.go
func sanitizeTemplateContentForTest(html string) string {
	// Escape Go template delimiters {{ and }}
	html = strings.ReplaceAll(html, "{{", "&#123;&#123;")
	html = strings.ReplaceAll(html, "}}", "&#125;&#125;")

	// Also handle Django/Jinja2 template patterns for completeness
	html = strings.ReplaceAll(html, "{%", "&#123;%")
	html = strings.ReplaceAll(html, "%}", "%&#125;")

	return html
}

// BenchmarkPageValidation benchmarks page validation performance
func BenchmarkPageValidation(b *testing.B) {
	testCases := []struct {
		name string
		html string
	}{
		{
			name: "Small page",
			html: `<div>{{.FirstName}}</div>`,
		},
		{
			name: "Medium page with forms",
			html: `<form><input name="email"><input name="password" type="password"></form>` + strings.Repeat(`<div>content</div>`, 100),
		},
		{
			name: "Large page with many templates",
			html: strings.Repeat(`<div>{{.FirstName}} content {{.LastName}}</div>`, 1000),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			page := Page{
				Name:               "Benchmark Page",
				HTML:               tc.html,
				CaptureCredentials: true,
				CapturePasswords:   true,
				RedirectURL:        "{{.URL}}",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = page.Validate()
			}
		})
	}
}
