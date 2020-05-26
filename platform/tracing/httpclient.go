package tracing

import (
	"net/http"

	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

// HTTPRoundTripper is a custom http.RoundTripper that adds the current trace to the headers
// before calling other services
type HTTPRoundTripper struct{}

// NewHTTPClient is a helper to create a new HTTP Client configured to use the tracing http.RoundTripper
func NewHTTPClient() *http.Client {
	return &http.Client{Transport: HTTPRoundTripper{}}
}

// RoundTrip implements the http.RoundTripper interface
func (HTTPRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if span := trace.FromContext(r.Context()); span != nil {
		httpformat := tracecontext.HTTPFormat{}
		httpformat.SpanContextToRequest(span.SpanContext(), r)
	}

	return http.DefaultTransport.RoundTrip(r)
}
