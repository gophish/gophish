package models

import (
	"strings"
	"testing"
	"testing/quick"
)

// TestValidateTemplateContent tests the preprocessing validation function
func TestValidateTemplateContent(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorText   string
	}{
		{
			name:        "CSS custom properties with assignment",
			input:       `:root { --color: {{theme = 'blue'}}; }`,
			expectError: true,
			errorText:   "template syntax error on line 1",
		},
		{
			name:        "JavaScript object with assignment",
			input:       `var config = {{api_key = "test", debug = true}};`,
			expectError: true,
			errorText:   "template syntax error on line 1",
		},
		{
			name:        "Valid Go template assignment",
			input:       `{{$var := .Field}}`,
			expectError: false, // Uses := which is valid Go template syntax
		},
		{
			name:        "Valid Go template comparison",
			input:       `{{if eq .Field "value"}}`,
			expectError: false, // Uses eq which is valid
		},
		{
			name:        "Django-style template (no = character)",
			input:       `{% if user.is_admin %}`,
			expectError: false, // No = character, so no conflict
		},
		{
			name:        "Mixed valid Go templates",
			input:       `<div>{{.FirstName}} {{.LastName}}</div>{{.Tracker}}`,
			expectError: false,
		},
		{
			name:        "Multiline with problematic pattern",
			input: `<style>
:root {
    --primary: {{color = 'blue'}};
    --secondary: {{alt_color = 'red'}};
}
</style>`,
			expectError: true,
			errorText:   "template syntax error on line 3", // Should identify the specific line
		},
		{
			name:        "Nested problematic patterns",
			input:       `{{outer = { inner = { value = 'test' } } }}`,
			expectError: true,
			errorText:   "template syntax error on line 1",
		},
		{
			name:        "No template patterns",
			input:       `<div>Regular HTML content</div>`,
			expectError: false,
		},
		{
			name:        "Empty input",
			input:       "",
			expectError: false,
		},
		{
			name:        "Template pattern without assignment",
			input:       `{{if .Condition}}content{{end}}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTemplateContent(tt.input)

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

// TestValidateTemplate tests the main ValidateTemplate function with various inputs
func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		errorText   string
	}{
		{
			name:        "Valid Gophish template variables",
			input:       `<div>Welcome {{.FirstName}} {{.LastName}}!</div>{{.Tracker}}`,
			expectError: false,
		},
		{
			name:        "Valid template with URL",
			input:       `<a href="{{.URL}}">Click here</a>`,
			expectError: false,
		},
		{
			name:        "Valid template with BaseURL",
			input:       `<form action="{{.BaseURL}}/submit">`,
			expectError: false,
		},
		{
			name:        "Valid template with RId",
			input:       `<input type="hidden" name="rid" value="{{.RId}}">`,
			expectError: false,
		},
		{
			name:        "CSS custom properties (should fail)",
			input:       `<style>:root { --color: {{theme = 'blue'}}; }</style>`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "JavaScript config (should fail)",
			input:       `<script>var config = {{api_key = "test"}};</script>`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Django template patterns (should fail)",
			input:       `{% if user = admin %}content{% endif %}`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Invalid Go template syntax",
			input:       `{{.InvalidField.}}`,
			expectError: true,
			errorText:   "template validation failed",
		},
		{
			name:        "Template with bad character (original error)",
			input:       `{{badSyntax = 'value'}}`,
			expectError: true,
			errorText:   "template syntax error",
		},
		{
			name:        "Escaped template patterns (should pass)",
			input:       `<style>:root { --color: &#123;&#123;theme = 'blue'&#125;&#125;; }</style>`,
			expectError: false,
		},
		{
			name:        "Complex valid template",
			input:       `<div>{{if .FirstName}}Hello {{.FirstName}}{{else}}Hello there{{end}}</div>`,
			expectError: false,
		},
		{
			name:        "Template with range",
			input:       `{{range .Items}}<div>{{.Name}}</div>{{end}}`,
			expectError: false,
		},
		{
			name:        "Template with variables and assignment",
			input:       `{{$name := .FirstName}}<div>{{$name}}</div>`,
			expectError: false,
		},
		{
			name:        "Empty template",
			input:       "",
			expectError: false,
		},
		{
			name:        "Regular HTML without templates",
			input:       `<form><input type="email" name="email"></form>`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.input)

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

// TestValidateTemplateEnhancedErrorMessages tests that enhanced error messages are provided
func TestValidateTemplateEnhancedErrorMessages(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectMsg string
	}{
		{
			name:      "Bad character error should mention imported HTML",
			input:     `{{invalid = 'syntax'}}`,
			expectMsg: "if this is imported HTML, it may contain invalid template syntax",
		},
		{
			name:      "Template preprocessing should provide line numbers",
			input:     `<div>{{theme = 'blue'}}</div>`,
			expectMsg: "template syntax error on line 1",
		},
		{
			name: "Multiline template should identify correct line",
			input: `<div>Content</div>
<style>{{color = 'red'}}</style>
<div>More content</div>`,
			expectMsg: "template syntax error on line 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.input)
			if err == nil {
				t.Error("Expected validation error but got none")
				return
			}

			if !strings.Contains(err.Error(), tt.expectMsg) {
				t.Errorf("Expected error message containing %q, got %q", tt.expectMsg, err.Error())
			}
		})
	}
}

// TestValidateTemplateRealWorldExamples tests with real-world problematic HTML patterns
func TestValidateTemplateRealWorldExamples(t *testing.T) {
	tests := []struct {
		name        string
		description string
		input       string
		expectError bool
	}{
		{
			name:        "Gmail-style CSS custom properties",
			description: "CSS custom properties commonly found in Gmail and other Google services",
			input: `<style>
				:root {
					--gm-border-color: {{borderColor = '#dadce0'}};
					--gm-primary-color: {{primaryColor = '#1a73e8'}};
				}
			</style>`,
			expectError: true,
		},
		{
			name:        "React component props",
			description: "JavaScript patterns from React applications",
			input: `<script>
				window.__INITIAL_STATE__ = {{
					user = { id = 123, name = "John" },
					config = { theme = "dark", locale = "en" }
				}};
			</script>`,
			expectError: true,
		},
		{
			name:        "Angular template syntax",
			description: "Angular-style template patterns",
			input: `<div>
				{{title = 'My App'}}
				{{user.name = 'Default User'}}
			</div>`,
			expectError: true,
		},
		{
			name:        "JSON-LD schema",
			description: "Structured data commonly found in web pages",
			input: `<script type="application/ld+json">
				{{
					"@type" = "WebPage",
					"name" = {{pageName = "Login Page"}},
					"url" = {{pageUrl = "https://example.com/login"}}
				}}
			</script>`,
			expectError: true,
		},
		{
			name:        "Vue.js template syntax",
			description: "Vue.js style templates that might be present in SPAs",
			input: `<div>
				{{message = 'Hello World'}}
				{{counter = 0}}
			</div>`,
			expectError: true,
		},
		{
			name:        "Valid Gophish templates (should pass)",
			description: "Legitimate Gophish template patterns",
			input: `<div>
				<h1>Welcome {{.FirstName}}!</h1>
				<p>Visit this link: {{.URL}}</p>
				{{.Tracker}}
			</div>`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.input)

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

// TestValidateTemplatePropertyBased uses property-based testing for template validation
func TestValidateTemplatePropertyBased(t *testing.T) {
	// Property: Valid Gophish template variables should always pass validation
	validGophishTemplates := func(firstName, lastName, position string) bool {
		template := `<div>Welcome {{.FirstName}} {{.LastName}}! Position: {{.Position}}</div>{{.Tracker}}<a href="{{.URL}}">Click</a>`
		err := ValidateTemplate(template)
		return err == nil
	}

	if err := quick.Check(validGophishTemplates, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property-based test for valid Gophish templates failed: %v", err)
	}

	// Property: Templates with assignment operators should fail preprocessing
	templatesWithAssignment := func(varName, value string) bool {
		if varName == "" || value == "" {
			return true // Skip empty inputs
		}
		template := `{{` + varName + ` = '` + value + `'}}`
		err := validateTemplateContent(template)
		return err != nil // Should always error
	}

	if err := quick.Check(templatesWithAssignment, &quick.Config{MaxCount: 100}); err != nil {
		t.Errorf("Property-based test for templates with assignment failed: %v", err)
	}
}

// TestValidateTemplateLargeInputs tests validation with large template content
func TestValidateTemplateLargeInputs(t *testing.T) {
	// Test with large valid template
	largeValidTemplate := strings.Repeat(`<div>{{.FirstName}} content {{.LastName}}</div>`, 1000)
	err := ValidateTemplate(largeValidTemplate)
	if err != nil {
		t.Errorf("Large valid template failed validation: %v", err)
	}

	// Test with large invalid template
	largeInvalidTemplate := strings.Repeat(`<div>{{invalid = 'pattern'}}</div>`, 100)
	err = ValidateTemplate(largeInvalidTemplate)
	if err == nil {
		t.Error("Large invalid template should have failed validation")
	}

	// Test with mixed large template
	mixedTemplate := strings.Repeat(`<div>{{.FirstName}}</div>`, 500) + 
		`<style>{{theme = 'blue'}}</style>` +
		strings.Repeat(`<div>{{.LastName}}</div>`, 500)
	err = ValidateTemplate(mixedTemplate)
	if err == nil {
		t.Error("Mixed large template with invalid pattern should have failed validation")
	}
}

// TestValidateTemplateSecurityPatterns tests security-related template patterns
func TestValidateTemplateSecurityPatterns(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		description string
	}{
		{
			name:        "Script injection attempt",
			input:       `{{script = '<script>alert("xss")</script>'}}`,
			expectError: true,
			description: "Should catch potential script injection in templates",
		},
		{
			name:        "SQL injection-like pattern",
			input:       `{{query = "SELECT * FROM users WHERE id = 1"}}`,
			expectError: true,
			description: "Should catch SQL-like patterns in templates",
		},
		{
			name:        "File path injection attempt",
			input:       `{{file = "../../../etc/passwd"}}`,
			expectError: true,
			description: "Should catch path traversal patterns",
		},
		{
			name:        "Command injection attempt",
			input:       `{{cmd = "rm -rf /"}}`,
			expectError: true,
			description: "Should catch command-like patterns",
		},
		{
			name:        "Valid Gophish security patterns",
			input:       `<form action="{{.URL}}" method="post"><input type="hidden" value="{{.RId}}"></form>`,
			expectError: false,
			description: "Legitimate Gophish security patterns should pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected validation error for %s", tt.description)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error for %s, got: %v", tt.description, err)
				}
			}
		})
	}
}

// TestValidateTemplateUnicodeAndSpecialChars tests validation with unicode and special characters
func TestValidateTemplateUnicodeAndSpecialChars(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "Unicode in invalid template",
			input:       `{{emoji = '🚀'}} {{chinese = '中文'}}`,
			expectError: true,
		},
		{
			name:        "Unicode in valid template",
			input:       `<div>Welcome {{.FirstName}}! 🚀 欢迎 {{.LastName}}</div>`,
			expectError: false,
		},
		{
			name:        "Special characters in invalid template",
			input:       `{{special = '@#$%^&*()_+-=[]{}|;:,.<>?'}}`,
			expectError: true,
		},
		{
			name:        "HTML entities in template",
			input:       `<div>&lt;{{.FirstName}}&gt; &amp; {{.LastName}}</div>`,
			expectError: false,
		},
		{
			name:        "Escaped HTML in invalid template",
			input:       `{{html = '&lt;script&gt;alert("xss")&lt;/script&gt;'}}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTemplate(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected validation error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no validation error but got: %v", err)
				}
			}
		})
	}
}

// BenchmarkValidateTemplate benchmarks template validation performance
func BenchmarkValidateTemplate(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Small valid template",
			input: `<div>{{.FirstName}}</div>`,
		},
		{
			name:  "Small invalid template",
			input: `<div>{{invalid = 'test'}}</div>`,
		},
		{
			name:  "Medium valid template",
			input: strings.Repeat(`<div>{{.FirstName}} content {{.LastName}}</div>`, 50),
		},
		{
			name:  "Medium invalid template",
			input: strings.Repeat(`<style>{{theme = 'blue'}}</style>`, 50),
		},
		{
			name:  "Large valid template",
			input: strings.Repeat(`<div>{{.FirstName}} content {{.LastName}}</div>`, 500),
		},
		{
			name:  "Large invalid template",
			input: strings.Repeat(`<script>{{config = 'value'}}</script>`, 500),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ValidateTemplate(tc.input)
			}
		})
	}
}

// BenchmarkValidateTemplateContent benchmarks the preprocessing function
func BenchmarkValidateTemplateContent(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Small content",
			input: `{{test = 'value'}}`,
		},
		{
			name:  "Medium content",
			input: strings.Repeat(`{{var = 'value'}} `, 100),
		},
		{
			name:  "Large content",
			input: strings.Repeat(`{{var = 'value'}} `, 1000),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				validateTemplateContent(tc.input)
			}
		})
	}
}