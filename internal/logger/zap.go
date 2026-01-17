package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewZapLogger creates a new zap-based logger
func NewZapLogger(level string) (Logger, error) {
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	zapLog, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		logger: zapLog,
		sugar:  zapLog.Sugar(),
	}, nil
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	l.sugar.Debugw(msg, toZapFields(fields)...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	l.sugar.Infow(msg, toZapFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	l.sugar.Warnw(msg, toZapFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	l.sugar.Errorw(msg, toZapFields(fields)...)
}

func (l *zapLogger) Fatal(msg string, fields ...Field) {
	l.sugar.Fatalw(msg, toZapFields(fields)...)
}

func (l *zapLogger) With(fields ...Field) Logger {
	newSugar := l.sugar
	for _, f := range fields {
		newSugar = newSugar.With(f.Key, f.Value)
	}
	return &zapLogger{
		logger: l.logger,
		sugar:  newSugar,
	}
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func toZapFields(fields []Field) []interface{} {
	zapFields := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		zapFields = append(zapFields, f.Key, f.Value)
	}
	return zapFields
}
