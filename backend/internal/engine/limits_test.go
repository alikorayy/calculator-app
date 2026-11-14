package engine

import (
	"errors"
	"testing"
)

func TestRange_ResultsWithinTheLimitAreAccepted(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"largest power of ten", "10^10000"},
		{"smallest power of ten", "10^-10000"},
		{"just below the upper limit", "10^9999*9.99"},
		{"negative value at the limit", "-(10^10000)"},
		{"intermediate results at the limit cancel out", "10^10000 - 10^10000 + 1"},
		{"zero times a huge number", "0*10^10000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseAndEvaluate(t, tt.expr); err != nil {
				t.Errorf("evaluate(%q) unexpected error: %v", tt.expr, err)
			}
		})
	}
}

func TestRange_ResultsOutsideTheLimitOverflow(t *testing.T) {
	tests := []struct {
		name string
		expr string
	}{
		{"product of two values at the limit", "10^10000 * 10^10000"},
		{"just above the upper limit", "10^10000 * 1.0000000001"},
		{"sum just above the upper limit", "10^10000 + 1"},
		{"negative value beyond the limit", "-(10^10000) * 10"},
		{"quotient below the lower limit", "10^-10000 / 10"},
		{"percentage below the lower limit", "10^-10000 * 10%"},
		{"division producing a huge value", "1 / 10^-10000 * 10"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseAndEvaluate(t, tt.expr)

			var engineErr *Error
			if !errors.As(err, &engineErr) {
				t.Fatalf("evaluate(%q) error = %v, want *Error", tt.expr, err)
			}
			if engineErr.Code != CodeOverflow {
				t.Errorf("Code = %q, want %q", engineErr.Code, CodeOverflow)
			}
		})
	}
}
