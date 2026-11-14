package engine

import "math/big"

// squareRoot returns √x.
//
// When x is a perfect square of a fraction (9, 0.25, 9/4) the root is exact.
// Otherwise the root is irrational and has no exact fraction, so it is computed
// with big.Float at sqrtPrecision bits (about 77 significant digits) and turned
// back into a *big.Rat. That fraction is extremely close to the true root, and
// all later arithmetic on it is exact again; the error stays far below the 16
// digits that are displayed.
//
// A negative x has no real square root and is a DOMAIN_ERROR.
func squareRoot(x *big.Rat) (*big.Rat, error) {
	if x.Sign() < 0 {
		return nil, newError(CodeDomainError, msgInvalidInput)
	}
	if root, ok := exactSquareRoot(x); ok {
		return root, nil
	}
	return approximateSquareRoot(x), nil
}

// exactSquareRoot returns √x when both the numerator and the denominator of x
// are perfect squares, e.g. √(9/4) = 3/2.
func exactSquareRoot(x *big.Rat) (*big.Rat, bool) {
	num, ok := integerSquareRoot(x.Num())
	if !ok {
		return nil, false
	}
	den, ok := integerSquareRoot(x.Denom())
	if !ok {
		return nil, false
	}
	return new(big.Rat).SetFrac(num, den), true
}

// integerSquareRoot returns √n and true when n is a perfect square.
func integerSquareRoot(n *big.Int) (*big.Int, bool) {
	root := new(big.Int).Sqrt(n)
	square := new(big.Int).Mul(root, root)
	return root, square.Cmp(n) == 0
}

// approximateSquareRoot computes √x in big.Float and converts it back to a fraction.
func approximateSquareRoot(x *big.Rat) *big.Rat {
	f := new(big.Float).SetPrec(sqrtPrecision).SetRat(x)
	f.Sqrt(f)
	root, _ := f.Rat(nil)
	return root
}
