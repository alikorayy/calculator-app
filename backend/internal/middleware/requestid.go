package middleware

import (
	"context"
	"crypto/rand"
	"net/http"
)

// requestIDKey is the context key under which the request ID is stored. An
// unexported type guarantees no other package can collide with it.
type requestIDKey struct{}

// RequestID gives every request an ID. A well-formed incoming X-Request-ID
// (for example one set by the reverse proxy) is kept, so one ID can follow a
// request across services; otherwise a new random ID is generated. The ID is
// returned in the X-Request-ID response header and stored in the request
// context, where RequestIDFromContext reads it.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if !isValidRequestID(id) {
			id = rand.Text()
		}

		w.Header().Set(HeaderRequestID, id)
		ctx := context.WithValue(r.Context(), requestIDKey{}, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFromContext returns the request ID stored by RequestID, or "" if
// there is none.
func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// isValidRequestID accepts short IDs made only of letters, digits, '-', '_'
// and '.'. Anything else, such as a newline that could forge a log line, is
// replaced by a generated ID.
func isValidRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLength {
		return false
	}
	for _, r := range id {
		if !isRequestIDChar(r) {
			return false
		}
	}
	return true
}

func isRequestIDChar(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	default:
		return r == '-' || r == '_' || r == '.'
	}
}
