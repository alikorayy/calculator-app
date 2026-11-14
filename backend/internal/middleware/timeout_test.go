package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	testTimeout     = 20 * time.Millisecond
	testContentType = "application/json"
)

// jsonErrorWriter imitates the API's ErrorWriter, including its Content-Type.
func jsonErrorWriter(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", testContentType)
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"error":{"code":%q,"message":%q}}`, code, message)
}

func TestTimeout_FastHandlersAreUnaffected(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", testContentType)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"result":"29"}`))
	})

	rec := httptest.NewRecorder()
	Timeout(testTimeout, jsonErrorWriter)(handler).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusCreated || rec.Body.String() != `{"result":"29"}` {
		t.Errorf("got %d %s, want the handler's own response", rec.Code, rec.Body.String())
	}
}

func TestTimeout_SlowHandlersGetA503AndACancelledContext(t *testing.T) {
	contextErr := make(chan error, 1)
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
		contextErr <- r.Context().Err()
		_, _ = w.Write([]byte("too late, this must be discarded"))
	})

	rec := httptest.NewRecorder()
	Timeout(testTimeout, jsonErrorWriter)(slow).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != testContentType {
		t.Errorf("Content-Type = %q, want %q", got, testContentType)
	}
	if want := `{"error":{"code":"TIMEOUT","message":"The request took too long to process"}}`; rec.Body.String() != want {
		t.Errorf("body = %s, want %s", rec.Body.String(), want)
	}
	if err := <-contextErr; !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("handler context error = %v, want context.DeadlineExceeded", err)
	}
}
