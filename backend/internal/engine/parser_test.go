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
		{"power", "2^3", "(2 ^ 3)"},
		{"power is right-associative", "2^3^2", "(2 ^ (3 ^ 2))"},
		{"power binds tighter than a sign", "-2^2", "(-(2 ^ 2))"},
		{"parenthesized negative base", "(-2)^2", "((-2) ^ 2)"},
		{"negative exponent", "2^-1", "(2 ^ (-1))"},
		{"power binds tighter than multiplication", "2*3^2", "(2 * (3 ^ 2))"},
		{"power before multiplication in source", "2^3*2", "((2 ^ 3) * 2)"},
		{"percent", "10%", "(10%)"},
		{"percent after an addition", "50+10%", "(50 + (10%))"},
		{"percent binds tighter than power", "10%^2", "((10%) ^ 2)"},
		{"repeated percent", "50%%", "((50%)%)"},
		{"percent of parentheses", "(10+20)%", "((10 + 20)%)"},
		{"square root symbol", "√(9)", "sqrt(9)"},
		{"square root name in any case", "SQRT(9)", "sqrt(9)"},
		{"square root in a sum", "sqrt(4)+1", "(sqrt(4) + 1)"},
		{"negated square root", "-√(4)", "(-sqrt(4))"},
		{"square root as a power base", "√(4)^2", "(sqrt(4) ^ 2)"},
		{"nested square roots", "√(√(16))", "sqrt(sqrt(16))"},
		{"square root of an expression", "√(3^2+4^2)", "sqrt(((3 ^ 2) + (4 ^ 2)))"},
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
		{"power without exponent", "2^", "Unexpected end of expression", 2},
		{"power without base", "^2", `Unexpected "^"`, 0},
		{"double power operator", "2^^3", `Unexpected "^"`, 2},
		{"percent without operand", "%", `Unexpected "%"`, 0},
		{"percent after an operator", "50+%", `Unexpected "%"`, 3},
		{"square root without parentheses", "√9", `Expected "(" after "√"`, 1},
		{"square root name at the end", "2+sqrt", `Expected "(" after "sqrt"`, 6},
		{"unclosed square root", "√(9", "Missing closing parenthesis", 3},
		{"empty square root", "√()", `Unexpected ")"`, 2},
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
