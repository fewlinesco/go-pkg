package web

import (
	"net/http"
)

// ErrorMessage represents a web error message as we want to make them consistent across all the API
type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewErrorMessage is a builder function
func NewErrorMessage(code int, message string) ErrorMessage {
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
	unmanagedMessage          = NewErrorMessage(500000, "Unmanaged error")
	notFoundMessage           = NewErrorMessage(400000, "Endpoint not found")
	badRequestMessage         = NewErrorMessage(400001, "Bad request")
	unmarshallableJSONMessage = NewErrorMessage(400002, "the body must be a valid JSON")
	missingBodyMessage        = NewErrorMessage(400003, "the body is empty")
	invalidRequestMessage     = NewErrorMessage(0000000, "one ore more of the input parameters was incorrect")
)

// NewErrUnmanagedResponse [deprecated] shouldn't be used outside this package. Define application specific errors instead
func NewErrUnmanagedResponse(traceid string) error {
	return &Error{
		HTTPCode:     http.StatusInternalServerError,
		ErrorMessage: unmanagedMessage,
	}
}

// NewErrBadRequestResponse [deprecated] shouldn't be used outside this package. Define application specific errors instead
func NewErrBadRequestResponse(details ErrorDetails) error {
	return &Error{
		HTTPCode:     http.StatusBadRequest,
		ErrorMessage: badRequestMessage,
		Details:      details,
	}
}

// NewErrNotFoundResponse [deprecated] shouldn't be used outside this package. Define application specific errors instead
func NewErrNotFoundResponse() error {
	return &Error{
		HTTPCode:     http.StatusNotFound,
		ErrorMessage: notFoundMessage,
	}
}

// newErrUnmarshallableJSON is returned if we are unable to unmarshal the request body to a struct
func newErrUnmarshallableJSON() error {
	return &Error{
		HTTPCode:     http.StatusBadRequest,
		ErrorMessage: unmarshallableJSONMessage,
	}
}

// newErrMissingRequestBody is returned when there is no body present in the request
func newErrMissingRequestBody() error {
	return &Error{
		HTTPCode:     http.StatusBadRequest,
		ErrorMessage: missingBodyMessage,
	}
}

// newErrInvalidRequest is returned when the request payload is ill-formed
// and can't be validated using the JSON schema
func newErrInvalidRequest(errorDetails ErrorDetails) error {
	return &Error{
		ErrorMessage: invalidRequestMessage,
		HTTPCode:     http.StatusBadRequest,
		Details:      errorDetails,
	}
}
