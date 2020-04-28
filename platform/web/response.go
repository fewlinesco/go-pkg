package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/fewlinesco/go-pkg/platform/monitoring"
)

func Respond(ctx context.Context, w http.ResponseWriter, data interface{}, statusCode int) error {
	v := ctx.Value(KeyValues).(*Values)
	v.StatusCode = statusCode

	if statusCode == http.StatusNoContent {
		w.WriteHeader(statusCode)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	return json.NewEncoder(w).Encode(data)
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error) error {
	webErr, ok := errors.Unwrap(err).(*Error)

	if !ok {
		v := ctx.Value(KeyValues).(*Values)

		err = NewErrUnmanagedResponse(v.TraceID)
		webErr = err.(*Error)

		monitoring.CaptureException(err).Log()
	}

	return Respond(ctx, w, webErr, webErr.HTTPCode)
}
