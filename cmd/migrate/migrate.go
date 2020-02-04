package migrate

import (
	"errors"
	"fmt"
	"github.com/fewlinesco/go-pkg/logging"
	"github.com/fewlinesco/go-pkg/sql"
	"strconv"
	"strings"
)

var (
	ErrInvalidParameter = errors.New("invalid parameter")
	ErrUnknownCommand   = errors.New("unknown command")

	commands = map[string]func(migrater sql.Migrater, logger logging.Logger, path string, params ...string) error{
		"up":    Up,
		"force": Force,
	}
)

func Migrate(command string, migrater sql.Migrater, logger logging.Logger, path string, params ...string) error {
	fn, ok := commands[command]
	if !ok {
		var validCmds []string
		for cmdName, _ := range commands {
			validCmds = append(validCmds, cmdName)
		}
		return fmt.Errorf("%w: '%v'. valid: %s", ErrUnknownCommand, command, strings.Join(validCmds, ", "))
	}

	return fn(migrater, logger, path, params...)
}

func Up(migrater sql.Migrater, logger logging.Logger, path string, params ...string) error {
	return sql.Up(migrater, logger, path)
}

func Force(migrater sql.Migrater, logger logging.Logger, path string, params ...string) error {
	if len(params) == 0 {
		return fmt.Errorf("%w: version number is missing", ErrInvalidParameter)
	}

	version, err := strconv.Atoi(params[0])
	if err != nil {
		return fmt.Errorf("%w: version number is not a number: %v", ErrInvalidParameter, err)
	}

	return sql.Force(migrater, logger, path, version)
}
