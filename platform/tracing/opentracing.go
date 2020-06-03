package tracing

import (
	"context"

	"contrib.go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

// Span represents an individual unit of work in the system
type Span trace.Span

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

// StartSpan creates a new span with the provided name
func StartSpan(ctx context.Context, name string) (context.Context, *trace.Span) {
	return trace.StartSpan(ctx, name)
}

// EndSpan ends the provided running span
func EndSpan(span *trace.Span) {
	span.End()
}
