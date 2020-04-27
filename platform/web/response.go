package web

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/getsentry/sentry-go"
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

		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("Trace-ID", v.TraceID)
		})
		sentry.CaptureException(err)

		err = NewErrUnmanagedResponse(v.TraceID)
		webErr = err.(*Error)
	}

	return Respond(ctx, w, webErr, webErr.HTTPCode)
}
