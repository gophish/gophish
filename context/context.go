// +build go1.7

package context

import (
	"net/http"

	"context"
)

func Get(r *http.Request, key interface{}) interface{} {
	return r.Context().Value(key)
}

func Set(r *http.Request, key, val interface{}) *http.Request {
	if val == nil {
		return r
	}

	return r.WithContext(context.WithValue(r.Context(), key, val))
}

func Clear(r *http.Request) {
	return
}
