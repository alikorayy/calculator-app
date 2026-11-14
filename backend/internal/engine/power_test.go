package engine

import (
	"errors"
	"math/big"
	"strings"
	"testing"
)

func TestPower_ExactResults(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{"integer power", "2^3", "8"},
		{"right-associative", "2^3^2", "512"},
		{"parentheses change the order", "(2^3)^2", "64"},
		{"sign applies after the power", "-2^2", "-4"},
		{"negative base, even exponent", "(-2)^2", "4"},
		{"negative base, odd exponent", "(-2)^3", "-8"},
		{"negative exponent", "2^-1", "1/2"},
		{"negative exponent on a fraction", "(2/3)^-2", "9/4"},
		{"decimal base stays exact", "0.1^3", "1/1000"},
		{"exponent is computed first", "2^(1+2)", "8"},
		{"exponent of zero", "5^0", "1"},
		{"zero to the power of zero", "0^0", "1"},
		{"zero to a positive power", "0^5", "0"},
		{"one to a huge power", "1^" + strings.Repeat("9", maxNumberDigits), "1"},
		{"minus one to a huge odd power", "(-1)^" + strings.Repeat("9", maxNumberDigits), "-1"},
		{"minus one to a huge even power", "(-1)^" + strings.Repeat("9", maxNumberDigits-1) + "8", "1"},

		// Fractional exponents are exact when the base is a perfect power.
		{"square root as a power", "4^0.5", "2"},
		{"cube root", "8^(1/3)", "2"},
		{"fourth root", "16^0.25", "2"},
		{"root and power combined", "27^(2/3)", "9"},
		{"fraction base", "(9/4)^0.5", "3/2"},
		{"negative fractional exponent", "4^-0.5", "1/2"},
		{"zero to a fractional power", "0^0.5", "0"},
		{"one to a fractional power", "1^0.3", "1"},
		{"huge perfect square", "(10^10000)^0.5", "1" + strings.Repeat("0", 5000)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAndEvaluate(t, tt.expr)
			if err != nil {
				t.Fatalf("evaluate(%q) unexpected error: %v", tt.expr, err)
			}
			if got != tt.want {
				t.Errorf("evaluate(%q) = %s, want %s", tt.expr, got, tt.want)
			}
		})
	}
}

func TestPower_IrrationalResultsAreAccurate(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		decimals int
		want     string
	}{
		{"power form of the square root of 2", "2^0.5", 70, sqrt2Digits},
		{"perfect-power numerator, non-power denominator", "(1/2)^0.5", 30, "0.707106781186547524400844362105"},
		// The Gelfond–Schneider constant 2^√2, an irrational power of an irrational exponent.
		{"irrational exponent", "2^√(2)", 30, "2.665144142690225188650297249873"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mustEvaluate(t, tt.expr).FloatString(tt.decimals); got != tt.want {
				t.Errorf("%s = %s\n     want %s", tt.expr, got, tt.want)
			}
		})
	}
}

func TestPower_IrrationalRootCubedIsWithinTolerance(t *testing.T) {
	got := mustEvaluate(t, "(2^(1/3))^3")

	diff := new(big.Rat).Sub(got, big.NewRat(2, 1))
	tolerance := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(70), nil))
	if diff.Abs(diff).Cmp(tolerance) > 0 {
		t.Errorf("(2^(1/3))^3 - 2 = %s, want |difference| < 1e-70", diff.FloatString(80))
	}
}

func TestApproximatePower_UnrepresentableResultIsOverflowNotPanic(t *testing.T) {
	// 1 + 10^-60 is invisible to float64; raised to ±10^70 it is e^(±10^10),
	// beyond even big.Float's exponent range.
	base := new(big.Rat).Add(big.NewRat(1, 1), powerOfTen(-60))
	tests := []struct {
		name     string
		exponent *big.Rat
	}{
		{"result too large becomes infinity", powerOfTen(70)},
		{"result too small becomes zero", new(big.Rat).Neg(powerOfTen(70))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := approximatePower(base, tt.exponent)

			var engineErr *Error
			if !errors.As(err, &engineErr) || engineErr.Code != CodeOverflow {
				t.Errorf("approximatePower = (%v, %v), want an OVERFLOW error", result, err)
			}
		})
	}
}

func TestPower_AtTheRangeLimit(t *testing.T) {
	got, err := parseAndEvaluate(t, "10^10000")
	if err != nil {
		t.Fatalf("10^10000 unexpected error: %v", err)
	}
	if wantDigits := maxResultExponent + 1; len(got) != wantDigits {
		t.Errorf("10^10000 has %d digits, want %d", len(got), wantDigits)
	}
}

func TestPower_Errors(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantCode Code
	}{
		{"zero to a negative power", "0^-1", CodeDivisionByZero},
		{"result too large", "10^10001", CodeOverflow},
		{"result too small", "10^-10001", CodeOverflow},
		{"huge exponent", "2^" + strings.Repeat("9", maxNumberDigits), CodeOverflow},
		{"tower of powers", "9^9^9", CodeOverflow},
		{"in range but too many digits to compute exactly", "1.0000000000000000000000000000001^1000", CodeLimitExceeded},
		{"negative base with a fractional exponent", "(-8)^(1/3)", CodeDomainError},
		{"minus one with a fractional exponent", "(-1)^0.5", CodeDomainError},
		{"zero to a negative fractional power", "0^-0.5", CodeDivisionByZero},
		{"fractional power too large", "10^10001.5", CodeOverflow},
		{"fractional power too small", "10^-10001.5", CodeOverflow},
		// Found by FuzzEvaluate: float64 sees this base as exactly 1, so a naive
		// estimate called the result small while big.Float overflowed to infinity.
		{"base very close to one with a huge fractional exponent", "1.000000000000001^1000^10.1", CodeOverflow},
		{"base very close to one with a huge negative fractional exponent", "1.000000000000001^-(1000^10.1)", CodeOverflow},
		{"error in the base", "(1/0)^2", CodeDivisionByZero},
		{"error in the exponent", "2^(1/0)", CodeDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAndEvaluate(t, tt.expr)

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("evaluate(%q) error = %v, want *Error", tt.expr, err)
			}
			if engineErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q (message: %s)", engineErr.Code, tt.wantCode, engineErr.Message)
			}
		})
	}
}
