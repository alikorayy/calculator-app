// Package httpapi exposes the calculator over HTTP: routing, request decoding
// and JSON responses. It contains no calculation logic.
package httpapi

import "net/http"

// NewRouter returns the HTTP handler with every route of the API registered.
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(routeHealth, handleHealth)
	return mux
}
