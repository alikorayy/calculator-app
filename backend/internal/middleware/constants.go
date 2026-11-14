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

// Security headers set on every response. The API only returns JSON, so it
// needs no scripts, styles, frames or other resources at all.
const (
	headerContentTypeOptions = "X-Content-Type-Options"
	valueNoSniff             = "nosniff"

	headerContentSecurityPolicy = "Content-Security-Policy"
	valueContentSecurityPolicy  = "default-src 'none'; frame-ancestors 'none'"

	headerReferrerPolicy = "Referrer-Policy"
	valueNoReferrer      = "no-referrer"

	headerCrossOriginResourcePolicy = "Cross-Origin-Resource-Policy"
	valueSameOrigin                 = "same-origin"
)

// Error codes and messages of the responses written by this package. They
// match the codes of the API's error envelope.
const (
	codeInternalError    = "INTERNAL_ERROR"
	messageInternalError = "An unexpected error occurred"

	codePayloadTooLarge    = "PAYLOAD_TOO_LARGE"
	messagePayloadTooLarge = "Request body is too large (maximum %d bytes)"

	codeTimeout    = "TIMEOUT"
	messageTimeout = "The request took too long to process"

	codeRateLimited    = "RATE_LIMITED"
	messageRateLimited = "Too many requests, please slow down"
)

// Rate limiting.
const (
	headerForwardedFor = "X-Forwarded-For"
	headerRetryAfter   = "Retry-After"
	// ipv6ClientPrefixBits is the size of the IPv6 network treated as one client.
	ipv6ClientPrefixBits = 64
)
