package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON_WritesStatusHeaderAndBody(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
	}
	rec := httptest.NewRecorder()

	writeJSON(rec, http.StatusCreated, payload{Name: "calculator"})

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
	if got := rec.Header().Get(headerContentType); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
	}
	if got, want := rec.Body.String(), `{"name":"calculator"}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

func TestWriteJSON_UnencodableBodyReturnsInternalError(t *testing.T) {
	rec := httptest.NewRecorder()

	// A channel has no JSON representation, so encoding fails.
	writeJSON(rec, http.StatusOK, make(chan int))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	var body ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Error.Code != codeInternalError {
		t.Errorf("error code = %q, want %q", body.Error.Code, codeInternalError)
	}
}

func TestWriteError_WritesEnvelope(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		code     string
		message  string
		wantBody string
	}{
		{
			name:     "bad request",
			status:   http.StatusBadRequest,
			code:     "INVALID_REQUEST",
			message:  "Request body is not valid JSON",
			wantBody: `{"error":{"code":"INVALID_REQUEST","message":"Request body is not valid JSON"}}`,
		},
		{
			name:     "unprocessable math",
			status:   http.StatusUnprocessableEntity,
			code:     "DIVISION_BY_ZERO",
			message:  "Cannot divide by zero",
			wantBody: `{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			WriteError(rec, tt.status, tt.code, tt.message)

			if rec.Code != tt.status {
				t.Errorf("status = %d, want %d", rec.Code, tt.status)
			}
			if got := rec.Header().Get(headerContentType); got != contentTypeJSON {
				t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}
