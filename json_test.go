package golog

import (
	"os"
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
	jl := NewJSONLoggerWithOptions(func(j *JSONLogger) {
		j.output = customOutput
	})

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
