package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newBufferLogger returns a JSON logger that adds request IDs and writes into buf.
func newBufferLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(NewRequestIDLogHandler(slog.NewJSONHandler(buf, nil)))
}

// decodeLogLines parses every JSON log line written to buf.
func decodeLogLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var lines []map[string]any
	for _, raw := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		var line map[string]any
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			t.Fatalf("log line is not JSON: %q", raw)
		}
		lines = append(lines, line)
	}
	return lines
}

// serveLogged runs one request through RequestID and Logging around handler and
// returns the decoded log lines and the response.
func serveLogged(t *testing.T, handler http.HandlerFunc) ([]map[string]any, *httptest.ResponseRecorder) {
	t.Helper()
	var buf bytes.Buffer
	chain := RequestID(Logging(newBufferLogger(&buf))(handler))

	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", nil))
	return decodeLogLines(t, &buf), rec
}

func TestLogging_WritesOneLinePerRequest(t *testing.T) {
	lines, rec := serveLogged(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	})

	if len(lines) != 1 {
		t.Fatalf("got %d log lines, want 1", len(lines))
	}
	line := lines[0]
	want := map[string]any{
		"level":         "INFO",
		"msg":           logMessageRequest,
		logKeyMethod:    http.MethodPost,
		logKeyPath:      "/api/v1/evaluate",
		logKeyStatus:    float64(http.StatusCreated),
		logKeyBytes:     float64(len("hello")),
		logKeyRequestID: rec.Header().Get(HeaderRequestID),
	}
	for key, value := range want {
		if line[key] != value {
			t.Errorf("%s = %v, want %v", key, line[key], value)
		}
	}
	if duration, ok := line[logKeyDurationMS].(float64); !ok || duration < 0 {
		t.Errorf("%s = %v, want a non-negative number", logKeyDurationMS, line[logKeyDurationMS])
	}
}

func TestLogging_StatusDefaultsTo200WhenOnlyTheBodyIsWritten(t *testing.T) {
	lines, _ := serveLogged(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	if got := lines[0][logKeyStatus]; got != float64(http.StatusOK) {
		t.Errorf("status = %v, want 200", got)
	}
}

func TestLogging_ServerErrorsAreLoggedAtErrorLevel(t *testing.T) {
	tests := []struct {
		status    int
		wantLevel string
	}{
		{http.StatusUnprocessableEntity, "INFO"},
		{http.StatusInternalServerError, "ERROR"},
		{http.StatusServiceUnavailable, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			lines, _ := serveLogged(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			})
			if got := lines[0]["level"]; got != tt.wantLevel {
				t.Errorf("level = %v, want %s", got, tt.wantLevel)
			}
		})
	}
}

func TestLogging_HandlerLogsShareTheRequestID(t *testing.T) {
	var buf bytes.Buffer
	logger := newBufferLogger(&buf)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.ErrorContext(r.Context(), "something failed inside the handler")
		w.WriteHeader(http.StatusInternalServerError)
	})

	rec := httptest.NewRecorder()
	RequestID(Logging(logger)(handler)).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	lines := decodeLogLines(t, &buf)
	if len(lines) != 2 {
		t.Fatalf("got %d log lines, want 2", len(lines))
	}
	id := rec.Header().Get(HeaderRequestID)
	for _, line := range lines {
		if line[logKeyRequestID] != id {
			t.Errorf("line %q has request_id %v, want %q", line["msg"], line[logKeyRequestID], id)
		}
	}
}

func TestStatusRecorder_KeepsTheFirstStatus(t *testing.T) {
	recorder := &statusRecorder{ResponseWriter: httptest.NewRecorder(), status: http.StatusOK}

	recorder.WriteHeader(http.StatusNotFound)
	recorder.WriteHeader(http.StatusInternalServerError)

	if recorder.status != http.StatusNotFound {
		t.Errorf("status = %d, want %d", recorder.status, http.StatusNotFound)
	}
}

func TestStatusRecorder_Unwrap(t *testing.T) {
	inner := httptest.NewRecorder()
	recorder := &statusRecorder{ResponseWriter: inner}

	if recorder.Unwrap() != inner {
		t.Error("Unwrap did not return the wrapped writer")
	}
}

func TestRequestIDLogHandler(t *testing.T) {
	const id = "abc-123"
	var buf bytes.Buffer
	logger := newBufferLogger(&buf).With("service", "calculator-api")
	ctx := context.WithValue(context.Background(), requestIDKey{}, id)

	logger.InfoContext(ctx, "with id")
	logger.Info("without id")
	logger.WithGroup("details").InfoContext(ctx, "grouped")

	lines := decodeLogLines(t, &buf)
	if got := lines[0][logKeyRequestID]; got != id {
		t.Errorf("request_id = %v, want %q", got, id)
	}
	if got := lines[0]["service"]; got != "calculator-api" {
		t.Errorf("attributes added with With were lost: service = %v", got)
	}
	if _, found := lines[1][logKeyRequestID]; found {
		t.Errorf("line logged without a request context has a request_id: %v", lines[1])
	}
	if group, _ := lines[2]["details"].(map[string]any); group[logKeyRequestID] != id {
		t.Errorf("grouped line lost the request ID: %v", lines[2])
	}
}
