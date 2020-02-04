package sql

import (
	"github.com/fewlinesco/go-pkg/logging"
)

type MigrationLogger struct {
	L logging.Logger
}

func (m *MigrationLogger) Printf(format string, v ...interface{}) {
	m.L.Infof(format, v...)
}

func (m *MigrationLogger) Verbose() bool { return false }
