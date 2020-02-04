package logging

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ZapLogger struct {
	L *zap.Logger
}

func NewZapLoggerForProduction() (Logger, func(), error) {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", ErrCantStart, err)
	}

	return NewZapLogger(zapLogger), func() { zapLogger.Sync() }, nil
}

func NewZapLogger(l *zap.Logger) Logger {
	return &ZapLogger{L: l.WithOptions(zap.AddCallerSkip(1))}
}

func (z *ZapLogger) With(fields ...Field) Logger {
	return z.with(fields...)
}

func (z *ZapLogger) Error(msg string) {
	z.L.Error(msg)
}

func (z *ZapLogger) Info(msg string) {
	z.L.Info(msg)
}
func (z *ZapLogger) Infof(msg string, vars ...interface{}) {
	z.L.Info(fmt.Sprintf(msg, vars...))
}

func (z *ZapLogger) with(fields ...Field) *ZapLogger {
	zapFields := make([]zapcore.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.String(field.GetName(), field.GetValue())
	}

	return &ZapLogger{L: z.L.With(zapFields...)}
}
