package loggingmock

import (
	"fmt"
	"github.com/fewlinesco/go-pkg/logging"
	"reflect"
	"testing"
)

type MockLoggerSeverity string

const (
	MockLoggerSeverityError MockLoggerSeverity = "error"
	MockLoggerSeverityInfo                     = "info"
)

type MockLogger struct {
	Lines []MockLoggerLine
}

func NewMockLogger() *MockContext {
	l := &MockLogger{Lines: []MockLoggerLine{}}

	return &MockContext{Logger: l, Fields: []logging.Field{}}
}

type MockContext struct {
	Logger *MockLogger
	Fields []logging.Field
}

func (l *MockContext) AssertLine(t *testing.T, linenumber int, severity MockLoggerSeverity, msg string, fields ...logging.Field) {
	line, err := l.Line(linenumber)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if line.Severity != severity {
		t.Errorf("log line: %d: unexpected severity. want: %s; have: %s", linenumber, string(severity), string(line.Severity))
	}

	if line.Message != msg {
		t.Errorf("log line: %d: unexpected message. want: %s; have: %s", linenumber, msg, line.Message)
	}

	if len(fields) != len(line.Fields) {
		t.Errorf("log line: %d: unexpected field numbers. want: %d; have: %d", linenumber, len(fields), len(line.Fields))
	}

	for _, expectedField := range fields {
		actualField, err := line.GetField(expectedField.GetName())
		if err != nil {
			t.Errorf("%s: %w", expectedField.GetName(), err)

			continue
		}

		expectedFieldType := reflect.TypeOf(expectedField).Name()
		actualFieldType := reflect.TypeOf(actualField).Name()
		if expectedFieldType != actualFieldType {
			t.Errorf("log line: %d: unexpected field type for %s. want: %s; have: %s", linenumber, expectedField.GetName(), expectedFieldType, actualFieldType)
		}

		if expectedField.GetValue() != actualField.GetValue() {
			t.Errorf("log line: %d: unexpected field value for %s. want: %s; have: %s", linenumber, expectedField.GetName(), expectedField.GetValue(), actualField.GetValue())
		}
	}
}

func (l *MockContext) Line(linenumber int) (MockLoggerLine, error) {
	if linenumber > len(l.Logger.Lines) || linenumber < 0 {
		return MockLoggerLine{}, fmt.Errorf("index out of bound. min: %d; max: %d", 0, len(l.Logger.Lines))
	}

	return l.Logger.Lines[linenumber], nil
}

func (l *MockContext) With(additionalFields ...logging.Field) logging.Logger {
	fields := append(l.Fields, additionalFields...)
	return &MockContext{Logger: l.Logger, Fields: fields}
}

func (l *MockContext) Error(msg string) {
	line := MockLoggerLine{Message: msg, Severity: MockLoggerSeverityError, Fields: l.Fields}
	l.Logger.Lines = append(l.Logger.Lines, line)
}

func (l *MockContext) Info(msg string) {
	line := MockLoggerLine{Message: msg, Severity: MockLoggerSeverityInfo, Fields: l.Fields}
	l.Logger.Lines = append(l.Logger.Lines, line)
}

func (l *MockContext) Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	line := MockLoggerLine{Message: msg, Severity: MockLoggerSeverityInfo, Fields: l.Fields}
	l.Logger.Lines = append(l.Logger.Lines, line)
}
