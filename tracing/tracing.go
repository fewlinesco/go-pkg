package tracing

import (
	"context"
	"errors"
	"fmt"
	"github.com/fewlinesco/go-pkg/logging"
	"github.com/opentracing/opentracing-go"
)

var ErrCantInitialize = errors.New("can't initialize tracer")

type Span struct {
	Name   string
	Span   opentracing.Span
	Logger logging.Logger
}

func StartSpanFromContext(ctx context.Context, logger logging.Logger, name string) (*Span, context.Context) {
	logger.Info(fmt.Sprintf("start span %s", name))
	span, newCtx := opentracing.StartSpanFromContext(ctx, name)

	return &Span{Span: span, Logger: logger, Name: name}, newCtx
}

func (s *Span) Finish() {
	s.Logger.Info(fmt.Sprintf("stop span %s", s.Name))

	s.Span.Finish()
}
