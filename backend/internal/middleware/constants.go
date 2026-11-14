package middleware

// HeaderRequestID carries the request ID, both on incoming requests (set by a
// proxy or client) and on every response.
const HeaderRequestID = "X-Request-ID"

// maxRequestIDLength bounds an incoming request ID so that a client cannot
// flood the logs through it.
const maxRequestIDLength = 64

// Keys of the structured log fields written by this package.
const (
	logKeyRequestID  = "request_id"
	logKeyMethod     = "method"
	logKeyPath       = "path"
	logKeyStatus     = "status"
	logKeyBytes      = "bytes"
	logKeyDurationMS = "duration_ms"
)

// logMessageRequest is the message of the one log line written per request.
const logMessageRequest = "request completed"

// Panic recovery.
const (
	logMessagePanic = "panic recovered"
	logKeyPanic     = "panic"
	logKeyStack     = "stack"
)

// Error codes and messages of the responses written by this package. They
// match the codes of the API's error envelope.
const (
	codeInternalError    = "INTERNAL_ERROR"
	messageInternalError = "An unexpected error occurred"
)
