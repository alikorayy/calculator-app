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
	default:
		return "?"
	}
}
