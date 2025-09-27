// package golog provides a small, fast JSON logger implementation and a
// minimal Logger interface suitable for replacing or wiring into standard
// library components.
//
// The design goals are:
//   - easy to initialize in one or two lines
//   - structured (JSON) output with stable fields
//   - low-allocation and concurrent-safe for common use cases
//
// Core concepts
//   - Logger interface: a tiny interface with Info/Warn/Error/Debug methods.
//     This allows you to pass the logger around or install a global logger via
//     SetLogger.
//   - JSONLogger: a ready-to-use implementation that writes newline-delimited
//     JSON objects. Create it with sensible defaults using NewJSONLogger or
//     customize via NewJSONLoggerWithOptions and Option helpers.
//
// Logger interface
// The package exposes a minimal Logger interface defined in `logger.go`:
//
//	type Logger interface {
//	    Info(msg string, additionalFields ...map[string]any)
//	    Warn(msg string, additionalFields ...map[string]any)
//	    Error(msg string, additionalFields ...map[string]any)
//	    Debug(msg string, additionalFields ...map[string]any)
//	}
//
// Use `SetLogger(l Logger)` to install a logger globally that adapter code can
// depend on.
//
// JSONLogger (usage)
// The JSON logger writes one JSON object per log call. Each object always
// contains the following core fields:
//   - timestamp: generated per entry in RFC3339Nano UTC format
//   - level: one of "debug", "info", "warn", "error"
//   - message: the string passed to the logging method
//
// In addition it merges:
//   - base fields: a map of fields attached to the logger at construction time
//     (for example, application name, environment, service id)
//   - per-call additional maps: zero or more `map[string]any` passed as the
//     final variadic parameter to Info/Warn/Error/Debug; later maps override
//     earlier values.
//
// Creating a logger
//
//	// default: writes to stdout with Info level
//	jl := NewJSONLogger()
//
// Customizing via options
//
//	jl := NewJSONLoggerWithOptions(
//	    WithLevel(DebugLevel),
//	    WithOutput(os.Stderr),
//	    WithBaseFields(map[string]any{"app": "api", "env": "prod"}),
//	)
//
// Convenience option helpers
//   - WithLevel(Level)           : set minimum log level (Debug/Info/Warn/Error)
//   - WithOutput(io.Writer)      : set writer (stdout, file, buffer)
//   - WithBaseFields(map[string]any) : add a set of base fields
//   - WithBaseField(key, value)  : add a single base field
//
// Logging calls
// Pass zero or more maps of additional fields. Each map is merged into the
// top-level JSON object (keys are normalized by trimming spaces, removing a
// trailing ':' and stripping surrounding quotes). Example:
//
//	jl.Info("user created", map[string]any{"userID": "1234"})
//
// Or pass multiple maps (later maps override earlier ones):
//
//	jl.Warn("disk low",
//	    map[string]any{"disk": "/dev/sda1"},
//	    map[string]any{"disk": "/dev/sda1", "free": 1024},
//	)
//
// Concurrency and performance notes
//   - Writes are protected by an internal mutex so each encoded JSON line is
//     written atomically. This prevents interleaving when multiple goroutines
//     call log methods concurrently.
//   - A sync.Pool of *bytes.Buffer is used to avoid allocating a fresh buffer
//     on every log call which reduces GC pressure in hot paths.
//   - Level filtering is a fast integer compare (if lv < jl.level { return }),
//     so suppressed log calls are cheap.
//   - If you need to change log level at runtime without races, consider
//     using an atomic-based helper (not provided by default) or ensure callers
//     set the level before concurrent usage.
//
// JSON encoding fallback
// If a value in the merged map can't be marshaled by encoding/json (for
// example a channel), the logger performs a best-effort fallback and writes a
// minimal JSON object containing the timestamp, level, message and an
// `error` field describing the marshal problem.
//
// Testing
// The package includes small tests that demonstrate expected behaviour
// (level filtering, merging maps and the fallback path). In tests you can
// capture output by constructing a logger with `WithOutput(bytes.NewBuffer(nil))`.
//
// Extending
//   - If you need maximum throughput and can tolerate eventual consistency,
//     consider using a single writer goroutine and a channel for log events.
//     That removes the per-line mutex at the cost of buffering and complexity.
//   - For extremely high performance JSON encoding, replace encoding/json with
//     a faster codec (bench first; encoding/json is perfectly fine for most
//     applications).
//
// Example: install global logger and use convenience wrappers
//
//	jl := NewJSONLoggerWithOptions(WithLevel(InfoLevel), WithBaseField("app", "svc"))
//	SetLogger(jl)
//
//	// elsewhere in code
//	Info("started", map[string]any{"pid": os.Getpid()})
//
// The package-level `Info`, `Warn`, `Error`, `Debug` functions call the
// installed global logger if one is set; otherwise they are no-ops. This makes
// it easy to instrument libraries safely and install a logger in your
// application bootstrap.
package golog
