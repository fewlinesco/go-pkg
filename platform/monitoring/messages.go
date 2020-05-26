package monitoring

import "github.com/getsentry/sentry-go"

// Message ??
type Message struct {
	Level    *sentry.Level
	Tags     map[string]string
	Contexts map[string]interface{}
	Message  string
}

// CaptureMessage ??
func CaptureMessage(message string) Message {
	return Message{
		Message:  message,
		Tags:     make(map[string]string),
		Contexts: make(map[string]interface{}),
	}
}

// SetLevel ??
func (exception Message) SetLevel(level sentry.Level) Message {
	exception.Level = &level

	return exception
}

// AddTag ??
func (exception Message) AddTag(key string, value string) Message {
	exception.Tags[key] = value

	return exception
}

// AddContext  ??
func (exception Message) AddContext(key string, context interface{}) Message {
	exception.Contexts[key] = context

	return exception
}

// Log ??
func (exception Message) Log() {
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

		sentry.CaptureMessage(exception.Message)
	})
}
