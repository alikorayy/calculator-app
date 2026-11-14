package httpapi

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/alikorayy/calculator-app/backend/internal/engine"
)

// CalculateRequest is the body of POST /api/v1/calculate.
type CalculateRequest struct {
	// Operation is one of add, subtract, multiply, divide, power, sqrt, percent.
	Operation *string `json:"operation"`
	// Operands are decimal numbers, each given as a JSON string ("12.5") or a
	// JSON number (12.5). They are kept as raw JSON so that a number's exact
	// text reaches the engine without passing through float64.
	Operands []json.RawMessage `json:"operands"`
}

// handleCalculate returns the handler for POST /api/v1/calculate.
func handleCalculate(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CalculateRequest
		if failure := decodeJSONBody(r, &req); failure != nil {
			writeDecodeFailure(w, failure)
			return
		}
		if req.Operation == nil {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, fmt.Sprintf(messageMissingField, fieldOperation))
			return
		}
		if req.Operands == nil {
			writeError(w, http.StatusBadRequest, codeInvalidRequest, fmt.Sprintf(messageMissingField, fieldOperands))
			return
		}
		operands, failure := operandTexts(req.Operands)
		if failure != nil {
			writeDecodeFailure(w, failure)
			return
		}

		result, err := engine.Calculate(engine.Operation(*req.Operation), operands)
		if err != nil {
			writeEngineError(w, logger, err)
			return
		}
		writeJSON(w, http.StatusOK, CalculationResponse{Result: result.Value, Approximate: result.Approximate})
	}
}

// operandTexts returns the text of each operand: the content of a JSON
// string, or the literal of a JSON number exactly as it was sent. Any other
// JSON value (null, true, an object, an array) is rejected. Whether the text is
// a valid decimal number is decided by the engine.
func operandTexts(raw []json.RawMessage) ([]string, *decodeFailure) {
	texts := make([]string, len(raw))
	for i, value := range raw {
		switch {
		case len(value) > 0 && value[0] == '"':
			// The decoder has already validated the JSON, so unquoting cannot fail.
			_ = json.Unmarshal(value, &texts[i])
		case len(value) > 0 && (value[0] == '-' || isASCIIDigit(value[0])):
			texts[i] = string(value)
		default:
			return nil, invalidRequest(fmt.Sprintf(messageInvalidOperandType, i+1))
		}
	}
	return texts, nil
}

func isASCIIDigit(b byte) bool {
	return b >= '0' && b <= '9'
}
