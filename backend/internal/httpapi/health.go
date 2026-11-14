package httpapi

import "net/http"

// HealthResponse is the body returned by the health check endpoint.
type HealthResponse struct {
	Status string `json:"status"`
}

// handleHealth reports that the process is up and able to serve requests.
func handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: healthStatusOK})
}
