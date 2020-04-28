package web

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"

	"github.com/fewlinesco/go-pkg/platform/monitoring"
)

type ctxKey int

const KeyValues ctxKey = 1

type Values struct {
	TraceID    string
	Now        time.Time
	StatusCode int
}

func NewRouter(logger *log.Logger, middlewares []Middleware) *Router {
	app := Router{
		Router:      mux.NewRouter(),
		logger:      logger,
		middlewares: middlewares,
	}

	app.Router.NotFoundHandler = app.defineHandler(func(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
		ctx, span := trace.StartSpan(ctx, "internal.platform.web.NotFoundHandler")
		defer span.End()

		return fmt.Errorf("%w", NewErrNotFoundResponse())
	})

	app.och = &ochttp.Handler{
		Handler:     app.Router,
		Propagation: &tracecontext.HTTPFormat{},
	}

	return &app
}

type Handler func(context.Context, http.ResponseWriter, *http.Request, map[string]string) error

type Router struct {
	*mux.Router
	logger      *log.Logger
	middlewares []Middleware
	och         *ochttp.Handler
	shutdown    chan os.Signal
}

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

func (a *Router) NewSubRouter(pathPrefix string, middleWares ...Middleware) *Router {
	subRouter := *a

	subRouter.Router = a.PathPrefix(pathPrefix).Subrouter()

	for _, m := range middleWares {
		subRouter.middlewares = append(subRouter.middlewares, m)
	}

	return &subRouter
}

func (a *Router) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.och.ServeHTTP(w, r)
}
