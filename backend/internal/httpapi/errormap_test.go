package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alikorayy/calculator-app/backend/internal/engine"
)

func TestStatusForEngineCode(t *testing.T) {
	tests := []struct {
		code       engine.Code
		wantStatus int
		wantCode   string
	}{
		{engine.CodeSyntaxError, http.StatusUnprocessableEntity, "SYNTAX_ERROR"},
		{engine.CodeDivisionByZero, http.StatusUnprocessableEntity, "DIVISION_BY_ZERO"},
		{engine.CodeUndefinedResult, http.StatusUnprocessableEntity, "UNDEFINED_RESULT"},
		{engine.CodeDomainError, http.StatusUnprocessableEntity, "DOMAIN_ERROR"},
		{engine.CodeOverflow, http.StatusUnprocessableEntity, "OVERFLOW"},
		{engine.CodeLimitExceeded, http.StatusUnprocessableEntity, "LIMIT_EXCEEDED"},
		{engine.CodeInvalidArgument, http.StatusBadRequest, "INVALID_REQUEST"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			status, code, ok := statusForEngineCode(tt.code)
			if !ok || status != tt.wantStatus || code != tt.wantCode {
				t.Errorf("statusForEngineCode(%s) = (%d, %q, %v), want (%d, %q, true)",
					tt.code, status, code, ok, tt.wantStatus, tt.wantCode)
			}
		})
	}
}

func TestWriteEngineError_UnexpectedErrorsAre500WithoutDetails(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"not an engine error", errors.New("database exploded")},
		{"engine error with an unmapped code", &engine.Error{Code: "NEW_CODE", Message: "secret detail", Position: engine.NoPosition}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			writeEngineError(context.Background(), rec, discardLogger(), tt.err)

			assertResponse(t, rec, http.StatusInternalServerError,
				`{"error":{"code":"INTERNAL_ERROR","message":"An unexpected error occurred"}}`)
		})
	}
}
