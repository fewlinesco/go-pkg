package web

import (
	"context"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/plugin/ochttp/propagation/tracecontext"
	"go.opencensus.io/trace"
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
	handler = wrapMiddleware(middlewares, handler)
	handler = wrapMiddleware(a.middlewares, handler)

	a.Router.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		ctx, span := trace.StartSpan(r.Context(), "internal.platform.web")
		defer span.End()

		v := Values{
			TraceID: span.SpanContext().TraceID.String(),
			Now:     time.Now(),
		}

		ctx = context.WithValue(ctx, KeyValues, &v)

		params := mux.Vars(r)

		if err := handler(ctx, w, r, params); err != nil {
			a.SignalShutdown()
			return
		}
	}).Methods(method)
}

func (a *Router) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

func (a *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.och.ServeHTTP(w, r)
}
