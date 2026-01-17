package logger

import "time"

// Logger defines the logging interface
type Logger interface {
	// Debug logs a debug message
	Debug(msg string, fields ...Field)

	// Info logs an info message
	Info(msg string, fields ...Field)

	// Warn logs a warning message
	Warn(msg string, fields ...Field)

	// Error logs an error message
	Error(msg string, fields ...Field)

	// Fatal logs a fatal message and exits
	Fatal(msg string, fields ...Field)

	// With returns a new logger with additional fields
	With(fields ...Field) Logger

	// Sync flushes any buffered log entries
	Sync() error
}

// Field represents a key-value pair in structured logging
type Field struct {
	Key   string
	Value interface{}
}

// Helper functions to create fields
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}
