package middleware

import (
	"fmt"
	"net/http"
)

// MaxBodySize returns middleware that limits request bodies to limit bytes,
// so a huge body cannot exhaust the server's memory.
//
// The limit is enforced in two places. A request that declares a larger
// Content-Length is answered with 413 right away, without reading the body.
// A request without a declared length (chunked transfer) gets a body reader
// that fails with *http.MaxBytesError once limit is passed; the handler that
// reads the body reports that as 413 as well.
func MaxBodySize(limit int64, writeError ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > limit {
				writeError(w, http.StatusRequestEntityTooLarge, codePayloadTooLarge, fmt.Sprintf(messagePayloadTooLarge, limit))
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, limit)
			next.ServeHTTP(w, r)
		})
	}
}
