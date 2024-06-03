//go:build go1.7
// +build go1.7

package context

import (
	"net/http"

	"context"
)

// Get retrieves a value from the request context
func Get(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

// Set stores a value on the request context
func Set(r *http.Request, key, val interface{}) *http.Request {
	if val == nil {
		return r
	}

	return r.WithContext(context.WithValue(r.Context(), key, val))
}

// Clear is a null operation, since this is handled automatically in Go > 1.7
func Clear(r *http.Request) {}
