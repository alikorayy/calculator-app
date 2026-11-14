package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckHealth(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"healthy", http.StatusOK, false},
		{"unhealthy", http.StatusServiceUnavailable, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			if err := checkHealth(server.URL); (err != nil) != tt.wantErr {
				t.Errorf("checkHealth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckHealth_UnreachableServer(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	url := server.URL
	server.Close()

	if err := checkHealth(url); err == nil {
		t.Error("checkHealth() of a closed server returned nil, want an error")
	}
}

func TestHealthcheckURL(t *testing.T) {
	if got, want := healthcheckURL(8080), "http://127.0.0.1:8080/healthz"; got != want {
		t.Errorf("healthcheckURL(8080) = %q, want %q", got, want)
	}
}
