package httpapi

// Routes served by the API, in Go's "METHOD /path" pattern form.
const (
	routeHealth = "GET /healthz"
)

// Response header values.
const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

// Error codes returned in the error envelope. Clients branch on these, so they never change.
const (
	codeInternalError = "INTERNAL_ERROR"
)

// Error messages returned in the error envelope.
const (
	messageInternalError = "An unexpected error occurred"
)

// Health check response values.
const (
	healthStatusOK = "ok"
)
