package golog

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoggerWithInfo(t *testing.T) {
	// Given
	type details map[string]any
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(InfoLevel),
		WithOutput(buf),
		WithBaseFields(map[string]any{
			"app":    "testapp",
			"env":    "test",
			"userID": "42",
		}),
	)

	// When
	jl.Info("info message", details{"orderID": "1001"})
	jl.Warn("warn message", details{"diskSpace": "low"})
	jl.Error("error message", details{"errorCode": 500})
	jl.Debug("debug message", details{"debugInfo": "details"})

	// Then
	levels := collectLevelsFromBuffer(buf)
	// expect info, warn, error (debug suppressed)
	if _, ok := levels["info"]; !ok {
		t.Fatalf("expected info to be present for level INFO")
	}
	if _, ok := levels["warn"]; !ok {
		t.Fatalf("expected warn to be present for level INFO")
	}
	if _, ok := levels["error"]; !ok {
		t.Fatalf("expected error to be present for level INFO")
	}
	if _, ok := levels["debug"]; ok {
		t.Fatalf("did not expect debug to be present for level INFO")
	}
}

func TestLoggerWithWarn(t *testing.T) {
	// Given
	type details map[string]any
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(WarnLevel),
		WithOutput(buf),
		WithBaseFields(map[string]any{
			"app": "testapp",
			"env": "test",
		}),
	)

	// When
	jl.Info("info message", details{"infoID": "1001"})
	jl.Warn("warn message", details{"diskSpace": "low"})
	jl.Error("error message", details{"errorCode": 500})
	jl.Debug("debug message", details{"debugInfo": "details"})

	// Then
	levels := collectLevelsFromBuffer(buf)
	// expect warn and error only
	if _, ok := levels["warn"]; !ok {
		t.Fatalf("expected warn to be present for level WARN")
	}
	if _, ok := levels["error"]; !ok {
		t.Fatalf("expected error to be present for level WARN")
	}
	if _, ok := levels["info"]; ok {
		t.Fatalf("did not expect info to be present for level WARN")
	}
	if _, ok := levels["debug"]; ok {
		t.Fatalf("did not expect debug to be present for level WARN")
	}
}

func TestLoggerWithError(t *testing.T) {
	// Given
	type details map[string]any
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(ErrorLevel),
		WithOutput(buf),
		WithBaseFields(map[string]any{
			"app": "testapp",
			"env": "test",
		}),
	)

	// When
	jl.Info("info message", details{"infoID": "1001"})
	jl.Warn("warn message", details{"diskSpace": "low"})
	jl.Error("error message", details{"errorCode": 500})
	jl.Debug("debug message", details{"debugInfo": "details"})

	// Then
	levels := collectLevelsFromBuffer(buf)
	// expect only error
	if _, ok := levels["error"]; !ok {
		t.Fatalf("expected error to be present for level ERROR")
	}
	if len(levels) != 1 {
		t.Fatalf("expected only error level for ERROR, got %v", levels)
	}
}

func TestLoggerWithDebug(t *testing.T) {
	// Given
	type details map[string]any
	buf := &bytes.Buffer{}
	jl := NewJSONLoggerWithOptions(
		WithLevel(DebugLevel),
		WithOutput(buf),
		WithBaseFields(map[string]any{
			"app": "testapp",
			"env": "test",
		}),
	)

	// When
	jl.Info("info message", details{"infoID": "1001"})
	jl.Warn("warn message", details{"diskSpace": "low"})
	jl.Error("error message", details{"errorCode": 500})
	jl.Debug("debug message", details{"debugInfo": "details"})

	// Then
	levels := collectLevelsFromBuffer(buf)
	// expect debug, info, warn, error
	for _, want := range []string{"debug", "info", "warn", "error"} {
		if _, ok := levels[want]; !ok {
			t.Fatalf("expected %s to be present for level DEBUG", want)
		}
	}
}

// collectLevelsFromBuffer parses newline-delimited JSON log lines from buf and
// returns a set of the `level` field values found.
//
// It is tolerant of malformed lines (those are skipped) and returns an empty
// set for an empty buffer.
func collectLevelsFromBuffer(buf *bytes.Buffer) map[string]struct{} {
	out := make(map[string]struct{})
	s := strings.TrimSpace(buf.String())
	if s == "" {
		return out
	}
	lines := strings.Split(s, "\n")
	for _, ln := range lines {
		var m map[string]any
		if err := json.Unmarshal([]byte(ln), &m); err != nil {
			// ignore parse errors in test context
			continue
		}
		if lv, ok := m["level"].(string); ok {
			out[lv] = struct{}{}
		}
	}
	return out
}
