package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var (
	ErrServerStopped = errors.New("server stopped")
	ErrServerFailed  = errors.New("server failed")
)

type Server struct {
	Name       string
	HTTPServer *http.Server
	failed     chan error
	stopped    chan error
}

type Router interface {
	Routes() http.Handler
}

func NewServer(name string, port int, router Router) *Server {
	address := fmt.Sprintf(":%d", port)

	return &Server{
		Name: name,
		HTTPServer: &http.Server{
			Handler:      router.Routes(),
			Addr:         address,
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		},
		failed:  make(chan error),
		stopped: make(chan error),
	}
}

func (s *Server) Start() {
	go func() {
		err := s.HTTPServer.ListenAndServe()

		switch err {
		case http.ErrServerClosed:
			s.stopped <- fmt.Errorf("%w: %v", ErrServerStopped, err)
		default:
			s.failed <- fmt.Errorf("%w: %v", ErrServerFailed, err)
		}
	}()
}

func (s *Server) Shutdown(ctx context.Context) {
	s.HTTPServer.Shutdown(ctx)
}

func (s *Server) Failed() <-chan error {
	return s.failed
}

func (s *Server) Stopped() <-chan error {
	return s.stopped
}
