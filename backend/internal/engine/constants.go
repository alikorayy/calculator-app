package engine

// Input limits. They keep every request cheap to evaluate and are reported as
// LIMIT_EXCEEDED when crossed.
const (
	// maxExpressionLength is the maximum number of characters in an expression.
	maxExpressionLength = 256
	// maxNumberDigits is the maximum number of significant digits in one number literal.
	maxNumberDigits = 32
)

// User-facing error messages. Format verbs are filled in with the offending input.
const (
	msgUnexpectedCharacter = "Unexpected character %q"
	msgInvalidNumber       = "Invalid number %q"
	msgUnknownName         = "Unknown name %q"
	msgExpressionTooLong   = "Expression is too long (maximum %d characters)"
	msgTooManyDigits       = "Number has too many digits (maximum %d)"
	msgEmptyExpression     = "Expression is empty"
	msgUnexpectedEnd       = "Unexpected end of expression"
	msgUnexpectedToken     = "Unexpected %q"
	msgMissingClosingParen = "Missing closing parenthesis"
	msgDivisionByZero      = "Cannot divide by zero"
	msgUndefinedResult     = "Result is undefined"
)

// Characters and names with a fixed meaning in an expression.
const (
	decimalSeparator = '.'
	functionSqrt     = "sqrt"
)
