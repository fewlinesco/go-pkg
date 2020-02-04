package sql

import (
	"errors"
	"fmt"
	"github.com/fewlinesco/go-pkg/logging"
	"github.com/golang-migrate/migrate/v4"
)

var (
	ErrNoMigrations            = errors.New("no migrations")
	ErrInvalidMigrationVersion = errors.New("invalid migration version")
	ErrMigrationTableLocked    = errors.New("database is locked. it requires a manual fix")
	ErrCantMigrate             = errors.New("can't migrate")
)

type Migrater interface {
	MigrateInstance(path string) (*migrate.Migrate, error)
}

func Force(migrater Migrater, logger logging.Logger, path string, version int) error {
	return runMigrationCommand(migrater, logger, path, func(m *migrate.Migrate) error { return m.Force(version) })
}

func Up(migrater Migrater, logger logging.Logger, path string) error {
	return runMigrationCommand(migrater, logger, path, func(m *migrate.Migrate) error { return m.Up() })
}

func runMigrationCommand(migrater Migrater, logger logging.Logger, path string, callback func(*migrate.Migrate) error) error {
	m, err := migrater.MigrateInstance(fmt.Sprintf("file://%s", path))
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

type MigrationLogger struct {
	L logging.Logger
}

func (m *MigrationLogger) Printf(format string, v ...interface{}) {
	m.L.Infof(format, v...)
}

func (m *MigrationLogger) Verbose() bool { return false }
