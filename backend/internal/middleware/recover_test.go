package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// plainErrorWriter is a minimal ErrorWriter that records what it was asked to write.
func plainErrorWriter(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, "%s: %s", code, message)
}

func TestRecover_TurnsAPanicIntoA500(t *testing.T) {
	var buf bytes.Buffer
	logger := newBufferLogger(&buf)
	panicking := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("nil map write in some handler")
	})

	rec := httptest.NewRecorder()
	RequestID(Recover(logger, plainErrorWriter)(panicking)).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
	if want := codeInternalError + ": " + messageInternalError; rec.Body.String() != want {
		t.Errorf("body = %q, want %q", rec.Body.String(), want)
	}

	lines := decodeLogLines(t, &buf)
	if len(lines) != 1 {
		t.Fatalf("got %d log lines, want 1", len(lines))
	}
	line := lines[0]
	if line["level"] != "ERROR" || line["msg"] != logMessagePanic {
		t.Errorf("log line = %v, want an ERROR %q", line, logMessagePanic)
	}
	if line[logKeyPanic] != "nil map write in some handler" {
		t.Errorf("panic = %v, want the panic value", line[logKeyPanic])
	}
	if stack, _ := line[logKeyStack].(string); !strings.Contains(stack, "TestRecover_TurnsAPanicIntoA500") {
		t.Errorf("stack trace does not point at the panicking code: %.200s", stack)
	}
	if line[logKeyRequestID] != rec.Header().Get(HeaderRequestID) {
		t.Errorf("panic log has request_id %v, want %q", line[logKeyRequestID], rec.Header().Get(HeaderRequestID))
	}
}

func TestRecover_PassesNormalRequestsThrough(t *testing.T) {
	var buf bytes.Buffer
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	rec := httptest.NewRecorder()
	Recover(newBufferLogger(&buf), plainErrorWriter)(handler).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if buf.Len() != 0 {
		t.Errorf("unexpected log output: %s", buf.String())
	}
}

func TestRecover_ReraisesErrAbortHandler(t *testing.T) {
	var buf bytes.Buffer
	aborting := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic(http.ErrAbortHandler)
	})

	defer func() {
		if recovered := recover(); recovered != http.ErrAbortHandler {
			t.Errorf("recovered %v, want http.ErrAbortHandler to be re-panicked", recovered)
		}
	}()
	Recover(newBufferLogger(&buf), plainErrorWriter)(aborting).
		ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
}
