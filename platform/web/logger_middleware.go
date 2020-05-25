package web

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/fewlinesco/go-pkg/platform/metrics"

	"go.opencensus.io/trace"
)

var (
	metricLatencyMs    = metrics.Float64("http_latency_ms", "The http latency in milliseconds", metrics.UnitMilliseconds)
	metricRequestTotal = metrics.Float64("http_request_total", "The total of http request using the response code as label", metrics.UnitDimensionless)

	metricTagResponseCode = metrics.MustNewTagKey("http/response_code")

	MetricViews = []*metrics.View{
		&metrics.View{
			Name:        "http/latency",
			Measure:     metricLatencyMs,
			Description: "The distribution of the latencies",
			Aggregation: metrics.ViewDistribution(0, 25, 100, 200, 400, 800, 10000),
		},
		&metrics.View{
			Name:        "http/requests",
			Measure:     metricRequestTotal,
			Description: "The number of requests",
			TagKeys:     []metrics.TagKey{metricTagResponseCode},
			Aggregation: metrics.ViewCount(),
		},
	}
)

func LoggerMiddleware(log *log.Logger) Middleware {
	return func(before Handler) Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
			ctx, span := trace.StartSpan(ctx, "internal.web.Logger")
			defer span.End()

			v := ctx.Value(KeyValues).(*Values)

			err := before(ctx, w, r, params)

			statuscode := v.StatusCode
			var message string
			if err != nil {
				message = err.Error()
				statuscode = 500
				if e, ok := errors.Unwrap(err).(*Error); ok {
					statuscode = e.HTTPCode
				}
			}

			elapsedTime := time.Since(v.Now)

			tags := []metrics.Tag{metrics.Tag{Key: metricTagResponseCode, Value: strconv.Itoa(statuscode)}}
			metrics.RecordWithTags(ctx, tags, metricLatencyMs.Measure(float64(elapsedTime.Milliseconds())))
			metrics.RecordWithTags(ctx, tags, metricRequestTotal.Measure(1))

			log.Printf(`method="%s" path="%s" traceid="%s" statuscode="%d" duration="%s" remoteaddr="%s" message="%s"`,
				r.Method, r.URL.Path,
				v.TraceID, statuscode,
				elapsedTime, r.RemoteAddr,
				message,
			)

			return err
		}

		return h
	}
}
