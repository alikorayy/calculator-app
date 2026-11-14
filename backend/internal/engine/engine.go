package engine

import (
	"fmt"
	"math/big"
	"strings"
)

// Result is the outcome of a successful calculation.
type Result struct {
	// Value is the result as a decimal string, rounded to 16 significant digits
	// and in scientific notation for very large or small magnitudes,
	// e.g. "29", "0.3333333333333333" or "1e+64".
	Value string
	// Approximate reports whether Value was rounded, i.e. is not the exact result.
	Approximate bool
}

// Operation is a single calculator operation accepted by Calculate.
type Operation string

// Operations supported by Calculate.
const (
	// OperationAdd returns a + b.
	OperationAdd Operation = "add"
	// OperationSubtract returns a - b.
	OperationSubtract Operation = "subtract"
	// OperationMultiply returns a × b.
	OperationMultiply Operation = "multiply"
	// OperationDivide returns a ÷ b.
	OperationDivide Operation = "divide"
	// OperationPower returns a ^ b.
	OperationPower Operation = "power"
	// OperationSqrt returns √a.
	OperationSqrt Operation = "sqrt"
	// OperationPercent returns a / 100.
	OperationPercent Operation = "percent"
)

// Evaluate computes a calculator expression such as "50 - 7 × 3" or
// "√(2) + 10%". Arithmetic is exact and the result is rounded only once, when
// it is formatted.
//
// Every problem with the expression is reported as an *Error with a stable
// Code; syntax errors also carry the Position of the offending character.
func Evaluate(expression string) (Result, error) {
	tree, err := parse(expression)
	if err != nil {
		return Result{}, err
	}
	return compute(tree)
}

// Calculate applies a single operation to operands given as decimal strings,
// such as "-12.5". Binary operations take two operands, sqrt and percent take
// one. It follows exactly the same rules as Evaluate.
//
// An unknown operation, a wrong number of operands or an operand that is not a
// decimal number is reported as an *Error with CodeInvalidArgument; every other
// problem uses the same codes as Evaluate.
func Calculate(op Operation, operands []string) (Result, error) {
	arity, known := operationArity(op)
	if !known {
		return Result{}, newError(CodeInvalidArgument, fmt.Sprintf(msgUnknownOperation, op))
	}
	if len(operands) != arity {
		return Result{}, newError(CodeInvalidArgument, fmt.Sprintf(msgWrongOperandCount, op, arity, len(operands)))
	}

	args := make([]node, len(operands))
	for i, text := range operands {
		value, err := parseOperand(text, i)
		if err != nil {
			return Result{}, err
		}
		args[i] = &numberNode{value: value}
	}
	return compute(operationTree(op, args))
}

// compute evaluates a tree and formats its value.
func compute(tree node) (Result, error) {
	value, err := evaluate(tree)
	if err != nil {
		return Result{}, err
	}
	text, approximate := format(value)
	return Result{Value: text, Approximate: approximate}, nil
}

// operationArity returns how many operands op takes, and false for an unknown op.
func operationArity(op Operation) (int, bool) {
	switch op {
	case OperationAdd, OperationSubtract, OperationMultiply, OperationDivide, OperationPower:
		return binaryArity, true
	case OperationSqrt, OperationPercent:
		return unaryArity, true
	default:
		return 0, false
	}
}

// operationTree builds the same tree the parser would build for op, so
// Calculate and Evaluate share one evaluator. op must be known and args must
// have the right length.
func operationTree(op Operation, args []node) node {
	switch op {
	case OperationAdd:
		return &binaryNode{op: tokenPlus, left: args[0], right: args[1]}
	case OperationSubtract:
		return &binaryNode{op: tokenMinus, left: args[0], right: args[1]}
	case OperationMultiply:
		return &binaryNode{op: tokenMultiply, left: args[0], right: args[1]}
	case OperationDivide:
		return &binaryNode{op: tokenDivide, left: args[0], right: args[1]}
	case OperationPower:
		return &binaryNode{op: tokenPower, left: args[0], right: args[1]}
	case OperationSqrt:
		return &sqrtNode{operand: args[0]}
	default:
		return &percentNode{operand: args[0]}
	}
}

// parseOperand converts one Calculate operand into an exact value. An operand
// is a number literal as in an expression, optionally preceded by '-':
// "12", "-0.5", ".5". Exponent notation such as "1e5" is not accepted.
func parseOperand(text string, index int) (*big.Rat, error) {
	literal := strings.TrimPrefix(text, "-")
	if !isNumberLiteral(literal) {
		return nil, newError(CodeInvalidArgument, fmt.Sprintf(msgInvalidOperand, index+1, text))
	}
	if significantDigits(literal) > maxNumberDigits {
		return nil, newError(CodeLimitExceeded, fmt.Sprintf(msgTooManyDigits, maxNumberDigits))
	}
	value, _ := new(big.Rat).SetString(text)
	return value, nil
}
