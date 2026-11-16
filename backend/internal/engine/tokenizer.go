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
	tokenPower
	tokenPercent
	tokenSquare
	tokenReciprocal
	tokenSqrt
	tokenLeftParen
	tokenRightParen
	tokenEOF
)

// token is one meaningful unit of an expression.
type token struct {
	kind tokenKind
	// text is the input exactly as the user typed it, e.g. "12.5", "×" or "sqrt".
	text string
	// pos is the 0-based character (rune) index of the token's first character.
	pos int
}

// tokenize splits expr into tokens, skipping whitespace. The returned slice
// always ends with a tokenEOF token positioned just past the last character.
//
// Unknown characters, unknown names and malformed numbers are reported as
// SYNTAX_ERROR; an expression or number that is too long is reported as
// LIMIT_EXCEEDED.
func tokenize(expr string) ([]token, error) {
	runes := []rune(expr)
	if len(runes) > maxExpressionLength {
		return nil, newError(CodeLimitExceeded, fmt.Sprintf(msgExpressionTooLong, maxExpressionLength))
	}

	var tokens []token
	for pos := 0; pos < len(runes); {
		r := runes[pos]
		switch {
		case unicode.IsSpace(r):
			pos++
			continue
		case isDigit(r) || r == decimalSeparator:
			number, err := scanNumber(runes, pos)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, number)
		case unicode.IsLetter(r):
			name, err := scanName(runes, pos)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, name)
		default:
			symbol, err := scanSymbol(runes, pos)
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, symbol)
		}
		pos += len([]rune(tokens[len(tokens)-1].text))
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
	if !isNumberLiteral(text) {
		return token{}, newSyntaxError(start, fmt.Sprintf(msgInvalidNumber, text))
	}
	if significantDigits(text) > maxNumberDigits {
		return token{}, newErrorAt(CodeLimitExceeded, start, fmt.Sprintf(msgTooManyDigits, maxNumberDigits))
	}
	return token{kind: tokenNumber, text: text, pos: start}, nil
}

// isNumberLiteral reports whether text consists only of digits and at most one
// decimal separator, with at least one digit: "12", "12.5", ".5" and "5." are
// number literals; "", ".", "1.2.3" and "-1" are not.
func isNumberLiteral(text string) bool {
	separators := 0
	for _, r := range text {
		switch {
		case r == decimalSeparator:
			separators++
		case !isDigit(r):
			return false
		}
	}
	return separators <= 1 && len(text) > separators
}

// significantDigits counts the digits of a literal, ignoring the decimal
// separator and leading zeros: "007.250" has 4.
func significantDigits(literal string) int {
	digits := strings.Replace(literal, string(decimalSeparator), "", 1)
	return len(strings.TrimLeft(digits, "0"))
}

// scanName reads the run of letters starting at start. The only known name is
// the square root function, matched case-insensitively ("sqrt", "SQRT").
func scanName(runes []rune, start int) (token, error) {
	end := start
	for end < len(runes) && unicode.IsLetter(runes[end]) {
		end++
	}

	text := string(runes[start:end])
	if !strings.EqualFold(text, functionSqrt) {
		return token{}, newSyntaxError(start, fmt.Sprintf(msgUnknownName, text))
	}
	return token{kind: tokenSqrt, text: text, pos: start}, nil
}

// scanSymbol reads the operator, symbol or parenthesis starting at start.
// Every symbol is one character, except the two-character reciprocal "⁻¹".
func scanSymbol(runes []rune, start int) (token, error) {
	if strings.HasPrefix(string(runes[start:]), reciprocalSymbol) {
		return token{kind: tokenReciprocal, text: reciprocalSymbol, pos: start}, nil
	}
	r := runes[start]
	kind, ok := symbolKind(r)
	if !ok {
		return token{}, newSyntaxError(start, fmt.Sprintf(msgUnexpectedCharacter, r))
	}
	return token{kind: kind, text: string(r), pos: start}, nil
}

// symbolKind maps a single-character operator, symbol or parenthesis to its
// token kind. Unicode operators as shown on calculator keypads (× ÷ −) are
// accepted alongside their ASCII equivalents.
func symbolKind(r rune) (tokenKind, bool) {
	switch r {
	case squareSymbol:
		return tokenSquare, true
	case '+':
		return tokenPlus, true
	case '-', '−':
		return tokenMinus, true
	case '*', '×':
		return tokenMultiply, true
	case '/', '÷':
		return tokenDivide, true
	case '^':
		return tokenPower, true
	case '%':
		return tokenPercent, true
	case '√':
		return tokenSqrt, true
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
