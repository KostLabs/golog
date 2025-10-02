package golog

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestJSONLogWriter_WriteLogEntry(t *testing.T) {
	// Given
	writer := NewJSONLogWriter()
	buf := &bytes.Buffer{}
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFormat := time.RFC3339
	level := "info"
	message := "test message"
	baseFields := map[string]any{"service": "test"}
	additionalFields := map[string]any{"key": "value"}

	// When
	err := writer.WriteLogEntry(buf, timestamp, timeFormat, level, message, baseFields, additionalFields)

	// Then
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, `"timestamp":"2024-01-01T12:00:00Z"`) {
		t.Errorf("expected timestamp in output, got %s", output)
	}
	if !strings.Contains(output, `"level":"info"`) {
		t.Errorf("expected level in output, got %s", output)
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Errorf("expected message in output, got %s", output)
	}
	if !strings.Contains(output, `"service":"test"`) {
		t.Errorf("expected service field in output, got %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("expected key field in output, got %s", output)
	}
	if !strings.HasSuffix(output, "}\n") {
		t.Errorf("expected output to end with '}\\n', got %s", output)
	}
}

func TestJSONLogWriter_WriteLogEntryWithNilFields(t *testing.T) {
	// Given
	writer := NewJSONLogWriter()
	buf := &bytes.Buffer{}
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFormat := time.RFC3339
	level := "info"
	message := "test message"

	// When
	err := writer.WriteLogEntry(buf, timestamp, timeFormat, level, message, nil, nil)

	// Then
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, `"timestamp":"2024-01-01T12:00:00Z"`) {
		t.Errorf("expected timestamp in output, got %s", output)
	}
	if !strings.Contains(output, `"level":"info"`) {
		t.Errorf("expected level in output, got %s", output)
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Errorf("expected message in output, got %s", output)
	}
}

func TestJSONLogWriter_WriteLogEntryWithUnsupportedType(t *testing.T) {
	// Given
	writer := NewJSONLogWriter()
	buf := &bytes.Buffer{}
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFormat := time.RFC3339
	level := "info"
	message := "test message"
	fieldsWithChan := map[string]any{"bad": make(chan int)}

	// When
	err := writer.WriteLogEntry(buf, timestamp, timeFormat, level, message, nil, fieldsWithChan)

	// Then
	if err == nil {
		t.Error("expected error for unsupported type, got nil")
	}
	
	output := buf.String()
	if !strings.Contains(output, `"error":"unsupported type for marshal"`) {
		t.Errorf("expected error field in output, got %s", output)
	}
}

func TestCompactJSONLogWriter(t *testing.T) {
	// Given
	writer := NewCompactJSONLogWriter()
	buf := &bytes.Buffer{}
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFormat := time.RFC3339
	level := "info"
	message := "test message"
	baseFields := map[string]any{"service": "test"}

	// When
	err := writer.WriteLogEntry(buf, timestamp, timeFormat, level, message, baseFields)

	// Then
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	output := buf.String()
	// Should be the same as regular JSON for now (CompactJSONLogWriter embeds JSONLogWriter)
	if !strings.Contains(output, `"service":"test"`) {
		t.Errorf("expected service field in output, got %s", output)
	}
}

func TestPrettyJSONLogWriter_WriteLogEntry(t *testing.T) {
	// Given
	indent := "  "
	writer := NewPrettyJSONLogWriter(indent)
	buf := &bytes.Buffer{}
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFormat := time.RFC3339
	level := "info"
	message := "test message"
	baseFields := map[string]any{"service": "test"}
	additionalFields := map[string]any{"key": "value"}

	// When
	err := writer.WriteLogEntry(buf, timestamp, timeFormat, level, message, baseFields, additionalFields)

	// Then
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "{\n") {
		t.Errorf("expected pretty JSON to start with '{\\n', got %s", output)
	}
	if !strings.Contains(output, `  "timestamp": "2024-01-01T12:00:00Z"`) {
		t.Errorf("expected indented timestamp in output, got %s", output)
	}
	if !strings.Contains(output, `  "level": "info"`) {
		t.Errorf("expected indented level in output, got %s", output)
	}
	if !strings.Contains(output, `  "message": "test message"`) {
		t.Errorf("expected indented message in output, got %s", output)
	}
	if !strings.Contains(output, `  "service": "test"`) {
		t.Errorf("expected indented service field in output, got %s", output)
	}
	if !strings.Contains(output, `  "key": "value"`) {
		t.Errorf("expected indented key field in output, got %s", output)
	}
	if !strings.HasSuffix(output, "\n}\n") {
		t.Errorf("expected output to end with '\\n}\\n', got %s", output)
	}
}

func TestPrettyJSONLogWriter_WithDefaultIndent(t *testing.T) {
	// Given
	writer := NewPrettyJSONLogWriter("")

	// When
	// (no action needed, testing constructor)

	// Then
	if writer.indent != "  " {
		t.Errorf("expected default indent to be '  ', got %q", writer.indent)
	}
}

func TestPrettyJSONLogWriter_WriteLogEntryWithNilFields(t *testing.T) {
	// Given
	writer := NewPrettyJSONLogWriter("  ")
	buf := &bytes.Buffer{}
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFormat := time.RFC3339
	level := "info"
	message := "test message"

	// When
	err := writer.WriteLogEntry(buf, timestamp, timeFormat, level, message, nil, nil)

	// Then
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, "{\n") {
		t.Errorf("expected pretty JSON to start with '{\\n', got %s", output)
	}
	if !strings.HasSuffix(output, "\n}\n") {
		t.Errorf("expected output to end with '\\n}\\n', got %s", output)
	}
}

func TestPrettyJSONLogWriter_WriteLogEntryWithUnsupportedType(t *testing.T) {
	// Given
	writer := NewPrettyJSONLogWriter("  ")
	buf := &bytes.Buffer{}
	timestamp := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	timeFormat := time.RFC3339
	level := "info"
	message := "test message"
	fieldsWithChan := map[string]any{"bad": make(chan int)}

	// When
	err := writer.WriteLogEntry(buf, timestamp, timeFormat, level, message, nil, fieldsWithChan)

	// Then
	if err == nil {
		t.Error("expected error for unsupported type, got nil")
	}
	
	output := buf.String()
	if !strings.Contains(output, `  "error": "unsupported type for marshal"`) {
		t.Errorf("expected indented error field in output, got %s", output)
	}
}

func TestJSONLogWriter_writeFields(t *testing.T) {
	// Given
	writer := NewJSONLogWriter()
	buf := &bytes.Buffer{}
	fields := map[string]any{
		"string_field": "value",
		"int_field":    42,
		"bool_field":   true,
		"nil_field":    nil,
	}

	// When
	err := writer.writeFields(buf, fields)

	// Then
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, `,"string_field":"value"`) {
		t.Errorf("expected string field in output, got %s", output)
	}
	if !strings.Contains(output, `,"int_field":42`) {
		t.Errorf("expected int field in output, got %s", output)
	}
	if !strings.Contains(output, `,"bool_field":true`) {
		t.Errorf("expected bool field in output, got %s", output)
	}
	if !strings.Contains(output, `,"nil_field":null`) {
		t.Errorf("expected nil field in output, got %s", output)
	}
}

func TestJSONLogWriter_writeFieldsWithKeyNormalization(t *testing.T) {
	// Given
	writer := NewJSONLogWriter()
	buf := &bytes.Buffer{}
	fields := map[string]any{
		`"quoted_key"`:   "value1",
		`'single_quoted'`: "value2",
		`key_with_colon:`: "value3",
		` spaced_key `:    "value4",
	}

	// When
	err := writer.writeFields(buf, fields)

	// Then
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	
	output := buf.String()
	if !strings.Contains(output, `"quoted_key":"value1"`) {
		t.Errorf("expected normalized quoted key in output, got %s", output)
	}
	if !strings.Contains(output, `"single_quoted":"value2"`) {
		t.Errorf("expected normalized single quoted key in output, got %s", output)
	}
	if !strings.Contains(output, `"key_with_colon":"value3"`) {
		t.Errorf("expected normalized colon key in output, got %s", output)
	}
	if !strings.Contains(output, `"spaced_key":"value4"`) {
		t.Errorf("expected normalized spaced key in output, got %s", output)
	}
}