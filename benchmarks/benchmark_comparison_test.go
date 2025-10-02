package benchmarks

import (
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/KostLabs/golog"
	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// setupLoggers creates loggers with identical configuration:
// - Timestamp: RFC3339Nano
// - Output: os.Stdout (or provided writer)
// - Level: info
func setupLoggers(output io.Writer) (*golog.JSONLogger, *slog.Logger, zerolog.Logger, *zap.Logger, *log.Logger, *logrus.Logger) {
	// Golog setup
	gologLogger := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.InfoLevel),
		golog.WithOutput(output),
	)

	// Slog setup with JSON handler
	slogLogger := slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Make timestamp consistent with RFC3339Nano
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().UTC().Format(time.RFC3339Nano))
			}
			// Make level consistent with other loggers
			if a.Key == slog.LevelKey {
				return slog.String("level", a.Value.String())
			}
			// Make message key consistent
			if a.Key == slog.MessageKey {
				return slog.String("message", a.Value.String())
			}
			return a
		},
	}))

	// Zerolog setup
	zerologLogger := zerolog.New(output).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
	// Configure zerolog to use RFC3339Nano and consistent field names
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.TimeFieldFormat = time.RFC3339Nano

	// Zap setup
	zapConfig := zap.NewProductionEncoderConfig()
	zapConfig.TimeKey = "timestamp"
	zapConfig.LevelKey = "level"
	zapConfig.MessageKey = "message"
	zapConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format(time.RFC3339Nano))
	}
	zapConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	zapCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(zapConfig),
		zapcore.AddSync(output),
		zapcore.InfoLevel,
	)
	zapLogger := zap.New(zapCore)

	// Apex/log setup
	apexLogger := &log.Logger{
		Handler: json.New(output),
		Level:   log.InfoLevel,
	}

	// Logrus setup
	logrusLogger := logrus.New()
	logrusLogger.SetOutput(output)
	logrusLogger.SetLevel(logrus.InfoLevel)
	logrusLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	return gologLogger, slogLogger, zerologLogger, zapLogger, apexLogger, logrusLogger
}

// BenchmarkGologInfo benchmarks golog Info logging
func BenchmarkGologInfo(b *testing.B) {
	gologLogger, _, _, _, _, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gologLogger.Info("test message")
		}
	})
}

// BenchmarkSlogInfo benchmarks slog Info logging
func BenchmarkSlogInfo(b *testing.B) {
	_, slogLogger, _, _, _, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slogLogger.Info("test message")
		}
	})
}

// BenchmarkZerologInfo benchmarks zerolog Info logging
func BenchmarkZerologInfo(b *testing.B) {
	_, _, zerologLogger, _, _, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologLogger.Info().Msg("test message")
		}
	})
}

// BenchmarkZapInfo benchmarks zap Info logging
func BenchmarkZapInfo(b *testing.B) {
	_, _, _, zapLogger, _, _ := setupLoggers(io.Discard)
	defer zapLogger.Sync()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zapLogger.Info("test message")
		}
	})
}

// BenchmarkApexInfo benchmarks apex/log Info logging
func BenchmarkApexInfo(b *testing.B) {
	_, _, _, _, apexLogger, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			apexLogger.Info("test message")
		}
	})
}

// BenchmarkLogrusInfo benchmarks logrus Info logging
func BenchmarkLogrusInfo(b *testing.B) {
	_, _, _, _, _, logrusLogger := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logrusLogger.Info("test message")
		}
	})
}

// BenchmarkGologInfoWithFields benchmarks golog Info logging with additional fields
func BenchmarkGologInfoWithFields(b *testing.B) {
	gologLogger, _, _, _, _, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			gologLogger.Info("test message", map[string]any{
				"user_id": 12345,
				"action":  "login",
				"ip":      "192.168.1.100",
			})
		}
	})
}

// BenchmarkSlogInfoWithFields benchmarks slog Info logging with additional fields
func BenchmarkSlogInfoWithFields(b *testing.B) {
	_, slogLogger, _, _, _, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			slogLogger.Info("test message",
				"user_id", 12345,
				"action", "login",
				"ip", "192.168.1.100",
			)
		}
	})
}

// BenchmarkZerologInfoWithFields benchmarks zerolog Info logging with additional fields
func BenchmarkZerologInfoWithFields(b *testing.B) {
	_, _, zerologLogger, _, _, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zerologLogger.Info().
				Int("user_id", 12345).
				Str("action", "login").
				Str("ip", "192.168.1.100").
				Msg("test message")
		}
	})
}

// BenchmarkZapInfoWithFields benchmarks zap Info logging with additional fields
func BenchmarkZapInfoWithFields(b *testing.B) {
	_, _, _, zapLogger, _, _ := setupLoggers(io.Discard)
	defer zapLogger.Sync()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			zapLogger.Info("test message",
				zap.Int("user_id", 12345),
				zap.String("action", "login"),
				zap.String("ip", "192.168.1.100"),
			)
		}
	})
}

// BenchmarkApexInfoWithFields benchmarks apex/log Info logging with additional fields
func BenchmarkApexInfoWithFields(b *testing.B) {
	_, _, _, _, apexLogger, _ := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			apexLogger.WithFields(log.Fields{
				"user_id": 12345,
				"action":  "login",
				"ip":      "192.168.1.100",
			}).Info("test message")
		}
	})
}

// BenchmarkLogrusInfoWithFields benchmarks logrus Info logging with additional fields
func BenchmarkLogrusInfoWithFields(b *testing.B) {
	_, _, _, _, _, logrusLogger := setupLoggers(io.Discard)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logrusLogger.WithFields(logrus.Fields{
				"user_id": 12345,
				"action":  "login",
				"ip":      "192.168.1.100",
			}).Info("test message")
		}
	})
}

// BenchmarkAllLoggersSimple runs a comparative benchmark of simple logging
func BenchmarkAllLoggersSimple(b *testing.B) {
	gologLogger, slogLogger, zerologLogger, zapLogger, apexLogger, logrusLogger := setupLoggers(io.Discard)
	defer zapLogger.Sync()

	b.Run("Golog", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				gologLogger.Info("test message")
			}
		})
	})

	b.Run("Slog", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				slogLogger.Info("test message")
			}
		})
	})

	b.Run("Zerolog", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				zerologLogger.Info().Msg("test message")
			}
		})
	})

	b.Run("Zap", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				zapLogger.Info("test message")
			}
		})
	})

	b.Run("Apex", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				apexLogger.Info("test message")
			}
		})
	})

	b.Run("Logrus", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logrusLogger.Info("test message")
			}
		})
	})
}

// BenchmarkAllLoggersWithFields runs a comparative benchmark of logging with fields
func BenchmarkAllLoggersWithFields(b *testing.B) {
	gologLogger, slogLogger, zerologLogger, zapLogger, apexLogger, logrusLogger := setupLoggers(io.Discard)
	defer zapLogger.Sync()

	b.Run("Golog", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				gologLogger.Info("test message", map[string]any{
					"user_id": 12345,
					"action":  "login",
					"ip":      "192.168.1.100",
				})
			}
		})
	})

	b.Run("Slog", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				slogLogger.Info("test message",
					"user_id", 12345,
					"action", "login",
					"ip", "192.168.1.100",
				)
			}
		})
	})

	b.Run("Zerolog", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				zerologLogger.Info().
					Int("user_id", 12345).
					Str("action", "login").
					Str("ip", "192.168.1.100").
					Msg("test message")
			}
		})
	})

	b.Run("Zap", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				zapLogger.Info("test message",
					zap.Int("user_id", 12345),
					zap.String("action", "login"),
					zap.String("ip", "192.168.1.100"),
				)
			}
		})
	})

	b.Run("Apex", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				apexLogger.WithFields(log.Fields{
					"user_id": 12345,
					"action":  "login",
					"ip":      "192.168.1.100",
				}).Info("test message")
			}
		})
	})

	b.Run("Logrus", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				logrusLogger.WithFields(logrus.Fields{
					"user_id": 12345,
					"action":  "login",
					"ip":      "192.168.1.100",
				}).Info("test message")
			}
		})
	})
}

// TestOutputConsistency verifies that all loggers produce similar JSON output format
func TestOutputConsistency(t *testing.T) {
	// This test ensures all loggers are configured consistently
	// Note: We use os.Stdout here to verify the actual output format
	// In practice, you might want to capture and compare the outputs

	gologLogger, slogLogger, zerologLogger, zapLogger, apexLogger, logrusLogger := setupLoggers(os.Stdout)
	defer zapLogger.Sync()

	t.Log("Testing output consistency - all loggers should produce similar JSON structure")
	t.Log("Expected fields: timestamp (RFC3339Nano), level (info), message")

	t.Log("\n=== Golog output ===")
	gologLogger.Info("test message", map[string]any{"test_field": "test_value"})

	t.Log("\n=== Slog output ===")
	slogLogger.Info("test message", "test_field", "test_value")

	t.Log("\n=== Zerolog output ===")
	zerologLogger.Info().Str("test_field", "test_value").Msg("test message")

	t.Log("\n=== Zap output ===")
	zapLogger.Info("test message", zap.String("test_field", "test_value"))

	t.Log("\n=== Apex output ===")
	apexLogger.WithField("test_field", "test_value").Info("test message")

	t.Log("\n=== Logrus output ===")
	logrusLogger.WithField("test_field", "test_value").Info("test message")
}
