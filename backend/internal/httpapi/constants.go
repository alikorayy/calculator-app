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

// Health check response values.
const (
	healthStatusOK = "ok"
)
