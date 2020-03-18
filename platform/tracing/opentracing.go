package tracing

import (
	"fmt"

	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"go.opencensus.io/trace"
)

type Config struct {
	LocalEndpoint string  `json:"local_endpoint"`
	ReporterURI   string  `json:"reporter_uri"`
	ServiceName   string  `json:"service_name"`
	Probability   float64 `json:"probability"`
}

var DefaultConfig = Config{
	LocalEndpoint: "0.0.0.0:8080",
	ReporterURI:   "http://localhost:9411/api/v2/spans",
	ServiceName:   "service-name",
	Probability:   0.05,
}

func Start(cfg Config) (func(), error) {
	localEndpoint, err := openzipkin.NewEndpoint(cfg.ServiceName, cfg.LocalEndpoint)
	if err != nil {
		return nil, fmt.Errorf("can't bind tracing to local endpoint: %w", err)
	}

	reporter := zipkinHTTP.NewReporter(cfg.ReporterURI)
	exporter := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(cfg.Probability),
	})

	return func() { reporter.Close() }, nil
}
