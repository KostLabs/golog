package golog

import (
	"bytes"
	"fmt"
	"strings"
	"time"
)

// LogWriter is an interface for writing structured log entries.
// It abstracts the JSON writing logic to eliminate code duplication.
type LogWriter interface {
	// WriteLogEntry writes a complete log entry with timestamp, level, message and fields
	WriteLogEntry(buf *bytes.Buffer, timestamp time.Time, timeFormat, level, message string,
		baseFields map[string]any, additionalFields ...map[string]any) error
}

// JSONLogWriter implements LogWriter for JSON output format.
type JSONLogWriter struct{}

// NewJSONLogWriter creates a new JSON log writer.
func NewJSONLogWriter() *JSONLogWriter {
	return &JSONLogWriter{}
}

// WriteLogEntry writes a JSON log entry to the buffer.
func (w *JSONLogWriter) WriteLogEntry(buf *bytes.Buffer, timestamp time.Time, timeFormat, level, message string,
	baseFields map[string]any, additionalFields ...map[string]any) error {

	var encodingError error

	// Start JSON object
	buf.WriteByte('{')

	// Write timestamp
	buf.WriteString(`"timestamp":"`)
	buf.WriteString(timestamp.UTC().Format(timeFormat))
	buf.WriteByte('"')

	// Write level
	buf.WriteString(`,"level":"`)
	buf.WriteString(level)
	buf.WriteByte('"')

	// Write message
	buf.WriteString(`,"message":`)
	fastQuote(buf, message)

	// Write base fields
	if err := w.writeFields(buf, baseFields); err != nil {
		encodingError = err
	}

	// Write additional fields
	for _, kvMap := range additionalFields {
		if err := w.writeFields(buf, kvMap); err != nil {
			encodingError = err
		}
	}

	// Add error field if encoding failed
	if encodingError != nil {
		buf.WriteByte(',')
		buf.WriteString(`"error":`)
		fastQuote(buf, encodingError.Error())
	}

	// Close JSON object and add newline
	buf.WriteByte('}')
	buf.WriteByte('\n')

	return encodingError
}

// writeFields writes a map of fields to the buffer.
func (w *JSONLogWriter) writeFields(buf *bytes.Buffer, fields map[string]any) error {
	if fields == nil {
		return nil
	}

	for k, v := range fields {
		// Simple key normalization inline
		key := k
		if len(k) > 0 && (k[0] == '"' || k[0] == '\'' || k[len(k)-1] == ':') {
			// Only normalize if necessary
			key = normalizeKey(k)
		}

		buf.WriteByte(',')
		fastQuote(buf, key)
		buf.WriteByte(':')

		if !encodeValue(buf, v) {
			// Fallback to string representation for unsupported types
			fastQuote(buf, fmt.Sprintf("%v", v))
			return errMarshalTypeUnsupported
		}
	}

	return nil
}

// CompactJSONLogWriter is a variant that writes more compact JSON (no spaces).
type CompactJSONLogWriter struct {
	*JSONLogWriter
}

// NewCompactJSONLogWriter creates a compact JSON log writer.
func NewCompactJSONLogWriter() *CompactJSONLogWriter {
	return &CompactJSONLogWriter{
		JSONLogWriter: NewJSONLogWriter(),
	}
}

// PrettyJSONLogWriter is a variant that writes prettier JSON with indentation.
type PrettyJSONLogWriter struct {
	indent string
}

// NewPrettyJSONLogWriter creates a pretty JSON log writer with specified indentation.
func NewPrettyJSONLogWriter(indent string) *PrettyJSONLogWriter {
	if indent == "" {
		indent = "  " // Default to 2 spaces
	}
	return &PrettyJSONLogWriter{indent: indent}
}

// WriteLogEntry writes a pretty-formatted JSON log entry.
func (w *PrettyJSONLogWriter) WriteLogEntry(buf *bytes.Buffer, timestamp time.Time, timeFormat, level, message string,
	baseFields map[string]any, additionalFields ...map[string]any) error {

	var encodingError error

	// Start JSON object with newline and indentation
	buf.WriteString("{\n")

	// Write timestamp with indentation
	buf.WriteString(w.indent)
	buf.WriteString(`"timestamp": "`)
	buf.WriteString(timestamp.UTC().Format(timeFormat))
	buf.WriteString("\",\n")

	// Write level with indentation
	buf.WriteString(w.indent)
	buf.WriteString(`"level": "`)
	buf.WriteString(level)
	buf.WriteString("\",\n")

	// Write message with indentation
	buf.WriteString(w.indent)
	buf.WriteString(`"message": `)
	fastQuote(buf, message)

	// Write base fields
	if err := w.writePrettyFields(buf, baseFields); err != nil {
		encodingError = err
	}

	// Write additional fields
	for _, kvMap := range additionalFields {
		if err := w.writePrettyFields(buf, kvMap); err != nil {
			encodingError = err
		}
	}

	// Add error field if encoding failed
	if encodingError != nil {
		buf.WriteString(",\n")
		buf.WriteString(w.indent)
		buf.WriteString(`"error": `)
		fastQuote(buf, encodingError.Error())
	}

	// Close JSON object with newline
	buf.WriteString("\n}\n")

	return encodingError
}

// writePrettyFields writes fields with pretty formatting.
func (w *PrettyJSONLogWriter) writePrettyFields(buf *bytes.Buffer, fields map[string]any) error {
	if fields == nil {
		return nil
	}

	for k, v := range fields {
		// Simple key normalization inline
		key := k
		if len(k) > 0 && (k[0] == '"' || k[0] == '\'' || k[len(k)-1] == ':') {
			key = normalizeKey(k)
		}

		buf.WriteString(",\n")
		buf.WriteString(w.indent)
		fastQuote(buf, key)
		buf.WriteString(": ")

		if !encodeValue(buf, v) {
			fastQuote(buf, fmt.Sprintf("%v", v))
			return errMarshalTypeUnsupported
		}
	}

	return nil
}

// normalizeKey performs simple key normalization inline for performance
func normalizeKey(k string) string {
	nk := strings.TrimSpace(k)
	nk = strings.TrimSuffix(nk, ":")
	nk = strings.Trim(nk, "\"'")
	return strings.TrimSpace(nk)
}
