package benchmarks

import (
	"io"
	"log/slog"
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

type benchmarkScenario struct {
	name string

	gologTyped []golog.Field
	slogArgs   []any
	zapFields  []zap.Field
	apexFields log.Fields
	logrus     logrus.Fields

	applyZerolog func(event *zerolog.Event) *zerolog.Event
}

var (
	scenarioSimple = benchmarkScenario{name: "Simple"}

	scenarioWithFields = benchmarkScenario{
		name: "WithFields",
		gologTyped: []golog.Field{
			golog.Int("user_id", 12345),
			golog.Str("action", "login"),
			golog.Str("ip", "192.168.1.100"),
			golog.Bool("success", true),
			golog.Int("latency_ms", 12),
		},
		slogArgs: []any{
			"user_id", 12345,
			"action", "login",
			"ip", "192.168.1.100",
			"success", true,
			"latency_ms", 12,
		},
		zapFields: []zap.Field{
			zap.Int("user_id", 12345),
			zap.String("action", "login"),
			zap.String("ip", "192.168.1.100"),
			zap.Bool("success", true),
			zap.Int("latency_ms", 12),
		},
		apexFields: log.Fields{
			"user_id":    12345,
			"action":     "login",
			"ip":         "192.168.1.100",
			"success":    true,
			"latency_ms": 12,
		},
		logrus: logrus.Fields{
			"user_id":    12345,
			"action":     "login",
			"ip":         "192.168.1.100",
			"success":    true,
			"latency_ms": 12,
		},
		applyZerolog: func(event *zerolog.Event) *zerolog.Event {
			return event.
				Int("user_id", 12345).
				Str("action", "login").
				Str("ip", "192.168.1.100").
				Bool("success", true).
				Int("latency_ms", 12)
		},
	}

	scenarioWithLargeFields = benchmarkScenario{
		name: "WithLargeFields",
		gologTyped: []golog.Field{
			golog.Int("user_id", 12345),
			golog.Str("action", "checkout"),
			golog.Str("ip", "192.168.1.100"),
			golog.Bool("success", true),
			golog.Int("latency_ms", 87),
			golog.Str("request_id", "req-123"),
			golog.Str("trace_id", "trace-abc"),
			golog.Str("service", "payments"),
			golog.Str("region", "eu-west-1"),
			golog.Int("retry", 0),
			golog.Int("bytes_in", 1024),
			golog.Int("bytes_out", 2048),
			golog.Bool("feature_flag_new", true),
		},
		slogArgs: []any{
			"user_id", 12345,
			"action", "checkout",
			"ip", "192.168.1.100",
			"success", true,
			"latency_ms", 87,
			"request_id", "req-123",
			"trace_id", "trace-abc",
			"service", "payments",
			"region", "eu-west-1",
			"retry", 0,
			"bytes_in", 1024,
			"bytes_out", 2048,
			"feature_flag_new", true,
		},
		zapFields: []zap.Field{
			zap.Int("user_id", 12345), zap.String("action", "checkout"), zap.String("ip", "192.168.1.100"),
			zap.Bool("success", true), zap.Int("latency_ms", 87), zap.String("request_id", "req-123"),
			zap.String("trace_id", "trace-abc"), zap.String("service", "payments"), zap.String("region", "eu-west-1"),
			zap.Int("retry", 0), zap.Int("bytes_in", 1024), zap.Int("bytes_out", 2048), zap.Bool("feature_flag_new", true),
		},
		apexFields: log.Fields{
			"user_id": 12345, "action": "checkout", "ip": "192.168.1.100", "success": true, "latency_ms": 87,
			"request_id": "req-123", "trace_id": "trace-abc", "service": "payments", "region": "eu-west-1", "retry": 0,
			"bytes_in": 1024, "bytes_out": 2048, "feature_flag_new": true,
		},
		logrus: logrus.Fields{
			"user_id": 12345, "action": "checkout", "ip": "192.168.1.100", "success": true, "latency_ms": 87,
			"request_id": "req-123", "trace_id": "trace-abc", "service": "payments", "region": "eu-west-1", "retry": 0,
			"bytes_in": 1024, "bytes_out": 2048, "feature_flag_new": true,
		},
		applyZerolog: func(event *zerolog.Event) *zerolog.Event {
			return event.
				Int("user_id", 12345).Str("action", "checkout").Str("ip", "192.168.1.100").Bool("success", true).Int("latency_ms", 87).
				Str("request_id", "req-123").Str("trace_id", "trace-abc").Str("service", "payments").Str("region", "eu-west-1").
				Int("retry", 0).Int("bytes_in", 1024).Int("bytes_out", 2048).Bool("feature_flag_new", true)
		},
	}

	scenarioWithExtraLargeFields = benchmarkScenario{
		name: "WithExtraLargeFields",
		gologTyped: []golog.Field{
			golog.Int("user_id", 12345), golog.Str("action", "checkout"), golog.Str("ip", "192.168.1.100"), golog.Bool("success", true), golog.Int("latency_ms", 87),
			golog.Str("request_id", "req-123"), golog.Str("trace_id", "trace-abc"), golog.Str("service", "payments"), golog.Str("region", "eu-west-1"), golog.Int("retry", 0),
			golog.Int("bytes_in", 1024), golog.Int("bytes_out", 2048), golog.Bool("feature_flag_new", true),
			golog.Int("cart_items", 4), golog.Float64("cart_total", 129.95), golog.Str("currency", "USD"), golog.Str("country", "US"), golog.Str("device", "ios"), golog.Str("app_version", "2.1.0"), golog.Str("experiment", "A"),
		},
		slogArgs: []any{
			"user_id", 12345, "action", "checkout", "ip", "192.168.1.100", "success", true, "latency_ms", 87,
			"request_id", "req-123", "trace_id", "trace-abc", "service", "payments", "region", "eu-west-1", "retry", 0,
			"bytes_in", 1024, "bytes_out", 2048, "feature_flag_new", true,
			"cart_items", 4, "cart_total", 129.95, "currency", "USD", "country", "US", "device", "ios", "app_version", "2.1.0", "experiment", "A",
		},
		zapFields: []zap.Field{
			zap.Int("user_id", 12345), zap.String("action", "checkout"), zap.String("ip", "192.168.1.100"), zap.Bool("success", true), zap.Int("latency_ms", 87),
			zap.String("request_id", "req-123"), zap.String("trace_id", "trace-abc"), zap.String("service", "payments"), zap.String("region", "eu-west-1"), zap.Int("retry", 0),
			zap.Int("bytes_in", 1024), zap.Int("bytes_out", 2048), zap.Bool("feature_flag_new", true),
			zap.Int("cart_items", 4), zap.Float64("cart_total", 129.95), zap.String("currency", "USD"), zap.String("country", "US"), zap.String("device", "ios"), zap.String("app_version", "2.1.0"), zap.String("experiment", "A"),
		},
		apexFields: log.Fields{
			"user_id": 12345, "action": "checkout", "ip": "192.168.1.100", "success": true, "latency_ms": 87,
			"request_id": "req-123", "trace_id": "trace-abc", "service": "payments", "region": "eu-west-1", "retry": 0,
			"bytes_in": 1024, "bytes_out": 2048, "feature_flag_new": true,
			"cart_items": 4, "cart_total": 129.95, "currency": "USD", "country": "US", "device": "ios", "app_version": "2.1.0", "experiment": "A",
		},
		logrus: logrus.Fields{
			"user_id": 12345, "action": "checkout", "ip": "192.168.1.100", "success": true, "latency_ms": 87,
			"request_id": "req-123", "trace_id": "trace-abc", "service": "payments", "region": "eu-west-1", "retry": 0,
			"bytes_in": 1024, "bytes_out": 2048, "feature_flag_new": true,
			"cart_items": 4, "cart_total": 129.95, "currency": "USD", "country": "US", "device": "ios", "app_version": "2.1.0", "experiment": "A",
		},
		applyZerolog: func(event *zerolog.Event) *zerolog.Event {
			return event.
				Int("user_id", 12345).Str("action", "checkout").Str("ip", "192.168.1.100").Bool("success", true).Int("latency_ms", 87).
				Str("request_id", "req-123").Str("trace_id", "trace-abc").Str("service", "payments").Str("region", "eu-west-1").Int("retry", 0).
				Int("bytes_in", 1024).Int("bytes_out", 2048).Bool("feature_flag_new", true).
				Int("cart_items", 4).Float64("cart_total", 129.95).Str("currency", "USD").Str("country", "US").Str("device", "ios").Str("app_version", "2.1.0").Str("experiment", "A")
		},
	}
)

// setupLoggers creates loggers with identical configuration:
// - Timestamp: RFC3339Nano
// - Output: provided writer
// - Level: info
func setupLoggers(output io.Writer) (*golog.JSONLogger, *slog.Logger, zerolog.Logger, *zap.Logger, *log.Logger, *logrus.Logger) {
	gologLogger := golog.NewJSONLoggerWithOptions(golog.WithLevel(golog.InfoLevel), golog.WithOutput(output))

	slogLogger := slog.New(slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String("timestamp", a.Value.Time().UTC().Format(time.RFC3339Nano))
			}
			if a.Key == slog.LevelKey {
				return slog.String("level", a.Value.String())
			}
			if a.Key == slog.MessageKey {
				return slog.String("message", a.Value.String())
			}
			return a
		},
	}))

	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerologLogger := zerolog.New(output).Level(zerolog.InfoLevel).With().Timestamp().Logger()

	zcfg := zap.NewProductionEncoderConfig()
	zcfg.TimeKey = "timestamp"
	zcfg.LevelKey = "level"
	zcfg.MessageKey = "message"
	zcfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format(time.RFC3339Nano))
	}
	zcfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	zcore := zapcore.NewCore(zapcore.NewJSONEncoder(zcfg), zapcore.AddSync(output), zapcore.InfoLevel)
	zapLogger := zap.New(zcore)

	apexLogger := &log.Logger{Handler: json.New(output), Level: log.InfoLevel}

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

func runScenario(b *testing.B, scenario benchmarkScenario, reportAllocs bool) {
	if reportAllocs {
		b.ReportAllocs()
	}

	_, slogLogger, zerologLogger, zapLogger, apexLogger, logrusLogger := setupLoggers(io.Discard)
	gologTypedNoLock := golog.NewJSONLoggerWithOptions(
		golog.WithLevel(golog.InfoLevel),
		golog.WithOutput(io.Discard),
		golog.WithWriteLock(false),
	)
	defer func() {
		_ = zapLogger.Sync()
	}()

	b.Run("Golog", func(b *testing.B) {
		if reportAllocs {
			b.ReportAllocs()
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if scenario.gologTyped == nil {
					gologTypedNoLock.Info("test message")
				} else {
					gologTypedNoLock.Info("test message", scenario.gologTyped...)
				}
			}
		})
	})

	b.Run("Slog", func(b *testing.B) {
		if reportAllocs {
			b.ReportAllocs()
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if scenario.slogArgs == nil {
					slogLogger.Info("test message")
				} else {
					slogLogger.Info("test message", scenario.slogArgs...)
				}
			}
		})
	})

	b.Run("Zerolog", func(b *testing.B) {
		if reportAllocs {
			b.ReportAllocs()
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				e := zerologLogger.Info()
				if scenario.applyZerolog != nil {
					e = scenario.applyZerolog(e)
				}
				e.Msg("test message")
			}
		})
	})

	b.Run("Zap", func(b *testing.B) {
		if reportAllocs {
			b.ReportAllocs()
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if scenario.zapFields == nil {
					zapLogger.Info("test message")
				} else {
					zapLogger.Info("test message", scenario.zapFields...)
				}
			}
		})
	})

	b.Run("Apex", func(b *testing.B) {
		if reportAllocs {
			b.ReportAllocs()
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if scenario.apexFields == nil {
					apexLogger.Info("test message")
				} else {
					apexLogger.WithFields(scenario.apexFields).Info("test message")
				}
			}
		})
	})

	b.Run("Logrus", func(b *testing.B) {
		if reportAllocs {
			b.ReportAllocs()
		}
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if scenario.logrus == nil {
					logrusLogger.Info("test message")
				} else {
					logrusLogger.WithFields(scenario.logrus).Info("test message")
				}
			}
		})
	})
}
