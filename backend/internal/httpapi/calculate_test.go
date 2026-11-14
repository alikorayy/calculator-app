package httpapi

import (
	"net/http"
	"testing"
)

const calculatePath = "/api/v1/calculate"

func TestCalculate_Success(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantBody string
	}{
		{"string operands", `{"operation":"divide","operands":["10","4"]}`, `{"result":"2.5","approximate":false}`},
		{"number operands", `{"operation":"divide","operands":[10,4]}`, `{"result":"2.5","approximate":false}`},
		{"mixed operands", `{"operation":"subtract","operands":["5",7]}`, `{"result":"-2","approximate":false}`},
		{"negative number", `{"operation":"multiply","operands":[-1.5,4]}`, `{"result":"-6","approximate":false}`},
		{"single operand", `{"operation":"sqrt","operands":["9"]}`, `{"result":"3","approximate":false}`},
		{"rounded result", `{"operation":"divide","operands":[1,3]}`, `{"result":"0.3333333333333333","approximate":true}`},
		// 2^53 + 1 cannot be represented by float64 (it would become ...992);
		// the number's exact text must reach the engine.
		{"number beyond float64 precision", `{"operation":"add","operands":[9007199254740993,0]}`, `{"result":"9007199254740993","approximate":false}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postJSON(t, calculatePath, contentTypeJSON, tt.body)
			assertResponse(t, rec, http.StatusOK, tt.wantBody)
		})
	}
}

func TestCalculate_Errors(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			"unknown operation", `{"operation":"modulo","operands":["1","2"]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Unknown operation \"modulo\""}}`,
		},
		{
			"wrong number of operands", `{"operation":"add","operands":["1"]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Wrong number of operands for \"add\": expected 2, got 1"}}`,
		},
		{
			"empty operand list", `{"operation":"sqrt","operands":[]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Wrong number of operands for \"sqrt\": expected 1, got 0"}}`,
		},
		{
			"operand that is not a number", `{"operation":"add","operands":["1","abc"]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Operand 2 is not a valid number: \"abc\""}}`,
		},
		{
			"exponent notation", `{"operation":"add","operands":[1e5,1]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Operand 1 is not a valid number: \"1e5\""}}`,
		},
		{
			"boolean operand", `{"operation":"add","operands":[true,1]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Operand 1 must be a string or a number"}}`,
		},
		{
			"null operand", `{"operation":"add","operands":[1,null]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Operand 2 must be a string or a number"}}`,
		},
		{
			"object operand", `{"operation":"sqrt","operands":[{"value":4}]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Operand 1 must be a string or a number"}}`,
		},
		{
			"missing operation", `{"operands":["1","2"]}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Field \"operation\" is required"}}`,
		},
		{
			"missing operands", `{"operation":"add"}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Field \"operands\" is required"}}`,
		},
		{
			"operands is not a list", `{"operation":"add","operands":"1,2"}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Field \"operands\" has the wrong type"}}`,
		},
		{
			"unknown field", `{"operation":"add","operands":["1","2"],"round":true}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Unknown field \"round\""}}`,
		},
		{
			"too many digits", `{"operation":"add","operands":["123456789012345678901234567890123","1"]}`, http.StatusUnprocessableEntity,
			`{"error":{"code":"LIMIT_EXCEEDED","message":"Number has too many digits (maximum 32)"}}`,
		},
		{
			"division by zero", `{"operation":"divide","operands":[1,0]}`, http.StatusUnprocessableEntity,
			`{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero"}}`,
		},
		{
			"square root of a negative number", `{"operation":"sqrt","operands":[-4]}`, http.StatusUnprocessableEntity,
			`{"error":{"code":"DOMAIN_ERROR","message":"Invalid input"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postJSON(t, calculatePath, contentTypeJSON, tt.body)
			assertResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}

func TestCalculate_RejectsNonJSONContentType(t *testing.T) {
	rec := postJSON(t, calculatePath, "text/plain", `{"operation":"add","operands":[1,2]}`)
	assertResponse(t, rec, http.StatusUnsupportedMediaType,
		`{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json"}}`)
}
