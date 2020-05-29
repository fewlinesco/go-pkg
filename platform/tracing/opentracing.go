package tracing

import (
	"contrib.go.opencensus.io/exporter/jaeger"
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
	ReporterURI:   "http://localhost:14268/api/traces",
	ServiceName:   "service-name",
	Probability:   0.05,
}

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

// AddAttribute can add a string attribute to the provided span
// Every value which is not an empty string will be changed to "[REDACTED]"
// This masks the actual value but indicates a certain key does have a value
// Empty values will be represented by ""
func AddAttribute(span *trace.Span, key string, value string) {
	if len(value) != 0 {
		value = "[REDACTED]"
	}

	attribute := trace.StringAttribute(key, value)

	span.AddAttributes(attribute)
}

// AddAttributeWithDisclosedData adds a string attribute to a span without concealing it's value
func AddAttributeWithDisclosedData(span *trace.Span, key string, value string) {
	attribute := trace.StringAttribute(key, value)
	span.AddAttributes(attribute)
}
