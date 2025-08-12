package logger

import (
	"context"

	"go.uber.org/zap"
)

type loggerKey struct{}

var LoggerKey = loggerKey{}

type requestIDKey struct{}

var RequestIDKey = requestIDKey{}

const (
	ServiceName = "service"
)

type Logger interface {
	Info(ctx context.Context, msg string, fields ...zap.Field)
	Error(ctx context.Context, msg string, fields ...zap.Field)
	Warn(ctx context.Context, msg string, fields ...zap.Field)
	Fatal(ctx context.Context, msg string, fields ...zap.Field)
}

type logger struct {
	serviceName string
	logger      *zap.Logger
}

func New(baseLogger *zap.Logger, serviceName string) Logger {
	if baseLogger == nil {
		panic("baseLogger cannot be nil")
	}
	return &logger{
		serviceName: serviceName,
		logger:      baseLogger,
	}
}

func (l *logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := l.addStandardFields(ctx, fields...)
	l.logger.Info(msg, allFields...)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := l.addStandardFields(ctx, fields...)
	l.logger.Warn(msg, allFields...)
}

func (l *logger) Error(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := l.addStandardFields(ctx, fields...)
	l.logger.Error(msg, allFields...)
}

func (l *logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	allFields := l.addStandardFields(ctx, fields...)
	l.logger.Fatal(msg, allFields...)
}

func (l *logger) addStandardFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	standardFields := make([]zap.Field, 0, 2+len(fields))

	standardFields = append(standardFields, zap.String(ServiceName, l.serviceName))

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		standardFields = append(standardFields, zap.String("requestID", requestID))
	}

	return append(standardFields, fields...)
}

func GetLoggerFromCtx(ctx context.Context) Logger {
	if l, ok := ctx.Value(LoggerKey).(Logger); ok && l != nil {
		return l
	}

	return New(zap.NewNop(), "no-op_logger")
}

func ContextWithLogger(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, LoggerKey, log)
}
