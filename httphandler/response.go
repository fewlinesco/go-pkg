package httphandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fewlinesco/go-pkg/erroring"
	"github.com/fewlinesco/go-pkg/logging"
	"net/http"
)

var ErrCantEncodeJSON = errors.New("can't encode json response")

type HTTPResponse interface {
	HTTPCode() int
}

func WriteJSONError(w http.ResponseWriter, logger logging.Logger, operation erroring.Operation, err error) error {
	systemErr, ok := err.(*erroring.Error)
	if !ok || systemErr.Operation != operation {
		return WriteJSON(w, InternalServerError.HTTPCode(), InternalServerError)
	}

	var response HTTPResponse

	switch systemErr.Kind {
	case erroring.KindMissingRequiredArguments, erroring.KindUnparsable:
		response = NewBadRequestError(systemErr.RelevantData)
	case erroring.KindUnprocessablePayload:
		response = NewUnprocessableEntityError(systemErr.RelevantData)
	case erroring.KindNotFound:
		response = NewNotFoundError()
	case erroring.KindInconsistentIndempotency:
		response = NewConflictError(systemErr.Error())
	default:
		response = InternalServerError
	}

	logging.LogError(logger.With(logging.HTTPCode(response.HTTPCode())), err)
	return WriteJSON(w, response.HTTPCode(), response)
}

func WriteJSON(w http.ResponseWriter, status int, data interface{}) error {
	json, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCantEncodeJSON, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(json)

	return nil
}
