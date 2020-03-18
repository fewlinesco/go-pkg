package web

import (
	"context"
	"net/http"

	"go.opencensus.io/trace"
)

func ErrorsMiddleware() Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.web.ErrorsMiddleware")
			defer span.End()

			if err := before(ctx, w, r, params); err != nil {
				if err := RespondError(ctx, w, err); err != nil {
					return err
				}
			}

			return nil
		}

		return h
	}
}
