package golog

import (
	"io"
	"testing"
	"time"
)

func BenchmarkDefaultLogger(b *testing.B) {
	jl := NewJSONLoggerWithOptions(WithOutput(io.Discard))

	b.Run("Direct.Info", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			jl.Info("user created", map[string]any{"user_id": i})
		}
	})

	b.Run("Package.Info", func(b *testing.B) {
		prev := logger
		SetLogger(jl)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			Info("user created", map[string]any{"user_id": i})
		}
		SetLogger(prev)
	})

	b.Run("Direct.MergeTwoMaps", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			jl.Info("merged", map[string]any{"a": i, "b": "x"}, map[string]any{"c": i})
		}
	})
}

func BenchmarkDebugLevelLogger(b *testing.B) {
	jl := NewJSONLoggerWithOptions(WithOutput(io.Discard), WithLevel(DebugLevel))

	b.Run("Direct.Info", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			jl.Info("user created", map[string]any{"user_id": i})
		}
	})

	b.Run("Package.Info", func(b *testing.B) {
		prev := logger
		SetLogger(jl)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			Info("user created", map[string]any{"user_id": i})
		}
		SetLogger(prev)
	})

	b.Run("Direct.MergeTwoMaps", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			jl.Info("merged", map[string]any{"a": i, "b": "x"}, map[string]any{"c": i})
		}
	})
}

func BenchmarkWithBaseFieldsLogger(b *testing.B) {
	jl := NewJSONLoggerWithOptions(WithOutput(io.Discard), WithBaseFields(map[string]any{"app": "benchApp", "env": "bench"}))

	b.Run("Direct.Info", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			jl.Info("user created", map[string]any{"user_id": i})
		}
	})

	b.Run("Package.Info", func(b *testing.B) {
		prev := logger
		SetLogger(jl)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			Info("user created", map[string]any{"user_id": i})
		}
		SetLogger(prev)
	})

	b.Run("Direct.MergeTwoMaps", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			jl.Info("merged", map[string]any{"a": i, "b": "x"}, map[string]any{"c": i})
		}
	})
}

func BenchmarkCustomTimeFormatLogger(b *testing.B) {
	jl := NewJSONLoggerWithOptions(WithOutput(io.Discard), WithCustomTimeFormat(time.RFC1123Z))

	b.Run("Direct.Info", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; b.Loop(); i++ {
			jl.Info("user created", map[string]any{"user_id": i})
		}
	})

	b.Run("Package.Info", func(b *testing.B) {
		prev := logger
		SetLogger(jl)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			Info("user created", map[string]any{"user_id": i})
		}
		SetLogger(prev)
	})

	b.Run("Direct.MergeTwoMaps", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			jl.Info("merged", map[string]any{"a": i, "b": "x"}, map[string]any{"c": i})
		}
	})
}
