package engine

import (
	"errors"
	"slices"
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
		wantMessage string
		wantPos     int
	}{
		{"lone decimal separator", ".", `Invalid number "."`, 0},
		{"two decimal separators", "2+1.2.3", `Invalid number "1.2.3"`, 2},
		{"unknown character", "2 & 3", "Unexpected character '&'", 2},
		{"letters", "abc", "Unexpected character 'a'", 0},
		{"non-ASCII digit", "1+٣", "Unexpected character '٣'", 2},
		// The no-break space is 2 bytes in UTF-8; positions must count characters, not bytes.
		{"position counts characters not bytes", " 1 &", "Unexpected character '&'", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tokenize(tt.expr)

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("tokenize(%q) error = %v, want *Error", tt.expr, err)
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
