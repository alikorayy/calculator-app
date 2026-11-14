package engine

// Input limits. They keep every request cheap to evaluate and are reported as
// LIMIT_EXCEEDED when crossed.
const (
	// maxExpressionLength is the maximum number of characters in an expression.
	maxExpressionLength = 256
	// maxNumberDigits is the maximum number of significant digits in one number literal.
	maxNumberDigits = 32
	// maxResultExponent bounds the magnitude of a power: its result must lie
	// between 10^-maxResultExponent and 10^maxResultExponent.
	maxResultExponent = 10000
	// maxExactPowerBits bounds the size, in bits, of the numerator or denominator
	// of an exactly computed power (about 30,000 decimal digits).
	maxExactPowerBits = 100000
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
	msgOutOfRange          = "Result is out of range"
	msgExponentTooLarge    = "Exponent is too large"
	msgFractionalExponent  = "Fractional exponents are not supported yet"
)

// Characters and names with a fixed meaning in an expression.
const (
	decimalSeparator = '.'
	functionSqrt     = "sqrt"
)
