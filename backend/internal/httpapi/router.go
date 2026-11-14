// Package httpapi exposes the calculator over HTTP: routing, request decoding
// and JSON responses. It contains no calculation logic.
package httpapi

import (
	"log/slog"
	"net/http"
)

// NewRouter returns the HTTP handler with every route of the API registered.
// Unexpected internal errors are logged to logger. Unknown paths and
// unsupported methods are answered with the JSON error envelope.
func NewRouter(logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(routeHealth, handleHealth)
	mux.HandleFunc(routeEvaluate, handleEvaluate(logger))
	return withJSONFallback(mux)
}
