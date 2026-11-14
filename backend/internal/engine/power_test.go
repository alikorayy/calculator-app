package engine

import (
	"errors"
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
		{"fractional exponent", "4^0.5", CodeDomainError},
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
