package web

import (
	"github.com/fewlinesco/go-pkg/platform/logging"
)

// Middleware is the type applications needs to conform to in order to define valid middlewares
type Middleware func(Handler) Handler

func wrapMiddleware(middlewares []Middleware, handler Handler) Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		middleware := middlewares[i]
		if middleware != nil {
			handler = middleware(handler)
		}
	}

	return handler
}

// DefaultMiddlewares contains the minimum middlewares every server should define for all its endpoints
func DefaultMiddlewares(logger *logging.Logger) []Middleware {
	return []Middleware{
		RecoveryMiddleware(),
		ErrorsMiddleware(),
		LoggerMiddleware(logger),
	}
}
