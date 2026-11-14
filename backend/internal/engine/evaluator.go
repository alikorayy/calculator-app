package engine

import (
	"fmt"
	"math/big"
)

// evaluate computes the exact value of the tree rooted at n. It walks the tree
// depth-first: both operands of a node are evaluated before the node itself,
// so the deepest operations run first.
//
// Every operation allocates a new *big.Rat, so values stored in the tree are
// never modified.
func evaluate(n node) (*big.Rat, error) {
	switch n := n.(type) {
	case *numberNode:
		return n.value, nil
	case *binaryNode:
		return evaluateBinary(n)
	default:
		// Only reachable if a new node type is added without teaching the evaluator about it.
		return nil, fmt.Errorf("engine: unsupported node type %T", n)
	}
}

// evaluateBinary evaluates both operands, then applies the operator.
func evaluateBinary(n *binaryNode) (*big.Rat, error) {
	left, err := evaluate(n.left)
	if err != nil {
		return nil, err
	}
	right, err := evaluate(n.right)
	if err != nil {
		return nil, err
	}

	switch n.op {
	case tokenPlus:
		return new(big.Rat).Add(left, right), nil
	case tokenMinus:
		return new(big.Rat).Sub(left, right), nil
	case tokenMultiply:
		return new(big.Rat).Mul(left, right), nil
	case tokenDivide:
		return divide(left, right)
	default:
		return nil, fmt.Errorf("engine: unsupported binary operator %d", n.op)
	}
}

// divide returns left / right. big.Rat panics on a zero divisor, so zero is
// rejected first: 0 / 0 has no defined value, any other x / 0 is a division by zero.
func divide(left, right *big.Rat) (*big.Rat, error) {
	if right.Sign() == 0 {
		if left.Sign() == 0 {
			return nil, newError(CodeUndefinedResult, msgUndefinedResult)
		}
		return nil, newError(CodeDivisionByZero, msgDivisionByZero)
	}
	return new(big.Rat).Quo(left, right), nil
}
