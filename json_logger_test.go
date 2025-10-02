package golog

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewJSONLoggerWithDefaults(t *testing.T) {
	// Given
	jl := NewJSONLogger()

	// When
	// (no action needed, as we're testing defaults)

	// Then
	if jl.level != InfoLevel {
		t.Errorf("expected default level to be InfoLevel, got %v", jl.level)
	}

	if jl.output != os.Stdout {
		t.Errorf("expected default output to be os.Stdout, got %v", jl.output)
	}

	if len(jl.baseFields) != 0 {
		t.Errorf("expected default baseFields to be empty, got %v", jl.baseFields)
	}

	if jl.bufferPool.New == nil {
		t.Error("expected bufferPool to have a New function, got nil")
	}

	if jl.writer == nil {
		t.Error("expected writer to be initialized, got nil")
	}

	if jl.timeFormat != time.RFC3339Nano {
		t.Errorf("expected default timeFormat to be RFC3339Nano, got %s", jl.timeFormat)
	}
}

func TestNewJSONLoggerWithLevel(t *testing.T) {
	// Given
	jl := NewJSONLoggerWithOptions(WithLevel(DebugLevel))

	// When
	// (no action needed, as we're testing level setting)

	// Then
	if jl.level != DebugLevel {
		t.Errorf("expected level to be DebugLevel, got %v", jl.level)
	}
}

func TestNewJSONLoggerWithOutput(t *testing.T) {
	// Given
	customOutput := os.Stderr
	jl := NewJSONLoggerWithOptions(WithOutput(customOutput))

	// When
	// (no action needed, as we're testing output setting)

	// Then
	if jl.output != customOutput {
		t.Errorf("expected output to be customOutput, got %v", jl.output)
	}
}

func TestNewJSONLoggerWithOptions(t *testing.T) {
	// Given
	environment := "development"
	customOutput := os.Stderr
	baseFields := map[string]any{
		"app":         "testapp",
		"environment": environment,
	}

	jl := NewJSONLoggerWithOptions(
		WithLevel(WarnLevel),
		WithOutput(customOutput),
		WithBaseFields(baseFields),
		WithCustomTimeFormat(time.RFC1123Z),
	)

	// When
	// (no action needed, as we're testing options application)

	// Then
	if jl.level != WarnLevel {
		t.Errorf("expected level to be WarnLevel, got %v", jl.level)
	}

	if jl.output != customOutput {
		t.Errorf("expected output to be customOutput, got %v", jl.output)
	}

	if len(jl.baseFields) != 2 || jl.baseFields["app"] != "testapp" || jl.baseFields["environment"] != environment {
		t.Errorf("expected baseFields to contain app=testapp and environment=%s, got %v", environment, jl.baseFields)
	}

	if jl.bufferPool.New == nil {
		t.Error("expected bufferPool to have a New function, got nil")
	}

	if jl.timeFormat != time.RFC1123Z {
		t.Errorf("expected timeFormat to be RFC1123Z, got %s", jl.timeFormat)
	}
}

func TestWithLogWriter(t *testing.T) {
	// Given
	customWriter := NewCompactJSONLogWriter()
	jl := NewJSONLoggerWithOptions(WithLogWriter(customWriter))

	// When
	// (no action needed, as we're testing writer setting)

	// Then
	if jl.writer != customWriter {
		t.Errorf("expected writer to be customWriter, got %v", jl.writer)
	}
}

func TestWithLogWriterNil(t *testing.T) {
	// Given
	originalWriter := NewJSONLogWriter()
	jl := &JSONLogger{writer: originalWriter}

	// When
	WithLogWriter(nil)(jl)

	// Then
	if jl.writer != originalWriter {
		t.Errorf("expected writer to remain unchanged when nil is passed, got %v", jl.writer)
	}
}

func TestWithPrettyJSON(t *testing.T) {
	// Given
	indent := "    "
	jl := NewJSONLoggerWithOptions(WithPrettyJSON(indent))

	// When
	// (no action needed, as we're testing pretty JSON setting)

	// Then
	prettyWriter, ok := jl.writer.(*PrettyJSONLogWriter)
	if !ok {
		t.Errorf("expected writer to be PrettyJSONLogWriter, got %T", jl.writer)
	}
	if prettyWriter.indent != indent {
		t.Errorf("expected indent to be %q, got %q", indent, prettyWriter.indent)
	}
}

func TestWithPrettyJSONDefaultIndent(t *testing.T) {
	// Given
	jl := NewJSONLoggerWithOptions(WithPrettyJSON(""))

	// When
	// (no action needed, as we're testing default indent)

	// Then
	prettyWriter, ok := jl.writer.(*PrettyJSONLogWriter)
	if !ok {
		t.Errorf("expected writer to be PrettyJSONLogWriter, got %T", jl.writer)
	}
	if prettyWriter.indent != "  " {
		t.Errorf("expected default indent to be '  ', got %q", prettyWriter.indent)
	}
}

func TestWithCompactJSON(t *testing.T) {
	// Given
	jl := NewJSONLoggerWithOptions(WithCompactJSON())

	// When
	// (no action needed, as we're testing compact JSON setting)

	// Then
	_, ok := jl.writer.(*CompactJSONLogWriter)
	if !ok {
		t.Errorf("expected writer to be CompactJSONLogWriter, got %T", jl.writer)
	}
}

func TestWithBaseField(t *testing.T) {
	// Given
	key := "service"
	value := "test-service"
	jl := NewJSONLoggerWithOptions(WithBaseField(key, value))

	// When
	// (no action needed, as we're testing base field setting)

	// Then
	if len(jl.baseFields) != 1 {
		t.Errorf("expected baseFields to have 1 entry, got %d", len(jl.baseFields))
	}
	if jl.baseFields[key] != value {
		t.Errorf("expected baseFields[%s] to be %v, got %v", key, value, jl.baseFields[key])
	}
}

func TestWithCustomTimeFormat(t *testing.T) {
	// Given
	customFormat := time.RFC822
	jl := NewJSONLoggerWithOptions(WithCustomTimeFormat(customFormat))

	// When
	// (no action needed, as we're testing time format setting)

	// Then
	if jl.timeFormat != customFormat {
		t.Errorf("expected timeFormat to be %s, got %s", customFormat, jl.timeFormat)
	}
}

func TestWithCustomTimeFormatEmpty(t *testing.T) {
	// Given
	originalFormat := time.RFC3339Nano
	jl := &JSONLogger{timeFormat: originalFormat}

	// When
	WithCustomTimeFormat("")(jl)

	// Then
	if jl.timeFormat != originalFormat {
		t.Errorf("expected timeFormat to remain unchanged when empty string is passed, got %s", jl.timeFormat)
	}
}

func TestJSONLoggerIntegration(t *testing.T) {
	// Given
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(DebugLevel),
		WithOutput(buf),
		WithBaseField("service", "test"),
	)

	// When
	jl.Info("test message", map[string]any{"key": "value"})

	// Then
	output := buf.String()
	if !strings.Contains(output, `"service":"test"`) {
		t.Errorf("expected output to contain service field, got %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("expected output to contain key field, got %s", output)
	}
	if !strings.Contains(output, `"message":"test message"`) {
		t.Errorf("expected output to contain message field, got %s", output)
	}
	if !strings.Contains(output, `"level":"info"`) {
		t.Errorf("expected output to contain level field, got %s", output)
	}
}

func TestPrettyJSONLoggerIntegration(t *testing.T) {
	// Given
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(InfoLevel),
		WithOutput(buf),
		WithPrettyJSON("  "),
		WithBaseField("service", "test"),
	)

	// When
	jl.Info("test message", map[string]any{"key": "value"})

	// Then
	output := buf.String()
	// Check for pretty formatting (newlines and indentation)
	if !strings.Contains(output, "{\n") {
		t.Errorf("expected pretty JSON to contain newlines and braces, got %s", output)
	}
	if !strings.Contains(output, `  "service": "test"`) {
		t.Errorf("expected pretty JSON to contain indented service field, got %s", output)
	}
}

func TestCompactJSONLoggerIntegration(t *testing.T) {
	// Given
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(InfoLevel),
		WithOutput(buf),
		WithCompactJSON(),
		WithBaseField("service", "test"),
	)

	// When
	jl.Info("test message", map[string]any{"key": "value"})

	// Then
	output := buf.String()
	// Should be compact (no extra spaces or newlines except the final one)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Errorf("expected compact JSON to be on a single line, got %d lines: %s", len(lines), output)
	}
}
