package web

import (
	"context"
	_ "expvar" // Register the expvar handlers
	"net/http"
	_ "net/http/pprof" // Register the pprof handlers
	"time"

	"github.com/fewlinesco/go-pkg/platform/logging"
)

// ServerConfig defines how to configure an HTTP server
type ServerConfig struct {
	Address         string `json:"address"`
	ReadTimeout     int    `json:"read_timeout"`
	WriteTimeout    int    `json:"write_timeout"`
	ShutdownTimeout int    `json:"shutdown_timeout"`
}

// DefaultServerConfig defines the default HTTP server configuration
var DefaultServerConfig = ServerConfig{
	Address:         ":8080",
	ReadTimeout:     30,
	WriteTimeout:    30,
	ShutdownTimeout: 45,
}

// DefaultMonitoringConfig defines the default HTTP monitoring server configuration
var DefaultMonitoringConfig = ServerConfig{
	Address:         ":8081",
	ReadTimeout:     30,
	WriteTimeout:    30,
	ShutdownTimeout: 45,
}

// NewServer creates a new HTTP server
func NewServer(config ServerConfig, router http.Handler) *http.Server {
	server := http.Server{
		Addr:         config.Address,
		Handler:      router,
		ReadTimeout:  time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.WriteTimeout) * time.Second,
	}

	return &server
}

// NewMonitoringServer creates a new monitoring server configured for metrics and healthz
func NewMonitoringServer(config ServerConfig, logger *logging.Logger, metricsHandler Handler, serviceCheckers []HealthzChecker) *http.Server {
	router := NewRouter(logger, nil)

	router.HandleFunc("GET", "/metrics", metricsHandler, DefaultMiddlewares(logger)...)
	router.HandleFunc("GET", "/ping", pingHandler, RecoveryMiddleware(logger), ErrorsMiddleware())
	router.HandleFunc("GET", "/healthz", HealthzHandler(serviceCheckers), DefaultMiddlewares(logger)...)
	return NewServer(config, router)
}

func pingHandler(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	w.WriteHeader(http.StatusOK)
	return nil
}
