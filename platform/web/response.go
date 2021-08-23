package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fewlinesco/go-pkg/platform/monitoring"
)

// Respond is a helper function in charge of sending back a JSON response to the client
func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {
	v := ctx.Value(KeyValues).(*Values)
	v.StatusCode = statusCode
	w.Header().Set("Fewlines-TraceID", v.TraceID)

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}

// RespondError is a helper function in charge of sending back a JSON response to the client based on an error.
// The error needs to be a wrapper around a web.Error, otherwise it will generate a 500 with a default message.
func RespondError(ctx context.Context, w http.ResponseWriter, err error) error {
	webErr, ok := errors.Unwrap(err).(*Error)
	if !ok {
		v := ctx.Value(KeyValues).(*Values)

		err = NewErrUnmanagedResponse(v.TraceID)
		webErr = err.(*Error)
	}

	if webErr.HTTPCode >= 500 {
		monitoring.CaptureException(err).Log()
	}

	return Respond(ctx, w, webErr, webErr.HTTPCode)
}
