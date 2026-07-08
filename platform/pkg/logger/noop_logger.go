package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewNop() *Logger {
	return &Logger{
		zapLogger: zap.NewNop(),
		level:     zap.NewAtomicLevelAt(zapcore.InfoLevel),
	}
}

func SetNopLogger() {
	globalLogger.Store(NewNop())
}
