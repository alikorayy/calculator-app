package engine

import (
	"math/big"
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name            string
		expr            string
		wantText        string
		wantApproximate bool
	}{
		// Exact results.
		{"integer", "50-7*3", "29", false},
		{"zero", "0", "0", false},
		{"negative zero", "-0", "0", false},
		{"computed zero", "1-1", "0", false},
		{"floating point classic", "0.1+0.2", "0.3", false},
		{"terminating decimal", "1/8", "0.125", false},
		{"trailing zeros are trimmed", "2.50*2", "5", false},
		{"negative decimal", "-10/4", "-2.5", false},
		{"integer with trailing zeros", "100", "100", false},
		{"decimal with integer part", "123.456", "123.456", false},

		// Rounded to 16 significant digits, half away from zero.
		{"one third", "1/3", "0.3333333333333333", true},
		{"two thirds rounds up", "2/3", "0.6666666666666667", true},
		{"negative two thirds rounds away from zero", "-2/3", "-0.6666666666666667", true},
		{"one seventh", "1/7", "0.1428571428571429", true},
		{"exact half rounds up", "1.0000000000000005", "1.000000000000001", true},
		{"negative exact half rounds away from zero", "-1.0000000000000005", "-1.000000000000001", true},
		{"irrational square root", "√(2)", "1.414213562373095", true},
		{"irrational power", "2^0.5", "1.414213562373095", true},
		{"tiny error of an irrational result is hidden", "√(2)*√(2)", "2", true},

		// Plain notation boundaries.
		{"largest plain integer", "1234567890123456", "1234567890123456", false},
		{"largest plain power of ten", "10^15", "1000000000000000", false},
		{"smallest plain decimal", "0.000001", "0.000001", false},

		// Scientific notation.
		{"first scientific power of ten", "10^16", "1e+16", false},
		{"seventeen digits are rounded", "12345678901234567", "1.234567890123457e+16", true},
		{"rounding carries into a new digit", "9999999999999999.5", "1e+16", true},
		{"product of two 32-digit numbers", "99999999999999999999999999999999^2", "1e+64", true},
		{"first scientific small number", "0.0000001", "1e-7", false},
		{"small number with digits", "0.00000012345", "1.2345e-7", false},
		{"negative scientific", "-(10^20)", "-1e+20", false},
		{"largest supported magnitude", "10^10000", "1e+10000", false},
		{"smallest supported magnitude", "10^-10000", "1e-10000", false},
		{"huge non-power of ten", "2^1000", "1.071508607186267e+301", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotApproximate := format(mustEvaluate(t, tt.expr))
			if gotText != tt.wantText || gotApproximate != tt.wantApproximate {
				t.Errorf("format(%s) = (%q, %v), want (%q, %v)",
					tt.expr, gotText, gotApproximate, tt.wantText, tt.wantApproximate)
			}
		})
	}
}

func TestFormat_ExponentNearPowersOfTen(t *testing.T) {
	// Values just below and at a power of ten must land on the correct side of it,
	// even though the exponent is first estimated with floating point.
	justBelow := new(big.Rat).Sub(powerOfTen(100), big.NewRat(1, 1))

	gotText, gotApproximate := format(justBelow)
	if want := "1e+100"; gotText != want || !gotApproximate {
		t.Errorf("format(10^100 - 1) = (%q, %v), want (%q, true)", gotText, gotApproximate, want)
	}

	gotText, _ = format(powerOfTen(-100))
	if want := "1e-100"; gotText != want {
		t.Errorf("format(10^-100) = %q, want %q", gotText, want)
	}

	exact := new(big.Rat).SetInt(new(big.Int).Sub(new(big.Int).Exp(big.NewInt(10), big.NewInt(15), nil), big.NewInt(1)))
	if gotText, _ := format(exact); gotText != strings.Repeat("9", 15) {
		t.Errorf("format(10^15 - 1) = %q, want fifteen nines", gotText)
	}
}
