package main

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alikorayy/calculator-app/backend/internal/config"
	"github.com/alikorayy/calculator-app/backend/internal/middleware"
)

const evaluatePath = "/api/v1/evaluate"

// newTestServer builds the full handler chain with the default configuration,
// logging into logs.
func newTestServer(t *testing.T, logs *bytes.Buffer, overrides map[string]string) http.Handler {
	t.Helper()
	cfg, err := config.Load(func(key string) (string, bool) {
		value, ok := overrides[key]
		return value, ok
	})
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	logger := slog.New(middleware.NewRequestIDLogHandler(slog.NewJSONHandler(logs, nil)))
	return newHandler(cfg, logger, newRateLimiter(cfg))
}

func evaluate(handler http.Handler, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, evaluatePath, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestHandler_EndToEnd(t *testing.T) {
	var logs bytes.Buffer
	rec := evaluate(newTestServer(t, &logs, nil), `{"expression":"50 - 7 × 3"}`)

	if rec.Code != http.StatusOK || rec.Body.String() != `{"result":"29","approximate":false}` {
		t.Fatalf("got %d %s, want 200 with result 29", rec.Code, rec.Body.String())
	}
	id := rec.Header().Get(middleware.HeaderRequestID)
	if id == "" {
		t.Error("response has no request ID")
	}
	for _, header := range []string{"X-Content-Type-Options", "Content-Security-Policy", "Referrer-Policy"} {
		if rec.Header().Get(header) == "" {
			t.Errorf("security header %s is missing", header)
		}
	}

	var line map[string]any
	if err := json.Unmarshal(logs.Bytes(), &line); err != nil {
		t.Fatalf("expected exactly one JSON log line, got %q", logs.String())
	}
	if line["request_id"] != id || line["status"] != float64(http.StatusOK) || line["path"] != evaluatePath {
		t.Errorf("log line = %v, want status 200, path %s and request_id %s", line, evaluatePath, id)
	}
}

func TestHandler_EarlyAnswersStillGetTheEnvelopeAndHeaders(t *testing.T) {
	var logs bytes.Buffer
	handler := newTestServer(t, &logs, map[string]string{
		config.EnvMaxBodyBytes:       "64",
		config.EnvRateLimitPerSecond: "1",
		config.EnvRateLimitBurst:     "2",
	})

	tooLarge := evaluate(handler, `{"expression":"`+strings.Repeat("1", 100)+`"}`)
	if tooLarge.Code != http.StatusRequestEntityTooLarge ||
		!strings.Contains(tooLarge.Body.String(), `"code":"PAYLOAD_TOO_LARGE"`) {
		t.Errorf("large body: got %d %s, want a 413 envelope", tooLarge.Code, tooLarge.Body.String())
	}

	// The 413 above used one token of the burst of 2; one more is allowed.
	evaluate(handler, `{"expression":"1"}`)
	limited := evaluate(handler, `{"expression":"1"}`)
	if limited.Code != http.StatusTooManyRequests ||
		!strings.Contains(limited.Body.String(), `"code":"RATE_LIMITED"`) {
		t.Errorf("third request: got %d %s, want a 429 envelope", limited.Code, limited.Body.String())
	}
	if limited.Header().Get("Retry-After") == "" || limited.Header().Get("X-Content-Type-Options") == "" {
		t.Errorf("429 response is missing Retry-After or security headers: %v", limited.Header())
	}
}
