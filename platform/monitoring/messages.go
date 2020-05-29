package monitoring

import "github.com/getsentry/sentry-go"

// Message stores all the information to dispatch an event
type Message struct {
	Level    *sentry.Level
	Tags     map[string]string
	Contexts map[string]interface{}
	Message  string
}

// CaptureMessage creates a new message which we want to send to  Sentry
func CaptureMessage(message string) Message {
	return Message{
		Message:  message,
		Tags:     make(map[string]string),
		Contexts: make(map[string]interface{}),
	}
}

// SetLevel sets the loglevel of the message
func (exception Message) SetLevel(level sentry.Level) Message {
	exception.Level = &level

	return exception
}

// AddTag sets a tag and it's value as additional information for the message
func (exception Message) AddTag(key string, value string) Message {
	exception.Tags[key] = value

	return exception
}

// AddContext adds more context to the message such as diagnostic information
func (exception Message) AddContext(key string, context interface{}) Message {
	exception.Contexts[key] = context

	return exception
}

// Log sends the created message to Sentry
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
