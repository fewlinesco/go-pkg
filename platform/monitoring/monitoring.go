package monitoring

import (
	"github.com/getsentry/sentry-go"
)

// Config defines how to configure an error tracker
type Config struct {
	AttachStacktrace bool    `json:"attach_stack_trace"`
	DSN              string  `json:"dsn"`
	ReleaseName      string  `json:"release_name"`
	Environment      string  `json:"environment"`
	Debug            bool    `json:"debug"`
	SampleRate       float64 `json:"sample_rate"`
}

// DefaultConfig represents the default values for the error tracker
var DefaultConfig = Config{
	AttachStacktrace: true,
	DSN:              "",
	ReleaseName:      "",
	Environment:      "development",
	Debug:            false,
	SampleRate:       0.8,
}

// LogLevel ??
type LogLevel struct {
	Debug, Info, Warning, Error, Fatal sentry.Level
}

// LogLevels ??
var LogLevels = LogLevel{
	Debug:   sentry.LevelDebug,
	Info:    sentry.LevelInfo,
	Warning: sentry.LevelWarning,
	Error:   sentry.LevelError,
	Fatal:   sentry.LevelFatal,
}

// CreateNewErrorMonitoring [deprecated] should use NewErrorMonitoring
func CreateNewErrorMonitoring(cfg Config) error {
	return NewErrorMonitoring(cfg)
}

// NewErrorMonitoring configures a new, global, error tracker
func NewErrorMonitoring(cfg Config) error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.DSN,
		Debug:       cfg.Debug,
		SampleRate:  cfg.SampleRate,
		Release:     cfg.ReleaseName,
		Environment: cfg.Environment,
	})

	if err != nil {
		return err
	}

	return nil
}

// AddTagToScope ???
func AddTagToScope(key string, tag string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, tag)
	})
}

// AddContextToScope ???
func AddContextToScope(key string, context interface{}) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetContext(key, context)
	})
}
