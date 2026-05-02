package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	sugar        *zap.SugaredLogger
	mu           sync.Mutex
)

type Config struct {
	Level  string
	Format string
}

func Init(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{Level: "info", Format: "console"}
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		level,
	)

	globalLogger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = globalLogger.Sugar()

	return nil
}

func ensureInit() {
	mu.Lock()
	defer mu.Unlock()
	if globalLogger == nil {
		if err := Init(nil); err != nil {
			globalLogger = zap.NewNop()
			sugar = globalLogger.Sugar()
		}
	}
	if sugar == nil {
		sugar = globalLogger.Sugar()
	}
}

func L() *zap.Logger {
	mu.Lock()
	l := globalLogger
	mu.Unlock()
	if l == nil {
		ensureInit()
		mu.Lock()
		l = globalLogger
		mu.Unlock()
	}
	return l
}

func S() *zap.SugaredLogger {
	mu.Lock()
	s := sugar
	mu.Unlock()
	if s == nil {
		ensureInit()
		mu.Lock()
		s = sugar
		mu.Unlock()
	}
	return s
}

func Sync() error {
	mu.Lock()
	l := globalLogger
	mu.Unlock()
	if l != nil {
		return l.Sync()
	}
	return nil
}

func Info(msg string, fields ...zap.Field) {
	L().Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	L().Error(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	L().Debug(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	L().Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	L().Fatal(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}
