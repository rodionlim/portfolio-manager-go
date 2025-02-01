//go:build !builtinassets

package ui

import (
	"net/http"
)

// AssetsHandler returns a no-op handler (here returning 404) when the build tag isn't specified.
func AssetsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}
