package engine_test

import (
	"errors"
	"testing"

	"github.com/alikorayy/calculator-app/backend/internal/engine"
)

// FuzzEvaluate checks the engine's core promise: for any input, Evaluate
// returns either a non-empty result or an *engine.Error, and never panics.
//
// Run it with: go test ./internal/engine -run '^$' -fuzz FuzzEvaluate -fuzztime 60s
func FuzzEvaluate(f *testing.F) {
	seeds := []string{
		"50 - 7 × 3", "0.1 + 0.2", "1/3", "-2^2", "2^3^2", "50 + 10%", "10%%",
		"√(2)", "sqrt(9/4)", "27^(2/3)", "2^0.5", "6/2(1+2)", "(1+2)(3+4)",
		"5/0", "0/0", "√(-4)", "(-8)^(1/3)", "10^10000 * 10", "9^9^9",
		"", "(", ")", "2 + * 3", "1.2.3", "√9", "((((((1))))))", "--5", "2 3",
	}
	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, expression string) {
		result, err := engine.Evaluate(expression)
		checkOutcome(t, expression, result, err)
	})
}

// FuzzCalculate checks the same promise for Calculate with arbitrary
// operations and operands.
//
// Run it with: go test ./internal/engine -run '^$' -fuzz FuzzCalculate -fuzztime 60s
func FuzzCalculate(f *testing.F) {
	f.Add("add", "0.1", "0.2")
	f.Add("divide", "1", "0")
	f.Add("power", "10", "10001")
	f.Add("sqrt", "-4", "")
	f.Add("percent", "50", "")
	f.Add("modulo", "1e5", "abc")

	f.Fuzz(func(t *testing.T, op, first, second string) {
		for _, operands := range [][]string{{first}, {first, second}} {
			result, err := engine.Calculate(engine.Operation(op), operands)
			checkOutcome(t, op, result, err)
		}
	})
}

func checkOutcome(t *testing.T, input string, result engine.Result, err error) {
	t.Helper()
	if err != nil {
		var engineErr *engine.Error
		if !errors.As(err, &engineErr) {
			t.Fatalf("%q: error %v (%T) is not an *engine.Error", input, err, err)
		}
		if engineErr.Code == "" || engineErr.Message == "" {
			t.Fatalf("%q: error without code or message: %+v", input, engineErr)
		}
		return
	}
	if result.Value == "" {
		t.Fatalf("%q: empty result without an error", input)
	}
}
