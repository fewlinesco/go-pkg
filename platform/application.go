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
	"github.com/jmoiron/sqlx"

	"github.com/fewlinesco/go-pkg/platform/database"
	"github.com/fewlinesco/go-pkg/platform/logging"
	"github.com/fewlinesco/go-pkg/platform/tracing"
	"github.com/fewlinesco/go-pkg/platform/web"
)

type ClassicalApplicationConfig struct {
	API        web.ServerConfig `json:"api"`
	Database   database.Config  `json:"database"`
	Monitoring web.ServerConfig `json:"monitoring"`
	Tracing    tracing.Config   `json:"tracing"`
}

type ClassicalApplication struct {
	Database     *sqlx.DB
	Logger       *log.Logger
	Router       *web.Router
	config       ClassicalApplicationConfig
	serverErrors chan error
}

var DefaultClassicalApplicationConfig = ClassicalApplicationConfig{
	API:        web.DefaultServerConfig,
	Monitoring: web.DefaultMonitoringConfig,
	Database:   database.DefaultConfig,
	Tracing:    tracing.DefaultConfig,
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

func (c *ClassicalApplication) Start(arguments []string, router *web.Router, migrations []darwin.Migration) error {
	var command string

	if len(arguments) > 0 {
		command = arguments[0]
		arguments = arguments[1:]
	}

	switch command {
	case "migrate":
		return c.StartMigrations(migrations)
	default:
		c.Router = router

		return c.StartServers()
	}
}

func (c *ClassicalApplication) StartMigrations(migrations []darwin.Migration) error {
	return database.Migrate(c.Database, migrations)
}

func (c *ClassicalApplication) StartServers() error {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	c.Logger.Println("start tracing endpoint")
	if err := tracing.Start(c.config.Tracing); err != nil {
		return err
	}

	defer func() {
		c.Logger.Println("stop tracing endpoint")
	}()

	go func() {
		c.Logger.Println("start monitoring server on ", c.config.Monitoring.Address)
		c.serverErrors <- web.NewMonitoringServer(c.config.Monitoring).ListenAndServe()
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
