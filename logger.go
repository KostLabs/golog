package golog

// Logger is the minimal logging interface used by this package. It mirrors the
// common leveled logging methods and accepts zero or more maps of additional
// fields which are merged into the emitted structured log entry.
//
// Example (implementing and installing a simple adapter in your service):
//
//	// package main
//	//
//	// import (
//	//     "fmt"
//	//     "your/import/path/stdlib"
//	// )
//	//
//	// // MyAdapter adapts application logging to the stdlib.Logger interface.
//	// type MyAdapter struct{}
//	//
//	// func (m *MyAdapter) Info(msg string, fields ...map[string]any) {
//	//     fmt.Println("INFO:", msg, fields)
//	// }
//	// func (m *MyAdapter) Warn(msg string, fields ...map[string]any) {
//	//     fmt.Println("WARN:", msg, fields)
//	// }
//	// func (m *MyAdapter) Error(msg string, fields ...map[string]any) {
//	//     fmt.Println("ERROR:", msg, fields)
//	// }
//	// func (m *MyAdapter) Debug(msg string, fields ...map[string]any) {
//	//     fmt.Println("DEBUG:", msg, fields)
//	// }
//	//
//	// func main() {
//	//     adapter := &MyAdapter{}
//	//     stdlib.SetLogger(adapter)
//	//     stdlib.Info("service started", map[string]any{"port": 8080})
//	// }
//
// The above shows a minimal adapter you can use when you want to route
// package-level logging to your application's logger implementation.
type Logger interface {
	Info(msg string, additionalFields ...map[string]any)
	Warn(msg string, additionalFields ...map[string]any)
	Error(msg string, additionalFields ...map[string]any)
	Debug(msg string, additionalFields ...map[string]any)
}

// logger is the package-level logger used by the helper functions. Install a
// custom logger with SetLogger. By default we initialize a sensible
// JSON logger so package helpers (Info/Warn/Error/Debug) are usable without
// explicit installation.
var logger Logger = NewJSONLogger()

// SetLogger installs a global Logger used by the package-level helpers. Call
// this from your application's bootstrap to route package-level logging calls
// to your logger implementation.
//
// Example (installing the provided JSON logger for package-level logging):
//
//	// jl := stdlib.NewJSONLogger()
//	// stdlib.SetLogger(jl)
//	// stdlib.Info("started", map[string]any{"env": "prod"})
//
// Use `SetLogger` during application initialization so other packages can
// call the package-level helpers without knowing the concrete logger.
func SetLogger(l Logger) {
	logger = l
}

// Package-level helper functions that forward to the installed logger.
// These allow consumers to call stdlib.Info(...), etc., without holding
// an explicit logger reference.
func Info(msg string, additionalFields ...map[string]any) {
	if logger == nil {
		return
	}
	logger.Info(msg, additionalFields...)
}

func Warn(msg string, additionalFields ...map[string]any) {
	if logger == nil {
		return
	}
	logger.Warn(msg, additionalFields...)
}

func Error(msg string, additionalFields ...map[string]any) {
	if logger == nil {
		return
	}
	logger.Error(msg, additionalFields...)
}

func Debug(msg string, additionalFields ...map[string]any) {
	if logger == nil {
		return
	}
	logger.Debug(msg, additionalFields...)
}

// Info logs a message at info level with optional additional fields.
// The fields are merged into the top-level JSON object for the entry.
//
// Example:
//
//	// jl := stdlib.NewJSONLogger()
//	// jl.Info("user created", map[string]any{"user_id": 123})
func (jl *JSONLogger) Info(msg string, additionalFields ...map[string]any) {
	jl.log(InfoLevel, "info", msg, additionalFields...)
}

// Warn logs a message at warn level with optional additional fields.
//
// Example:
//
//	// jl := stdlib.NewJSONLogger()
//	// jl.Warn("high memory usage", map[string]any{"heap_mb": 512})
func (jl *JSONLogger) Warn(msg string, additionalFields ...map[string]any) {
	jl.log(WarnLevel, "warn", msg, additionalFields...)
}

// Error logs a message at error level with optional additional fields.
//
// Example:
//
//	// jl := stdlib.NewJSONLogger()
//	// jl.Error("failed to connect to db", map[string]any{"db": "primary"})
func (jl *JSONLogger) Error(msg string, additionalFields ...map[string]any) {
	jl.log(ErrorLevel, "error", msg, additionalFields...)
}

// Debug logs a message at debug level with optional additional fields.
// Debug messages are filtered out unless the logger's level is set to Debug.
//
// Example (enable debug by creating the logger with DebugLevel):
//
//	// jl := stdlib.NewJSONLoggerWithOptions(WithLevel(DebugLevel))
//	// jl.Debug("cache miss", map[string]any{"key": "user:123"})
func (jl *JSONLogger) Debug(msg string, additionalFields ...map[string]any) {
	jl.log(DebugLevel, "debug", msg, additionalFields...)
}
