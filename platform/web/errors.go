package web

import (
	"net/http"
)

// ErrorMessage represents a web error message as we want to make them consistent across all the API
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorMessage is a builder function
func NewErrorMessage(code string, message string) ErrorMessage {
	return ErrorMessage{Code: code, Message: message}
}

// ErrorDetails defines a list of key-value object representing the error details.
// It can be used like this:
// ErrorDetails{"name" => "name is required"}
type ErrorDetails map[string]string

// Error represents a JSON response to send back to the user.
// The applications handlers need to return either nil or wrap one web.Error if they want to return
// JSON errors to the clients
type Error struct {
	ErrorMessage
	HTTPCode int          `json:"-"`
	Details  ErrorDetails `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return string(e.Message)
}

var (
	// UnmanagedErrorMessage is the error message we return when an unexpected error took place
	UnmanagedErrorMessage = NewErrorMessage("500000", "Unmanaged error")
	// NotFoundMessage is the error message we return when a not found error took place
	NotFoundMessage = NewErrorMessage("400000", "Endpoint not found")
	// BadRequestMessage is the error message we return when we were unable to process the request
	BadRequestMessage = NewErrorMessage("400001", "Bad request")
	// InvalidJSONMessage is the error message we return when an unexpected error took place
	InvalidJSONMessage = NewErrorMessage("400002", "the body must be a valid JSON")
	// MissingBodyMessage is the error message we return when the request body is empty
	MissingBodyMessage = NewErrorMessage("400003", "the body is empty")
	// InvalidRequestBodyContentMessage is the error message we return when the request body contains one or more invalid input parameters
	InvalidRequestBodyContentMessage = NewErrorMessage("100001", "one ore more of the input parameters was incorrect")
	// InvalidJSONSchemaFilePath is the error message we return when we were unable to find a json schema at the specified path
	InvalidJSONSchemaFilePath = NewErrorMessage("100005", "the provided file path for the json schema is invalid")
	// RequestBodyTooLarge is the error message we return when the request body is larger than the limit set when using http.MaxBytesReader()
	RequestBodyTooLarge = NewErrorMessage("100006", "the request body is larger than the set size limit")
)

// NewErrUnmanagedResponse [deprecated] shouldn't be used outside this package. Define application specific errors instead
func NewErrUnmanagedResponse(traceid string) error {
	return &Error{
		HTTPCode:     http.StatusInternalServerError,
		ErrorMessage: UnmanagedErrorMessage,
		Details: map[string]string{
			"traceid": traceid,
		},
	}
}

// NewErrBadRequestResponse [deprecated] shouldn't be used outside this package. Define application specific errors instead
func NewErrBadRequestResponse(details ErrorDetails) error {
	return &Error{
		HTTPCode:     http.StatusBadRequest,
		ErrorMessage: BadRequestMessage,
		Details:      details,
	}
}

// NewErrNotFoundResponse [deprecated] shouldn't be used outside this package. Define application specific errors instead
func NewErrNotFoundResponse() error {
	return &Error{
		HTTPCode:     http.StatusNotFound,
		ErrorMessage: NotFoundMessage,
	}
}

// NewErrInvalidJSON is returned if we are unable to unmarshal the request body to a struct
func NewErrInvalidJSON() error {
	return &Error{
		HTTPCode:     http.StatusBadRequest,
		ErrorMessage: InvalidJSONMessage,
	}
}

// NewErrMissingRequestBody is returned when there is no body present in the request
func NewErrMissingRequestBody() error {
	return &Error{
		HTTPCode:     http.StatusBadRequest,
		ErrorMessage: MissingBodyMessage,
	}
}

// NewErrInvalidRequestBodyContent is returned when the request payload is ill-formed
// and can't be validated using the JSON schema
func NewErrInvalidRequestBodyContent(errorDetails ErrorDetails) error {
	return &Error{
		ErrorMessage: InvalidRequestBodyContentMessage,
		HTTPCode:     http.StatusBadRequest,
		Details:      errorDetails,
	}
}

// NewErrInvalidJSONSchemaFilePath is returned when the file path provided
// to the DecodeWithJSONSchema function contains an error
func NewErrInvalidJSONSchemaFilePath() error {
	return &Error{
		ErrorMessage: InvalidJSONSchemaFilePath,
		HTTPCode:     http.StatusInternalServerError,
	}
}

// NewErrRequestBodyTooLarge is returned when the request body is larger than the limit set by using http.MaxBytesReader()
var ErrRequestBodyTooLarge = &Error{
	HTTPCode:     http.StatusRequestEntityTooLarge,
	ErrorMessage: RequestBodyTooLarge,
}
