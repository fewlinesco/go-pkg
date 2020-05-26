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

// ApplicationConfig represents a minimal API configuration that can be override / augmented by the aplication
type ApplicationConfig struct {
	API             web.ServerConfig  `json:"api"`
	Monitoring      web.ServerConfig  `json:"monitoring"`
	Tracing         tracing.Config    `json:"tracing"`
	ErrorMonitoring monitoring.Config `json:"error_monitoring"`
}

// ClassicalApplicationConfig represents a classical API configuration including a SQL Database and a broker that can be override / augmented by the aplication
type ClassicalApplicationConfig struct {
	ApplicationConfig
	Database database.Config `json:"database"`
}

// Application represents a minimal API
type Application struct {
	HealthzHandler web.Handler
	Logger         *log.Logger
	Router         *web.Router
	config         ApplicationConfig
	serverErrors   chan error
}

// ClassicalApplication represents a classical API including a SQL Database and a broker
type ClassicalApplication struct {
	Application
	config   ClassicalApplicationConfig
	Database *database.DB
}

// DefaultApplicationConfig are sane default configuration for any minimal application
var DefaultApplicationConfig = ApplicationConfig{
	API:             web.DefaultServerConfig,
	Monitoring:      web.DefaultMonitoringConfig,
	Tracing:         tracing.DefaultConfig,
	ErrorMonitoring: monitoring.DefaultConfig,
}

// DefaultClassicalApplicationConfig are sane default configuration for any classical application
var DefaultClassicalApplicationConfig = ClassicalApplicationConfig{
	ApplicationConfig: DefaultApplicationConfig,
	Database:          database.DefaultConfig,
}

// DefaultClassicalApplicationMetricViews are defaults metrics generated for any classical application
func DefaultClassicalApplicationMetricViews() []*metrics.View {
	var views []*metrics.View

	views = append(views, database.MetricViews...)
	views = append(views, web.MetricViews...)

	return views
}

// ReadConfiguration reads a file and unmarshal it to the given cfg struct
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

// NewClassicalApplication creates a classical application
func NewClassicalApplication(config ClassicalApplicationConfig) (*ClassicalApplication, error) {
	db, err := database.Connect(config.Database)
	if err != nil {
		return nil, err
	}

	return &ClassicalApplication{
		Database: db,
		config:   config,
		Application: Application{
			config:       config.ApplicationConfig,
			Logger:       logging.NewDefaultLogger(),
			serverErrors: make(chan error, 2),
		},
	}, nil
}

// NewDBLessApplication creates a minimal application
func NewDBLessApplication(config ApplicationConfig) (*Application, error) {
	return &Application{

		Logger: logging.NewDefaultLogger(),

		config:       config,
		serverErrors: make(chan error, 2),
	}, nil
}

// Start spawns the HTTP and Monitoring servers
func (a *Application) Start(name string, arguments []string, router *web.Router, serviceCheckers []web.HealthzChecker) error {
	a.Router = router

	return a.StartServers(name, serviceCheckers)
}

// Start spawns the HTTP and Monitoring servers or run migrations if the first argument is "migrate"
func (c *ClassicalApplication) Start(name string, arguments []string, router *web.Router, metricViews []*metrics.View, serviceCheckers []web.HealthzChecker, migrations []darwin.Migration) error {
	var command string

	if len(arguments) > 0 {
		command = arguments[0]
		arguments = arguments[1:]
	}

	if err := metrics.RegisterViews(metricViews...); err != nil {
		return err
	}

	defaultServiceCheckers := []web.HealthzChecker{database.HealthCheck(c.Database)}
	serviceCheckers = append(defaultServiceCheckers, serviceCheckers...)

	switch command {
	case "migrate":
		return c.StartMigrations(migrations)
	default:
		return c.Application.Start(name, arguments, router, serviceCheckers)
	}
}

// StartMigrations runs the migrations
func (c *ClassicalApplication) StartMigrations(migrations []darwin.Migration) error {
	return database.Migrate(c.Database, migrations)
}

// StartServers spawns the HTTP and Monitoring server
func (a *Application) StartServers(name string, serviceCheckers []web.HealthzChecker) error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	a.Logger.Println("start tracing endpoint")
	if err := tracing.Start(a.config.Tracing); err != nil {
		return err
	}

	defer func() {
		a.Logger.Println("stop tracing endpoint")
	}()

	metricsHandler, err := metrics.CreateHandler(name)
	if err != nil {
		return fmt.Errorf("can't create metrics handler: %v", err)
	}

	go func() {
		a.Logger.Println("start monitoring server on ", a.config.Monitoring.Address)
		a.serverErrors <- web.NewMonitoringServer(a.config.Monitoring, a.Logger, web.WrapNetHTTPHandler("metrics", metricsHandler), serviceCheckers).ListenAndServe()
	}()

	if err := monitoring.CreateNewErrorMonitoring(a.config.ErrorMonitoring); err != nil {
		return err
	}

	defer func() {
		if err := recover(); err != nil {
			sentry.CurrentHub().Recover(err)
		}

		sentry.Flush(time.Duration(a.config.API.ShutdownTimeout) * time.Second)
	}()

	api := web.NewServer(a.config.API, a.Router)
	go func() {
		a.Logger.Println("start api server on ", a.config.API.Address)
		a.serverErrors <- api.ListenAndServe()
	}()

	select {
	case err := <-a.serverErrors:
		return fmt.Errorf("server failed: %v", err)

	case sig := <-shutdown:
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(a.config.API.ShutdownTimeout)*time.Second)
		defer cancel()

		err := api.Shutdown(ctx)
		if err != nil {
			a.Logger.Printf("graceful shutdown did not complete in %v : %v", a.config.API.ShutdownTimeout, err)
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
