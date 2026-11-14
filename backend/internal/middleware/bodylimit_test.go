package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testBodyLimit = 16

// readingHandler reads the whole body and reports what happened through its results.
func readingHandler(body *string, readErr *error) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		*body, *readErr = string(data), err
		w.WriteHeader(http.StatusOK)
	})
}

func TestMaxBodySize_AcceptsBodiesUpToTheLimit(t *testing.T) {
	for _, size := range []int{0, testBodyLimit - 1, testBodyLimit} {
		var body string
		var readErr error
		sent := strings.Repeat("a", size)

		rec := httptest.NewRecorder()
		MaxBodySize(testBodyLimit, plainErrorWriter)(readingHandler(&body, &readErr)).
			ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(sent)))

		if rec.Code != http.StatusOK || readErr != nil || body != sent {
			t.Errorf("size %d: status %d, read error %v, body %q", size, rec.Code, readErr, body)
		}
	}
}

func TestMaxBodySize_RejectsADeclaredLengthAboveTheLimitWithoutReading(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", testBodyLimit+1)))
	rec := httptest.NewRecorder()
	MaxBodySize(testBodyLimit, plainErrorWriter)(next).ServeHTTP(rec, req)

	if called {
		t.Error("the handler ran for a body that was too large")
	}
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", rec.Code)
	}
	if want := "PAYLOAD_TOO_LARGE: Request body is too large (maximum 16 bytes)"; rec.Body.String() != want {
		t.Errorf("body = %q, want %q", rec.Body.String(), want)
	}
}

func TestMaxBodySize_StopsAnUndeclaredBodyAtTheLimit(t *testing.T) {
	var body string
	var readErr error
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", testBodyLimit*4)))
	req.ContentLength = -1 // unknown length, as with chunked transfer encoding

	MaxBodySize(testBodyLimit, plainErrorWriter)(readingHandler(&body, &readErr)).
		ServeHTTP(httptest.NewRecorder(), req)

	var maxBytesErr *http.MaxBytesError
	if !errors.As(readErr, &maxBytesErr) {
		t.Fatalf("read error = %v, want *http.MaxBytesError", readErr)
	}
	if len(body) > testBodyLimit {
		t.Errorf("handler read %d bytes, more than the limit %d", len(body), testBodyLimit)
	}
}
