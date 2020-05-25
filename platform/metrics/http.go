package metrics

import (
	"fmt"
	"net/http"

	"contrib.go.opencensus.io/exporter/prometheus"
)

// CreateHandler creates a HTTP handler we can mount to display metrics
// The namespace is the prefix used in front of all metrics
func CreateHandler(namespace string) (http.Handler, error) {
	pe, err := prometheus.NewExporter(prometheus.Options{Namespace: namespace})

	if err != nil {
		return nil, fmt.Errorf("can't create metric exporter: %v", err)
	}

	return pe, nil
}
