package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFallback_UnmatchedRequestsUseJSONEnvelope(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		path      string
		wantCode  int
		wantBody  string
		wantAllow string
	}{
		{
			name:     "unknown path",
			method:   http.MethodGet,
			path:     "/anything",
			wantCode: http.StatusNotFound,
			wantBody: `{"error":{"code":"NOT_FOUND","message":"The requested resource does not exist"}}`,
		},
		{
			name:      "wrong method on known path",
			method:    http.MethodPost,
			path:      "/healthz",
			wantCode:  http.StatusMethodNotAllowed,
			wantBody:  `{"error":{"code":"METHOD_NOT_ALLOWED","message":"The method is not allowed for the requested resource"}}`,
			wantAllow: "GET, HEAD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, nil)

			newTestRouter().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
			if got := rec.Header().Get(headerContentType); got != contentTypeJSON {
				t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("Allow = %q, want %q", got, tt.wantAllow)
			}
		})
	}
}

func TestFallback_PassesThroughOtherStatuses(t *testing.T) {
	rec := httptest.NewRecorder()
	// An unclean path makes the mux answer with a redirect, which must not be rewritten.
	req := httptest.NewRequest(http.MethodGet, "//anything", nil)

	newTestRouter().ServeHTTP(rec, req)

	// The exact redirect code is the standard library's choice (301 in Go 1.25,
	// 307 in later releases); what matters is that it is passed through.
	if rec.Code < http.StatusMultipleChoices || rec.Code >= http.StatusBadRequest {
		t.Errorf("status = %d, want a 3xx redirect", rec.Code)
	}
	if got, want := rec.Header().Get("Location"), "/anything"; got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
}
