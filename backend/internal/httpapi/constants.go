package httpapi

// Routes served by the API, in Go's "METHOD /path" pattern form.
const (
	routeHealth    = "GET /healthz"
	routeEvaluate  = "POST /api/v1/evaluate"
	routeCalculate = "POST /api/v1/calculate"
)

// Header names and values.
const (
	headerContentType = "Content-Type"
	contentTypeJSON   = "application/json"
)

// Error codes returned in the error envelope. Clients branch on these, so they
// never change. Codes of calculation errors come from the engine package.
const (
	codeInternalError        = "INTERNAL_ERROR"
	codeNotFound             = "NOT_FOUND"
	codeMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	codeInvalidRequest       = "INVALID_REQUEST"
	codeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
)

// Error messages returned in the error envelope. Format verbs are filled in
// with the offending field.
const (
	messageInternalError        = "An unexpected error occurred"
	messageNotFound             = "The requested resource does not exist"
	messageMethodNotAllowed     = "The method is not allowed for the requested resource"
	messageUnsupportedMediaType = "Content-Type must be application/json"
	messageEmptyBody            = "Request body is empty"
	messageInvalidJSON          = "Request body is not valid JSON"
	messageNotAnObject          = "Request body must be a JSON object"
	messageTrailingData         = "Request body must contain a single JSON object"
	messageUnknownField         = "Unknown field %s"
	messageWrongFieldType       = "Field %q has the wrong type"
	messageMissingField         = "Field %q is required"
	messageInvalidOperandType   = "Operand %d must be a string or a number"
)

// jsonUnknownFieldPrefix starts the error text encoding/json returns for a
// field rejected by DisallowUnknownFields; the package has no typed error for it.
const jsonUnknownFieldPrefix = "json: unknown field "

// Request field names, as used in error messages.
const (
	fieldExpression = "expression"
	fieldOperation  = "operation"
	fieldOperands   = "operands"
)

// Health check response values.
const (
	healthStatusOK = "ok"
)
