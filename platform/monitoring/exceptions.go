package monitoring

import (
	"github.com/getsentry/sentry-go"
)

// Exception stores all the information to dispatch an event which describes an error
type Exception struct {
	Level    *sentry.Level
	Tags     map[string]string
	Contexts map[string]interface{}
	Err      error
}

// CaptureException creates a new event from an error which we'll send to Sentry
func CaptureException(exception error) Exception {
	return Exception{
		Err:      exception,
		Tags:     make(map[string]string),
		Contexts: make(map[string]interface{}),
	}
}

// SetLevel sets the loglevel of the exception
func (exception Exception) SetLevel(level sentry.Level) Exception {
	exception.Level = &level

	return exception
}

// AddTag sets a tag and it's value as additional information for the exception
func (exception Exception) AddTag(key string, value string) Exception {
	exception.Tags[key] = value

	return exception
}

// AddContext adds more context to the exception such as diagnostic information
func (exception Exception) AddContext(key string, context interface{}) Exception {
	exception.Contexts[key] = context

	return exception
}

// Log sends the created exception event to Sentry
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
