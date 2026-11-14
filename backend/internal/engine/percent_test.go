package engine

import (
	"errors"
	"testing"
)

func TestPercent_Semantics(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		// On its own, or inside * and /, b% is b/100.
		{"standalone percent", "10%", "1/10"},
		{"percent of a number", "200*10%", "20"},
		{"dividing by a percent", "50/10%", "500"},
		{"percent in a power", "10%^2", "1/100"},
		{"repeated percent", "50%%", "1/200"},
		{"negative percent", "-10%", "-1/10"},

		// As the whole right operand of + or -, b% is a percentage of the left operand.
		{"adding a percentage", "50+10%", "55"},
		{"subtracting a percentage", "50-10%", "45"},
		{"keypad minus", "50−10%", "45"},
		{"percentage of a parenthesized sum", "(10+20)+10%", "33"},
		{"percentage of everything to the left", "100+10%+10%", "121"},
		{"decimal percentage stays exact", "80+12.5%", "90"},

		// When the right operand is more than a percentage, the special rule does not apply.
		{"percent inside a product on the right", "50+10%*2", "251/5"},
		{"percent on the left of a sum", "10%+50", "501/10"},
		{"percent of a percent on the right", "50+10%%", "1001/20"},
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

func TestPercent_Errors(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantCode Code
	}{
		{"error inside a percentage", "(1/0)%", CodeDivisionByZero},
		{"error inside a percentage of the left operand", "5+(1/0)%", CodeDivisionByZero},
		{"dividing by zero percent", "5/0%", CodeDivisionByZero},
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
