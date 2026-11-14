package engine

import (
	"math"
	"math/big"
)

// power returns base^exponent exactly for an integer exponent.
//
// Special cases follow common calculator behaviour: 0^0 = 1, 0^n = 0 for n > 0,
// and 0^n for n < 0 is a division by zero. Bases 1 and -1 are answered without
// computing, so even a huge exponent is fine. For every other base the size of
// the result is estimated first, and the power is refused before any expensive
// work if it would be out of range (OVERFLOW) or too large to compute exactly
// (LIMIT_EXCEEDED).
func power(base, exponent *big.Rat) (*big.Rat, error) {
	if !exponent.IsInt() {
		return nil, newError(CodeDomainError, msgFractionalExponent)
	}
	n := exponent.Num()

	switch {
	case base.Sign() == 0:
		return powerOfZero(n)
	case isPlusOrMinusOne(base):
		return powerOfPlusOrMinusOne(base, n), nil
	}

	if err := checkPowerSize(base, n); err != nil {
		return nil, err
	}
	return exactPower(base, n), nil
}

func powerOfZero(n *big.Int) (*big.Rat, error) {
	switch n.Sign() {
	case 0:
		return big.NewRat(1, 1), nil
	case 1:
		return new(big.Rat), nil
	default:
		return nil, newError(CodeDivisionByZero, msgDivisionByZero)
	}
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

// checkPowerSize estimates base^n before computing it. float64 is used here
// only to estimate sizes, never to compute a result.
//
// Two different limits apply:
//   - magnitude: the result must lie between 10^-maxResultExponent and
//     10^maxResultExponent, otherwise it is OVERFLOW;
//   - exact size: numerator and denominator grow by a factor of n, so a base
//     with many digits and a large exponent can be in range yet need millions
//     of digits to represent exactly (e.g. 1.000…001^1000); that is LIMIT_EXCEEDED.
func checkPowerSize(base *big.Rat, n *big.Int) error {
	absN, _ := new(big.Float).SetInt(n).Float64()
	absN = math.Abs(absN)

	magnitude := absN * math.Abs(log10Abs(base.Num())-log10Abs(base.Denom()))
	if magnitude > maxResultExponent {
		return newError(CodeOverflow, msgOutOfRange)
	}

	widestPart := max(base.Num().BitLen(), base.Denom().BitLen())
	if absN*float64(widestPart) > maxExactPowerBits {
		return newError(CodeLimitExceeded, msgExponentTooLarge)
	}
	return nil
}

// exactPower computes base^n with integer arithmetic on the numerator and
// denominator. Callers must have checked the size with checkPowerSize.
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

// log10Abs estimates log10(|x|) for x != 0 from its 64 most significant bits,
// so it works for integers far beyond the float64 range.
func log10Abs(x *big.Int) float64 {
	const mantissaBits = 64
	shift := max(x.BitLen()-mantissaBits, 0)
	top := new(big.Int).Rsh(new(big.Int).Abs(x), uint(shift))
	return math.Log10(float64(top.Uint64())) + float64(shift)*math.Log10(2)
}
