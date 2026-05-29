package golog

import (
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Level represents a logging severity threshold.
//
// Higher values mean higher severity.
// A logger configured with a given level writes entries whose level is
// greater than or equal to that configured level.
type Level int32

const (
	// DebugLevel enables debug, info, warn, and error logs.
	DebugLevel Level = 0 + iota
	// InfoLevel enables info, warn, and error logs.
	InfoLevel
	// WarnLevel enables warn and error logs.
	WarnLevel
	// ErrorLevel enables only error logs.
	ErrorLevel
)

// JSONLogger is a small, fast, concurrent-safe JSON logger implementation.
// Create one with NewJSONLogger or NewJSONLoggerWithOptions. Use the Option
// helpers to customize level, output and base fields.
type JSONLogger struct {
	output     io.Writer
	baseFields map[string]any
	// level is stored as int32 and accessed atomically so concurrent SetLevel
	// calls are safe without a mutex, and the field is still comparable as Level
	// in internal tests via a direct cast.
	level      Level
	mutex      sync.Mutex
	bufferPool sync.Pool
	// lockWrites protects output writes for concurrency-safe ordering by default.
	// You can disable it with WithWriteLock(false) for maximum throughput when
	// writing to a thread-safe sink.
	lockWrites bool
	// timeFormat controls how timestamps are rendered. Defaults to
	// time.RFC3339Nano but can be changed with WithCustomTimeFormat.
	timeFormat string
	// baseFieldsCache holds a pre-encoded JSON fragment of all base fields,
	// e.g. `,"service":"api","version":"1.0"`. Built once on first log call.
	baseFieldsCache []byte
	baseFieldsOnce  sync.Once
}

// Option configures the JSONLogger.
type Option func(*JSONLogger)

// NewJSONLogger returns a logger with sensible defaults:
//   - Level: InfoLevel
//   - Output: os.Stdout
//   - No base fields
func NewJSONLogger() *JSONLogger {
	l := &JSONLogger{
		output:     os.Stdout,
		baseFields: make(map[string]any),
		level:      InfoLevel,
		lockWrites: true,
		timeFormat: time.RFC3339Nano,
		bufferPool: sync.Pool{
			New: func() any {
				// Pre-allocate a reusable byte slice for the hot path.
				slice := make([]byte, 0, 512)
				return &slice
			},
		},
	}
	return l
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
	return func(jsonLogger *JSONLogger) { atomic.StoreInt32((*int32)(&jsonLogger.level), int32(logLevel)) }
}

// WithOutput sets the writer for the logger (stdout, file, buffer, etc.).
func WithOutput(writer io.Writer) Option {
	return func(jsonLogger *JSONLogger) {
		jsonLogger.output = writer
	}
}

// WithWriteLock enables/disables per-write mutex locking.
//
// By default locking is enabled for safe concurrent writes to non-thread-safe
// outputs (for example bytes.Buffer, files without external synchronization,
// or custom writers).
//
// Disable only when your output is already safe for concurrent writes and you
// prefer maximum throughput over strict write serialization.
func WithWriteLock(enabled bool) Option {
	return func(jsonLogger *JSONLogger) {
		jsonLogger.lockWrites = enabled
	}
}

// WithBaseFields adds the provided fields to the logger's base fields. These
// fields are included in every emitted log entry.
func WithBaseFields(fields map[string]any) Option {
	return func(jsonLogger *JSONLogger) {
		for key, value := range fields {
			jsonLogger.baseFields[key] = value
		}
		// Reset cache so it will be rebuilt on next log call.
		jsonLogger.baseFieldsOnce = sync.Once{}
	}
}

// WithBaseField adds a single base field key/value that will be included in
// every log entry.
func WithBaseField(key string, value any) Option {
	return func(jsonLogger *JSONLogger) {
		jsonLogger.baseFields[key] = value
		// Reset cache so it will be rebuilt on next log call.
		jsonLogger.baseFieldsOnce = sync.Once{}
	}
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

// buildBaseFieldsCache pre-encodes all base fields into a reusable []byte fragment.
// Called once via sync.Once before the first log entry is written.
func (jsonLogger *JSONLogger) buildBaseFieldsCache() {
	if len(jsonLogger.baseFields) == 0 {
		jsonLogger.baseFieldsCache = nil
		return
	}
	cache := make([]byte, 0, 128)
	for fieldKey, fieldValue := range jsonLogger.baseFields {
		cache = append(cache, ',')
		cache = appendQuoteBytes(cache, fieldKey)
		cache = append(cache, ':')
		var ok bool
		cache, ok = appendValueBytes(cache, fieldValue)
		if !ok {
			cache = appendQuoteBytes(cache, "<unsupported>")
		}
	}
	jsonLogger.baseFieldsCache = cache
}

// logFields writes a JSON entry using typed Field values.
func (jsonLogger *JSONLogger) logFields(logLevel Level, levelString, message string, fields []Field) {
	if Level(atomic.LoadInt32((*int32)(&jsonLogger.level))) > logLevel {
		return
	}

	jsonLogger.baseFieldsOnce.Do(jsonLogger.buildBaseFieldsCache)

	bufPtr := jsonLogger.bufferPool.Get().(*[]byte)
	buffer := (*bufPtr)[:0]

	timeFormat := jsonLogger.timeFormat

	buffer = append(buffer, '{')
	buffer = append(buffer, `"timestamp":"`...)
	var tsBuf [64]byte
	now := time.Now().UTC()
	if timeFormat == time.RFC3339Nano {
		buffer = append(buffer, appendRFC3339NanoUTC(tsBuf[:0], now)...)
	} else {
		buffer = now.AppendFormat(buffer, timeFormat)
	}
	buffer = append(buffer, '"')
	buffer = append(buffer, `,"level":"`...)
	buffer = append(buffer, levelString...)
	buffer = append(buffer, '"')
	buffer = append(buffer, `,"message":`...)
	buffer = appendQuoteBytes(buffer, message)

	if jsonLogger.baseFieldsCache != nil {
		buffer = append(buffer, jsonLogger.baseFieldsCache...)
	}

	for i := range fields {
		buffer = appendFieldBytes(buffer, fields[i])
	}

	buffer = append(buffer, '}', '\n')

	if jsonLogger.lockWrites {
		jsonLogger.mutex.Lock()
		_, _ = jsonLogger.output.Write(buffer)
		jsonLogger.mutex.Unlock()
	} else {
		_, _ = jsonLogger.output.Write(buffer)
	}

	*bufPtr = buffer[:0]
	jsonLogger.bufferPool.Put(bufPtr)
}

func appendRFC3339NanoUTC(dst []byte, t time.Time) []byte {
	year, month, day := t.Date()
	hour, minute, sec := t.Clock()
	nsec := t.Nanosecond()

	dst = append4(dst, year)
	dst = append(dst, '-')
	dst = append2(dst, int(month))
	dst = append(dst, '-')
	dst = append2(dst, day)
	dst = append(dst, 'T')
	dst = append2(dst, hour)
	dst = append(dst, ':')
	dst = append2(dst, minute)
	dst = append(dst, ':')
	dst = append2(dst, sec)

	if nsec != 0 {
		dst = append(dst, '.')
		start := len(dst)
		var frac [9]byte
		for i := 8; i >= 0; i-- {
			frac[i] = byte('0' + (nsec % 10))
			nsec /= 10
		}
		dst = append(dst, frac[:]...)
		for len(dst) > start && dst[len(dst)-1] == '0' {
			dst = dst[:len(dst)-1]
		}
	}

	return append(dst, 'Z')
}

func append2(dst []byte, value int) []byte {
	return append(dst, byte('0'+value/10), byte('0'+value%10))
}

func append4(dst []byte, value int) []byte {
	return append(
		dst,
		byte('0'+(value/1000)%10),
		byte('0'+(value/100)%10),
		byte('0'+(value/10)%10),
		byte('0'+value%10),
	)
}
