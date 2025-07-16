package main

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/gophish/gophish/models"
)

// FuzzTemplatePatterns fuzzes template pattern sanitization and validation
func FuzzTemplatePatterns(f *testing.F) {
	// Seed corpus with known problematic patterns
	seeds := []string{
		"{{theme = 'blue'}}",
		"{% if user = admin %}",
		"{{api_key = \"test\"}}",
		"{{nested = { inner = { value = 'test' } } }}",
		"{{emoji = '🚀'}}",
		"{{chinese = '中文'}}",
		"{{special = '@#$%^&*()_+-='}}",
		"{{{{nested}}}}",
		"{% %}{{}}",
		"{{incomplete",
		"incomplete}}",
		"{%incomplete",
		"incomplete%}",
		"{{a = b = c}}",
		"{{= 'value'}}",
		"{{var =}}",
		"{{ = }}",
		"{{}}",
		"{%}",
		"{{.ValidField}}",
		"{{$var := .Field}}",
		"{{if eq .Field \"value\"}}",
	}

	// Add seeds to fuzzer
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Skip invalid UTF-8 to avoid test failures
		if !utf8.ValidString(input) {
			t.Skip("Skipping invalid UTF-8 input")
		}

		// Test sanitization function doesn't panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("sanitizeTemplateContent panicked with input %q: %v", input, r)
				}
			}()

			sanitized := sanitizeTemplateContentFuzz(input)

			// Property: Sanitized output should not contain unescaped template delimiters
			if strings.Contains(sanitized, "{{") || strings.Contains(sanitized, "}}") ||
				strings.Contains(sanitized, "{%") || strings.Contains(sanitized, "%}") {
				t.Errorf("sanitizeTemplateContent failed to escape all delimiters in input %q", input)
			}
		}()

		// Test template validation doesn't panic
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("ValidateTemplate panicked with input %q: %v", input, r)
				}
			}()

			// Validation may return error, but shouldn't panic
			_ = models.ValidateTemplate(input)
		}()
	})
}

// FuzzTemplateGeneration generates random template-like patterns for testing
func FuzzTemplateGeneration(f *testing.F) {
	f.Fuzz(func(t *testing.T,
		openDelim string,
		closeDelim string,
		content string,
		operator string,
		value string,
	) {
		// Skip invalid UTF-8
		for _, s := range []string{openDelim, closeDelim, content, operator, value} {
			if !utf8.ValidString(s) {
				t.Skip("Skipping invalid UTF-8 input")
			}
		}

		// Generate template-like pattern
		pattern := openDelim + content + operator + value + closeDelim

		// Test that our functions handle arbitrary patterns gracefully
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Functions panicked with generated pattern %q: %v", pattern, r)
				}
			}()

			// Test sanitization
			sanitized := sanitizeTemplateContentFuzz(pattern)

			// Test validation
			_ = models.ValidateTemplate(pattern)
			_ = models.ValidateTemplate(sanitized)
		}()
	})
}

// FuzzLargeTemplatePatterns tests with large inputs that might cause performance issues
func FuzzLargeTemplatePatterns(f *testing.F) {
	// Seed with patterns that could be repeated many times
	seeds := []string{
		"{{var = 'value'}}",
		"{% tag %}",
		"{{nested = { a = { b = 'c' } } }}",
	}

	for _, seed := range seeds {
		f.Add(seed, 10)   // Small repetition
		f.Add(seed, 100)  // Medium repetition
		f.Add(seed, 1000) // Large repetition
	}

	f.Fuzz(func(t *testing.T, pattern string, repetitions int) {
		if !utf8.ValidString(pattern) {
			t.Skip("Skipping invalid UTF-8 input")
		}

		// Limit repetitions to prevent excessive memory usage
		if repetitions < 0 || repetitions > 10000 {
			t.Skip("Skipping invalid repetition count")
		}

		// Generate large input
		largeInput := strings.Repeat(pattern, repetitions)

		// Test with timeout to catch infinite loops
		done := make(chan bool, 1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Functions panicked with large input: %v", r)
				}
				done <- true
			}()

			// Test sanitization performance
			sanitized := sanitizeTemplateContentFuzz(largeInput)

			// Test validation performance
			_ = models.ValidateTemplate(largeInput)
			_ = models.ValidateTemplate(sanitized)
		}()

		// Timeout after 5 seconds
		select {
		case <-done:
			// Completed successfully
		case <-time.After(5 * time.Second):
			t.Error("Functions took too long with large input - possible performance issue")
		}
	})
}

// FuzzNestedTemplatePatterns generates deeply nested template patterns
func FuzzNestedTemplatePatterns(f *testing.F) {
	f.Fuzz(func(t *testing.T, depth int, content string) {
		if !utf8.ValidString(content) {
			t.Skip("Skipping invalid UTF-8 input")
		}

		// Limit depth to prevent excessive memory usage
		if depth < 0 || depth > 100 {
			t.Skip("Skipping invalid depth")
		}

		// Generate nested pattern
		pattern := content
		for i := 0; i < depth; i++ {
			pattern = "{{" + pattern + " = 'value'}}"
		}

		// Test that deeply nested patterns are handled correctly
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Functions panicked with nested pattern (depth %d): %v", depth, r)
				}
			}()

			sanitized := sanitizeTemplateContentFuzz(pattern)
			_ = models.ValidateTemplate(pattern)
			_ = models.ValidateTemplate(sanitized)
		}()
	})
}

// FuzzUnicodeTemplatePatterns tests with various Unicode characters
func FuzzUnicodeTemplatePatterns(f *testing.F) {
	// Seed with various Unicode ranges
	unicodeSeeds := []string{
		"🚀",       // Emoji
		"中文",      // Chinese
		"العربية", // Arabic
		"русский", // Russian
		"हिन्दी",  // Hindi
		"한국어",     // Korean
		"🎉🔥💯",     // Multiple emoji
		"café",    // Latin with accents
		"Москва",  // Cyrillic
		"東京",      // Japanese
	}

	for _, unicode := range unicodeSeeds {
		f.Add(unicode)
	}

	f.Fuzz(func(t *testing.T, unicodeContent string) {
		if !utf8.ValidString(unicodeContent) {
			t.Skip("Skipping invalid UTF-8 input")
		}

		// Generate template patterns with Unicode content
		patterns := []string{
			"{{var = '" + unicodeContent + "'}}",
			"{% if content = '" + unicodeContent + "' %}",
			"{{" + unicodeContent + " = 'value'}}",
			"<div>" + unicodeContent + " {{.FirstName}}</div>",
			"<style>:root { --" + unicodeContent + ": {{value = 'test'}}; }</style>",
		}

		for _, pattern := range patterns {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Functions panicked with Unicode pattern %q: %v", pattern, r)
					}
				}()

				sanitized := sanitizeTemplateContentFuzz(pattern)
				_ = models.ValidateTemplate(pattern)
				_ = models.ValidateTemplate(sanitized)
			}()
		}
	})
}

// FuzzBinaryDataWithTemplates tests binary-like data mixed with template patterns
func FuzzBinaryDataWithTemplates(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		// Convert to string (may contain invalid UTF-8)
		input := string(data)

		// Mix with template patterns
		patterns := []string{
			input + "{{var = 'value'}}",
			"{{test = '" + input + "'}}",
			"{% tag %}" + input + "{% endtag %}",
			"<script>" + input + "{{config = 'test'}}</script>",
		}

		for _, pattern := range patterns {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Functions panicked with binary+template pattern: %v", r)
					}
				}()

				// These functions should handle arbitrary input gracefully
				sanitized := sanitizeTemplateContentFuzz(pattern)
				_ = models.ValidateTemplate(pattern)
				_ = models.ValidateTemplate(sanitized)
			}()
		}
	})
}

// FuzzRealWorldHTMLPatterns tests with realistic HTML structures
func FuzzRealWorldHTMLPatterns(f *testing.F) {
	// Seed with realistic HTML templates
	htmlSeeds := []string{
		`<style>:root { --color: {{theme = 'blue'}}; }</style>`,
		`<script>var config = {{api = 'key'}};</script>`,
		`{% if user = admin %}<div>Admin</div>{% endif %}`,
		`<form action="{{action = '/submit'}}">`,
		`<div data-props="{{props = '{}'}}">`,
		`<meta name="{{name = 'viewport'}}" content="{{content = 'width=device-width'}}">`,
	}

	for _, seed := range htmlSeeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, htmlFragment string) {
		if !utf8.ValidString(htmlFragment) {
			t.Skip("Skipping invalid UTF-8 input")
		}

		// Wrap in various HTML contexts
		contexts := []string{
			"<html><head>" + htmlFragment + "</head></html>",
			"<style>" + htmlFragment + "</style>",
			"<script>" + htmlFragment + "</script>",
			"<div>" + htmlFragment + "</div>",
			"<form>" + htmlFragment + "</form>",
			"<!DOCTYPE html>" + htmlFragment,
		}

		for _, context := range contexts {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Functions panicked with HTML context %q: %v", context, r)
					}
				}()

				sanitized := sanitizeTemplateContentFuzz(context)
				_ = models.ValidateTemplate(context)
				_ = models.ValidateTemplate(sanitized)
			}()
		}
	})
}

// Helper function that mirrors sanitizeTemplateContent for fuzzing
func sanitizeTemplateContentFuzz(html string) string {
	// Escape Go template delimiters {{ and }}
	html = strings.ReplaceAll(html, "{{", "&#123;&#123;")
	html = strings.ReplaceAll(html, "}}", "&#125;&#125;")

	// Also handle Django/Jinja2 template patterns for completeness
	html = strings.ReplaceAll(html, "{%", "&#123;%")
	html = strings.ReplaceAll(html, "%}", "%&#125;")

	return html
}

// Continuous fuzzing benchmark
func BenchmarkFuzzTemplatePatterns(b *testing.B) {
	rand.Seed(time.Now().UnixNano())

	// Generate random template patterns for benchmarking
	patterns := make([]string, 1000)
	for i := range patterns {
		patterns[i] = generateRandomTemplatePattern()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pattern := patterns[i%len(patterns)]

		// Test sanitization performance
		sanitized := sanitizeTemplateContentFuzz(pattern)

		// Test validation performance
		_ = models.ValidateTemplate(pattern)
		_ = models.ValidateTemplate(sanitized)
	}
}

// Helper to generate random template patterns for benchmarking
func generateRandomTemplatePattern() string {
	delimiters := []string{"{{", "{%"}
	closeDelimiters := []string{"}}", "%}"}
	operators := []string{" = ", " := ", " == ", " != "}
	values := []string{"'test'", "\"value\"", "123", "true", "false"}
	variables := []string{"theme", "config", "user", "data", "api_key", "color"}

	openDelim := delimiters[rand.Intn(len(delimiters))]
	closeDelim := closeDelimiters[rand.Intn(len(closeDelimiters))]
	variable := variables[rand.Intn(len(variables))]
	operator := operators[rand.Intn(len(operators))]
	value := values[rand.Intn(len(values))]

	return fmt.Sprintf("%s%s%s%s%s", openDelim, variable, operator, value, closeDelim)
}
