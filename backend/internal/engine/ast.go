package engine

import (
	"fmt"
	"math/big"
)

// node is one element of the abstract syntax tree (AST) built by the parser.
// String renders the node fully parenthesized, e.g. "(1 + (2 * 3))", which
// makes the tree's structure visible in tests and logs.
type node interface {
	String() string
}

// numberNode is a number literal, already converted to an exact rational.
type numberNode struct {
	value *big.Rat
}

func (n *numberNode) String() string {
	return n.value.RatString()
}

// negateNode flips the sign of its operand, e.g. -5 or -(1 + 2).
type negateNode struct {
	operand node
}

func (n *negateNode) String() string {
	return fmt.Sprintf("(-%s)", n.operand)
}

// percentNode is an operand followed by '%', e.g. 10%.
type percentNode struct {
	operand node
}

func (n *percentNode) String() string {
	return fmt.Sprintf("(%s%%)", n.operand)
}

// percentChangeNode is "base + rate%" or "base - rate%": base increased or
// decreased by a percentage of itself, e.g. 50 + 10% = 55. It acts as a single
// operand, so operators after it apply to the changed value: 50 + 10% × 2 = 110.
type percentChangeNode struct {
	op   tokenKind
	base node
	rate *percentNode
}

func (n *percentChangeNode) String() string {
	return fmt.Sprintf("(%s %s %s of it)", n.base, operatorSymbol(n.op), n.rate)
}

// sqrtNode is the square root of its operand, e.g. √(9).
type sqrtNode struct {
	operand node
}

func (n *sqrtNode) String() string {
	return fmt.Sprintf("sqrt(%s)", n.operand)
}

// binaryNode applies an operator to two operands, e.g. left + right.
type binaryNode struct {
	op    tokenKind
	left  node
	right node
}

func (n *binaryNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.left, operatorSymbol(n.op), n.right)
}

// operatorSymbol returns the canonical ASCII symbol of a binary operator.
func operatorSymbol(kind tokenKind) string {
	switch kind {
	case tokenPlus:
		return "+"
	case tokenMinus:
		return "-"
	case tokenMultiply:
		return "*"
	case tokenDivide:
		return "/"
	case tokenPower:
		return "^"
	default:
		return "?"
	}
}
