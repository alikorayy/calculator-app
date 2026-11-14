package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/alikorayy/calculator-app/backend/internal/engine"
)

// statusForEngineCode is the single place where engine error codes are mapped
// to HTTP. A request that is well formed but whose math cannot be done is
// 422 Unprocessable Entity, so the client can show the message to the user;
// a malformed Calculate call is 400 Bad Request, like every other client bug,
// and is reported with the generic INVALID_REQUEST code.
//
// ok is false for a code this table does not know, which is an internal bug.
func statusForEngineCode(code engine.Code) (status int, publicCode string, ok bool) {
	switch code {
	case engine.CodeSyntaxError,
		engine.CodeDivisionByZero,
		engine.CodeUndefinedResult,
		engine.CodeDomainError,
		engine.CodeOverflow,
		engine.CodeLimitExceeded:
		return http.StatusUnprocessableEntity, string(code), true
	case engine.CodeInvalidArgument:
		return http.StatusBadRequest, codeInvalidRequest, true
	default:
		return 0, "", false
	}
}

// writeEngineError writes an error returned by the engine. Anything that is not
// an *engine.Error with a known code is unexpected: it is logged and answered
// with a generic 500, without leaking internal details to the client.
func writeEngineError(w http.ResponseWriter, logger *slog.Logger, err error) {
	var engineErr *engine.Error
	if !errors.As(err, &engineErr) {
		logger.Error("unexpected engine error", "error", err)
		writeError(w, http.StatusInternalServerError, codeInternalError, messageInternalError)
		return
	}

	status, code, ok := statusForEngineCode(engineErr.Code)
	if !ok {
		logger.Error("engine error code has no HTTP mapping", "code", engineErr.Code, "error", err)
		writeError(w, http.StatusInternalServerError, codeInternalError, messageInternalError)
		return
	}

	body := ErrorBody{Code: code, Message: engineErr.Message}
	if engineErr.Position != engine.NoPosition {
		position := engineErr.Position
		body.Position = &position
	}
	writeJSON(w, status, ErrorResponse{Error: body})
}
