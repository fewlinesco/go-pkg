package monitoring

import (
	"github.com/getsentry/sentry-go"
)

type Exception struct {
	Level      *sentry.Level
	Tags       map[string]string
	Contexts   map[string]interface{}
	Err        error
}

func CaptureException(exception error) Exception {
	return Exception{
		Err:        exception,
		Tags:       make(map[string]string, 0),
		Contexts:   make(map[string]interface{}, 0),
	}
}

func (exception Exception) SetLevel(level sentry.Level) Exception {
	exception.Level = &level

	return exception
}

func (exception Exception) AddTag(key string, value string) Exception {
	exception.Tags[key] = value

	return exception
}

func (exception Exception) AddContext(key string, context interface{}) Exception {
	exception.Contexts[key] = context

	return exception
}

func (exception Exception) Log() {
	sentry.WithScope(func(scope *sentry.Scope) {
		for key, tag := range exception.Tags {
			scope.SetTag(key, tag)
		}

		for key, context := range exception.Contexts {
			scope.SetContext(key, context)
		}

		if exception.Level != nil {
			scope.SetLevel(*exception.Level)
		}

		sentry.CaptureException(exception.Err)
	})
}
