package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurityHeaders_AreSetOnEveryResponse(t *testing.T) {
	want := map[string]string{
		"X-Content-Type-Options":       "nosniff",
		"Content-Security-Policy":      "default-src 'none'; frame-ancestors 'none'",
		"Referrer-Policy":              "no-referrer",
		"Cross-Origin-Resource-Policy": "same-origin",
	}

	for _, status := range []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(status)
			}))

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

			for name, value := range want {
				if got := rec.Header().Get(name); got != value {
					t.Errorf("%s = %q, want %q", name, got, value)
				}
			}
		})
	}
}
