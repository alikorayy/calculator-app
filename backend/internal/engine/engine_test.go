package engine_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/alikorayy/calculator-app/backend/internal/engine"
)

func TestEvaluate_Results(t *testing.T) {
	tests := []struct {
		expr string
		want engine.Result
	}{
		{"50 - 7 × 3", engine.Result{Value: "29"}},
		{"0.1 + 0.2", engine.Result{Value: "0.3"}},
		{"1/3", engine.Result{Value: "0.3333333333333333", Approximate: true}},
		{"50 + 10%", engine.Result{Value: "55"}},
		// Percent follows the Google calculator.
		{"50 + 10% × 2", engine.Result{Value: "110"}},
		{"50 + (10%)", engine.Result{Value: "50.1"}},
		{"50 + 10%%", engine.Result{Value: "0.55"}},
		{"-2^2", engine.Result{Value: "-4"}},
		{"6/2(1+2)", engine.Result{Value: "9"}},
		{"√(9/4)", engine.Result{Value: "1.5"}},
		{"√(2)", engine.Result{Value: "1.414213562373095", Approximate: true}},
		{"27^(2/3)", engine.Result{Value: "9"}},
		{"10^20", engine.Result{Value: "1e+20"}},
		// The x² and 1/x keys apply to the whole operand before them, as on Windows.
		{"3²⁻¹", engine.Result{Value: "0.1111111111111111", Approximate: true}},
		{"3⁻¹²", engine.Result{Value: "0.1111111111111111", Approximate: true}},
		{"3⁻¹⁻¹", engine.Result{Value: "3"}},
		{"3²^3", engine.Result{Value: "729"}},
		{"2^3²", engine.Result{Value: "512"}},
		{"-2²", engine.Result{Value: "-4"}},
		// In range, but too large to compute as an exact fraction: computed
		// approximately. Expected values come from 120-digit decimal arithmetic.
		{"√(2)^400", engine.Result{Value: "1.60693804425899e+60", Approximate: true}},
		{"1.1^50000", engine.Result{Value: "4.307823600783607e+2069", Approximate: true}},
		{"(-1.1)^50001", engine.Result{Value: "-4.738605960861967e+2069", Approximate: true}},
		{"(-1.1)^50000", engine.Result{Value: "4.307823600783607e+2069", Approximate: true}},
		{"0.9^-50000", engine.Result{Value: "7.490797101273443e+2287", Approximate: true}},
		{"1.0000000000000001^1000000", engine.Result{Value: "1.0000000001", Approximate: true}},
		{"1.0000000000000000000000000000001^1000", engine.Result{Value: "1", Approximate: true}},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			got, err := engine.Evaluate(tt.expr)
			if err != nil {
				t.Fatalf("Evaluate(%q) unexpected error: %v", tt.expr, err)
			}
			if got != tt.want {
				t.Errorf("Evaluate(%q) = %+v, want %+v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestEvaluate_Errors(t *testing.T) {
	tests := []struct {
		expr         string
		wantCode     engine.Code
		wantPosition int
	}{
		{"2 + * 3", engine.CodeSyntaxError, 4},
		{"", engine.CodeSyntaxError, 0},
		{"5 / 0", engine.CodeDivisionByZero, engine.NoPosition},
		{"0 / 0", engine.CodeUndefinedResult, engine.NoPosition},
		{"√(-4)", engine.CodeDomainError, engine.NoPosition},
		{"0⁻¹", engine.CodeDivisionByZero, engine.NoPosition},
		{"10^99999", engine.CodeOverflow, engine.NoPosition},
		{strings.Repeat("1", 300), engine.CodeLimitExceeded, engine.NoPosition},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			_, err := engine.Evaluate(tt.expr)
			assertEngineError(t, err, tt.wantCode, tt.wantPosition)
		})
	}
}

func TestCalculate_Results(t *testing.T) {
	tests := []struct {
		name     string
		op       engine.Operation
		operands []string
		want     engine.Result
	}{
		{"add", engine.OperationAdd, []string{"0.1", "0.2"}, engine.Result{Value: "0.3"}},
		{"subtract", engine.OperationSubtract, []string{"5", "7"}, engine.Result{Value: "-2"}},
		{"multiply", engine.OperationMultiply, []string{"-1.5", "4"}, engine.Result{Value: "-6"}},
		{"divide", engine.OperationDivide, []string{"10", "4"}, engine.Result{Value: "2.5"}},
		{"divide with rounding", engine.OperationDivide, []string{"1", "3"}, engine.Result{Value: "0.3333333333333333", Approximate: true}},
		{"power", engine.OperationPower, []string{"2", "10"}, engine.Result{Value: "1024"}},
		{"fractional power", engine.OperationPower, []string{"2", "0.5"}, engine.Result{Value: "1.414213562373095", Approximate: true}},
		{"sqrt", engine.OperationSqrt, []string{"9"}, engine.Result{Value: "3"}},
		{"percent", engine.OperationPercent, []string{"50"}, engine.Result{Value: "0.5"}},
		{"leading decimal separator", engine.OperationAdd, []string{".5", "5."}, engine.Result{Value: "5.5"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Calculate(tt.op, tt.operands)
			if err != nil {
				t.Fatalf("Calculate(%s, %q) unexpected error: %v", tt.op, tt.operands, err)
			}
			if got != tt.want {
				t.Errorf("Calculate(%s, %q) = %+v, want %+v", tt.op, tt.operands, got, tt.want)
			}
		})
	}
}

func TestCalculate_Errors(t *testing.T) {
	tests := []struct {
		name     string
		op       engine.Operation
		operands []string
		wantCode engine.Code
	}{
		{"unknown operation", "modulo", []string{"1", "2"}, engine.CodeInvalidArgument},
		{"empty operation", "", []string{"1", "2"}, engine.CodeInvalidArgument},
		{"too few operands", engine.OperationAdd, []string{"1"}, engine.CodeInvalidArgument},
		{"too many operands", engine.OperationSqrt, []string{"1", "2"}, engine.CodeInvalidArgument},
		{"no operands", engine.OperationDivide, nil, engine.CodeInvalidArgument},
		{"not a number", engine.OperationAdd, []string{"1", "abc"}, engine.CodeInvalidArgument},
		{"empty operand", engine.OperationAdd, []string{"", "1"}, engine.CodeInvalidArgument},
		{"lone minus sign", engine.OperationAdd, []string{"-", "1"}, engine.CodeInvalidArgument},
		{"double minus sign", engine.OperationAdd, []string{"--1", "1"}, engine.CodeInvalidArgument},
		{"two decimal separators", engine.OperationAdd, []string{"1.2.3", "1"}, engine.CodeInvalidArgument},
		{"exponent notation", engine.OperationAdd, []string{"1e5", "1"}, engine.CodeInvalidArgument},
		{"fraction notation", engine.OperationAdd, []string{"1/2", "1"}, engine.CodeInvalidArgument},
		{"hexadecimal", engine.OperationAdd, []string{"0x10", "1"}, engine.CodeInvalidArgument},
		{"surrounding whitespace", engine.OperationAdd, []string{" 1", "1"}, engine.CodeInvalidArgument},
		{"too many digits", engine.OperationAdd, []string{strings.Repeat("9", 33), "1"}, engine.CodeLimitExceeded},
		{"division by zero", engine.OperationDivide, []string{"1", "0"}, engine.CodeDivisionByZero},
		{"square root of a negative number", engine.OperationSqrt, []string{"-4"}, engine.CodeDomainError},
		{"power out of range", engine.OperationPower, []string{"10", "10001"}, engine.CodeOverflow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.Calculate(tt.op, tt.operands)
			assertEngineError(t, err, tt.wantCode, engine.NoPosition)
		})
	}
}

func TestCalculate_InvalidOperandMessageNamesTheOperand(t *testing.T) {
	_, err := engine.Calculate(engine.OperationAdd, []string{"1", "abc"})

	var engineErr *engine.Error
	if !errors.As(err, &engineErr) {
		t.Fatalf("error = %v, want *engine.Error", err)
	}
	if want := `Operand 2 is not a valid number: "abc"`; engineErr.Message != want {
		t.Errorf("Message = %q, want %q", engineErr.Message, want)
	}
}

func assertEngineError(t *testing.T, err error, wantCode engine.Code, wantPosition int) {
	t.Helper()
	var engineErr *engine.Error
	if !errors.As(err, &engineErr) {
		t.Fatalf("error = %v, want *engine.Error", err)
	}
	if engineErr.Code != wantCode {
		t.Errorf("Code = %q, want %q (message: %s)", engineErr.Code, wantCode, engineErr.Message)
	}
	if engineErr.Position != wantPosition {
		t.Errorf("Position = %d, want %d", engineErr.Position, wantPosition)
	}
}
