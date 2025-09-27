# GoLogger

GoLogger is a small, efficient, structured JSON logger for Go programs. It's
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

Create and use a `JSONLogger` directly:

```go
package main

import (
		"io"
		"os"

		"github.com/KostLabs/golog"
)

func main() {
		// Simple logger with defaults (Info level, os.Stdout)
		jl := golog.NewJSONLogger()

		// Write an entry directly
		jl.Info("user created", map[string]any{"user_id": 123})

		// Create a logger with options (debug level, discard output, base fields)
		jl2 := golog.NewJSONLoggerWithOptions(
				golog.WithLevel(golog.DebugLevel),
				golog.WithOutput(io.Discard),
				golog.WithBaseFields(map[string]any{"app": "myapp", "env": "dev"}),
		)

		jl2.Debug("cache miss", map[string]any{"key": "user:123"})

		// Install a package-level logger so other packages can call golog.Info(...)
		golog.SetLogger(jl2)
		golog.Info("service started", map[string]any{"port": 8080})
}
```

API note: you can pass zero or more `map[string]any` maps to the logging
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
- Fast encoder: `coder.go` provides a fast, reflection-free encoder for
	common types (strings, numbers, bool, time.Time, `map[string]any` and
	`[]any`). The logger tries the fast encoder first and falls back to a
	reflect-based marshaler (also provided in `coder.go`) for values that the
	fast path doesn't support. This gives predictable performance while
	remaining correct for arbitrary types.

## Benchmarks (representative)

Benchmarks were run on macOS/arm64 (Apple M3 Pro). These are representative
results from repository benchmarks; your hardware and Go version will
influence numbers.

Benchmark summary (directed at typical structured log calls):

| Benchmark | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| DefaultLogger / Direct.Info | ~984 ns/op | 144 B/op | 6 allocs/op |
| DefaultLogger / Package.Info | ~1064 ns/op | 488 B/op | 10 allocs/op |
| DefaultLogger / Direct.MergeTwoMaps | ~1172 ns/op | 192 B/op | 12 allocs/op |

How to reproduce locally:

```bash
cd /path/to/loggerv2
go test -bench . -benchmem -run '^$'
```

For CPU/memory profiles you can build the test binary and run it with
profiling flags:

```bash
go test -c -o golog.test
./golog.test -test.bench . -test.benchmem -test.run '^$' -test.memprofile=mem.prof -test.cpuprofile=cpu.prof
go tool pprof -top ./golog.test mem.prof
```

## When to use the fast encoder or tune further

- If your log values are mostly primitive scalars (strings, ints, booleans)
	and small maps, the fast encoder provides the best throughput.
- If you log complex, heavily nested or arbitrary types frequently, the
	reflect fallback may still dominate CPU/allocs. In that case, prefer
	logging well-known scalar fields or structs (or add MarshalJSON on types
	you control) to reduce reflection cost.

Suggested next steps if you need more speed:

- Avoid logging large or deep structures directly.
- Emit typed, compact fields instead of large maps when possible.
- Consider rendering timestamps as integers (UnixNano) if string formatting
	overhead matters.

## Contributing

Patches welcome. Add tests for any new encoder behavior and keep benchmark
changes isolated so we can compare before/after.

## License

See repository LICENSE (if present).

