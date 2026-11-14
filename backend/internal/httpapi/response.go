package httpapi

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse is the envelope of every error response returned by the API.
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody describes a single error: a stable machine-readable code, a
// human-readable message and, for syntax errors, the 0-based character position
// in the expression. Position is a pointer because 0 is a valid position; it
// is omitted from the JSON when there is none.
type ErrorBody struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Position *int   `json:"position,omitempty"`
}

// writeJSON encodes body as JSON and writes it with the given status code.
// If body cannot be encoded, a 500 error envelope is written instead, so the
// client never receives a half-written or empty response.
func writeJSON(w http.ResponseWriter, status int, body any) {
	payload, err := json.Marshal(body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, codeInternalError, messageInternalError)
		return
	}

	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(status)
	// A write error means the client has gone away; there is nobody left to report it to.
	_, _ = w.Write(payload)
}

// writeError writes an error response using the standard error envelope.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Error: ErrorBody{Code: code, Message: message}})
}
