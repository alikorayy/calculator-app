package engine

// Input limits. They keep every request cheap to evaluate and are reported as
// LIMIT_EXCEEDED when crossed.
const (
	// maxExpressionLength is the maximum number of characters in an expression.
	maxExpressionLength = 256
	// maxNumberDigits is the maximum number of significant digits in one number literal.
	maxNumberDigits = 32
	// maxNestingDepth is the maximum number of nested parentheses.
	maxNestingDepth = 32
	// maxResultExponent bounds the magnitude of every result, including
	// intermediate ones: a non-zero value must lie between 10^-maxResultExponent
	// and 10^maxResultExponent.
	maxResultExponent = 10000
	// rangeCheckMargin is how close (in powers of ten) an estimated magnitude
	// must be to maxResultExponent before the range is checked exactly.
	rangeCheckMargin = 1
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
	msgNestedTooDeeply     = "Parentheses are nested too deeply (maximum %d levels)"
	msgEmptyExpression     = "Expression is empty"
	msgUnexpectedEnd       = "Unexpected end of expression"
	msgUnexpectedToken     = "Unexpected %q"
	msgMissingClosingParen = "Missing closing parenthesis"
	msgDivisionByZero      = "Cannot divide by zero"
	msgUndefinedResult     = "Result is undefined"
	msgOutOfRange          = "Result is out of range"
	msgExponentTooLarge    = "Exponent is too large"
	msgInvalidInput        = "Invalid input"
	msgExpectedParenAfter  = "Expected \"(\" after %q"
)

// Precision settings for results that cannot be represented exactly.
const (
	// irrationalPrecision is the big.Float precision, in bits, used for
	// irrational square roots and powers. 256 bits is about 77 significant
	// decimal digits, far more than the 16 digits that are displayed.
	irrationalPrecision = 256
	// maxExactRootDegree is the largest root degree q for which base^(p/q) is
	// checked for an exact result. It covers the exponents people actually type,
	// such as 0.5 (q = 2), 1/3 (q = 3) or 0.25 (q = 4).
	maxExactRootDegree = 64
	// rootGuardBits is the extra big.Float precision used when estimating an
	// integer root, so the estimate is off by at most one.
	rootGuardBits = 64
)

// percentBase is the divisor that turns a percentage into a fraction: 10% = 10/100.
const percentBase = 100

// Characters and names with a fixed meaning in an expression.
const (
	decimalSeparator = '.'
	functionSqrt     = "sqrt"
)
