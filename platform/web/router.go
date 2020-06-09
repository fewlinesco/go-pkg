package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"

	"github.com/fewlinesco/go-pkg/platform/logging"
	"github.com/fewlinesco/go-pkg/platform/monitoring"
)

type ctxKey int

// KeyValues is the key used to store/fetch the web values stored in the context.
// It needs to be public but we use it wisely. It's not a good practice to get data from there
// in the application context
const KeyValues ctxKey = 1

// Values represents all the web values stored in the context
type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

// Handler is the type application needs to conform to in order to handle HTTP requests
type Handler func(context.Context, http.ResponseWriter, *http.Request, map[string]string) error

// WrapNetHTTPHandler is a simple handler that wraps a classical and existing HTTP handler.
func WrapNetHTTPHandler(name string, h http.Handler) Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		_, span := trace.StartSpan(ctx, fmt.Sprintf("internal.platform.web.WrapNetHTTP.%s", name))
		defer span.End()

		h.ServeHTTP(w, r)

		return nil
	}
}

// NewRouter creates a new Router with a list of default middlewares that will be applied to all routes
func NewRouter(logger *logging.Logger, middlewares []Middleware) *Router {
	app := Router{
		Router:      mux.NewRouter(),
		logger:      logger,
		middlewares: middlewares,
	}

	app.Router.NotFoundHandler = app.defineHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		_, span := trace.StartSpan(ctx, "internal.platform.web.NotFoundHandler")
		defer span.End()

		return fmt.Errorf("%w", NewErrNotFoundResponse())
	})

	app.och = &ochttp.Handler{
		Handler:     app.Router,
		Propagation: &tracecontext.HTTPFormat{},
	}

	return &app
}

// Router represents the application routes
type Router struct {
	*mux.Router
	logger      *logging.Logger
	middlewares []Middleware
	och         *ochttp.Handler
	shutdown    chan os.Signal
}

// HandleFunc is a way to add a new router to the router. The middlewares will be added to the default middlewares set on the server. It's not a replacement.
func (a *Router) HandleFunc(method string, path string, handler Handler, middlewares ...Middleware) {
	a.Router.HandleFunc(path, a.defineHandler(handler, middlewares...)).Methods(method)
}

func (a *Router) defineHandler(handler Handler, middlewares ...Middleware) http.HandlerFunc {
	handler = wrapMiddleware(middlewares, handler)
	handler = wrapMiddleware(a.middlewares, handler)

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := trace.StartSpan(r.Context(), "internal.platform.web")
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID.String(),
			Now:     time.Now(),
		}

		ctx = context.WithValue(ctx, KeyValues, &v)

		monitoring.AddTagToScope("Trace-ID", v.TraceID)

		params := mux.Vars(r)
		queryvalues := r.URL.Query()
		for key := range queryvalues {
			params[key] = queryvalues.Get(key)
		}

		if err := handler(ctx, w, r, params); err != nil {
			a.SignalShutdown()
			return
		}
	}
}

// NewSubRouter creates a scoped subrouter on a path with its own set of default middlewares and routes.
func (a *Router) NewSubRouter(pathPrefix string, middleWares ...Middleware) *Router {
	subRouter := *a

	subRouter.Router = a.PathPrefix(pathPrefix).Subrouter()

	subRouter.middlewares = append(subRouter.middlewares, middleWares...)

	return &subRouter
}

// SignalShutdown asks the HTTP server to stop
func (a *Router) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.och.ServeHTTP(w, r)
}
