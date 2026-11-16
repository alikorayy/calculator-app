package engine

import (
	"fmt"
	"math/big"
)

// evaluate computes the exact value of the tree rooted at n. It walks the tree
// depth-first: both operands of a node are evaluated before the node itself,
// so the deepest operations run first.
//
// Every intermediate result must stay within the supported range (see
// checkRange), so a chain of cheap operations such as 10^10000 * 10^10000
// cannot grow past the limit either.
//
// Every operation allocates a new *big.Rat, so values stored in the tree are
// never modified.
func evaluate(n node) (*big.Rat, error) {
	value, err := evaluateNode(n)
	if err != nil {
		return nil, err
	}
	if err := checkRange(value); err != nil {
		return nil, err
	}
	return value, nil
}

// evaluateNode computes the value of a single node, evaluating its children first.
func evaluateNode(n node) (*big.Rat, error) {
	switch n := n.(type) {
	case *numberNode:
		return n.value, nil
	case *negateNode:
		operand, err := evaluate(n.operand)
		if err != nil {
			return nil, err
		}
		return new(big.Rat).Neg(operand), nil
	case *percentNode:
		operand, err := evaluate(n.operand)
		if err != nil {
			return nil, err
		}
		return new(big.Rat).Quo(operand, big.NewRat(percentBase, 1)), nil
	case *percentChangeNode:
		return evaluatePercentChange(n)
	case *sqrtNode:
		operand, err := evaluate(n.operand)
		if err != nil {
			return nil, err
		}
		return squareRoot(operand)
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
	case tokenPower:
		return power(left, right)
	default:
		return nil, fmt.Errorf("engine: unsupported binary operator %d", n.op)
	}
}

// evaluatePercentChange computes base ± base·rate, e.g. 50 + 10% = 50 + 5 = 55.
func evaluatePercentChange(n *percentChangeNode) (*big.Rat, error) {
	base, err := evaluate(n.base)
	if err != nil {
		return nil, err
	}
	rate, err := evaluate(n.rate)
	if err != nil {
		return nil, err
	}
	change := new(big.Rat).Mul(base, rate)
	if n.op == tokenMinus {
		return new(big.Rat).Sub(base, change), nil
	}
	return new(big.Rat).Add(base, change), nil
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
