package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fewlinesco/go-pkg/platform/logging"
	"github.com/fewlinesco/go-pkg/platform/metrics"
)

// APIServer describes what an API server looks like
type APIServer struct {
	Name   string
	Config ServerConfig
	Server *http.Server
}

// APIServerError describes how an API server error looks like so we can use it to shutdown other servers in the service
type APIServerError struct {
	Origin string
	Error  error
}

// Start starts an API server and sends a message to the error channel in case it would have to close down
func (s *APIServer) Start(errorChannel chan APIServerError) {
	err := s.Server.ListenAndServe()

	errorChannel <- APIServerError{
		Origin: s.Name,
		Error:  err,
	}
}

// CreateMetricsHandler returns a http handler which can be used on the monitoring server to look at any collected metrics
func CreateMetricsHandler(namespace string) (http.Handler, error) {
	var handler http.Handler
	if err := metrics.RegisterViews(MetricViews...); err != nil {
		return handler, fmt.Errorf("can't create metrics views for %s: %v", namespace, err)
	}

	handler, err := metrics.CreateHandler(namespace)
	if err != nil {
		return handler, fmt.Errorf("can't create metrics handler: %v", err)
	}

	return handler, nil
}

// CreateMonitoringAPIServer creates a new monitoring server which can run along the other API servers in the service
func CreateMonitoringAPIServer(serverConfig ServerConfig, logger *logging.Logger, metricsHandler http.Handler, serviceCheckers []HealthzChecker) APIServer {
	return APIServer{
		Name:   "monitoring",
		Config: serverConfig,
		Server: NewMonitoringServer(serverConfig, logger, WrapNetHTTPHandler("metrics", metricsHandler), serviceCheckers),
	}
}

// MonitorAPIServerShutdown listens to the shutdown channel and error channel on which other servers push any errors in which case it will attempt
// to gracefully shutdown other servers in the service
func MonitorAPIServerShutdown(logger *logging.Logger, servers []APIServer, serverErr chan APIServerError, shutdown chan os.Signal) error {
	select {
	case apiErr := <-serverErr:
		logger.Printf("the '%s' server exited with an error: %v", apiErr.Origin, apiErr.Error)
		for _, server := range servers {
			if server.Name != apiErr.Origin {
				if err := HandleAPIServerShutdown(server, logger); err != nil {
					return err
				}
			}
		}

	case sig := <-shutdown:
		logger.Printf("shutdown signal received: %v", sig)
		for _, server := range servers {
			if err := HandleAPIServerShutdown(server, logger); err != nil {
				return err
			}
		}
	}

	return nil
}

// HandleAPIServerShutdown instructs an API server to stop and shut down gracefully
func HandleAPIServerShutdown(server APIServer, logger *logging.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(server.Config.ShutdownTimeout)*time.Second)
	defer cancel()

	err := server.Server.Shutdown(ctx)
	if err != nil {
		logger.Printf("graceful shutdown of '%s' server did not complete in %v seconds: %v", server.Name, server.Config.ShutdownTimeout, err)
		if err = server.Server.Close(); err != nil {
			return fmt.Errorf("could not stop %s server gracefully: %v", server.Name, err)
		}
	}

	logger.Printf("the '%s' server exited gracefully", server.Name)
	return nil
}
