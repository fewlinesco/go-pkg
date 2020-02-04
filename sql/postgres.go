package sql

import (
	"errors"
	"fmt"
	"github.com/fewlinesco/go-pkg/logging"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PostgresClient struct {
	DBx *sqlx.DB
}

func NewPostgresClient(c Config) (*PostgresClient, error) {
	dbx, err := sqlx.Connect("postgres", c.ConnectionString("postgres"))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCantConnect, err)
	}

	return &PostgresClient{DBx: dbx}, nil
}

func (c *PostgresClient) NamedExec(query string, arg interface{}) error {
	tx, err := c.DBx.Beginx()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnkown, err)
	}

	_, err = tx.NamedExec(query, arg)

	if err != nil {
		var pqerr *pq.Error
		if errors.As(err, &pqerr) {
			if string(pqerr.Code) == "23505" {
				return fmt.Errorf("%w", &ConstraintError{Constraint: pqerr.Constraint, Table: pqerr.Table})
			}
		}

		return fmt.Errorf("%w: %v", ErrUnkown, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrUnkown, err)
	}

	return nil
}

func (c *PostgresClient) Force(logger logging.Logger, path string, version int) error {
	return c.runMigrationCommand(logger, path, func(m *migrate.Migrate) error { return m.Force(version) })
}

func (c *PostgresClient) Migrate(logger logging.Logger, path string) error {
	return c.runMigrationCommand(logger, path, func(m *migrate.Migrate) error { return m.Up() })
}

func (c *PostgresClient) runMigrationCommand(logger logging.Logger, path string, callback func(*migrate.Migrate) error) error {
	driver, err := postgres.WithInstance(c.DBx.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCantConnect, err)
	}

	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", path), "postgres", driver)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCantMigrate, err)
	}

	m.Log = &MigrationLogger{L: logger}

	if err := callback(m); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("%w", ErrNoMigrations)
		}

		if errors.Is(err, migrate.ErrNilVersion) || errors.Is(err, migrate.ErrInvalidVersion) {
			return fmt.Errorf("%w: %v", ErrInvalidMigrationVersion, err)
		}

		if errors.Is(err, migrate.ErrLocked) {
			return fmt.Errorf("%w", ErrMigrationTableLocked)
		}

		return fmt.Errorf("%w: %v", ErrCantMigrate, err)
	}

	return nil
}
