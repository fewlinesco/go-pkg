package sql

import (
	"errors"
	"fmt"
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

func (c *PostgresClient) MigrateInstance(path string) (*migrate.Migrate, error) {
	driver, err := postgres.WithInstance(c.DBx.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCantConnect, err)
	}

	return migrate.NewWithDatabaseInstance(path, "postgres", driver)
}
