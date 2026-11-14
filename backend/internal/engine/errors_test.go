package engine

import (
	"errors"
	"fmt"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *Error
		want string
	}{
		{
			name: "without position",
			err:  &Error{Code: CodeDivisionByZero, Message: "Cannot divide by zero", Position: NoPosition},
			want: "DIVISION_BY_ZERO: Cannot divide by zero",
		},
		{
			name: "with position",
			err:  &Error{Code: CodeSyntaxError, Message: "Unexpected ')'", Position: 5},
			want: "SYNTAX_ERROR: Unexpected ')' at position 5",
		},
		{
			name: "position zero is a real position",
			err:  &Error{Code: CodeSyntaxError, Message: "Unexpected '*'", Position: 0},
			want: "SYNTAX_ERROR: Unexpected '*' at position 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestError_FoundThroughWrapping(t *testing.T) {
	original := &Error{Code: CodeOverflow, Message: "Result is too large", Position: NoPosition}
	wrapped := fmt.Errorf("evaluate: %w", original)

	var engineErr *Error
	if !errors.As(wrapped, &engineErr) {
		t.Fatalf("errors.As did not find *Error in %v", wrapped)
	}
	if engineErr.Code != CodeOverflow {
		t.Errorf("Code = %q, want %q", engineErr.Code, CodeOverflow)
	}
}
