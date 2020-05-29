package web

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opencensus.io/trace"
)

// RecoveryMiddleware recovers panic errors to send a classical 500. It also sends the error to Sentry
func RecoveryMiddleware() Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.web.RecoveryMiddleware")

			defer func() {
				if err := recover(); err != nil {
					v := ctx.Value(KeyValues).(*Values)

					sentry.CurrentHub().Recover(err)
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
