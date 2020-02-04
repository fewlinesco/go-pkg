package sql

import (
	"errors"
	"fmt"
	"net/url"
)

var (
	ErrNoMigrations            = errors.New("no migrations")
	ErrInvalidMigrationVersion = errors.New("invalid migration version")
	ErrMigrationTableLocked    = errors.New("database is locked. it requires a manual fix")
	ErrCantMigrate             = errors.New("can't migrate")

	ErrCantConnect = errors.New("can't connect to postgres")

	ErrUnkown = errors.New("unknown error")
)

type ConstraintError struct {
	Constraint string
	Table      string
}

func (c ConstraintError) Error() string {
	return fmt.Sprintf("constraint error on field '%s' for table '%s'", c.Constraint, c.Table)
}

type Config struct {
	User     string            `json:"user"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Host     string            `json:"host"`
	Port     string            `json:"port"`
	Options  map[string]string `json:"options"`
}

func (c Config) ConnectionString(driver string) string {
	options := url.Values{}
	for k, v := range c.Options {
		options.Set(k, v)
	}

	connection := url.URL{}
	connection.Scheme = driver
	connection.User = url.UserPassword(c.User, c.Password)
	connection.Host = fmt.Sprintf("%s:%s", c.Host, c.Port)
	connection.Path = c.Database
	connection.RawQuery = options.Encode()

	return connection.String()
}
