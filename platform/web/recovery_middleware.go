package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fewlinesco/go-pkg/platform/logging"
	"github.com/getsentry/sentry-go"
	"go.opencensus.io/trace"
)

// RecoveryMiddleware recovers panic errors to send a classical 500. It also sends the error to Sentry
func RecoveryMiddleware(logger *logging.Logger) Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) (err error) {
			ctx, span := trace.StartSpan(ctx, "internal.web.RecoveryMiddleware")
			defer span.End()

			v := ctx.Value(KeyValues).(*Values)
			elapsedTime := time.Since(v.Now)

			defer func() {

				if recoverErr := recover(); recoverErr != nil {

					v := ctx.Value(KeyValues).(*Values)

					sentry.CurrentHub().Recover(err)
					sentry.Flush(2 * time.Second)

					logger.PrintRequestResponse(
						logging.RequestAttribute(r.Method, r.URL.Path, http.StatusInternalServerError),
						logging.TraceAttribute(v.TraceID),
						logging.DurationAttribute(elapsedTime),
						logging.RemoteAddressAttribute(r.RemoteAddr),
						fmt.Errorf("a panic has been recovered: %v", recoverErr).Error(),
					)
					_ = Respond(ctx, w, NewErrUnmanagedResponse(v.TraceID), http.StatusInternalServerError)
					err = nil
				}
				span.End()
			}()

			return before(ctx, w, r, params)
		}

		return h
	}
}
