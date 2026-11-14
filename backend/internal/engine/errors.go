package engine

import "fmt"

// Code identifies the kind of failure. Codes are part of the public API contract
// and never change once released.
type Code string

// Error codes returned by the engine.
const (
	// CodeSyntaxError means the expression is not well formed, e.g. "2 + * 3".
	CodeSyntaxError Code = "SYNTAX_ERROR"
	// CodeDivisionByZero means a non-zero value was divided by zero, e.g. "5 / 0".
	CodeDivisionByZero Code = "DIVISION_BY_ZERO"
	// CodeUndefinedResult means the result is mathematically undefined, e.g. "0 / 0".
	CodeUndefinedResult Code = "UNDEFINED_RESULT"
	// CodeDomainError means an operation has no real result, e.g. "√(-4)".
	CodeDomainError Code = "DOMAIN_ERROR"
	// CodeOverflow means the result is too large to represent, e.g. "10^99999".
	CodeOverflow Code = "OVERFLOW"
	// CodeLimitExceeded means the input is too long, too deeply nested or has too many digits.
	CodeLimitExceeded Code = "LIMIT_EXCEEDED"
)

// NoPosition is the Position of an Error that does not point at a specific
// place in the expression.
const NoPosition = -1

// Error is the single error type returned by the engine.
type Error struct {
	// Code is the stable, machine-readable kind of failure.
	Code Code
	// Message is a short, human-readable description safe to show to end users.
	Message string
	// Position is the 0-based character index in the expression where the
	// problem was found, or NoPosition.
	Position int
}

// Error implements the error interface. The text is meant for logs; clients
// should use Code, Message and Position instead.
func (e *Error) Error() string {
	if e.Position == NoPosition {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s at position %d", e.Code, e.Message, e.Position)
}

// newSyntaxError returns a SYNTAX_ERROR pointing at the given character position.
func newSyntaxError(pos int, message string) *Error {
	return &Error{Code: CodeSyntaxError, Message: message, Position: pos}
}
