package engine

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

// format renders an exact value the way a calculator displays it: rounded to
// displayDigits significant digits (half away from zero), without trailing
// zeros, in plain notation for moderate magnitudes and in scientific notation
// ("1.5e+20", "1e-7") outside them.
//
// approximate is true when rounding dropped non-zero digits, i.e. the text is
// not the exact value (1/3 → "0.3333333333333333", √2 → "1.414213562373095").
func format(value *big.Rat) (text string, approximate bool) {
	if value.Sign() == 0 {
		return "0", false
	}

	abs := new(big.Rat).Abs(value)
	exponent := decimalExponent(abs)
	digits, exponent, approximate := roundToSignificantDigits(abs, exponent)

	text = placeDecimalPoint(digits, exponent)
	if value.Sign() < 0 {
		text = "-" + text
	}
	return text, approximate
}

// decimalExponent returns e such that 10^e <= abs < 10^(e+1). The estimate
// from log10Abs is corrected with exact comparisons, so rounding errors in
// the estimate cannot produce a wrong exponent.
func decimalExponent(abs *big.Rat) int {
	exponent := int(math.Floor(log10Abs(abs.Num()) - log10Abs(abs.Denom())))
	for abs.Cmp(powerOfTen(exponent)) < 0 {
		exponent--
	}
	for abs.Cmp(powerOfTen(exponent+1)) >= 0 {
		exponent++
	}
	return exponent
}

// roundToSignificantDigits returns the first displayDigits significant digits
// of abs, rounded half away from zero, with trailing zeros removed. Rounding
// can carry into a new leading digit (9.99…9 → 10), in which case the
// returned exponent is one larger.
func roundToSignificantDigits(abs *big.Rat, exponent int) (digits string, newExponent int, approximate bool) {
	scaled := new(big.Rat).Mul(abs, powerOfTen(displayDigits-1-exponent))
	quotient, remainder := new(big.Int).QuoRem(scaled.Num(), scaled.Denom(), new(big.Int))

	doubledRemainder := new(big.Int).Lsh(remainder, 1)
	if doubledRemainder.Cmp(scaled.Denom()) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if len(quotient.String()) > displayDigits {
		quotient.Quo(quotient, big.NewInt(10))
		exponent++
	}

	return strings.TrimRight(quotient.String(), "0"), exponent, remainder.Sign() != 0
}

// placeDecimalPoint renders significant digits with the given decimal
// exponent, choosing plain or scientific notation.
func placeDecimalPoint(digits string, exponent int) string {
	switch {
	case exponent >= scientificUpperExponent || exponent <= scientificLowerExponent:
		return scientificNotation(digits, exponent)
	case exponent < 0:
		return "0." + strings.Repeat("0", -exponent-1) + digits
	case len(digits) <= exponent+1:
		return digits + strings.Repeat("0", exponent+1-len(digits))
	default:
		return digits[:exponent+1] + "." + digits[exponent+1:]
	}
}

// scientificNotation renders digits as d.ddd…e±x, e.g. "1.5e+20" or "1e-7".
func scientificNotation(digits string, exponent int) string {
	mantissa := digits[:1]
	if len(digits) > 1 {
		mantissa += "." + digits[1:]
	}
	return fmt.Sprintf("%se%+d", mantissa, exponent)
}

// powerOfTen returns 10^exponent as an exact fraction; exponent may be negative.
func powerOfTen(exponent int) *big.Rat {
	magnitude := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(max(exponent, -exponent))), nil)
	if exponent < 0 {
		return new(big.Rat).SetFrac(big.NewInt(1), magnitude)
	}
	return new(big.Rat).SetInt(magnitude)
}
