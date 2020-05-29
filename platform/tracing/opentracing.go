package tracing

import (
	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// Config represents the JSON config applications can define in order to configure tracing
type Config struct {
	LocalEndpoint string  `json:"local_endpoint"`
	ReporterURI   string  `json:"reporter_uri"`
	ServiceName   string  `json:"service_name"`
	Probability   float64 `json:"probability"`
}

// DefaultConfig are the sane defaults all applications should use
var DefaultConfig = Config{
	LocalEndpoint: "0.0.0.0:8080",
	ReporterURI:   "http://localhost:14268/api/traces",
	ServiceName:   "service-name",
	Probability:   0.05,
}

// Start configures and registers a new tracing exporter
func Start(cfg Config) error {
	exporter, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: cfg.ReporterURI,
		AgentEndpoint:     cfg.LocalEndpoint,
		Process: jaeger.Process{
			ServiceName: cfg.ServiceName,
		},
	})

	if err != nil {
		return err
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(cfg.Probability),
	})

	return nil
}
