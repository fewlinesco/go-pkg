package monitoring

import (
	"github.com/getsentry/sentry-go"
)

type Config struct {
	AttachStacktrace bool `json:"attach_stack_trace"`
	DSN string  `json:"dsn"`
	ReleaseName   string  `json:"release_name"`
	Environment   string  `json:"environment"`
	Debug   bool `json:"debug"`
	SampleRate float64 `json:"sample_rate"`
}

var DefaultConfig = Config{
	AttachStacktrace: true,
	DSN: "",
	ReleaseName:   "",
	Environment:   "development",
	Debug:  false,
	SampleRate: 0.8,
}

func CreateNewErrorMonitoring(cfg Config) error {
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
