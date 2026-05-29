package main

import (
	"errors"
	"os"

	"github.com/KostLabs/golog"
)

func main() {
	// -------------------------------------------------------------------------
	// 1. Simple Usage (Zero Configuration)
	// -------------------------------------------------------------------------
	golog.Info("service started")
	golog.Error("connection failed",
		golog.Str("host", "db.example.com"),
		golog.Int("retry_count", 3),
	)

	// -------------------------------------------------------------------------
	// 2. Basic Custom Logger — Debug level & explicit output
	// -------------------------------------------------------------------------
	basicLogger := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.DebugLevel),
		golog.WithOutput(os.Stdout),
	)
	golog.SetLogger(basicLogger)

	golog.Info("payment processed",
		golog.Int("user_id", 12345),
		golog.Float64("amount", 99.99),
		golog.Str("currency", "USD"),
	)

	// -------------------------------------------------------------------------
	// 3. Highly Customized Logger
	// -------------------------------------------------------------------------
	customLogger := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.DebugLevel),
		golog.WithOutput(os.Stderr),
		golog.WithBaseField("service", "payment-api"),
		golog.WithBaseField("version", "1.2.3"),
		golog.WithCustomTimeFormat("2006-01-02 15:04:05"),
	)
	golog.SetLogger(customLogger)

	golog.Info("payment processed",
		golog.Int("user_id", 12345),
		golog.Float64("amount", 99.99),
		golog.Str("currency", "USD"),
	)

	// -------------------------------------------------------------------------
	// 4. Instance-Based Logging
	// -------------------------------------------------------------------------
	authLogger := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.DebugLevel),
		golog.WithOutput(os.Stdout),
		golog.WithBaseField("component", "auth"),
	)

	authLogger.Info("authentication successful",
		golog.Int("user_id", 12345),
		golog.Str("method", "oauth2"),
	)

	// -------------------------------------------------------------------------
	// 5. Package-Level Helpers
	// -------------------------------------------------------------------------
	appLogger := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.DebugLevel),
		golog.WithOutput(os.Stdout),
	)
	golog.SetLogger(appLogger)

	golog.Info("server starting")
	golog.Debug("cache hit", golog.Str("key", "user:123"))
	golog.Warn("high memory usage", golog.Int("heap_mb", 512))

	dbErr := errors.New("dial tcp: connection refused")
	golog.Error("database error", golog.Str("error", dbErr.Error()))

	// -------------------------------------------------------------------------
	// 6. All four log levels on an instance logger
	// -------------------------------------------------------------------------
	logger := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.DebugLevel),
		golog.WithOutput(os.Stdout),
		golog.WithBaseField("app", "example"),
	)

	logger.Debug("debug details", golog.Str("trace_id", "abc-123"))
	logger.Info("request received", golog.Str("path", "/api/pay"), golog.Str("method", "POST"))
	logger.Warn("slow query", golog.Int("duration_ms", 850), golog.Str("query", "SELECT *"))
	logger.Error("upstream timeout", golog.Str("service", "fraud-check"), golog.Int("timeout_ms", 3000))
}
