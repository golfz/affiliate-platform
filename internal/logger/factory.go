package logger

var globalLogger Logger

// Init initializes the global logger
func Init(level string) error {
	logger, err := NewZapLogger(level)
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// Get returns the global logger
func Get() Logger {
	if globalLogger == nil {
		// Fallback to a basic logger if not initialized
		logger, _ := NewZapLogger("info")
		return logger
	}
	return globalLogger
}
