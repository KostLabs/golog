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
//   - Field API: typed fields (Str/Int/Bool/...) for higher-throughput
//     logging via Info/Warn/Error/Debug.
//
// Logger interface
// The package exposes a minimal Logger interface defined in `logger.go`:
//
//	type Logger interface {
//	    Info(message string, fields ...Field)
//	    Warn(message string, fields ...Field)
//	    Error(message string, fields ...Field)
//	    Debug(message string, fields ...Field)
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
//   - per-call typed fields: zero or more Field values passed to
//     Info/Warn/Error/Debug; later fields with the same key override earlier
//     ones.
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
//   - WithWriteLock(bool)         : enable/disable output write lock
//   - WithBaseFields(map[string]any) : add a set of base fields
//   - WithBaseField(key, value)  : add a single base field
//
// Logging calls
// Pass zero or more typed fields. Each field is merged into the top-level JSON
// object. Example:
//
//	jl.Info("user created", Str("userID", "1234"), Str("role", "admin"))
//
// For performance-sensitive code paths use typed fields:
//
//	jl.Info("user login", Int("user_id", 42), Str("ip", "127.0.0.1"), Bool("success", true))
//
// Concurrency and performance notes
//   - Writes are protected by an internal mutex so each encoded JSON line is
//     written atomically. This prevents interleaving when multiple goroutines
//     call log methods concurrently.
//   - A sync.Pool of reusable []byte buffers is used to avoid fresh allocations
//     on every log call.
//   - Level filtering uses atomic reads/writes, so suppressed log calls are
//     cheap and runtime level changes are race-safe.
//   - Output writes are lock-protected by default for safe serialized writes.
//     You can disable locking with WithWriteLock(false) when writing to a
//     thread-safe sink and optimizing for throughput.
//
// Unsupported values
// If a field value can't be encoded by the fast encoder (for example a channel),
// golog writes "<unsupported>" for that field value and continues encoding the
// rest of the log entry.
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
//	Info("started", Int("pid", os.Getpid()))
//
// The package-level `Info`, `Warn`, `Error`, `Debug` functions call the
// installed global logger if one is set; otherwise they are no-ops. This makes
// it easy to instrument libraries safely and install a logger in your
// application bootstrap.
package golog
