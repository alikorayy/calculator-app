package middleware

// HeaderRequestID carries the request ID, both on incoming requests (set by a
// proxy or client) and on every response.
const HeaderRequestID = "X-Request-ID"

// maxRequestIDLength bounds an incoming request ID so that a client cannot
// flood the logs through it.
const maxRequestIDLength = 64
