package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"syscall"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey string

const (
	traceIDKey contextKey = "trace_id"
	userIDKey  contextKey = "user_id"
)

var globalLogger atomic.Pointer[Logger]

func init() {
	globalLogger.Store(NewNop())
}

type Logger struct {
	zapLogger *zap.Logger
	level     zap.AtomicLevel
}

func Init(level string, asJSON bool) error {
	return InitWithConfig(Config{
		Level: level,
		JSON:  asJSON,
	})
}

func InitWithConfig(config Config) error {
	log, err := New(config)
	if err != nil {
		return err
	}

	globalLogger.Store(log)

	return nil
}

func New(config Config) (*Logger, error) {
	zapConfig, level, err := buildZapConfig(config)
	if err != nil {
		return nil, err
	}

	zapLogger, err := zapConfig.Build(zap.AddCallerSkip(2))
	if err != nil {
		return nil, fmt.Errorf("build zap logger: %w", err)
	}

	return &Logger{
		zapLogger: zapLogger,
		level:     level,
	}, nil
}

func buildZapConfig(config Config) (zap.Config, zap.AtomicLevel, error) {
	levelText := config.Level
	if levelText == "" {
		levelText = defaultLevel
	}

	level, err := zap.ParseAtomicLevel(levelText)
	if err != nil {
		return zap.Config{}, zap.AtomicLevel{}, fmt.Errorf("parse log level: %w", err)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.Level = level
	zapConfig.Development = config.Development
	zapConfig.Encoding = "console"
	if config.JSON {
		zapConfig.Encoding = "json"
	}
	zapConfig.OutputPaths = []string{"stdout"}
	zapConfig.ErrorOutputPaths = []string{"stderr"}
	zapConfig.EncoderConfig = buildEncoderConfig()
	if config.ServiceName != "" {
		zapConfig.InitialFields = map[string]any{
			"service": config.ServiceName,
		}
	}

	return zapConfig, level, nil
}

func buildEncoderConfig() zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.MessageKey = "message"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	return encoderConfig
}

func SetLevel(levelText string) error {
	level, err := zap.ParseAtomicLevel(levelText)
	if err != nil {
		return fmt.Errorf("parse log level: %w", err)
	}

	Default().level.SetLevel(level.Level())

	return nil
}

func InitForBenchmark() {
	SetNopLogger()
}

func Default() *Logger {
	log := globalLogger.Load()
	if log == nil {
		return NewNop()
	}

	return log
}

func Sync() error {
	return Default().Sync()
}

func With(fields ...zap.Field) *Logger {
	return Default().With(fields...)
}

func WithContext(ctx context.Context) *Logger {
	return Default().WithContext(ctx)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

func Debug(ctx context.Context, msg string, fields ...zap.Field) {
	Default().Debug(ctx, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...zap.Field) {
	Default().Info(ctx, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...zap.Field) {
	Default().Warn(ctx, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	Default().Error(ctx, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	Default().Fatal(ctx, msg, fields...)
}

func (l *Logger) Sync() error {
	if l == nil || l.zapLogger == nil {
		return nil
	}

	if err := l.zapLogger.Sync(); err != nil && !isIgnorableSyncError(err) {
		return err
	}

	return nil
}

func (l *Logger) With(fields ...zap.Field) *Logger {
	if l == nil || l.zapLogger == nil {
		return NewNop()
	}

	return &Logger{
		zapLogger: l.zapLogger.With(fields...),
		level:     l.level,
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	return l.With(fieldsFromContext(ctx)...)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	if l == nil || l.zapLogger == nil {
		return
	}

	l.zapLogger.Debug(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if l == nil || l.zapLogger == nil {
		return
	}

	l.zapLogger.Info(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	if l == nil || l.zapLogger == nil {
		return
	}

	l.zapLogger.Warn(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (l *Logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	if l == nil || l.zapLogger == nil {
		return
	}

	l.zapLogger.Error(msg, append(fieldsFromContext(ctx), fields...)...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if l == nil || l.zapLogger == nil {
		return
	}

	l.zapLogger.Fatal(msg, append(fieldsFromContext(ctx), fields...)...)
}

func fieldsFromContext(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	fields := make([]zap.Field, 0, 2)

	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String(string(traceIDKey), traceID))
	}

	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String(string(userIDKey), userID))
	}

	return fields
}

func isIgnorableSyncError(err error) bool {
	return errors.Is(err, syscall.EINVAL) || errors.Is(err, os.ErrInvalid)
}
