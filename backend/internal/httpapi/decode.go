package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
)

// decodeFailure describes why a request body was rejected, ready to be
// written as an error response.
type decodeFailure struct {
	status  int
	code    string
	message string
}

// decodeJSONBody strictly decodes a JSON request body into dst: the
// Content-Type must be application/json, the body must be exactly one JSON
// object, and fields that dst does not declare are rejected. It returns nil on
// success.
//
// A body cut off by http.MaxBytesReader (see middleware.MaxBodySize) is
// reported as 413 rather than as invalid JSON.
func decodeJSONBody(r *http.Request, dst any) *decodeFailure {
	if !hasJSONContentType(r) {
		return &decodeFailure{http.StatusUnsupportedMediaType, codeUnsupportedMediaType, messageUnsupportedMediaType}
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return describeDecodeError(err)
	}
	err := decoder.Decode(&struct{}{})
	switch {
	case errors.Is(err, io.EOF):
		return nil
	case isBodyTooLarge(err):
		return describeDecodeError(err)
	default:
		return invalidRequest(messageTrailingData)
	}
}

// isBodyTooLarge reports whether err comes from reading past the body limit.
func isBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}

// hasJSONContentType reports whether the request declares a JSON body.
// Parameters such as "; charset=utf-8" are allowed.
func hasJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get(headerContentType))
	return err == nil && mediaType == contentTypeJSON
}

// describeDecodeError turns an encoding/json error into a client-facing message.
func describeDecodeError(err error) *decodeFailure {
	var typeErr *json.UnmarshalTypeError
	var maxBytesErr *http.MaxBytesError
	switch {
	case errors.As(err, &maxBytesErr):
		return &decodeFailure{http.StatusRequestEntityTooLarge, codePayloadTooLarge, fmt.Sprintf(messagePayloadTooLarge, maxBytesErr.Limit)}
	case errors.Is(err, io.EOF):
		return invalidRequest(messageEmptyBody)
	case errors.As(err, &typeErr) && typeErr.Field == "":
		return invalidRequest(messageNotAnObject)
	case errors.As(err, &typeErr):
		return invalidRequest(fmt.Sprintf(messageWrongFieldType, typeErr.Field))
	case strings.HasPrefix(err.Error(), jsonUnknownFieldPrefix):
		field := strings.TrimPrefix(err.Error(), jsonUnknownFieldPrefix)
		return invalidRequest(fmt.Sprintf(messageUnknownField, field))
	default:
		return invalidRequest(messageInvalidJSON)
	}
}

func invalidRequest(message string) *decodeFailure {
	return &decodeFailure{http.StatusBadRequest, codeInvalidRequest, message}
}

// writeDecodeFailure writes f as an error response.
func writeDecodeFailure(w http.ResponseWriter, f *decodeFailure) {
	WriteError(w, f.status, f.code, f.message)
}
