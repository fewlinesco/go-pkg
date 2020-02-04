package tracing

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	jaegermetrics "github.com/uber/jaeger-lib/metrics"
)

type JaegerConfig struct {
	AgentHost string
	AgentPort string
}

func NewJaegerTracer(serviceName string, config JaegerConfig) (opentracing.Tracer, func(), error) {
	cfg := jaegercfg.Configuration{
		Sampler:  &jaegercfg.SamplerConfig{Type: jaeger.SamplerTypeConst, Param: 1},
		Reporter: &jaegercfg.ReporterConfig{LocalAgentHostPort: fmt.Sprintf("%s:%s", config.AgentHost, config.AgentPort)},
	}

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := jaegermetrics.NullFactory

	client, closer, err := cfg.New(serviceName, jaegercfg.Logger(jLogger), jaegercfg.Metrics(jMetricsFactory))

	cfg.ServiceName = serviceName

	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrCantInitialize, err)
	}

	return client, func() { closer.Close() }, nil
}
