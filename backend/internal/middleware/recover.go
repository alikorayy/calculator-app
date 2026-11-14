package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// ErrorWriter writes an error response in the API's format. Middleware takes
// it as a parameter so that this package does not depend on the HTTP layer
// that defines the format.
type ErrorWriter func(w http.ResponseWriter, status int, code, message string)

// Recover returns middleware that turns a panic in any later handler into a
// logged error and a 500 response, instead of a dropped connection. The log
// line includes the panic value and the stack trace; the client only gets the
// generic message, so no internal details leak.
//
// http.ErrAbortHandler is re-panicked: it is the standard library's way to
// abort a response on purpose and must reach the server unchanged.
func Recover(logger *slog.Logger, writeError ErrorWriter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				recovered := recover()
				if recovered == nil {
					return
				}
				if err, ok := recovered.(error); ok && errors.Is(err, http.ErrAbortHandler) {
					panic(recovered)
				}

				logger.ErrorContext(r.Context(), logMessagePanic,
					slog.Any(logKeyPanic, recovered),
					slog.String(logKeyStack, string(debug.Stack())),
				)
				writeError(w, http.StatusInternalServerError, codeInternalError, messageInternalError)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
