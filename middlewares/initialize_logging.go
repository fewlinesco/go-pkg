package middlewares

import (
	"context"
	"github.com/fewlinesco/go-pkg/logging"
	"net/http"
)

type LoggingKeyType int

const LoggingKey LoggingKeyType = 0

type InitializeLogging struct {
	Logger logging.Logger
}

func (m InitializeLogging) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), LoggingKey, m.Logger.With())

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetLogger(r *http.Request) logging.Logger {
	value := r.Context().Value(LoggingKey)
	logger, ok := value.(logging.Logger)
	if !ok {
		return nil
	}

	return logger
}
