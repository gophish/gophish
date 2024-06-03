//go:build !go1.7
// +build !go1.7

package context

import (
	"net/http"

	"github.com/gorilla/context"
)

func Get(r *http.Request, key interface{}) interface{} {
	return context.Get(r, key)
}

func Set(r *http.Request, key, val interface{}) *http.Request {
	if val == nil {
		return r
	}

	context.Set(r, key, val)
	return r
}

func Clear(r *http.Request) {
	context.Clear(r)
}
