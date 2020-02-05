package middlewares

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"net/http"
)

type TracingKeyType int

const TracingKey TracingKeyType = 0

type InitializeTracing struct {
	Tracer opentracing.Tracer
}

func (m InitializeTracing) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := GetLogger(r)

		var serverSpan opentracing.Span

		wireContext, err := m.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))

		if err != nil {
			logger.Info(fmt.Sprintf("no incoming trace to start with: %v", err))
		}

		serverSpan = opentracing.StartSpan("start http trace", ext.RPCServerOption(wireContext))
		defer serverSpan.Finish()

		ctx := opentracing.ContextWithSpan(r.Context(), serverSpan)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
