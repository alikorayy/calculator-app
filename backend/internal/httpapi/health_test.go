package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth_ReturnsOK(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	newTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get(headerContentType); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
	}

	var body HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Status != healthStatusOK {
		t.Errorf("status field = %q, want %q", body.Status, healthStatusOK)
	}
}

func TestHealth_RejectsOtherMethods(t *testing.T) {
	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(method, "/healthz", nil)

			newTestRouter().ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s /healthz status = %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}
