package engine

// User-facing error messages. Format verbs are filled in with the offending input.
const (
	msgUnexpectedCharacter = "Unexpected character %q"
	msgInvalidNumber       = "Invalid number %q"
	msgEmptyExpression     = "Expression is empty"
	msgUnexpectedEnd       = "Unexpected end of expression"
	msgUnexpectedToken     = "Unexpected %q"
	msgMissingClosingParen = "Missing closing parenthesis"
)

// Characters with a fixed meaning in a number literal.
const (
	decimalSeparator = '.'
)
