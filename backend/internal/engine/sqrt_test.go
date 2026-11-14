package engine

import (
	"errors"
	"math/big"
	"strings"
	"testing"
)

// sqrt2Digits is √2 to 70 decimal places.
const sqrt2Digits = "1.4142135623730950488016887242096980785696718753769480731766797379907325"

func TestSquareRoot_ExactForPerfectSquares(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{"perfect square", "√(9)", "3"},
		{"zero", "√(0)", "0"},
		{"decimal perfect square", "√(0.25)", "1/2"},
		{"fraction of perfect squares", "√(9/4)", "3/2"},
		{"one over a perfect square", "√(1/9)", "1/3"},
		{"nested roots", "√(√(16))", "2"},
		{"square of a root stays exact", "√(9)^2", "9"},
		{"pythagoras", "√(3^2+4^2)", "5"},
		{"sign applies after the root", "-√(4)", "-2"},
		{"huge perfect square", "√(10^10000)", "1" + strings.Repeat("0", 5000)},
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

func TestSquareRoot_IrrationalIsAccurateFarBeyondDisplayedDigits(t *testing.T) {
	root := mustEvaluate(t, "√(2)")

	const checkedDecimals = 70
	if got := root.FloatString(checkedDecimals); got != sqrt2Digits {
		t.Errorf("√2 = %s\n          want %s", got, sqrt2Digits)
	}
}

func TestSquareRoot_IrrationalRootsCombineWithinTolerance(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want *big.Rat
	}{
		{"root squared", "√(2)*√(2)", big.NewRat(2, 1)},
		{"perfect-square numerator, non-square denominator", "√(1/2)*√(2)", big.NewRat(1, 1)},
		{"non-square numerator, perfect-square denominator", "√(2/9)*3*√(2)", big.NewRat(2, 1)},
	}
	tolerance := new(big.Rat).SetFrac(big.NewInt(1), new(big.Int).Exp(big.NewInt(10), big.NewInt(70), nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mustEvaluate(t, tt.expr)

			diff := new(big.Rat).Sub(got, tt.want)
			if diff.Abs(diff).Cmp(tolerance) > 0 {
				t.Errorf("%s - %s = %s, want |difference| < 1e-70", tt.expr, tt.want.RatString(), diff.FloatString(80))
			}
		})
	}
}

func TestSquareRoot_Errors(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantCode Code
	}{
		{"negative number", "√(-4)", CodeDomainError},
		{"small negative number", "√(-0.0001)", CodeDomainError},
		{"negative expression", "√(1-2)", CodeDomainError},
		{"error inside the root", "√(1/0)", CodeDivisionByZero},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAndEvaluate(t, tt.expr)

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("evaluate(%q) error = %v, want *Error", tt.expr, err)
			}
			if engineErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", engineErr.Code, tt.wantCode)
			}
		})
	}
}

// mustEvaluate parses and evaluates expr, failing the test on any error.
func mustEvaluate(t *testing.T, expr string) *big.Rat {
	t.Helper()
	tree, err := parse(expr)
	if err != nil {
		t.Fatalf("parse(%q) unexpected error: %v", expr, err)
	}
	value, err := evaluate(tree)
	if err != nil {
		t.Fatalf("evaluate(%q) unexpected error: %v", expr, err)
	}
	return value
}
