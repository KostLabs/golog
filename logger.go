package golog

// Logger is the minimal typed logging interface used by this package.
//
// It mirrors common leveled methods and accepts zero or more typed Field
// values. Using Field avoids map allocations in hot paths.
type Logger interface {
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
	Error(message string, fields ...Field)
	Debug(message string, fields ...Field)
}

// logger is the package-level logger used by helper functions.
// Install a custom logger with SetLogger.
var logger Logger = NewJSONLogger()

// SetLogger installs a global Logger used by package-level helpers.
func SetLogger(l Logger) {
	logger = l
}

// Info logs a message at info level via the installed package-level logger.
// If no logger is installed, the call is a no-op.
func Info(message string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Info(message, fields...)
}

// Warn logs a message at warn level via the installed package-level logger.
// If no logger is installed, the call is a no-op.
func Warn(message string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Warn(message, fields...)
}

// Error logs a message at error level via the installed package-level logger.
// If no logger is installed, the call is a no-op.
func Error(message string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Error(message, fields...)
}

// Debug logs a message at debug level via the installed package-level logger.
// If no logger is installed, the call is a no-op.
func Debug(message string, fields ...Field) {
	if logger == nil {
		return
	}
	logger.Debug(message, fields...)
}

// Info logs a message at info level with optional typed fields.
func (jsonLogger *JSONLogger) Info(message string, fields ...Field) {
	jsonLogger.logFields(InfoLevel, "info", message, fields)
}

// Warn logs a message at warn level with optional typed fields.
func (jsonLogger *JSONLogger) Warn(message string, fields ...Field) {
	jsonLogger.logFields(WarnLevel, "warn", message, fields)
}

// Error logs a message at error level with optional typed fields.
func (jsonLogger *JSONLogger) Error(message string, fields ...Field) {
	jsonLogger.logFields(ErrorLevel, "error", message, fields)
}

// Debug logs a message at debug level with optional typed fields.
func (jsonLogger *JSONLogger) Debug(message string, fields ...Field) {
	jsonLogger.logFields(DebugLevel, "debug", message, fields)
}
