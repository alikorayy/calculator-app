package engine

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func tok(kind tokenKind, text string, pos int) token {
	return token{kind: kind, text: text, pos: pos}
}

func TestTokenize_Valid(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want []token
	}{
		{
			name: "simple addition",
			expr: "1+2",
			want: []token{
				tok(tokenNumber, "1", 0), tok(tokenPlus, "+", 1), tok(tokenNumber, "2", 2), tok(tokenEOF, "", 3),
			},
		},
		{
			name: "every operator and parentheses",
			expr: "(1-2)*3/4",
			want: []token{
				tok(tokenLeftParen, "(", 0), tok(tokenNumber, "1", 1), tok(tokenMinus, "-", 2),
				tok(tokenNumber, "2", 3), tok(tokenRightParen, ")", 4), tok(tokenMultiply, "*", 5),
				tok(tokenNumber, "3", 6), tok(tokenDivide, "/", 7), tok(tokenNumber, "4", 8),
				tok(tokenEOF, "", 9),
			},
		},
		{
			name: "whitespace is skipped but counted in positions",
			expr: " 12.5 *  3 ",
			want: []token{
				tok(tokenNumber, "12.5", 1), tok(tokenMultiply, "*", 6), tok(tokenNumber, "3", 9), tok(tokenEOF, "", 11),
			},
		},
		{
			name: "leading and trailing decimal separator",
			expr: ".5+5.",
			want: []token{
				tok(tokenNumber, ".5", 0), tok(tokenPlus, "+", 2), tok(tokenNumber, "5.", 3), tok(tokenEOF, "", 5),
			},
		},
		{
			name: "number text is kept exactly as typed",
			expr: "007.250",
			want: []token{tok(tokenNumber, "007.250", 0), tok(tokenEOF, "", 7)},
		},
		{
			name: "unicode keypad operators map to the same kinds as ASCII",
			expr: "8×2÷4−1",
			want: []token{
				tok(tokenNumber, "8", 0), tok(tokenMultiply, "×", 1), tok(tokenNumber, "2", 2),
				tok(tokenDivide, "÷", 3), tok(tokenNumber, "4", 4), tok(tokenMinus, "−", 5),
				tok(tokenNumber, "1", 6), tok(tokenEOF, "", 7),
			},
		},
		{
			name: "power and percent",
			expr: "2^3+50%",
			want: []token{
				tok(tokenNumber, "2", 0), tok(tokenPower, "^", 1), tok(tokenNumber, "3", 2),
				tok(tokenPlus, "+", 3), tok(tokenNumber, "50", 4), tok(tokenPercent, "%", 6),
				tok(tokenEOF, "", 7),
			},
		},
		{
			name: "square root symbol",
			expr: "√(9)",
			want: []token{
				tok(tokenSqrt, "√", 0), tok(tokenLeftParen, "(", 1), tok(tokenNumber, "9", 2),
				tok(tokenRightParen, ")", 3), tok(tokenEOF, "", 4),
			},
		},
		{
			name: "square root name in any case keeps the typed text",
			expr: "sqrt(4)+SQRT(9)",
			want: []token{
				tok(tokenSqrt, "sqrt", 0), tok(tokenLeftParen, "(", 4), tok(tokenNumber, "4", 5),
				tok(tokenRightParen, ")", 6), tok(tokenPlus, "+", 7), tok(tokenSqrt, "SQRT", 8),
				tok(tokenLeftParen, "(", 12), tok(tokenNumber, "9", 13), tok(tokenRightParen, ")", 14),
				tok(tokenEOF, "", 15),
			},
		},
		{
			name: "leading zeros do not count towards the digit limit",
			expr: "000" + strings.Repeat("9", maxNumberDigits),
			want: []token{
				tok(tokenNumber, "000"+strings.Repeat("9", maxNumberDigits), 0),
				tok(tokenEOF, "", maxNumberDigits+3),
			},
		},
		{
			name: "expression at the length limit",
			expr: strings.Repeat(" ", maxExpressionLength-1) + "1",
			want: []token{tok(tokenNumber, "1", maxExpressionLength-1), tok(tokenEOF, "", maxExpressionLength)},
		},
		{
			name: "empty expression yields only EOF",
			expr: "",
			want: []token{tok(tokenEOF, "", 0)},
		},
		{
			name: "whitespace-only expression yields only EOF",
			expr: "   ",
			want: []token{tok(tokenEOF, "", 3)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenize(tt.expr)
			if err != nil {
				t.Fatalf("tokenize(%q) unexpected error: %v", tt.expr, err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("tokenize(%q)\n got: %v\nwant: %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestTokenize_Invalid(t *testing.T) {
	tests := []struct {
		name        string
		expr        string
		wantCode    Code
		wantMessage string
		wantPos     int
	}{
		{"lone decimal separator", ".", CodeSyntaxError, `Invalid number "."`, 0},
		{"two decimal separators", "2+1.2.3", CodeSyntaxError, `Invalid number "1.2.3"`, 2},
		{"unknown character", "2 & 3", CodeSyntaxError, "Unexpected character '&'", 2},
		{"unknown name", "2+abc", CodeSyntaxError, `Unknown name "abc"`, 2},
		{"name glued to sqrt", "sqrtx(4)", CodeSyntaxError, `Unknown name "sqrtx"`, 0},
		{"non-ASCII digit", "1+٣", CodeSyntaxError, "Unexpected character '٣'", 2},
		// × is 2 bytes in UTF-8; positions must count characters, not bytes.
		{"position counts characters not bytes", "2×3 &", CodeSyntaxError, "Unexpected character '&'", 4},
		{
			"number with too many digits", "1+" + strings.Repeat("9", maxNumberDigits+1),
			CodeLimitExceeded, "Number has too many digits (maximum 32)", 2,
		},
		{
			"decimal digits count too", strings.Repeat("1", maxNumberDigits/2) + "." + strings.Repeat("1", maxNumberDigits/2+1),
			CodeLimitExceeded, "Number has too many digits (maximum 32)", 0,
		},
		{
			"expression too long", strings.Repeat("1", maxExpressionLength+1),
			CodeLimitExceeded, "Expression is too long (maximum 256 characters)", NoPosition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tokenize(tt.expr)

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("tokenize(%q) error = %v, want *Error", tt.expr, err)
			}
			if engineErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", engineErr.Code, tt.wantCode)
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
