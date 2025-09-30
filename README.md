# golog

`golog` is a small, efficient, structured JSON logger for Go programs. It's
designed to be easy to integrate into applications: create a logger, add a
couple of base fields and use package-level helpers (or keep a logger
reference). The library favors low allocations and provides a fast internal
encoder with a safe fallback.

This repository provides:

- A compact JSON logger implementation (`JSONLogger`).
- A minimal `Logger` interface and package-level helpers (`Info`, `Warn`,
	`Error`, `Debug`) so your code can call logging without holding a logger
	reference.
- A small, fast encoder (`coder.go`) which writes JSON directly into pooled
	buffers for common shapes and falls back to a reflect-based marshaler for
	complex types.

The project aims for predictable performance in hot paths while keeping the
API ergonomic.

## Quick start

The package exposes helpers (`golog.Info`, `golog.Debug`, etc.) for callers
that don't want to hold a logger reference. For convenience the package
installs a default `JSONLogger` on import, so package-level helpers work
out-of-the-box and write to stdout using sensible defaults. If you need to
customize behavior, create a logger and call `SetLogger` to replace the
default.

### JSON Logger

`JSONLogger` can be used either as a standalone instance (call methods on the
returned logger) or installed as the package-level logger so package helpers
(`golog.Info`, `golog.Debug`, etc.) forward to it. The examples below show
both patterns.

Defaults for `NewJSONLogger`:
- Info log level
- Output to `os.Stdout`
- No base fields
- Timestamp in RFC3339Nano format

Example — use package-level helpers with defaults (no setup):

```go
package main

import (
	"github.com/KostLabs/golog"
)

func main() {
	// Optional: define details with map[string]any type to avoid repeating
	// the type in every log
	type details map[string]any

	// package-level helpers use a sensible default JSON logger, so this
	// works without explicit setup.
	golog.Info("service started", details{"port": 8080})
}
```

`NewJSONLoggerWithOptions` allows to define the options you would like to override.

- `WithLevel` - set minimum log level (debug, info, warn, error).
- `WithOutput` - change the destination (stdout, file, buffer).
- `WithBaseFields` / `WithBaseField` - fields that are included on every log
	entry.
- `WithCustomTimeFormat` - change timestamp formatting for the `timestamp`
	field (defaults to RFC3339Nano).


Example — create a configured logger with self-defined options and install it
as the repo-level logger:

```go
package main

import (
	"io"
	"github.com/KostLabs/golog"
)

func main() {
	type details map[string]any

	jl := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.DebugLevel),
		golog.WithOutput(io.Discard),
		golog.WithBaseField("service", "payments"),
	)
	golog.SetLogger(jl)

	// Now package-level helpers forward to the installed logger
	golog.Info("service started", details{"port": 8080})
	golog.Debug("cache miss", details{"key": "user:123"})
}
```

**API note**: you can pass zero or more `map[string]any` maps to the logging
methods; they are merged into the emitted JSON object (later maps override
earlier ones).

## Logger interface

The package exposes a minimal `Logger` interface so your application can
adapt its own logger implementation or use the provided one:

```go
type Logger interface {
		Info(msg string, additionalFields ...map[string]any)
		Warn(msg string, additionalFields ...map[string]any)
		Error(msg string, additionalFields ...map[string]any)
		Debug(msg string, additionalFields ...map[string]any)
}
```

Use `SetLogger` to install a global logger that the package-level helpers
forward to.

## Options

Create loggers with `NewJSONLoggerWithOptions(...)` and configure:

- `WithLevel(Level)` — set minimum log level (Debug, Info, Warn, Error).
- `WithOutput(io.Writer)` — change the destination (stdout, file, buffer).
- `WithBaseFields(map[string]any)` / `WithBaseField(k, v)` — fields that
	are included on every log entry.
- `WithCustomTimeFormat(format string)` — change timestamp formatting for
	the `timestamp` field (defaults to RFC3339Nano).

Example:

```go
jl := NewJSONLoggerWithOptions(
		WithLevel(DebugLevel),
		WithOutput(os.Stderr),
		WithBaseField("service", "payments"),
)
```

## Implementation notes

- Buffer pooling: the logger uses a `sync.Pool` of `bytes.Buffer` to avoid
	repeated allocations for temporary buffers.
- Map pooling: temporary maps used to merge base fields and provided maps are
	also pooled to reduce allocations.
- Fast encoder: `encoder.go` provides a fast, reflection-free encoder for
	common types (strings, numbers, bool, time.Time, `map[string]any` and
	`[]any`). The logger tries the fast encoder first and falls back to a
	reflect-based marshaler (provided in `marshal.go`) for values that the
	fast path doesn't support. This gives predictable performance while
	remaining correct for arbitrary types.

## Benchmarks

Benchmarks were run on macOS/arm64 (Apple M3 Pro). These are representative
results from repository benchmarks; your hardware and Go version will
influence numbers.

Benchmark summary (directed at typical structured log calls):

```
go test -bench . -benchmem -run '^$'
goos: darwin
goarch: arm64
pkg: github.com/KostLabs/golog
cpu: Apple M3 Pro
BenchmarkDefaultLogger/Direct.Info-11    1686820               709.2 ns/op            96 B/op          5 allocs/op
BenchmarkDefaultLogger/Package.Info-11           1502918               798.1 ns/op           440 B/op          9 allocs/op
BenchmarkDefaultLogger/Direct.MergeTwoMaps-11    1485727               809.9 ns/op           112 B/op          7 allocs/op
BenchmarkDebugLevelLogger/Direct.Info-11         1711009               701.3 ns/op            96 B/op          5 allocs/op
BenchmarkDebugLevelLogger/Package.Info-11        1504086               798.1 ns/op           440 B/op          9 allocs/op
BenchmarkDebugLevelLogger/Direct.MergeTwoMaps-11                 1479944               809.6 ns/op           112 B/op       7 allocs/op
BenchmarkWithBaseFieldsLogger/Direct.Info-11                     1411914               850.0 ns/op            96 B/op       5 allocs/op
BenchmarkWithBaseFieldsLogger/Package.Info-11                    1267008               947.5 ns/op           440 B/op       9 allocs/op
BenchmarkWithBaseFieldsLogger/Direct.MergeTwoMaps-11             1253274               958.7 ns/op           112 B/op       7 allocs/op
BenchmarkCustomTimeFormatLogger/Direct.Info-11                   1491144               802.8 ns/op            96 B/op       5 allocs/op
BenchmarkCustomTimeFormatLogger/Package.Info-11                  1337048               898.4 ns/op           440 B/op       9 allocs/op
BenchmarkCustomTimeFormatLogger/Direct.MergeTwoMaps-11           1318813               908.7 ns/op           112 B/op       7 allocs/op
PASS
ok      github.com/KostLabs/golog       16.440s
```

How to reproduce locally:

```bash
cd /path/to/loggerv2
go test -bench . -benchmem -run '^$'
```

## Contributing

Patches welcome. Add tests for any new encoder behavior and keep benchmark
changes isolated so we can compare before/after.

## License

See repository LICENSE (if present).

