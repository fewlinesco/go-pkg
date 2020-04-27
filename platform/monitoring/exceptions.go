package monitoring

import "github.com/getsentry/sentry-go"

type LogLevel struct {
	Debug, Info, Warning, Error, Fatal sentry.Level
}

type Exception struct {
	Level    *sentry.Level
	Tags     map[string]string
	Contexts map[string]string
	Err      error
}

var LogLevels = LogLevel{
	Debug:   sentry.LevelDebug,
	Info:    sentry.LevelInfo,
	Warning: sentry.LevelWarning,
	Error:   sentry.LevelError,
	Fatal:   sentry.LevelFatal,
}

func CaptureException(exception error) Exception {
	return Exception{
		Err: exception,
		Tags: make(map[string]string, 0),
		Contexts: make(map[string]string, 0),
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

func (exception Exception) AddContext(key string, context string) Exception {
	exception.Contexts[key] = context

	return exception
}

func (exception Exception) Log() {
	sentry.WithScope(func(scope *sentry.Scope) {
		for tagName, tag := range exception.Tags {
			scope.SetTag(tagName, tag)
		}

		for contextName, context := range exception.Contexts {
			scope.SetContext(contextName, context)
		}

		if &exception.Level != nil {
			scope.SetLevel(*exception.Level)
		}
	})

	sentry.CaptureException(exception.Err)
}
