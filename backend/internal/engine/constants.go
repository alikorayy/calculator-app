package engine

// User-facing error messages. Format verbs are filled in with the offending input.
const (
	msgUnexpectedCharacter = "Unexpected character %q"
	msgInvalidNumber       = "Invalid number %q"
)

// Characters with a fixed meaning in a number literal.
const (
	decimalSeparator = '.'
)
