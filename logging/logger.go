package logging

import (
	"errors"
	"github.com/fewlinesco/go-pkg/erroring"
	"github.com/fewlinesco/go-pkg/logging/internal"
	"strings"
)

var (
	ErrCantStart = errors.New("can't start logger")
)

type Field interface {
	GetName() string
	GetValue() string
}

func Int(name string, value int) Field {
	return internal.IntField{Name: name, Value: value}
}

func String(name string, value string) Field {
	return internal.StringField{Name: name, Value: value}
}

func HTTPCode(value int) Field {
	return internal.IntField{Name: "http_code", Value: value}
}

func Stacktrace(stacktrace []string) Field {
	return internal.StringField{Name: "stacktrace", Value: strings.Join(stacktrace, ": ")}
}

func Kind(kind erroring.Kind) Field {
	return internal.StringField{Name: "kind", Value: string(kind)}
}

func Operation(operation erroring.Operation) Field {
	return internal.StringField{Name: "operation", Value: string(operation)}
}

func Source(source erroring.Source) Field {
	return internal.StringField{Name: "source", Value: string(source)}
}

func StringPtr(name string, value *string) Field {
	var v string
	if value != nil {
		v = *value
	}

	return internal.StringField{Name: name, Value: v}
}

func LogError(logger Logger, err error) {
	sysErr, ok := err.(*erroring.Error)
	if !ok {
		LogError(logger, &erroring.Error{
			Operation: erroring.Operation("unknown"),
			Kind:      erroring.KindUnexpected,
			Source:    erroring.SourceUnknown,
			Err:       err,
		})

		return
	}

	l := logger.With(
		Operation(sysErr.Operation),
		Kind(sysErr.Kind),
		Stacktrace(sysErr.Stacktrace()),
	)

	switch sysErr.Source {
	case erroring.SourceClient:
		l.Info(sysErr.Error())
		return
	default:
		l.Error(sysErr.Error())
		return
	}
}

type Logger interface {
	With(...Field) Logger

	Error(string)
	Info(string)
	Infof(string, ...interface{})
}

func NewDefaultLogger() (Logger, func(), error) {
	return NewZapLoggerForProduction()
}
