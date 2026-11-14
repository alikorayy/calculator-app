package httpapi

import (
	"encoding/json"
	"net/http"
)

// HealthResponse is the body returned by the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// handleHealth reports that the process is up and able to serve requests.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set(headerContentType, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{Status: healthStatusOK})
}
