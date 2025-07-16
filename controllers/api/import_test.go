package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"testing/quick"
)

// TestSanitizeTemplateContent tests the sanitizeTemplateContent function with various inputs
func TestSanitizeTemplateContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CSS custom properties",
			input:    `:root { --var: {{value = 'test'}}; }`,
			expected: `:root { --var: &#123;&#123;value = 'test'&#125;&#125;; }`,
		},
		{
			name:     "JavaScript JSON config",
			input:    `var config = {{key = "value"}};`,
			expected: `var config = &#123;&#123;key = "value"&#125;&#125;;`,
		},
		{
			name:     "Django/Jinja2 templates",
			input:    `{% if condition = true %}content{% endif %}`,
			expected: `&#123;% if condition = true %&#125;content&#123;% endif %&#125;`,
		},
		{
			name:     "Valid Go templates should be escaped too",
			input:    `{{.FirstName}}`,
			expected: `&#123;&#123;.FirstName&#125;&#125;`,
		},
		{
			name:     "Mixed content",
			input:    `<style>:root{--color:{{theme='blue'}}}</style><script>var x={{key='val'}}</script>{% tag %}`,
			expected: `<style>:root{--color:&#123;&#123;theme='blue'&#125;&#125;}</style><script>var x=&#123;&#123;key='val'&#125;&#125;</script>&#123;% tag %&#125;`,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "No template patterns",
			input:    `<div>Regular HTML content</div>`,
			expected: `<div>Regular HTML content</div>`,
		},
		{
			name:     "Nested patterns",
			input:    `{{outer {{inner = value}} }}`,
			expected: `&#123;&#123;outer &#123;&#123;inner = value&#125;&#125; &#125;&#125;`,
		},
		{
			name:     "Unicode characters",
			input:    `{{emoji = '🚀'}} {{chinese = '中文'}}`,
			expected: `&#123;&#123;emoji = '🚀'&#125;&#125; &#123;&#123;chinese = '中文'&#125;&#125;`,
		},
		{
			name:     "Special characters in templates",
			input:    `{{special = '@#$%^&*()_+-=[]{}|;:,.<>?'}}`,
			expected: `&#123;&#123;special = '@#$%^&*()_+-=[]{}|;:,.<>?'&#125;&#125;`,
		},
		{
			name:     "Malformed templates",
			input:    `{{{incomplete {{another = test} {% malformed`,
			expected: `&#123;&#123;{incomplete &#123;&#123;another = test} &#123;% malformed`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTemplateContent(tt.input)
			if result != tt.expected {
				t.Errorf("sanitizeTemplateContent() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestSanitizeTemplateContentLargeInput tests performance with large inputs
func TestSanitizeTemplateContentLargeInput(t *testing.T) {
	// Create a large HTML document with many template patterns
	var builder strings.Builder
	builder.WriteString("<html><head><style>")

	// Add 1000 CSS custom properties with template patterns
	for i := 0; i < 1000; i++ {
		builder.WriteString(fmt.Sprintf("--var%d: {{value%d = 'test%d'}}; ", i, i, i))
	}

	builder.WriteString("</style><script>")

	// Add 1000 JavaScript template patterns
	for i := 0; i < 1000; i++ {
		builder.WriteString(fmt.Sprintf("var config%d = {{key%d = 'value%d'}}; ", i, i, i))
	}

	builder.WriteString("</script></head><body>")

	// Add Django-style templates
	for i := 0; i < 500; i++ {
		builder.WriteString(fmt.Sprintf("{%% if condition%d = true %%}content%d{%% endif %%} ", i, i))
	}

	builder.WriteString("</body></html>")

	input := builder.String()

	// Test that it doesn't panic and produces expected output
	result := sanitizeTemplateContent(input)

	// Verify that all patterns were escaped
	if strings.Contains(result, "{{") || strings.Contains(result, "}}") ||
		strings.Contains(result, "{%") || strings.Contains(result, "%}") {
		t.Error("sanitizeTemplateContent() failed to escape all template patterns in large input")
	}

	// Verify expected escapes are present
	if !strings.Contains(result, "&#123;&#123;") || !strings.Contains(result, "&#125;&#125;") {
		t.Error("sanitizeTemplateContent() failed to apply expected escapes in large input")
	}
}

// TestSanitizeTemplateContentPropertyBased uses property-based testing
func TestSanitizeTemplateContentPropertyBased(t *testing.T) {
	property := func(input string) bool {
		result := sanitizeTemplateContent(input)

		// Property: Result should not contain unescaped template delimiters
		hasUnescapedTemplates := strings.Contains(result, "{{") ||
			strings.Contains(result, "}}") ||
			strings.Contains(result, "{%") ||
			strings.Contains(result, "%}")

		return !hasUnescapedTemplates
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 1000}); err != nil {
		t.Errorf("Property-based test failed: %v", err)
	}
}

// TestImportSiteHandlerSuccess tests the ImportSite HTTP handler with successful cases
func TestImportSiteHandlerSuccess(t *testing.T) {
	tests := []struct {
		name         string
		htmlFile     string
		expectEscape bool
	}{
		{
			name:         "Gmail problematic HTML",
			htmlFile:     "../../testdata/gmail_problematic.html",
			expectEscape: true,
		},
		{
			name:         "Django template HTML",
			htmlFile:     "../../testdata/django_template.html",
			expectEscape: true,
		},
		{
			name:         "React app HTML",
			htmlFile:     "../../testdata/react_app.html",
			expectEscape: true,
		},
		{
			name:         "Valid Gophish template",
			htmlFile:     "../../testdata/valid_gophish_template.html",
			expectEscape: true, // Even valid templates get escaped during import
		},
		{
			name:         "Mixed templates",
			htmlFile:     "../../testdata/mixed_templates.html",
			expectEscape: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Read test HTML file
			htmlContent, err := os.ReadFile(tt.htmlFile)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.htmlFile, err)
			}

			// Create a test server that serves the HTML content
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/html")
				w.Write(htmlContent)
			}))
			defer testServer.Close()

			// Create request to ImportSite handler
			reqBody := cloneRequest{
				URL:              testServer.URL,
				IncludeResources: false,
			}
			reqJSON, _ := json.Marshal(reqBody)

			req := httptest.NewRequest("POST", "/api/import/site", bytes.NewReader(reqJSON))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			// Create server instance and call handler
			server := &Server{}
			server.ImportSite(w, req)

			// Check response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status %d, got %d. Response: %s", http.StatusOK, w.Code, w.Body.String())
				return
			}

			var response cloneResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// Verify HTML was processed
			if response.HTML == "" {
				t.Error("Expected non-empty HTML in response")
			}

			// Verify sanitization was applied
			if tt.expectEscape {
				originalHasTemplates := strings.Contains(string(htmlContent), "{{") || strings.Contains(string(htmlContent), "{%")
				responseHasTemplates := strings.Contains(response.HTML, "{{") || strings.Contains(response.HTML, "{%")

				if originalHasTemplates && responseHasTemplates {
					t.Error("Expected template patterns to be escaped in response")
				}

				// Verify escaped patterns are present
				if originalHasTemplates && (!strings.Contains(response.HTML, "&#123;&#123;") && !strings.Contains(response.HTML, "&#123;%")) {
					t.Error("Expected escaped template patterns in response")
				}
			}

			// Verify base href was added
			if !strings.Contains(response.HTML, fmt.Sprintf(`<base href="%s">`, testServer.URL)) {
				t.Error("Expected base href to be added to HTML")
			}
		})
	}
}

// TestImportSiteHandlerErrors tests error cases for ImportSite handler
func TestImportSiteHandlerErrors(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Invalid method",
			method:         "GET",
			requestBody:    cloneRequest{URL: "http://example.com"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Method not allowed",
		},
		{
			name:           "Invalid JSON",
			method:         "POST",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Error decoding JSON Request",
		},
		{
			name:           "Empty URL",
			method:         "POST",
			requestBody:    cloneRequest{URL: ""},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "No URL Specified",
		},
		{
			name:           "Invalid URL",
			method:         "POST",
			requestBody:    cloneRequest{URL: "not-a-url"},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/import/site", bytes.NewReader(reqBody))
			if tt.method == "POST" {
				req.Header.Set("Content-Type", "application/json")
			}

			w := httptest.NewRecorder()

			server := &Server{}
			server.ImportSite(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}

				if message, exists := response["message"].(string); !exists || !strings.Contains(message, tt.expectedError) {
					t.Errorf("Expected error message containing %q, got %q", tt.expectedError, message)
				}
			}
		})
	}
}

// TestImportSiteFormProcessing tests that form processing still works after sanitization
func TestImportSiteFormProcessing(t *testing.T) {
	htmlWithForms := `<!DOCTYPE html>
<html>
<head>
    <style>:root { --color: {{theme = 'blue'}}; }</style>
</head>
<body>
    <form action="/login" method="post">
        <input type="email" name="email">
        <input type="password" name="password">
        <button type="submit">Login</button>
    </form>
    <form action="/register" method="post">
        <input type="text" name="username">
        <input type="email" name="email">
    </form>
</body>
</html>`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlWithForms))
	}))
	defer testServer.Close()

	reqBody := cloneRequest{
		URL:              testServer.URL,
		IncludeResources: false,
	}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/import/site", bytes.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	server := &Server{}
	server.ImportSite(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response cloneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify template patterns were escaped
	if strings.Contains(response.HTML, "{{theme = 'blue'}}") {
		t.Error("Expected template patterns to be escaped")
	}

	// Verify forms were processed correctly
	expectedHiddenInputs := []string{
		`<input type="hidden" name="__original_url" value="` + testServer.URL + `/login"/>`,
		`<input type="hidden" name="__original_url" value="` + testServer.URL + `/register"/>`,
	}

	for _, expected := range expectedHiddenInputs {
		if !strings.Contains(response.HTML, expected) {
			t.Errorf("Expected hidden input %q to be added to form", expected)
		}
	}

	// Verify base href was added
	expectedBase := fmt.Sprintf(`<base href="%s">`, testServer.URL)
	if !strings.Contains(response.HTML, expectedBase) {
		t.Errorf("Expected base href %q to be added", expectedBase)
	}
}

// BenchmarkSanitizeTemplateContent benchmarks the sanitization function
func BenchmarkSanitizeTemplateContent(b *testing.B) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name:  "Small HTML",
			input: `<div>{{simple = 'test'}}</div>`,
		},
		{
			name:  "Medium HTML",
			input: strings.Repeat(`<style>:root{--var{{i}}:{{value{{i}} = 'test{{i}}'}}}</style>`, 100),
		},
		{
			name:  "Large HTML",
			input: strings.Repeat(`<style>:root{--var{{i}}:{{value{{i}} = 'test{{i}}'}}}</style><script>var x{{i}}={{key{{i}} = 'val{{i}}'}}</script>`, 1000),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sanitizeTemplateContent(tc.input)
			}
		})
	}
}

// BenchmarkImportSiteHandler benchmarks the full ImportSite handler
func BenchmarkImportSiteHandler(b *testing.B) {
	// Create test HTML content
	htmlContent := `<!DOCTYPE html><html><head><style>:root{--color:{{theme='blue'}}}</style></head><body><form><input name="test"></form></body></html>`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(htmlContent))
	}))
	defer testServer.Close()

	reqBody := cloneRequest{
		URL:              testServer.URL,
		IncludeResources: false,
	}
	reqJSON, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/api/import/site", bytes.NewReader(reqJSON))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		server := &Server{}
		server.ImportSite(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Handler failed with status %d", w.Code)
		}
	}
}
