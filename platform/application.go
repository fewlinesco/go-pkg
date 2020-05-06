package platform

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GuiaBolso/darwin"
	"github.com/getsentry/sentry-go"

	"github.com/fewlinesco/go-pkg/platform/database"
	"github.com/fewlinesco/go-pkg/platform/logging"
	"github.com/fewlinesco/go-pkg/platform/metrics"
	"github.com/fewlinesco/go-pkg/platform/monitoring"
	"github.com/fewlinesco/go-pkg/platform/tracing"
	"github.com/fewlinesco/go-pkg/platform/web"
)

type ClassicalApplicationConfig struct {
	API             web.ServerConfig  `json:"api"`
	Database        database.Config   `json:"database"`
	Monitoring      web.ServerConfig  `json:"monitoring"`
	Tracing         tracing.Config    `json:"tracing"`
	ErrorMonitoring monitoring.Config `json:"error_monitoring"`
}

type ClassicalApplication struct {
	Database       *database.DB
	HealthzHandler web.Handler
	Logger         *log.Logger
	Router         *web.Router
	config         ClassicalApplicationConfig
	serverErrors   chan error
}

var DefaultClassicalApplicationConfig = ClassicalApplicationConfig{
	API:             web.DefaultServerConfig,
	Monitoring:      web.DefaultMonitoringConfig,
	Database:        database.DefaultConfig,
	Tracing:         tracing.DefaultConfig,
	ErrorMonitoring: monitoring.DefaultConfig,
}

func DefaultClassicalApplicationMetricViews() []*metrics.View {
	var views []*metrics.View

	views = append(views, database.MetricViews...)
	views = append(views, web.MetricViews...)

	return views
}

func ReadConfiguration(filepath string, cfg interface{}) error {
	cfgfile, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("can't open %s file: %v", filepath, err)
	}

	if err := json.NewDecoder(cfgfile).Decode(cfg); err != nil {
		return fmt.Errorf("can't parse %s file: %v", filepath, err)
	}

	return nil
}

func NewClassicalApplication(config ClassicalApplicationConfig) (*ClassicalApplication, error) {
	db, err := database.Connect(config.Database)
	if err != nil {
		return nil, err
	}

	return &ClassicalApplication{
		Database: db,
		Logger:   logging.NewDefaultLogger(),

		config:       config,
		serverErrors: make(chan error, 2),
	}, nil
}

func (c *ClassicalApplication) Start(name string, arguments []string, router *web.Router, metricViews []*metrics.View, healthzHandler web.Handler, migrations []darwin.Migration) error {
	var command string

	if len(arguments) > 0 {
		command = arguments[0]
		arguments = arguments[1:]
	}

	if err := metrics.RegisterViews(metricViews...); err != nil {
		return err
	}

	switch command {
	case "migrate":
		return c.StartMigrations(migrations)
	default:
		c.Router = router

		return c.StartServers(name, healthzHandler)
	}
}

func (c *ClassicalApplication) StartMigrations(migrations []darwin.Migration) error {
	return database.Migrate(c.Database, migrations)
}

func (c *ClassicalApplication) StartServers(name string, healthzHandler web.Handler) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	c.Logger.Println("start tracing endpoint")
	if err := tracing.Start(c.config.Tracing); err != nil {
		return err
	}

	defer func() {
		c.Logger.Println("stop tracing endpoint")
	}()

	metricsHandler, err := metrics.CreateHandler(name)
	if err != nil {
		return fmt.Errorf("can't create metrics handler: %v", err)
	}

	go func() {
		c.Logger.Println("start monitoring server on ", c.config.Monitoring.Address)
		c.serverErrors <- web.NewMonitoringServer(c.config.Monitoring, c.Logger, web.WrapNetHTTPHandler("metrics", metricsHandler), healthzHandler).ListenAndServe()
	}()

	if err := monitoring.CreateNewErrorMonitoring(c.config.ErrorMonitoring); err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			sentry.CurrentHub().Recover(err)
		}

		sentry.Flush(time.Duration(c.config.API.ShutdownTimeout) * time.Second)
	}()

	api := web.NewServer(c.config.API, c.Router)
	go func() {
		c.Logger.Println("start api server on ", c.config.API.Address)
		c.serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-c.serverErrors:
		return fmt.Errorf("server failed: %v", err)

	case sig := <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.config.API.ShutdownTimeout)*time.Second)
		defer cancel()

		err := api.Shutdown(ctx)
		if err != nil {
			c.Logger.Printf("graceful shutdown did not complete in %v : %v", c.config.API.ShutdownTimeout, err)
			err = api.Close()
		}

		switch {
		case sig == syscall.SIGSTOP:
			return fmt.Errorf("integrity issue caused shutdown")
		case err != nil:
			return fmt.Errorf("could not stop server gracefully: %v", err)
		}
	}

	return nil
}
