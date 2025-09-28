package golog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestLogMergesBaseAndProvidedMaps(t *testing.T) {
	// Given
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(DebugLevel),
		WithOutput(buf),
		WithBaseFields(map[string]any{"app": "original", "env": "test"}),
	)

	// When
	jl.Info("user created", map[string]any{"userID": "u123", "app": "override"})

	// Then
	s := strings.TrimSpace(buf.String())
	if s == "" {
		t.Fatal("expected log output, got empty buffer")
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(strings.SplitN(s, "\n", 2)[0]), &got); err != nil {
		t.Fatalf("failed to unmarshal log JSON: %v", err)
	}

	if got["app"] != "override" {
		t.Fatalf("expected app=override, got %v", got["app"])
	}
	if got["userID"] != "u123" {
		t.Fatalf("expected userID=u123, got %v", got["userID"])
	}
	if got["message"] != "user created" {
		t.Fatalf("expected message=user created, got %v", got["message"])
	}
	// timestamp parseable
	if ts, ok := got["timestamp"].(string); !ok {
		t.Fatalf("timestamp missing or not a string: %v", got["timestamp"])
	} else {
		if _, err := time.Parse(time.RFC3339Nano, ts); err != nil {
			t.Fatalf("timestamp not parseable RFC3339Nano: %v", err)
		}
	}
}

func TestLogLevelFilteringSanity(t *testing.T) {
	// Given
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(WarnLevel),
		WithOutput(buf),
	)

	// When
	jl.Info("info should be suppressed", map[string]any{"k": "v"})
	jl.Error("error should show", map[string]any{"err": "boom"})

	// Then
	levels := collectLevelsFromBuffer(buf)
	if _, ok := levels["error"]; !ok {
		t.Fatalf("expected error present")
	}
	if _, ok := levels["info"]; ok {
		t.Fatalf("did not expect info present")
	}
}

func TestLogEncodeFallbackWritesMinimalJSON(t *testing.T) {
	// Given
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(DebugLevel),
		WithOutput(buf),
	)

	// When
	jl.Info("bad payload", map[string]any{"bad": make(chan int)})

	// Then
	s := strings.TrimSpace(buf.String())
	if s == "" {
		t.Fatal("expected fallback output in buffer")
	}

	// fallback emits a single JSON-ish line; attempt to decode it
	first := strings.SplitN(s, "\n", 2)[0]
	var got map[string]any
	if err := json.Unmarshal([]byte(first), &got); err != nil {
		t.Fatalf("fallback output not valid JSON: %v -- %s", err, first)
	}
	if got["message"] != "bad payload" {
		t.Fatalf("expected message=bad payload, got %v", got["message"])
	}
	if _, ok := got["error"]; !ok {
		t.Fatalf("expected error field in fallback output, got %v", got)
	}
}
