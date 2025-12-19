package httplog

import (
	"encoding/json"
	"net/http"
	"strings"
)

const redactedValue = "[REDACTED]"

// Redactor handles sensitive data masking.
type Redactor struct {
	keywords map[string]bool
	enabled  bool
}

// NewRedactor creates a redactor with the given keywords.
func NewRedactor(keywords []string) *Redactor {
	if len(keywords) == 0 {
		return &Redactor{enabled: false}
	}

	keywordMap := make(map[string]bool)
	for _, keyword := range keywords {
		keywordMap[strings.ToLower(keyword)] = true
	}

	return &Redactor{
		keywords: keywordMap,
		enabled:  true,
	}
}

// RedactHeaders masks sensitive headers based on configured keywords.
func (r *Redactor) RedactHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	for key, values := range headers {
		if r.enabled && r.shouldRedact(key) {
			result[key] = redactedValue
		} else {
			result[key] = strings.Join(values, ", ")
		}
	}

	return result
}

// RedactBody masks sensitive JSON fields based on configured keywords.
// Returns the original body if it's not valid JSON or redaction is disabled.
func (r *Redactor) RedactBody(body string) string {
	if !r.enabled || body == "" {
		return body
	}

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		// Not JSON, return as-is
		return body
	}

	// Redact sensitive keys recursively
	r.redactMap(data)

	// Re-serialize
	redacted, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(redacted)
}

// shouldRedact checks if a key matches any redaction keyword.
func (r *Redactor) shouldRedact(key string) bool {
	// Normalize key
	normalizedKey := strings.ToLower(key)
	normalizedKey = strings.ReplaceAll(normalizedKey, "-", "")
	normalizedKey = strings.ReplaceAll(normalizedKey, "_", "")

	// Check for exact match with normalized keywords
	for keyword := range r.keywords {
		normalizedKeyword := strings.ReplaceAll(keyword, "_", "")
		normalizedKeyword = strings.ReplaceAll(normalizedKeyword, "-", "")

		if normalizedKey == normalizedKeyword || strings.Contains(normalizedKey, normalizedKeyword) {
			return true
		}
	}

	return false
}

// redactMap recursively redacts sensitive keys in a map.
func (r *Redactor) redactMap(data map[string]interface{}) {
	for key, value := range data {
		if r.shouldRedact(key) {
			data[key] = redactedValue
			continue
		}

		// Recurse into nested objects
		switch v := value.(type) {
		case map[string]interface{}:
			r.redactMap(v)
		case []interface{}:
			r.redactSlice(v)
		}
	}
}

// redactSlice recursively redacts sensitive keys in a slice.
func (r *Redactor) redactSlice(data []interface{}) {
	for _, item := range data {
		if m, ok := item.(map[string]interface{}); ok {
			r.redactMap(m)
		}
	}
}
