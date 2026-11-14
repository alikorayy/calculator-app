package engine

import (
	"math"
	"math/big"

	"github.com/ALTree/bigfloat"
)

// power returns base^exponent.
//
// Integer exponents are computed exactly. A fractional exponent p/q is also
// exact when the base is a perfect q-th power (4^0.5 = 2, 27^(2/3) = 9);
// otherwise the result is irrational and is computed with big.Float at
// irrationalPrecision, like an irrational square root.
//
// Special cases follow common calculator behavior: 0^0 = 1, 0^x = 0 for x > 0,
// 0^x for x < 0 is a division by zero, and a negative base with a fractional
// exponent ((-8)^0.5) has no real result and is a DOMAIN_ERROR. Before any
// expensive work the size of the result is estimated, and powers that would be
// out of range (OVERFLOW) or too large to compute exactly (LIMIT_EXCEEDED) are
// refused.
func power(base, exponent *big.Rat) (*big.Rat, error) {
	switch {
	case base.Sign() == 0:
		return powerOfZero(exponent)
	case exponent.IsInt():
		return integerPower(base, exponent.Num())
	default:
		return fractionalPower(base, exponent)
	}
}

func powerOfZero(exponent *big.Rat) (*big.Rat, error) {
	switch exponent.Sign() {
	case 0:
		return big.NewRat(1, 1), nil
	case 1:
		return new(big.Rat), nil
	default:
		return nil, newError(CodeDivisionByZero, msgDivisionByZero)
	}
}

// integerPower returns base^n exactly. base must not be zero.
func integerPower(base *big.Rat, n *big.Int) (*big.Rat, error) {
	if isPlusOrMinusOne(base) {
		return powerOfPlusOrMinusOne(base, n), nil
	}
	exponent := new(big.Rat).SetInt(n)
	if err := checkPowerMagnitude(base, exponent); err != nil {
		return nil, err
	}
	if err := checkExactPowerSize(base, n); err != nil {
		return nil, err
	}
	return exactPower(base, n), nil
}

// fractionalPower returns base^exponent for a non-integer exponent. base must
// not be zero.
func fractionalPower(base, exponent *big.Rat) (*big.Rat, error) {
	switch {
	case base.Sign() < 0:
		return nil, newError(CodeDomainError, msgInvalidInput)
	case isPlusOrMinusOne(base):
		return big.NewRat(1, 1), nil
	}
	if err := checkPowerMagnitude(base, exponent); err != nil {
		return nil, err
	}
	if root, ok := exactRoot(base, exponent.Denom()); ok {
		return integerPower(root, exponent.Num())
	}
	return approximatePower(base, exponent)
}

func isPlusOrMinusOne(r *big.Rat) bool {
	return r.IsInt() && r.Num().CmpAbs(big.NewInt(1)) == 0
}

// powerOfPlusOrMinusOne returns 1, or -1 when a negative base meets an odd exponent.
func powerOfPlusOrMinusOne(base *big.Rat, n *big.Int) *big.Rat {
	if base.Sign() < 0 && n.Bit(0) == 1 {
		return big.NewRat(-1, 1)
	}
	return big.NewRat(1, 1)
}

// checkPowerMagnitude estimates |base^exponent| = 10^(exponent·log10|base|)
// before computing it and reports OVERFLOW when it is clearly outside
// 10^-maxResultExponent .. 10^maxResultExponent. Results close to the limit
// are computed and then decided exactly by checkRange. float64 is used here
// only to estimate sizes, never to compute a result.
func checkPowerMagnitude(base, exponent *big.Rat) error {
	absExponent, _ := exponent.Float64()
	magnitude := math.Abs(absExponent) * math.Abs(preciseLog10(new(big.Rat).Abs(base)))
	// Written as "not within" so that a NaN estimate is rejected too.
	if !(magnitude <= maxResultExponent+rangeCheckMargin) {
		return newError(CodeOverflow, msgOutOfRange)
	}
	return nil
}

// preciseLog10 returns log10(x) for x > 0. It is computed in big.Float because
// float64 cannot tell a base such as 1.000000000000001 from 1, and a tiny
// logarithm multiplied by a huge exponent is still a huge result.
func preciseLog10(x *big.Rat) float64 {
	ln, _ := bigfloat.Log(new(big.Float).SetPrec(irrationalPrecision).SetRat(x)).Float64()
	return ln / math.Ln10
}

// checkExactPowerSize reports LIMIT_EXCEEDED when base^n would need too many
// bits to represent exactly. Numerator and denominator grow by a factor of n,
// so a base with many digits and a large exponent can be in range yet need
// millions of digits (e.g. 1.000…001^1000).
func checkExactPowerSize(base *big.Rat, n *big.Int) error {
	absN, _ := new(big.Float).SetInt(n).Float64()
	widestPart := max(base.Num().BitLen(), base.Denom().BitLen())
	if math.Abs(absN)*float64(widestPart) > maxExactPowerBits {
		return newError(CodeLimitExceeded, msgExponentTooLarge)
	}
	return nil
}

// exactPower computes base^n with integer arithmetic on the numerator and
// denominator. Callers must have checked the size first.
func exactPower(base *big.Rat, n *big.Int) *big.Rat {
	absN := new(big.Int).Abs(n)
	num := new(big.Int).Exp(base.Num(), absN, nil)
	den := new(big.Int).Exp(base.Denom(), absN, nil)

	result := new(big.Rat).SetFrac(num, den)
	if n.Sign() < 0 {
		result.Inv(result)
	}
	return result
}

// exactRoot returns the degree-th root of a positive base when both its
// numerator and denominator are perfect powers, e.g. the cube root of 8/27 is
// 2/3. Degrees above maxExactRootDegree are not tried.
func exactRoot(base *big.Rat, degree *big.Int) (*big.Rat, bool) {
	if !degree.IsInt64() || degree.Int64() > maxExactRootDegree {
		return nil, false
	}
	num, ok := integerRoot(base.Num(), degree.Int64())
	if !ok {
		return nil, false
	}
	den, ok := integerRoot(base.Denom(), degree.Int64())
	if !ok {
		return nil, false
	}
	return new(big.Rat).SetFrac(num, den), true
}

// integerRoot returns the degree-th root of a positive n and true when n is a
// perfect power. The root is estimated in big.Float with enough precision to
// be off by at most one, then verified exactly with integer arithmetic.
func integerRoot(n *big.Int, degree int64) (*big.Int, bool) {
	precision := uint(n.BitLen())/uint(degree) + rootGuardBits
	x := new(big.Float).SetPrec(precision).SetInt(n)
	reciprocal := new(big.Float).SetPrec(precision).Quo(big.NewFloat(1), new(big.Float).SetInt64(degree))

	estimate, _ := bigfloat.Pow(x, reciprocal).Int(nil)
	for _, candidate := range []*big.Int{estimate, new(big.Int).Add(estimate, big.NewInt(1))} {
		if new(big.Int).Exp(candidate, big.NewInt(degree), nil).Cmp(n) == 0 {
			return candidate, true
		}
	}
	return nil, false
}

// approximatePower computes an irrational base^exponent in big.Float and
// converts it back to a fraction. base must be positive.
//
// checkPowerMagnitude normally rejects results that are out of range before
// this runs. As a second line of defense, a result that big.Float itself
// cannot represent (infinite, or so small that it becomes zero) is reported
// as OVERFLOW instead of being converted.
func approximatePower(base, exponent *big.Rat) (*big.Rat, error) {
	x := new(big.Float).SetPrec(irrationalPrecision).SetRat(base)
	y := new(big.Float).SetPrec(irrationalPrecision).SetRat(exponent)
	f := bigfloat.Pow(x, y)
	if f.IsInf() || f.Sign() == 0 {
		return nil, newError(CodeOverflow, msgOutOfRange)
	}
	result, _ := f.Rat(nil)
	return result, nil
}
