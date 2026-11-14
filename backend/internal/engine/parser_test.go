package engine

import (
	"errors"
	"testing"
)

func TestParse_BuildsTreeWithPrecedence(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{"single number", "42", "42"},
		{"decimal becomes exact fraction", "0.5", "1/2"},
		{"trailing zeros do not matter", "2.50", "5/2"},
		{"addition", "1+2", "(1 + 2)"},
		{"multiplication before addition", "1+2*3", "(1 + (2 * 3))"},
		{"multiplication first in source", "1*2+3", "((1 * 2) + 3)"},
		{"case study example", "50 - 7 * 3", "(50 - (7 * 3))"},
		{"subtraction is left-associative", "10-4-3", "((10 - 4) - 3)"},
		{"division is left-associative", "8/4/2", "((8 / 4) / 2)"},
		{"parentheses override precedence", "(1+2)*3", "((1 + 2) * 3)"},
		{"redundant parentheses disappear", "((2))", "2"},
		{"nested parentheses", "2*(3+(4-1))", "(2 * (3 + (4 - 1)))"},
		{"negative number", "-5", "(-5)"},
		{"negative operand after an operator", "2*-3", "(2 * (-3))"},
		{"unary minus binds tighter than multiplication", "-2*3", "((-2) * 3)"},
		{"subtracting a negative number", "10 - -3", "(10 - (-3))"},
		{"keypad minus sign", "1−−2", "(1 - (-2))"},
		{"negated parentheses", "-(1+2)", "(-(1 + 2))"},
		{"double negation", "--5", "(-(-5))"},
		{"unary plus is dropped", "+5", "5"},
		{"mixed signs", "-+-5", "(-(-5))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree, err := parse(tt.expr)
			if err != nil {
				t.Fatalf("parse(%q) unexpected error: %v", tt.expr, err)
			}
			if got := tree.String(); got != tt.want {
				t.Errorf("parse(%q) = %s, want %s", tt.expr, got, tt.want)
			}
		})
	}
}

func TestParse_SyntaxErrors(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		wantMessage string
		wantPos     int
	}{
		{"empty", "", "Expression is empty", 0},
		{"whitespace only", "   ", "Expression is empty", 0},
		{"operator at the end", "1+", "Unexpected end of expression", 2},
		{"operator at the start", "*1", `Unexpected "*"`, 0},
		{"two operators in a row", "2*/3", `Unexpected "/"`, 2},
		{"two numbers in a row", "1 2", `Unexpected "2"`, 2},
		{"unclosed parenthesis", "(1+2", "Missing closing parenthesis", 4},
		{"extra closing parenthesis", "1+2)", `Unexpected ")"`, 3},
		{"empty parentheses", "()", `Unexpected ")"`, 1},
		{"sign without operand", "-", "Unexpected end of expression", 1},
		{"sign before an operator", "2*-/3", `Unexpected "/"`, 3},
		{"tokenizer error is passed through", "2&3", "Unexpected character '&'", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parse(tt.expr)

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("parse(%q) error = %v, want *Error", tt.expr, err)
			}
			if engineErr.Code != CodeSyntaxError {
				t.Errorf("Code = %q, want %q", engineErr.Code, CodeSyntaxError)
			}
			if engineErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", engineErr.Message, tt.wantMessage)
			}
			if engineErr.Position != tt.wantPos {
				t.Errorf("Position = %d, want %d", engineErr.Position, tt.wantPos)
			}
		})
	}
}
