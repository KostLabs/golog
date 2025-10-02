# golog

`golog` is lightweight, high-performance, **zero-dependency** Go logging framework designed for structured JSON logging for modern cloud-native systems.

## Problem

Most of the famour Go loggers either have complex API's and poor performance (Logrus, Apex) or are very fast but have clunky API's and/or external dependencies (Zap, Zerolog). Some of them are not regularly updated or maintained. The ones which are maintained have updates available only in master branch and not in tagged releases, which makes them insecure to use in production.

## Solution

The solution is `golog`. An intuitive, developer-friendly structured logger with minimal configuration and zero dependnencies. It is designed to be fast, efficient and easily maintainable. If maintainer will stop maintaining it, it will be easy for community to fork and continue the maintenance.

Additionally it offers human-friendly API, flexible output formats (standard, pretty, compact JSON), built-in key normalization and error handling.

## Benefits

Main benefits in front of most popular loggers like Zap, Zerolog, Slog, Logrus and Apex Log are:

- **Zero dependencies** - no external packages required
- **Developer-Friendly API** - no need for wrappers or chaining
- **Built-in key normalization** - handles common key formatting issues
- **Built-in error handling** - automatic fallbacks for unsupported types
- **Multiple output formats** - standard JSON, pretty JSON, compact JSON

## Benchmarks

The main goal of `golog` is to provide high-performance logging framework, while the API remains simple and developer-friendly. Below are the benchmarks comparing `golog` with other popular loggers.

```
go test -bench . -benchmem -run='^$'
goos: darwin
goarch: arm64
pkg: benchmarks
cpu: Apple M3 Pro
BenchmarkGologInfo-11                    7844247               145.1 ns/op            32 B/op          1 allocs/op
BenchmarkSlogInfo-11                     5131329               233.2 ns/op            48 B/op          3 allocs/op
BenchmarkZerologInfo-11                 45882885                24.98 ns/op            0 B/op          0 allocs/op
BenchmarkZapInfo-11                     20638976                59.24 ns/op           32 B/op          1 allocs/op
BenchmarkApexInfo-11                     1584324               756.4 ns/op           216 B/op          4 allocs/op
BenchmarkLogrusInfo-11                    816307              1384 ns/op             871 B/op         19 allocs/op
BenchmarkGologInfoWithFields-11          4954791               250.2 ns/op           381 B/op          5 allocs/op
BenchmarkSlogInfoWithFields-11           4479103               265.0 ns/op            48 B/op          3 allocs/op
BenchmarkZerologInfoWithFields-11       38103000                30.04 ns/op            0 B/op          0 allocs/op
BenchmarkZapInfoWithFields-11            9110072               131.4 ns/op           225 B/op          2 allocs/op
BenchmarkApexInfoWithFields-11            715658              1551 ns/op            1185 B/op         16 allocs/op
BenchmarkLogrusInfoWithFields-11          466314              2500 ns/op            1818 B/op         29 allocs/op
BenchmarkAllLoggersSimple/Golog-11       8325763               139.2 ns/op            32 B/op          1 allocs/op
BenchmarkAllLoggersSimple/Slog-11        5184996               232.5 ns/op            48 B/op          3 allocs/op
BenchmarkAllLoggersSimple/Zerolog-11    38801517                28.28 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersSimple/Zap-11        20168942                60.25 ns/op           32 B/op          1 allocs/op
BenchmarkAllLoggersSimple/Apex-11        1541712               788.2 ns/op           216 B/op          4 allocs/op
BenchmarkAllLoggersSimple/Logrus-11       820881              1400 ns/op             871 B/op         19 allocs/op
BenchmarkAllLoggersWithFields/Golog-11           4722756               260.1 ns/op           381 B/op          5 allocs/op
BenchmarkAllLoggersWithFields/Slog-11            4476350               267.3 ns/op            48 B/op          3 allocs/op
BenchmarkAllLoggersWithFields/Zerolog-11        33964786                34.67 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersWithFields/Zap-11             8598110               140.3 ns/op           225 B/op          2 allocs/op
BenchmarkAllLoggersWithFields/Apex-11             704889              1554 ns/op            1185 B/op         16 allocs/op
BenchmarkAllLoggersWithFields/Logrus-11           462268              2512 ns/op            1819 B/op         29 allocs/op
```

Run the following commands to see the benchmarks:

```bash
# Clone the repository
git clone https://github.com/KostLabs/golog
cd golog/benchmarks

# Run comparative benchmarks
go test -bench=BenchmarkAllLoggers -benchmem -run='^$'
```

## Usage

### Simple Usage (Zero Configuration)

```go
package main

import "github.com/KostLabs/golog"

func main() {
	type fields map[string]any // optional type alias for convenience

	// Works immediately - no setup required!
	golog.Info("service started", fields{"port": 8080})
	golog.Error("connection failed", fields{
		"host":       "db.example.com",
		"retry_count": 3,
	})
}
```

### Advanced Configuration

`golog` offers advanced configuration options with methods like `WithLevel`, `WithOutput`, `WithBaseField`, `WithPrettyJSON`, and `WithCustomTimeFormat`. The methods can be chained for a fluent configuration experience.

#### Example - Basic Custom Logger with Debug Log Level & Custom Output

```go
package main

import (
    "os"
    "github.com/KostLabs/golog"
)

func main() {
    // Create a basic custom logger
    logger := golog.NewJSONLoggerWithOptions(
        golog.WithLevel(golog.DebugLevel),
        golog.WithOutput(os.Stdout),
    )

    // Set as global logger
    golog.SetLogger(logger)

    // Now all package-level calls use your configuration
    golog.Info("payment processed", map[string]any{
        "user_id": 12345,
        "amount": 99.99,
        "currency": "USD",
    })
}
```

#### Example - Highly Customized Logger

```go
package main

import (
    "os"
    "github.com/KostLabs/golog"
)

func main() {
    // Create a highly customized logger
    logger := golog.NewJSONLoggerWithOptions(
        golog.WithLevel(golog.DebugLevel),
        golog.WithOutput(os.Stderr),
        golog.WithBaseField("service", "payment-api"),
        golog.WithBaseField("version", "1.2.3"),
        golog.WithPrettyJSON("  "), // Pretty formatting
        golog.WithCustomTimeFormat("2006-01-02 15:04:05"),
    )
    
    // Set as global logger
    golog.SetLogger(logger)
    
    // Now all package-level calls use your configuration
    golog.Info("payment processed", map[string]any{
        "user_id": 12345,
        "amount": 99.99,
        "currency": "USD",
    })
}
```

### Usage Patterns

#### Instance-Based Logging
```go
logger := golog.NewJSONLoggerWithOptions(
    golog.WithLevel(golog.DebugLevel),
    golog.WithBaseField("component", "auth"),
)

logger.Info("authentication successful", map[string]any{
    "user_id": 12345,
    "method": "oauth2",
})
```

#### Package-Level Helpers
```go
// Configure once
golog.SetLogger(myCustomLogger)

// Use anywhere in your application
golog.Info("server starting")
golog.Debug("cache hit", map[string]any{"key": "user:123"})
golog.Error("database error", map[string]any{"error": err.Error()})
```

## Migration Guide

Migrating from other popular Go loggers to `golog` is straightforward due to its simple and intuitive API. Below are examples of how to translate common logging patterns from Logrus, Zap, and Slog to `golog`.

### From Logrus
```go
// Before (Logrus)
logger := logrus.New()
logger.WithFields(logrus.Fields{"key": "value"}).Info("message")

// After (golog)
logger := golog.NewJSONLogger()
logger.Info("message", map[string]any{"key": "value"})
```

### From Zap
```go
// Before (Zap)
logger.Info("message", zap.String("key", "value"), zap.Int("count", 42))

// After (golog)
logger.Info("message", map[string]any{"key": "value", "count": 42})
```

### From Slog
```go
// Before (Slog)
slog.Info("message", "key", "value", "count", 42)

// After (golog)
golog.Info("message", map[string]any{"key": "value", "count": 42})
```
