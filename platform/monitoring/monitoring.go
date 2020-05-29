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

// LogLevel defines the levels at which events can be logged
type LogLevel struct {
	Debug, Info, Warning, Error, Fatal sentry.Level
}

// LogLevels maps the different levels you can do to the corresponding Sentry levels
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

// AddTagToScope adds a piece of data, as a key value pair, to the sentry scope you're currently in
// This data will be searchable in Sentry
func AddTagToScope(key string, tag string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag(key, tag)
	})
}

// AddContextToScope adds more information to the current sentry scope you're in
// this is mainly intended for adding additional diagnostic information to an event
// this information is usually not searchable in sentry
func AddContextToScope(key string, context interface{}) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetContext(key, context)
	})
}
