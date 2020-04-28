package web

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opencensus.io/trace"
)

func RecoveryMiddleware() Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.web.RecoveryMiddleware")

			defer func() {
				err := recover()
				if err != nil {
					v := ctx.Value(KeyValues).(*Values)

					sentry.Flush(2 * time.Second)

					_ = Respond(ctx, w, NewErrUnmanagedResponse(v.TraceID), http.StatusInternalServerError)
				}
				span.End()
			}()

			_ = before(ctx, w, r, params)

			return nil
		}

		return h
	}
}
