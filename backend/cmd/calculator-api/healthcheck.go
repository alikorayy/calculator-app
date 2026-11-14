package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/alikorayy/calculator-app/backend/internal/httpapi"
)

// healthcheckCommand is the argument that makes the binary check a running
// server instead of starting one: "calculator-api healthcheck". The runtime
// image has no shell or curl, so the container health check uses the binary
// itself.
const healthcheckCommand = "healthcheck"

// healthcheckTimeout bounds how long the health check waits for an answer.
const healthcheckTimeout = 2 * time.Second

// healthcheckURL is the health endpoint of the server running in the same container.
func healthcheckURL(port int) string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", port, httpapi.HealthPath)
}

// checkHealth reports an error unless url answers 200 OK within the timeout.
func checkHealth(url string) error {
	client := &http.Client{Timeout: healthcheckTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("healthcheck: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthcheck: %s answered %d", url, resp.StatusCode)
	}
	return nil
}
