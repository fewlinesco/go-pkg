package monitoring

import "github.com/getsentry/sentry-go"

type Message struct {
	Level      *sentry.Level
	Tags       map[string]string
	Contexts   map[string]interface{}
	Additional map[string]string
	Message    string
}

func CaptureMessage(message string) Message {
	return Message{
		Message:    message,
		Tags:       make(map[string]string, 0),
		Contexts:   make(map[string]interface{}, 0),
		Additional: make(map[string]string, 0),
	}
}

func (exception Message) SetLevel(level sentry.Level) Message {
	exception.Level = &level

	return exception
}

func (exception Message) AddTag(key string, value string) Message {
	exception.Tags[key] = value

	return exception
}

func (exception Message) AddContext(key string, context interface{}) Message {
	exception.Contexts[key] = context

	return exception
}

func (exception Message) AddAdditional(key string, data string) Message {
	exception.Additional[key] = data

	return exception
}

func (exception Message) Log() {
	sentry.WithScope(func(scope *sentry.Scope) {
		for key, tag := range exception.Tags {
			scope.SetTag(key, tag)
		}

		for key, context := range exception.Contexts {
			scope.SetContext(key, context)
		}

		for key, data := range exception.Additional {
			scope.SetExtra(key, data)
		}

		if exception.Level != nil {
			scope.SetLevel(*exception.Level)
		}

		sentry.CaptureMessage(exception.Message)
	})
}
