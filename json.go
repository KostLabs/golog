package golog

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type Level int

const (
	DebugLevel Level = 0 + iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// JSONLogger is a small, fast, concurrent-safe JSON logger implementation.
// Create one with NewJSONLogger or NewJSONLoggerWithOptions. Use the Option
// helpers to customize level, output and base fields.
type JSONLogger struct {
	output     io.Writer
	baseFields map[string]any
	level      Level
	mutex      sync.Mutex
	bufferPool sync.Pool
	// mapPool holds reusable maps for building log entries to reduce
	// allocations when creating temporary maps per log call.
	mapPool sync.Pool
	// timeFormat controls how timestamps are rendered. Defaults to
	// time.RFC3339Nano but can be changed with WithCustomTimeFormat.
	timeFormat string
}

// Option configures the JSONLogger.
type Option func(*JSONLogger)

// NewJSONLogger returns a logger with sensible defaults:
//   - Level: InfoLevel
//   - Output: os.Stdout
//   - No base fields
func NewJSONLogger() *JSONLogger {
	return &JSONLogger{
		output:     os.Stdout,
		baseFields: make(map[string]any),
		level:      InfoLevel,
		timeFormat: time.RFC3339Nano,
		bufferPool: sync.Pool{
			New: func() any { return &bytes.Buffer{} },
		},
		mapPool: sync.Pool{
			New: func() any { return make(map[string]any, 8) },
		},
	}
}

// NewJSONLoggerWithOptions creates a logger and applies functional options.
// Use the Option helpers WithLevel, WithOutput, WithBaseFields and
// WithBaseField to configure the logger.
func NewJSONLoggerWithOptions(opts ...Option) *JSONLogger {
	jl := NewJSONLogger()
	for _, opt := range opts {
		opt(jl)
	}

	return jl
}

// WithLevel sets the minimum level for the logger. Logs with lower severity
// than the configured level are dropped.
func WithLevel(l Level) Option {
	return func(jl *JSONLogger) { jl.level = l }
}

// WithOutput sets the writer for the logger (stdout, file, buffer, etc.).
func WithOutput(w io.Writer) Option {
	return func(jl *JSONLogger) { jl.output = w }
}

// WithBaseFields adds the provided fields to the logger's base fields. These
// fields are included in every emitted log entry.
func WithBaseFields(fields map[string]any) Option {
	return func(jl *JSONLogger) {
		for k, v := range fields {
			jl.baseFields[k] = v
		}
	}
}

// WithBaseField adds a single base field key/value that will be included in
// every log entry.
func WithBaseField(k string, v any) Option {
	return func(jl *JSONLogger) { jl.baseFields[k] = v }
}

// WithCustomTimeFormat sets a custom time format for the timestamp field.
// If not set, the logger uses RFC3339Nano.
func WithCustomTimeFormat(format string) Option {
	return func(jl *JSONLogger) {
		if format == "" {
			return
		}

		jl.timeFormat = format
	}
}

// log builds a JSON object from baseFields + message fields and writes it.
func (jl *JSONLogger) log(lv Level, levelStr, msg string, keyValuePairs ...map[string]any) {
	if lv < jl.level {
		return
	}

	// try to reuse a temporary map from the pool to avoid allocating a new
	// map for each log invocation. Clear it before use and put it back when
	// finished.
	m := jl.mapPool.Get().(map[string]any)
	// clear any previous contents
	for k := range m {
		delete(m, k)
	}
	// pre-size isn't required because pooled maps will retain capacity.
	for k, v := range jl.baseFields {
		m[k] = v
	}

	// use configured time format
	tf := jl.timeFormat
	if tf == "" {
		tf = time.RFC3339Nano
	}

	m["timestamp"] = time.Now().UTC().Format(tf)

	// Keep only a machine-friendly plain `level` field.
	m["level"] = levelStr
	m["message"] = msg

	// merge provided maps into the output map
	if len(keyValuePairs) > 0 {
		mergeMaps(m, keyValuePairs...)
	}

	buf := jl.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	// Try fast encoder first; if it cannot handle some type, fall back to
	// encoding/json which handles all cases.
	if FastEncode(buf, m) {
		// ensure newline suffix like json.Encoder.Encode does
		if buf.Len() == 0 || buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteByte('\n')
		}
	} else {
		if err := MarshalToBuffer(buf, m); err != nil {
			// if marshal fails (unsupported type), write a minimal JSON error
			jl.mutex.Lock()
			tf := jl.timeFormat
			if tf == "" {
				tf = time.RFC3339Nano
			}

			if _, writeErr := fmt.Fprintf(jl.output,
				"{\"timestamp\":\"%s\",\"level\":\"%s\",\"message\":\"%s\",\"error\":\"%s\"}\n",
				time.Now().UTC().Format(tf), levelStr, msg, err.Error()); writeErr != nil {
				// best-effort fallback: write write-errors to stderr
				fmt.Fprintln(os.Stderr, "json logger write error:", writeErr)
			}

			jl.mutex.Unlock()
			jl.bufferPool.Put(buf)
			return
		}

		// ensure newline suffix
		if buf.Len() == 0 || buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteByte('\n')
		}
	}

	jl.mutex.Lock()
	_, _ = jl.output.Write(buf.Bytes())
	jl.mutex.Unlock()
	jl.bufferPool.Put(buf)

	// clear and return the temporary map to the pool
	for k := range m {
		delete(m, k)
	}
	jl.mapPool.Put(m)
}
