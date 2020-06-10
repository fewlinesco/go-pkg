package logging

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Logger follow the standard go Logger and only give access to `Println`
type Logger struct {
	logger *zap.Logger
}

// RemoteAddressAttribute represents a remote IP address
type RemoteAddressAttribute string

// DurationAttribute represents an elapsed time
type DurationAttribute time.Duration

// TraceAttribute represents a tracing identifier
type TraceAttribute string

// RequestAttributes represents the request information we want to log
type RequestAttributes struct {
	method     string
	path       string
	statusCode int
}

// RequestAttribute is a helper function building a RequestAttributes struct
func RequestAttribute(method string, path string, statusCode int) RequestAttributes {
	return RequestAttributes{method: method, path: path, statusCode: statusCode}
}

// NewDefaultLogger creates a new logger with a default configuration
func NewDefaultLogger() (*Logger, error) {
	zLogger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("can't create a logger: %v", err)
	}

	logger := Logger{logger: zLogger.WithOptions(zap.AddCallerSkip(1))}

	return &logger, nil
}

// Printf prints the logs
func (l *Logger) Printf(f string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(f, v...))
}

// Println prints the logs
func (l *Logger) Println(v ...interface{}) {
	l.logger.Info(fmt.Sprint(v...))
}

// PrintRequestResponse craft and log a message for an http call response
func (l *Logger) PrintRequestResponse(r RequestAttributes, t TraceAttribute, d DurationAttribute, a RemoteAddressAttribute, msg string) {
	l.logger.Info(msg,
		zap.String("method", r.method),
		zap.String("path", r.path),
		zap.Int("statuscode", r.statusCode),
		zap.String("traceid", string(t)),
		zap.Int64("duration", time.Duration(d).Milliseconds()),
		zap.String("remoteaddr", string(a)),
	)
}

// Sync ensure the logs are flushed. It should be called before shutdown at least
func (l *Logger) Sync() error {
	return l.logger.Sync()
}
