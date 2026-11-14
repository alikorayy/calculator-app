package engine

import (
	"math"
	"math/big"
)

// checkRange reports OVERFLOW when r is non-zero and its magnitude lies outside
// 10^-maxResultExponent .. 10^maxResultExponent.
//
// The magnitude is first estimated with log10Abs. Only when the estimate is
// within rangeCheckMargin of a limit is the exact comparison done, so the
// expensive 10^maxResultExponent is built only for results near the edge.
func checkRange(r *big.Rat) error {
	if r.Sign() == 0 {
		return nil
	}

	magnitude := math.Abs(log10Abs(r.Num()) - log10Abs(r.Denom()))
	switch {
	case magnitude < maxResultExponent-rangeCheckMargin:
		return nil
	case magnitude > maxResultExponent+rangeCheckMargin:
		return newError(CodeOverflow, msgOutOfRange)
	}

	abs := new(big.Rat).Abs(r)
	if abs.Cmp(powerOfTen(maxResultExponent)) > 0 || abs.Cmp(powerOfTen(-maxResultExponent)) < 0 {
		return newError(CodeOverflow, msgOutOfRange)
	}
	return nil
}

// log10Abs estimates log10(|x|) for x != 0 from its 64 most significant bits,
// so it works for integers far beyond the float64 range. float64 is used only
// to estimate sizes, never to compute a result.
func log10Abs(x *big.Int) float64 {
	const mantissaBits = 64
	shift := max(x.BitLen()-mantissaBits, 0)
	top := new(big.Int).Rsh(new(big.Int).Abs(x), uint(shift))
	return math.Log10(float64(top.Uint64())) + float64(shift)*math.Log10(2)
}
