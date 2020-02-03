package httphandler

import (
	"github.com/fewlinesco/kit/erroring"
	"net/http"
)

type BasicHTTPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (r BasicHTTPError) HTTPCode() int {
	return r.Code
}

var (
	InternalServerError = BasicHTTPError{Code: http.StatusInternalServerError, Message: "internal server error"}
	UnauthorizedError   = BasicHTTPError{Code: http.StatusUnauthorized, Message: "unauthorized error"}
)

func NewBadRequestError(detail map[string]string) BadRequestError {
	return BadRequestError{
		Code:    http.StatusBadRequest,
		Message: "bad request error",
		Detail:  detail,
	}
}

type BadRequestError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Detail  map[string]string `json:"detail,omitempty"`
}

func (b BadRequestError) HTTPCode() int {
	return b.Code
}

func NewUnprocessableEntityError(detail map[string]string) UnprocessableEntityError {
	return UnprocessableEntityError{
		Code:    http.StatusUnprocessableEntity,
		Message: "invalid payload",
		Detail:  detail,
	}
}

type UnprocessableEntityError struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Detail  map[string]string `json:"detail,omitempty"`
}

func (r UnprocessableEntityError) HTTPCode() int {
	return r.Code
}

type NotFoundHTTPError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Resource   string `json:"resource"`
	Identifier string `json:"identifier"`
}

func (r NotFoundHTTPError) HTTPCode() int {
	return r.Code
}

func NewNotFoundError(resource string, code string) NotFoundHTTPError {
	return NotFoundHTTPError{
		Code:       http.StatusBadRequest,
		Message:    "resource not found",
		Resource:   resource,
		Identifier: code,
	}
}

func FromError(err error) HTTPResponse {
	switch e := err.(type) {
	case erroring.BusinessError:
		return UnprocessableEntityError{
			Code:    http.StatusUnprocessableEntity,
			Message: e.Summary(),
			Detail:  e.Detail(),
		}
	}

	return InternalServerError
}
