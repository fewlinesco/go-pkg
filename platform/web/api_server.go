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

type ApiServer struct {
	Name   string
	Config ServerConfig
	Server *http.Server
}

type apiServerError struct {
	Origin string
	Error  error
}

func (s *ApiServer) Start(errorChannel chan apiServerError) {
	go func() {
		err := s.Server.ListenAndServe()

		errorChannel <- apiServerError{
			Origin: s.Name,
			Error:  err,
		}
	}()
}

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

func CreateMonitoringAPIServer(serverConfig ServerConfig, logger *logging.Logger, metricsHandler http.Handler, serviceCheckers []HealthzChecker) ApiServer {
	return ApiServer{
		Name:   "monitoring",
		Config: serverConfig,
		Server: NewMonitoringServer(serverConfig, logger, WrapNetHTTPHandler("metrics", metricsHandler), serviceCheckers),
	}
}

func MonitorAPIServerShutdown(logger *logging.Logger, servers []ApiServer, serverErr chan apiServerError, shutdown chan os.Signal) error {
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

func HandleAPIServerShutdown(server ApiServer, logger *logging.Logger) error {
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
