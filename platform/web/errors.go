package web

import (
	"net/http"
)

type ErrorMessage string
type ErrorDetails map[string]string

type Error struct {
	Code    int          `json:"code"`
	Message ErrorMessage `json:"message"`
	Details ErrorDetails `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return string(e.Message)
}

const (
	UnmanagedMessage  ErrorMessage = "[500000] Unmanaged error"
	NotFoundMessage                = "[400000] Endpoint not found"
	BadRequestMessage              = "[400001] Bad request"
)

func NewErrUnmanagedResponse(traceid string) error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: UnmanagedMessage,
	}
}

func NewErrBadRequestResponse(details ErrorDetails) error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: BadRequestMessage,
		Details: details,
	}
}

func NewErrNotFoundResponse() error {
	return &Error{
		Code:    http.StatusNotFound,
		Message: NotFoundMessage,
	}
}
