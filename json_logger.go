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
func NewJSONLoggerWithOptions(options ...Option) *JSONLogger {
	jsonLogger := NewJSONLogger()
	for _, option := range options {
		option(jsonLogger)
	}

	return jsonLogger
}

// WithLevel sets the minimum level for the logger. Logs with lower severity
// than the configured level are dropped.
func WithLevel(logLevel Level) Option {
	return func(jsonLogger *JSONLogger) { jsonLogger.level = logLevel }
}

// WithOutput sets the writer for the logger (stdout, file, buffer, etc.).
func WithOutput(writer io.Writer) Option {
	return func(jsonLogger *JSONLogger) { jsonLogger.output = writer }
}

// WithBaseFields adds the provided fields to the logger's base fields. These
// fields are included in every emitted log entry.
func WithBaseFields(fields map[string]any) Option {
	return func(jsonLogger *JSONLogger) {
		for key, value := range fields {
			jsonLogger.baseFields[key] = value
		}
	}
}

// WithBaseField adds a single base field key/value that will be included in
// every log entry.
func WithBaseField(key string, value any) Option {
	return func(jsonLogger *JSONLogger) { jsonLogger.baseFields[key] = value }
}

// WithCustomTimeFormat sets a custom time format for the timestamp field.
// If not set, the logger uses RFC3339Nano.
func WithCustomTimeFormat(timeFormat string) Option {
	return func(jsonLogger *JSONLogger) {
		if timeFormat == "" {
			return
		}

		jsonLogger.timeFormat = timeFormat
	}
}

// log builds a JSON object from baseFields + message fields and writes it.
func (jsonLogger *JSONLogger) log(logLevel Level, levelString, message string, keyValuePairs ...map[string]any) {
	if logLevel < jsonLogger.level {
		return
	}

	buffer := jsonLogger.bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()

	// Use configured time format
	timeFormat := jsonLogger.timeFormat
	if timeFormat == "" {
		timeFormat = time.RFC3339Nano
	}

	buffer.WriteByte('{')

	// Write timestamp
	buffer.WriteString(`"timestamp":"`)
	buffer.WriteString(time.Now().UTC().Format(timeFormat))
	buffer.WriteByte('"')

	// Write level
	buffer.WriteString(`,"level":"`)
	buffer.WriteString(levelString)
	buffer.WriteByte('"')

	// Write message
	buffer.WriteString(`,"message":`)
	fastQuote(buffer, message)

	// Write base fields directly (optimized)
	for fieldKey, fieldValue := range jsonLogger.baseFields {
		buffer.WriteByte(',')
		fastQuote(buffer, fieldKey)
		buffer.WriteByte(':')
		if !encodeValue(buffer, fieldValue) {
			fastQuote(buffer, "<unsupported>")
		}
	}

	// Write additional fields directly (optimized)
	for _, keyValueMap := range keyValuePairs {
		if keyValueMap == nil {
			continue
		}
		for fieldKey, fieldValue := range keyValueMap {
			buffer.WriteByte(',')
			if len(fieldKey) > 0 && (fieldKey[0] == '"' || fieldKey[0] == '\'' || fieldKey[len(fieldKey)-1] == ':') {
				fastQuote(buffer, normalizeKeyInline(fieldKey))
			} else {
				fastQuote(buffer, fieldKey)
			}

			buffer.WriteByte(':')
			if !encodeValue(buffer, fieldValue) {
				fastQuote(buffer, "<unsupported>")
			}
		}
	}

	buffer.WriteByte('}')
	buffer.WriteByte('\n')

	// Write with minimal locking
	jsonLogger.mutex.Lock()
	_, _ = jsonLogger.output.Write(buffer.Bytes())
	jsonLogger.mutex.Unlock()
	jsonLogger.bufferPool.Put(buffer)
}

// normalizeKeyInline performs key normalization without allocation when possible
func normalizeKeyInline(keyString string) string {
	if len(keyString) <= 2 {
		return keyString
	}

	startIndex := 0
	endIndex := len(keyString)

	// Trim quotes
	if keyString[0] == '"' || keyString[0] == '\'' {
		startIndex++
	}
	if endIndex > startIndex && (keyString[endIndex-1] == '"' || keyString[endIndex-1] == '\'') {
		endIndex--
	}

	// Trim colon
	if endIndex > startIndex && keyString[endIndex-1] == ':' {
		endIndex--
	}

	// Trim spaces (simple case)
	for startIndex < endIndex && keyString[startIndex] == ' ' {
		startIndex++
	}

	for endIndex > startIndex && keyString[endIndex-1] == ' ' {
		endIndex--
	}

	if startIndex == 0 && endIndex == len(keyString) {
		return keyString
	}

	return keyString[startIndex:endIndex]
}
