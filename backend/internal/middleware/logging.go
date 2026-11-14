package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"slices"
	"time"
)

// Logging returns middleware that writes one structured log line per request
// once it has been handled: method, path, status, response size and duration.
// Server errors (5xx) are logged at ERROR level, everything else at INFO.
//
// Successful requests to quietPaths are logged at DEBUG instead. This is meant
// for health checks, which an orchestrator calls every few seconds: they would
// otherwise flood the log, yet a failing health check stays visible.
//
// It must run inside RequestID so the line can carry the request ID; the ID is
// added by the handler created with NewRequestIDLogHandler.
func Logging(logger *slog.Logger, quietPaths ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(recorder, r)

			level := requestLogLevel(recorder.status, slices.Contains(quietPaths, r.URL.Path))
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

// requestLogLevel chooses the level of a request's log line from its status
// and whether its path is quiet.
func requestLogLevel(status int, quiet bool) slog.Level {
	switch {
	case status >= http.StatusInternalServerError:
		return slog.LevelError
	case quiet && status < http.StatusBadRequest:
		return slog.LevelDebug
	default:
		return slog.LevelInfo
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
