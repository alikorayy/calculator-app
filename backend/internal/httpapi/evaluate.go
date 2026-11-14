package httpapi

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alikorayy/calculator-app/backend/internal/engine"
)

// EvaluateRequest is the body of POST /api/v1/evaluate.
type EvaluateRequest struct {
	// Expression is the calculator expression, e.g. "50 - 7 × 3". It is a
	// pointer so that a missing field can be told apart from an empty string.
	Expression *string `json:"expression"`
}

// CalculationResponse is the body of a successful calculation.
type CalculationResponse struct {
	// Result is the value as a decimal string, e.g. "29" or "0.3333333333333333".
	Result string `json:"result"`
	// Approximate is true when Result was rounded.
	Approximate bool `json:"approximate"`
}

// handleEvaluate returns the handler for POST /api/v1/evaluate.
func handleEvaluate(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EvaluateRequest
		if failure := decodeJSONBody(r, &req); failure != nil {
			writeDecodeFailure(w, failure)
			return
		}
		if req.Expression == nil {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, fmt.Sprintf(messageMissingField, fieldExpression))
			return
		}

		result, err := engine.Evaluate(*req.Expression)
		if err != nil {
			writeEngineError(r.Context(), w, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, CalculationResponse{Result: result.Value, Approximate: result.Approximate})
	}
}
