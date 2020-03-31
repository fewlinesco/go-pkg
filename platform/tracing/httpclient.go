package tracing

import (
	"net/http"

	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
)

type HTTPRoundTripper struct{}

func NewHTTPClient() *http.Client {
	return &http.Client{Transport: HTTPRoundTripper{}}
}

func (HTTPRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	if span := trace.FromContext(r.Context()); span != nil {
		httpformat := tracecontext.HTTPFormat{}
		httpformat.SpanContextToRequest(span.SpanContext(), r)
	}

	return http.DefaultTransport.RoundTrip(r)
}
