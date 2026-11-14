package httpapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const evaluatePath = "/api/v1/evaluate"

func TestEvaluate_Success(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantBody    string
	}{
		{"exact result", contentTypeJSON, `{"expression":"50 - 7 × 3"}`, `{"result":"29","approximate":false}`},
		{"rounded result", contentTypeJSON, `{"expression":"1/3"}`, `{"result":"0.3333333333333333","approximate":true}`},
		{"content type with charset", "application/json; charset=utf-8", `{"expression":"0.1+0.2"}`, `{"result":"0.3","approximate":false}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postJSON(t, evaluatePath, tt.contentType, tt.body)
			assertResponse(t, rec, http.StatusOK, tt.wantBody)
		})
	}
}

func TestEvaluate_CalculationErrorsAre422(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantBody string
	}{
		{
			"syntax error carries a position",
			`{"expression":"2 + * 3"}`,
			`{"error":{"code":"SYNTAX_ERROR","message":"Unexpected \"*\"","position":4}}`,
		},
		{
			"position zero is included",
			`{"expression":"*3"}`,
			`{"error":{"code":"SYNTAX_ERROR","message":"Unexpected \"*\"","position":0}}`,
		},
		{
			"empty expression",
			`{"expression":""}`,
			`{"error":{"code":"SYNTAX_ERROR","message":"Expression is empty","position":0}}`,
		},
		{
			"division by zero has no position",
			`{"expression":"5/0"}`,
			`{"error":{"code":"DIVISION_BY_ZERO","message":"Cannot divide by zero"}}`,
		},
		{
			"square root of a negative number",
			`{"expression":"√(-4)"}`,
			`{"error":{"code":"DOMAIN_ERROR","message":"Invalid input"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postJSON(t, evaluatePath, contentTypeJSON, tt.body)
			assertResponse(t, rec, http.StatusUnprocessableEntity, tt.wantBody)
		})
	}
}

func TestEvaluate_MalformedRequests(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		wantStatus  int
		wantBody    string
	}{
		{
			"missing content type", "", `{"expression":"1"}`, http.StatusUnsupportedMediaType,
			`{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json"}}`,
		},
		{
			"wrong content type", "text/plain", `{"expression":"1"}`, http.StatusUnsupportedMediaType,
			`{"error":{"code":"UNSUPPORTED_MEDIA_TYPE","message":"Content-Type must be application/json"}}`,
		},
		{
			"empty body", contentTypeJSON, ``, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Request body is empty"}}`,
		},
		{
			"not JSON", contentTypeJSON, `expression=1`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Request body is not valid JSON"}}`,
		},
		{
			"truncated JSON", contentTypeJSON, `{"expression":`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Request body is not valid JSON"}}`,
		},
		{
			"array instead of object", contentTypeJSON, `["1+1"]`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Request body must be a JSON object"}}`,
		},
		{
			"unknown field", contentTypeJSON, `{"expression":"1","precision":5}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Unknown field \"precision\""}}`,
		},
		{
			"wrong field type", contentTypeJSON, `{"expression":42}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Field \"expression\" has the wrong type"}}`,
		},
		{
			"missing field", contentTypeJSON, `{}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Field \"expression\" is required"}}`,
		},
		{
			"null field", contentTypeJSON, `{"expression":null}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Field \"expression\" is required"}}`,
		},
		{
			"two JSON objects", contentTypeJSON, `{"expression":"1"}{"expression":"2"}`, http.StatusBadRequest,
			`{"error":{"code":"INVALID_REQUEST","message":"Request body must contain a single JSON object"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := postJSON(t, evaluatePath, tt.contentType, tt.body)
			assertResponse(t, rec, tt.wantStatus, tt.wantBody)
		})
	}
}

func TestEvaluate_BodyCutOffByTheSizeLimitIs413(t *testing.T) {
	const object = `{"expression":"1"}`
	tests := []struct {
		name  string
		body  string
		limit int64
	}{
		{"limit reached inside the object", `{"expression":"` + strings.Repeat("1", 100) + `"}`, 20},
		{"limit reached after the object", object + strings.Repeat(" ", 100), int64(len(object)) + 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, evaluatePath, strings.NewReader(tt.body))
			req.Header.Set(headerContentType, contentTypeJSON)
			rec := httptest.NewRecorder()
			req.Body = http.MaxBytesReader(rec, req.Body, tt.limit)

			newTestRouter().ServeHTTP(rec, req)

			assertResponse(t, rec, http.StatusRequestEntityTooLarge, fmt.Sprintf(
				`{"error":{"code":"PAYLOAD_TOO_LARGE","message":"Request body is too large (maximum %d bytes)"}}`, tt.limit))
		})
	}
}

func TestEvaluate_OnlyPostIsAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, evaluatePath, nil))

	assertResponse(t, rec, http.StatusMethodNotAllowed,
		`{"error":{"code":"METHOD_NOT_ALLOWED","message":"The method is not allowed for the requested resource"}}`)
	if got := rec.Header().Get("Allow"); got != http.MethodPost {
		t.Errorf("Allow = %q, want %q", got, http.MethodPost)
	}
}

func assertResponse(t *testing.T, rec *httptest.ResponseRecorder, wantStatus int, wantBody string) {
	t.Helper()
	if rec.Code != wantStatus {
		t.Errorf("status = %d, want %d", rec.Code, wantStatus)
	}
	if got := rec.Header().Get(headerContentType); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
	}
	if got := rec.Body.String(); got != wantBody {
		t.Errorf("body = %s\n want %s", got, wantBody)
	}
}
