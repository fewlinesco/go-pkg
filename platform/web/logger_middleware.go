package web

import (
	"context"
	"log"
	"net/http"
	"time"

	"go.opencensus.io/trace"
)

func LoggerMiddleware(log *log.Logger) Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.web.Logger")
			defer span.End()

			v := ctx.Value(KeyValues).(*Values)

			err := before(ctx, w, r, params)

			var message string
			if err != nil {
				message = err.Error()
			}

			log.Printf(`method="%s" path="%s" traceid="%s" statuscode="%d" duration="%s" remoteaddr="%s" message="%s"`,
				r.Method, r.URL.Path,
				v.TraceID, v.StatusCode,
				time.Since(v.Now), r.RemoteAddr,
				message,
			)

			return err
		}

		return h
	}
}
