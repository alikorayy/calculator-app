package middleware

import (
	"bytes"
	"net/http"
	"time"
)

// Timeout returns middleware that answers with 503 when a handler runs longer
// than limit. It builds on http.TimeoutHandler, which buffers the handler's
// response and discards it if the deadline passes first.
//
// A goroutine cannot be stopped from outside, so a slow handler keeps running
// until it returns; what the deadline does is cancel the request's context, so
// handlers that check r.Context() can stop early, and the client gets its
// answer on time either way.
//
// http.TimeoutHandler only writes a fixed text body, so the error envelope is
// rendered once, up front, with writeError. Its headers (the JSON
// Content-Type) are set before the handler runs; a handler that answers in
// time replaces them with its own.
func Timeout(limit time.Duration, writeError ErrorWriter) func(http.Handler) http.Handler {
	timeoutResponse := renderError(writeError, http.StatusServiceUnavailable, codeTimeout, messageTimeout)

	return func(next http.Handler) http.Handler {
		timeoutHandler := http.TimeoutHandler(next, limit, timeoutResponse.body.String())
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for name, values := range timeoutResponse.header {
				w.Header()[name] = values
			}
			timeoutHandler.ServeHTTP(w, r)
		})
	}
}

// renderedResponse is an http.ResponseWriter that keeps a response in memory.
type renderedResponse struct {
	header http.Header
	body   bytes.Buffer
}

func (r *renderedResponse) Header() http.Header         { return r.header }
func (r *renderedResponse) Write(b []byte) (int, error) { return r.body.Write(b) }
func (r *renderedResponse) WriteHeader(int)             {}

// renderError captures the response writeError produces, so it can be reused.
func renderError(writeError ErrorWriter, status int, code, message string) *renderedResponse {
	rendered := &renderedResponse{header: http.Header{}}
	writeError(rendered, status, code, message)
	return rendered
}
