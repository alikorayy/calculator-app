package engine

import (
	"errors"
	"strings"
	"testing"
)

func TestParse_NestingLimit(t *testing.T) {
	nested := func(depth int, open string) string {
		return strings.Repeat(open, depth) + "1" + strings.Repeat(")", depth)
	}

	t.Run("at the limit", func(t *testing.T) {
		for _, open := range []string{"(", "√("} {
			if _, err := parse(nested(maxNestingDepth, open)); err != nil {
				t.Errorf("%d levels of %q: unexpected error: %v", maxNestingDepth, open, err)
			}
		}
	})

	t.Run("one level too deep", func(t *testing.T) {
		tests := []struct {
			open    string
			wantPos int
		}{
			{"(", maxNestingDepth},
			{"√(", 2*maxNestingDepth + 1},
		}
		for _, tt := range tests {
			_, err := parse(nested(maxNestingDepth+1, tt.open))

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("%q nested too deeply: error = %v, want *Error", tt.open, err)
			}
			if engineErr.Code != CodeLimitExceeded {
				t.Errorf("%q: Code = %q, want %q", tt.open, engineErr.Code, CodeLimitExceeded)
			}
			if engineErr.Position != tt.wantPos {
				t.Errorf("%q: Position = %d, want %d", tt.open, engineErr.Position, tt.wantPos)
			}
		}
	})

	t.Run("sequential parentheses do not add up", func(t *testing.T) {
		sequential := strings.Repeat("(1)+", maxNestingDepth+1) + "1"
		if _, err := parse(sequential); err != nil {
			t.Errorf("sequential parentheses: unexpected error: %v", err)
		}
	})
}

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
		{"square key", "3²", "(3 ^ 2)"},
		{"reciprocal key", "4⁻¹", "(4 ^ -1)"},
		{"postfix keys apply to the whole operand before them", "3²⁻¹", "((3 ^ 2) ^ -1)"},
		{"reciprocal twice", "3⁻¹⁻¹", "((3 ^ -1) ^ -1)"},
		{"power of a squared base", "3²^3", "((3 ^ 2) ^ 3)"},
		{"square of an exponent", "2^3²", "(2 ^ (3 ^ 2))"},
		{"square binds tighter than a sign", "-2²", "(-(2 ^ 2))"},
		{"square of a percentage", "5%²", "((5%) ^ 2)"},
		{"square of parentheses", "(1+2)²", "((1 + 2) ^ 2)"},
		{"implicit multiplication after a square", "3²(2)", "((3 ^ 2) * 2)"},
		{"square root symbol", "√(9)", "sqrt(9)"},
		{"square root name in any case", "SQRT(9)", "sqrt(9)"},
		{"square root in a sum", "sqrt(4)+1", "(sqrt(4) + 1)"},
		{"negated square root", "-√(4)", "(-sqrt(4))"},
		{"square root as a power base", "√(4)^2", "(sqrt(4) ^ 2)"},
		{"nested square roots", "√(√(16))", "sqrt(sqrt(16))"},
		{"square root of an expression", "√(3^2+4^2)", "sqrt(((3 ^ 2) + (4 ^ 2)))"},
		{"implicit multiplication before a parenthesis", "2(3+1)", "(2 * (3 + 1))"},
		{"implicit multiplication between parentheses", "(1+2)(3+4)", "((1 + 2) * (3 + 4))"},
		{"implicit multiplication after a parenthesis", "(2)3", "(2 * 3)"},
		{"implicit multiplication before a square root", "2√(9)", "(2 * sqrt(9))"},
		{"implicit multiplication after a percent", "10%(5)", "((10%) * 5)"},
		{"implicit multiplication is left to right with division", "6/2(1+2)", "((6 / 2) * (1 + 2))"},
		{"implicit multiplication binds looser than power", "2(3)^2", "(2 * (3 ^ 2))"},
		{"implicit multiplication after a power", "2^3(4)", "((2 ^ 3) * 4)"},
		{"implicit multiplication after a sign", "-2(3)", "((-2) * 3)"},
		{"implicit multiplication inside a sum", "1+2(3)", "(1 + (2 * 3))"},
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
		{"square without operand", "²", `Unexpected "²"`, 0},
		{"reciprocal after an operator", "2+⁻¹", `Unexpected "⁻¹"`, 2},
		{"number after a square is not implicit multiplication", "3²2", `Unexpected "2"`, 2},
		{"square root without parentheses", "√9", `Expected "(" after "√"`, 1},
		{"square root name at the end", "2+sqrt", `Expected "(" after "sqrt"`, 6},
		{"unclosed square root", "√(9", "Missing closing parenthesis", 3},
		{"empty square root", "√()", `Unexpected ")"`, 2},
		{"number after a number is not implicit multiplication", "2 3", `Unexpected "3"`, 2},
		{"number after a percent is not implicit multiplication", "10%5", `Unexpected "5"`, 3},
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
