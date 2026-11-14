package engine

import (
	"errors"
	"math/big"
	"testing"
)

// parseAndEvaluate runs the full pipeline on expr and fails the test on a parse error.
func parseAndEvaluate(t *testing.T, expr string) (string, error) {
	t.Helper()
	tree, err := parse(expr)
	if err != nil {
		t.Fatalf("parse(%q) unexpected error: %v", expr, err)
	}
	value, err := evaluate(tree)
	if err != nil {
		return "", err
	}
	return value.RatString(), nil
}

func TestEvaluate_ExactArithmetic(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{"addition", "1+2", "3"},
		{"case study example", "50 - 7 * 3", "29"},
		{"left-associative subtraction", "10-4-3", "3"},
		{"parentheses", "(1+2)*3", "9"},
		{"division keeps an exact fraction", "1/3", "1/3"},
		{"one third times three is exactly one", "1/3*3", "1"},
		{"0.1 + 0.2 is exactly 0.3", "0.1+0.2", "3/10"},
		{"0.7 + 0.1 is exactly 0.8", "0.7+0.1", "4/5"},
		{"negative result", "2-5", "-3"},
		{"zero divided by a number", "0/5", "0"},
		{"numbers beyond 64 bits stay exact", "99999999999999999999*99999999999999999999", "9999999999999999999800000000000000000001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAndEvaluate(t, tt.expr)
			if err != nil {
				t.Fatalf("evaluate(%q) unexpected error: %v", tt.expr, err)
			}
			if got != tt.want {
				t.Errorf("evaluate(%q) = %s, want %s", tt.expr, got, tt.want)
			}
		})
	}
}

func TestEvaluate_DivisionErrors(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantCode Code
	}{
		{"number divided by zero", "5/0", CodeDivisionByZero},
		{"divided by an expression that is zero", "1/(2-2)", CodeDivisionByZero},
		{"zero divided by zero", "0/0", CodeUndefinedResult},
		{"computed zero divided by zero", "(3-3)/(1-1)", CodeUndefinedResult},
		{"error in the right operand stops evaluation", "1 + 2 * (4/0)", CodeDivisionByZero},
		{"error in the left operand stops evaluation", "(4/0) + 1", CodeDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAndEvaluate(t, tt.expr)

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("evaluate(%q) error = %v, want *Error", tt.expr, err)
			}
			if engineErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", engineErr.Code, tt.wantCode)
			}
			if engineErr.Position != NoPosition {
				t.Errorf("Position = %d, want NoPosition", engineErr.Position)
			}
		})
	}
}

// unknownNode is a node type the evaluator does not support.
type unknownNode struct{}

func (unknownNode) String() string { return "?" }

func TestEvaluate_UnsupportedTreeReturnsErrorInsteadOfPanicking(t *testing.T) {
	one := &numberNode{value: big.NewRat(1, 1)}
	tests := []struct {
		name string
		tree node
	}{
		{"unknown node type", unknownNode{}},
		{"unknown operator", &binaryNode{op: tokenLeftParen, left: one, right: one}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := evaluate(tt.tree); err == nil {
				t.Errorf("evaluate(%s) expected an error, got nil", tt.tree)
			}
		})
	}
}
