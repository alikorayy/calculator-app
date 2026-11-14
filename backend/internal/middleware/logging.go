package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Logging returns middleware that writes one structured log line per request
// once it has been handled: method, path, status, response size and duration.
// Server errors (5xx) are logged at ERROR level, everything else at INFO.
//
// It must run inside RequestID so the line can carry the request ID; the ID is
// added by the handler created with NewRequestIDLogHandler.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			level := slog.LevelInfo
			if recorder.status >= http.StatusInternalServerError {
				level = slog.LevelError
			}
			logger.LogAttrs(r.Context(), level, logMessageRequest,
				slog.String(logKeyMethod, r.Method),
				slog.String(logKeyPath, r.URL.Path),
				slog.Int(logKeyStatus, recorder.status),
				slog.Int(logKeyBytes, recorder.bytes),
				slog.Float64(logKeyDurationMS, float64(time.Since(start))/float64(time.Millisecond)),
			)
		})
	}
}

// statusRecorder remembers the status code and the number of body bytes a
// handler writes, while passing everything through to the real writer.
type statusRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(status int) {
	if !s.wroteHeader {
		s.status = status
		s.wroteHeader = true
	}
	s.ResponseWriter.WriteHeader(status)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	s.wroteHeader = true
	n, err := s.ResponseWriter.Write(b)
	s.bytes += n
	return n, err
}

// Unwrap gives http.ResponseController access to the underlying writer.
func (s *statusRecorder) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

// NewRequestIDLogHandler wraps a slog.Handler so that every record logged with
// a request's context (logger.InfoContext, logger.LogAttrs, ...) carries that
// request's ID. All log lines of one request can then be found by its ID.
func NewRequestIDLogHandler(next slog.Handler) slog.Handler {
	return requestIDLogHandler{Handler: next}
}

type requestIDLogHandler struct {
	slog.Handler
}

func (h requestIDLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if id := RequestIDFromContext(ctx); id != "" {
		record.AddAttrs(slog.String(logKeyRequestID, id))
	}
	return h.Handler.Handle(ctx, record)
}

func (h requestIDLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return requestIDLogHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h requestIDLogHandler) WithGroup(name string) slog.Handler {
	return requestIDLogHandler{Handler: h.Handler.WithGroup(name)}
}
