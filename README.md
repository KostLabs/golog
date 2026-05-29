# golog

`golog` is lightweight, high-performance, **zero-dependency** Go logging framework designed for structured JSON logging for modern cloud-native systems.

## Problem

Most of the famous Go loggers either have complex API's and poor performance (Logrus, Apex) or are very fast but have clunky API's and/or external dependencies (Zap, Zerolog). Some of them are not regularly updated or maintained. The ones which are maintained have updates available only in master branch and not in tagged releases, which makes them insecure to use in production.

## Solution

The solution is `golog`. An intuitive, developer-friendly structured logger with minimal configuration and zero dependencies. It is designed to be fast, efficient and easily maintainable. If maintainer will stop maintaining it, it will be easy for community to fork and continue the maintenance.

## Benefits

Main benefits in front of most popular loggers like Zap, Zerolog, Slog, Logrus and Apex Log are:

- **Zero dependencies** - no external packages required
- **Developer-Friendly API** - no need for wrappers or chaining
- **Built-in key normalization** - handles common key formatting issues
- **Built-in error handling** - automatic fallbacks for unsupported types

## Benchmarks

The main goal of `golog` is to provide high-performance logging framework, while the API remains simple and developer-friendly. Below are the benchmarks comparing `golog` with other popular loggers.

```
goos: darwin
goarch: arm64
pkg: benchmarks
cpu: Apple M3 Pro
BenchmarkAllLoggersSimple/Golog-12              23851167                49.71 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersSimple/Slog-12                6871609               175.9 ns/op            48 B/op          3 allocs/op
BenchmarkAllLoggersSimple/Zerolog-12            79315244                23.59 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersSimple/Zap-12                27729715                43.80 ns/op           32 B/op          1 allocs/op
BenchmarkAllLoggersSimple/Apex-12                1851841               646.4 ns/op           216 B/op          4 allocs/op
BenchmarkAllLoggersSimple/Logrus-12              1000000              1163 ns/op             872 B/op         19 allocs/op
BenchmarkAllLoggersWithFields/Golog-12          47756599                29.59 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersWithFields/Slog-12            5805918               210.5 ns/op            48 B/op          3 allocs/op
BenchmarkAllLoggersWithFields/Zerolog-12        40578582                44.23 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersWithFields/Zap-12            19405268                63.27 ns/op           32 B/op          1 allocs/op
BenchmarkAllLoggersWithFields/Apex-12             781440              1479 ns/op             993 B/op         18 allocs/op
BenchmarkAllLoggersWithFields/Logrus-12           474498              2461 ns/op            2293 B/op         35 allocs/op
BenchmarkAllLoggersWithLargeFields/Golog-12     31537002                37.50 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersWithLargeFields/Slog-12       3384866               356.7 ns/op           369 B/op          4 allocs/op
BenchmarkAllLoggersWithLargeFields/Zerolog-12   35600738                33.32 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersWithLargeFields/Zap-12       13241354                93.57 ns/op           32 B/op          1 allocs/op
BenchmarkAllLoggersWithLargeFields/Apex-12        379323              3256 ns/op            2236 B/op         37 allocs/op
BenchmarkAllLoggersWithLargeFields/Logrus-12      248450              4815 ns/op            4175 B/op         55 allocs/op
BenchmarkAllLoggersWithExtraLargeFields/Golog-12                21859084                56.06 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersWithExtraLargeFields/Slog-12                  2231878               532.8 ns/op           701 B/op          5 allocs/op
BenchmarkAllLoggersWithExtraLargeFields/Zerolog-12              26186888                45.58 ns/op            0 B/op          0 allocs/op
BenchmarkAllLoggersWithExtraLargeFields/Zap-12                  11174211               109.0 ns/op            32 B/op          1 allocs/op
BenchmarkAllLoggersWithExtraLargeFields/Apex-12                   222519              5367 ns/op            3968 B/op         53 allocs/op
BenchmarkAllLoggersWithExtraLargeFields/Logrus-12                 175506              6877 ns/op            5878 B/op         69 allocs/op
PASS
ok      benchmarks      33.504s
```

Run the following commands to see the benchmarks:

```bash
# Clone the repository
git clone https://github.com/KostLabs/golog
cd golog/benchmarks

# CPU benchmarks
go test -bench '^BenchmarkCPU' -run='^$'

# Memory benchmarks
go test -bench '^BenchmarkMemory' -benchmem -run='^$'

# Full suite (CPU + memory)
go test -bench . -run='^$'
```

The visualized results can be visible on [GitHub Pages](https://kostlabs.github.io/golog/).

## Usage

### Installation

```bash
go get github.com/KostLabs/golog
```

### Simple Usage (Zero Configuration)

```go
package main

import "github.com/KostLabs/golog"

func main() {
	// Optional type alias for convenience
	type fields map[string]any

	golog.Info("service started")
	golog.Error("connection failed", fields{
		"host":       "db.example.com",
		"retry_count": 3,
	})
}
```

### Advanced Configuration

`golog` offers advanced configuration options with methods like `WithLevel`, `WithOutput`, `WithBaseField`, `WithBaseFields` and `WithCustomTimeFormat`. The methods can be chained for a fluent configuration experience.

#### Example - Basic Custom Logger with Debug Log Level & Custom Output

```go
package main

import (
    "os"
    "github.com/KostLabs/golog"
)

func main() {
	// Optional type alias for convenience
	type fields map[string]any

    logger := golog.NewJSONLoggerWithOptions(
        golog.WithLevel(golog.DebugLevel),
        golog.WithOutput(os.Stdout),
    )

    golog.SetLogger(logger)
    golog.Info("payment processed", fields{
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
	// Optional type alias for convenience
	type fields map[string]any
    logger := golog.NewJSONLoggerWithOptions(
        golog.WithLevel(golog.DebugLevel),
        golog.WithOutput(os.Stderr),
        golog.WithBaseFields(
			fields{
				"service": "payment-api",
				"version": "1.2.3",
			},
		),
        golog.WithCustomTimeFormat("2006-01-02 15:04:05"),
    )
    
    // Set as global logger
    golog.SetLogger(logger)
    
    // Now all package-level calls use your configuration
    golog.Info("payment processed", fields{
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
