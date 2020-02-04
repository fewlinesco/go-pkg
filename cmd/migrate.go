package cmd

import (
	"errors"
	"github.com/fewlinesco/go-pkg/cmd/migrate"
	"github.com/fewlinesco/go-pkg/logging"
	"github.com/fewlinesco/go-pkg/sql"
	"os"
)

func Migrate(command string, migrater sql.Migrater, logger logging.Logger, path string, params ...string) {
	err := migrate.Migrate(command, migrater, logger, path, params...)

	if errors.Is(err, sql.ErrNoMigrations) {
		logger.Info(err.Error())
		os.Exit(0)
	}

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
