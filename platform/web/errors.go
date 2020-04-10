package web

import (
	"net/http"
)

type ErrorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewErrorMessage(code int, message string) ErrorMessage {
	return ErrorMessage{Code: code, Message: message}
}

type ErrorDetails map[string]string

type Error struct {
	ErrorMessage
	HTTPCode int          `json:"-"`
	Details  ErrorDetails `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return string(e.Message)
}

var (
	unmanagedMessage  = NewErrorMessage(500000, "Unmanaged error")
	notFoundMessage   = NewErrorMessage(400000, "Endpoint not found")
	badRequestMessage = NewErrorMessage(400001, "Bad request")
)

func NewErrUnmanagedResponse(traceid string) error {
	return &Error{
		HTTPCode:     http.StatusInternalServerError,
		ErrorMessage: unmanagedMessage,
	}
}

func NewErrBadRequestResponse(details ErrorDetails) error {
	return &Error{
		HTTPCode:     http.StatusBadRequest,
		ErrorMessage: badRequestMessage,
		Details:      details,
	}
}

func NewErrNotFoundResponse() error {
	return &Error{
		HTTPCode:     http.StatusNotFound,
		ErrorMessage: notFoundMessage,
	}
}
