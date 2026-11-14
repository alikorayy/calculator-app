package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// serveWithRequestID runs one request through RequestID and returns the ID the
// handler saw in its context together with the recorded response.
func serveWithRequestID(t *testing.T, incomingID string) (seenID string, rec *httptest.ResponseRecorder) {
	t.Helper()
	handler := RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seenID = RequestIDFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if incomingID != "" {
		req.Header.Set(HeaderRequestID, incomingID)
	}
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return seenID, rec
}

func TestRequestID_GeneratesAnIDWhenMissing(t *testing.T) {
	seenID, rec := serveWithRequestID(t, "")

	if seenID == "" {
		t.Fatal("handler saw no request ID")
	}
	if got := rec.Header().Get(HeaderRequestID); got != seenID {
		t.Errorf("response header = %q, context = %q; want the same ID", got, seenID)
	}
}

func TestRequestID_GeneratedIDsAreUnique(t *testing.T) {
	first, _ := serveWithRequestID(t, "")
	second, _ := serveWithRequestID(t, "")

	if first == second {
		t.Errorf("two requests got the same ID %q", first)
	}
}

func TestRequestID_KeepsAValidIncomingID(t *testing.T) {
	const incoming = "proxy-generated_ID.42"

	seenID, rec := serveWithRequestID(t, incoming)

	if seenID != incoming {
		t.Errorf("context ID = %q, want %q", seenID, incoming)
	}
	if got := rec.Header().Get(HeaderRequestID); got != incoming {
		t.Errorf("response header = %q, want %q", got, incoming)
	}
}

func TestRequestID_ReplacesAnInvalidIncomingID(t *testing.T) {
	tests := []struct {
		name     string
		incoming string
	}{
		{"too long", strings.Repeat("a", maxRequestIDLength+1)},
		{"forged log line", "abc\nlevel=ERROR msg=fake"},
		{"spaces", "my id"},
		{"non-ASCII", "kimlik-ğ"},
		{"quotes", `abc"def`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seenID, _ := serveWithRequestID(t, tt.incoming)

			if seenID == tt.incoming || seenID == "" {
				t.Errorf("context ID = %q, want a newly generated ID", seenID)
			}
		})
	}
}

func TestRequestIDFromContext_EmptyWithoutMiddleware(t *testing.T) {
	if got := RequestIDFromContext(context.Background()); got != "" {
		t.Errorf("RequestIDFromContext = %q, want empty", got)
	}
}
