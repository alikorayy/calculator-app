package engine

import (
	"fmt"
	"strings"
	"unicode"
)

// tokenKind classifies a token.
type tokenKind int

const (
	tokenNumber tokenKind = iota
	tokenPlus
	tokenMinus
	tokenMultiply
	tokenDivide
	tokenLeftParen
	tokenRightParen
	tokenEOF
)

// token is one meaningful unit of an expression.
type token struct {
	kind tokenKind
	// text is the input exactly as the user typed it, e.g. "12.5" or "+".
	text string
	// pos is the 0-based character (rune) index of the token's first character.
	pos int
}

// tokenize splits expr into tokens, skipping whitespace. The returned slice
// always ends with a tokenEOF token positioned just past the last character.
// Characters that cannot start a token and malformed numbers are reported as
// SYNTAX_ERROR with their position.
func tokenize(expr string) ([]token, error) {
	runes := []rune(expr)
	var tokens []token

	for pos := 0; pos < len(runes); {
		r := runes[pos]
		switch {
		case unicode.IsSpace(r):
			pos++
		case isDigit(r) || r == decimalSeparator:
			number, err := scanNumber(runes, pos)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, number)
			pos += len([]rune(number.text))
		default:
			kind, ok := operatorKind(r)
			if !ok {
				return nil, newSyntaxError(pos, fmt.Sprintf(msgUnexpectedCharacter, r))
			}
			tokens = append(tokens, token{kind: kind, text: string(r), pos: pos})
			pos++
		}
	}

	return append(tokens, token{kind: tokenEOF, pos: len(runes)}), nil
}

// scanNumber reads the number literal starting at start. A literal is a run of
// digits containing at most one decimal separator and at least one digit, so
// "12", "12.5", ".5" and "5." are valid while "." and "1.2.3" are not.
func scanNumber(runes []rune, start int) (token, error) {
	end := start
	for end < len(runes) && (isDigit(runes[end]) || runes[end] == decimalSeparator) {
		end++
	}

	text := string(runes[start:end])
	separators := strings.Count(text, string(decimalSeparator))
	if separators > 1 || separators == len(text) {
		return token{}, newSyntaxError(start, fmt.Sprintf(msgInvalidNumber, text))
	}
	return token{kind: tokenNumber, text: text, pos: start}, nil
}

// operatorKind maps a single-character operator or parenthesis to its token kind.
func operatorKind(r rune) (tokenKind, bool) {
	switch r {
	case '+':
		return tokenPlus, true
	case '-':
		return tokenMinus, true
	case '*':
		return tokenMultiply, true
	case '/':
		return tokenDivide, true
	case '(':
		return tokenLeftParen, true
	case ')':
		return tokenRightParen, true
	default:
		return 0, false
	}
}

// isDigit reports whether r is an ASCII digit. unicode.IsDigit is deliberately
// not used: it also accepts digits from other scripts (e.g. "٣"), which the
// arithmetic layer cannot parse.
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
