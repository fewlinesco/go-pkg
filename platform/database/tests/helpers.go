package tests

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/GuiaBolso/darwin"
	"github.com/fewlinesco/go-pkg/platform/database"
	"github.com/jmoiron/sqlx"
)

func migrate(cfg database.Config, t *testing.T) func() {
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("could not connect to DB: %#v, with config: %#v", err, cfg)
	}
	defer db.Close()

	err = database.Migrate(db, []darwin.Migration{
		{
			Version:     1,
			Description: "Create test data table",
			Script: `
				DROP TABLE IF EXISTS test_data;
				CREATE TABLE test_data(
					id UUID PRIMARY KEY,
					code VARCHAR(63),
					number INTEGER DEFAULT NULL
				)`,
		},
	})

	if err != nil {
		t.Fatalf("could not migrate the database: %#v", err)
	}

	return func() {
		db, err := database.Connect(cfg)
		if err != nil {
			t.Fatalf("could not connect to DB: %#v", err)
		}
		defer db.Close()

		_, err = db.ExecContext(context.Background(), `DROP TABLE IF EXISTS test_data; DROP TABLE IF EXISTS darwin_migrations;`)
		if err != nil {
			t.Fatalf("could not clean the database: %#v", err)
		}
	}
}

func connect(config database.Config) (*sqlx.DB, error) {
	if config.URL == "" {
		options := make(url.Values)
		for key, value := range config.Options {
			options.Add(key, value)
		}

		connectionURL := url.URL{
			Scheme:   config.Scheme,
			User:     url.UserPassword(config.Username, config.Password),
			Host:     fmt.Sprintf("%s:%d", config.Host, config.Port),
			Path:     config.Database,
			RawQuery: options.Encode(),
		}
		config.URL = connectionURL.String()
	}
	return sqlx.Connect(config.Driver, config.URL)
}
