package loggingmock

import (
	"fmt"
	"github.com/fewlinesco/go-pkg/logging"
)

type MockLoggerLine struct {
	Message  string
	Fields   []logging.Field
	Severity MockLoggerSeverity
}

func (l *MockLoggerLine) GetField(name string) (logging.Field, error) {
	for _, f := range l.Fields {
		if f.GetName() == name {
			return f, nil
		}
	}

	return nil, fmt.Errorf("field not found")
}
