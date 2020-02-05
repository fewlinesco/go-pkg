package tracing

import (
	"github.com/opentracing/opentracing-go"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/opentracer"
	ddtracer "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

func NewDatadogTracer(serviceName string) (opentracing.Tracer, func(), error) {
	tracer := opentracer.New(ddtracer.WithServiceName(serviceName))

	return tracer, ddtracer.Stop, nil
}
