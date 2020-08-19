package web

import (
	"context"
	"net/http"

	"github.com/fewlinesco/go-pkg/platform/tracing"
)

// ErrorsMiddleware is in charge of sending the JSON response to the client in case of business errors
func ErrorsMiddleware() Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := tracing.StartSpan(ctx, "platform.web.ErrorMiddleware")
			defer span.End()

			if err := before(ctx, w, r, params); err != nil {
				tracing.AddAttributeWithDisclosedData(span, "error", err.Error())

				if err := RespondError(ctx, w, err); err != nil {
					return err
				}
			}

			return nil
		}

		return h
	}
}
