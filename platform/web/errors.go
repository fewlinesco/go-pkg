package web

import (
	"net/http"
)

type ErrorMessage string

type Error struct {
	Code    int               `json:"code"`
	Message ErrorMessage      `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return string(e.Message)
}

const (
	UnmanagedMessage  ErrorMessage = "[500000] Unmanaged error"
	BadRequestMessage              = "[400001] Bad request"
)

func NewErrUnmanagedResponse(traceid string) error {
	return &Error{
		Code:    http.StatusInternalServerError,
		Message: UnmanagedMessage,
	}
}

func NewErrBadRequestResponse(details map[string]string) error {
	return &Error{
		Code:    http.StatusBadRequest,
		Message: BadRequestMessage,
		Details: details,
	}
}
