package golog

import (
	"bytes"
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
	writer     LogWriter
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
		writer:     NewJSONLogWriter(),
		timeFormat: time.RFC3339Nano,
		bufferPool: sync.Pool{
			New: func() any {
				// Pre-allocate with reasonable capacity to reduce allocations
				return bytes.NewBuffer(make([]byte, 0, 256))
			},
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

// WithLogWriter sets a custom LogWriter implementation for the logger.
// This allows you to customize the output format (e.g., pretty JSON, compact JSON).
func WithLogWriter(writer LogWriter) Option {
	return func(jl *JSONLogger) {
		if writer != nil {
			jl.writer = writer
		}
	}
}

// WithPrettyJSON configures the logger to use pretty-formatted JSON output.
func WithPrettyJSON(indent string) Option {
	return func(jl *JSONLogger) {
		jl.writer = NewPrettyJSONLogWriter(indent)
	}
}

// WithCompactJSON configures the logger to use compact JSON output.
func WithCompactJSON() Option {
	return func(jl *JSONLogger) {
		jl.writer = NewCompactJSONLogWriter()
	}
}

// log builds a JSON object from baseFields + message fields and writes it.
func (jl *JSONLogger) log(lv Level, levelStr, msg string, keyValuePairs ...map[string]any) {
	if lv < jl.level {
		return
	}

	buf := jl.bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	// Use configured time format
	tf := jl.timeFormat
	if tf == "" {
		tf = time.RFC3339Nano
	}

	// Use the LogWriter to write the log entry
	jl.writer.WriteLogEntry(buf, time.Now(), tf, levelStr, msg, jl.baseFields, keyValuePairs...)

	// Write to output with minimal locking
	jl.mutex.Lock()
	_, _ = jl.output.Write(buf.Bytes())
	jl.mutex.Unlock()
	jl.bufferPool.Put(buf)
}
