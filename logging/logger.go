package logging

import (
	"errors"
	"github.com/fewlinesco/go-pkg/logging/internal"
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

func StringPtr(name string, value *string) Field {
	var v string
	if value != nil {
		v = *value
	}

	return internal.StringField{Name: name, Value: v}
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
