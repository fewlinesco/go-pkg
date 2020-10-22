package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opencensus.io/trace"
)

// RecoveryMiddleware recovers panic errors to send a classical 500. It also sends the error to Sentry
func RecoveryMiddleware() Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) (err error) {
			ctx, span := trace.StartSpan(ctx, "internal.web.RecoveryMiddleware")

			defer func() {
				if err := recover(); err != nil {
					v := ctx.Value(KeyValues).(*Values)

					sentry.CurrentHub().Recover(err)
					sentry.Flush(2 * time.Second)

					err = fmt.Errorf("a panic has been recovered: %w: %v ", NewErrUnmanagedResponse(v.TraceID), err)
				}
				span.End()
			}()

			return before(ctx, w, r, params)
		}

		return h
	}
}
