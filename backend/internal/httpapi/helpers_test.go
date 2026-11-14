package httpapi

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// discardLogger drops every log record so tests stay quiet.
func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// newTestRouter returns the API router with a silent logger.
func newTestRouter() http.Handler {
	return NewRouter(discardLogger())
}

// postJSON sends body to path through the router with the given Content-Type
// and returns the recorded response.
func postJSON(t *testing.T, path, contentType, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set(headerContentType, contentType)
	}
	rec := httptest.NewRecorder()
	newTestRouter().ServeHTTP(rec, req)
	return rec
}
