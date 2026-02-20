package logger

import (
	"testing"

	"go.uber.org/zap"
)

func TestInit_DefaultConfig(t *testing.T) {
	err := Init(nil)
	if err != nil {
		t.Fatalf("Init with nil config failed: %v", err)
	}

	if globalLogger == nil {
		t.Error("Expected globalLogger to be initialized")
	}
	if sugar == nil {
		t.Error("Expected sugar to be initialized")
	}
}

func TestInit_ConsoleFormat(t *testing.T) {
	cfg := &Config{Level: "info", Format: "console"}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init with console format failed: %v", err)
	}
}

func TestInit_JsonFormat(t *testing.T) {
	cfg := &Config{Level: "info", Format: "json"}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init with json format failed: %v", err)
	}
}

func TestInit_DebugLevel(t *testing.T) {
	cfg := &Config{Level: "debug", Format: "console"}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init with debug level failed: %v", err)
	}
}

func TestInit_ErrorLevel(t *testing.T) {
	cfg := &Config{Level: "error", Format: "console"}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init with error level failed: %v", err)
	}
}

func TestInit_InvalidLevel(t *testing.T) {
	cfg := &Config{Level: "invalid", Format: "console"}
	err := Init(cfg)
	if err != nil {
		t.Fatalf("Init should not fail with invalid level, it should default to info: %v", err)
	}
}

func TestL(t *testing.T) {
	logger := L()
	if logger == nil {
		t.Fatal("Expected logger instance, got nil")
	}
}

func TestL_AutoInit(t *testing.T) {
	globalLogger = nil
	logger := L()
	if logger == nil {
		t.Fatal("L() should auto-initialize if not already initialized")
	}
}

func TestS(t *testing.T) {
	s := S()
	if s == nil {
		t.Fatal("Expected sugared logger instance, got nil")
	}
}

func TestS_AutoInit(t *testing.T) {
	sugar = nil
	s := S()
	if s == nil {
		t.Fatal("S() should auto-initialize if not already initialized")
	}
}

func TestSync(t *testing.T) {
	Init(nil)
	_ = Sync()
}

func TestSync_NilLogger(t *testing.T) {
	globalLogger = nil
	err := Sync()
	if err != nil {
		t.Errorf("Sync with nil logger should return nil, got: %v", err)
	}
}

func TestInfo(t *testing.T) {
	Init(nil)
	Info("test info message", zap.String("key", "value"))
}

func TestError(t *testing.T) {
	Init(nil)
	Error("test error message", zap.String("key", "value"))
}

func TestDebug(t *testing.T) {
	cfg := &Config{Level: "debug", Format: "console"}
	Init(cfg)
	Debug("test debug message", zap.String("key", "value"))
}

func TestWarn(t *testing.T) {
	Init(nil)
	Warn("test warn message", zap.String("key", "value"))
}

func TestWith(t *testing.T) {
	Init(nil)
	logger := With(zap.String("service", "test"))
	if logger == nil {
		t.Fatal("Expected logger from With(), got nil")
	}
}
