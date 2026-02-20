package calcite

import (
	"context"
	"testing"
	"time"
)

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		Address: "localhost:9090",
		Timeout: 30 * time.Second,
	}

	if cfg.Address != "localhost:9090" {
		t.Errorf("Expected Address 'localhost:9090', got '%s'", cfg.Address)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Expected Timeout 30s, got %v", cfg.Timeout)
	}
}

func TestConfig_DefaultTimeout(t *testing.T) {
	cfg := &Config{
		Address: "localhost:9090",
		Timeout: 0,
	}

	if cfg.Timeout != 0 {
		t.Errorf("Expected Timeout 0, got %v", cfg.Timeout)
	}
}

func TestClient_Close_Nil(t *testing.T) {
	c := &Client{conn: nil}
	err := c.Close()
	if err != nil {
		t.Errorf("Expected no error on close with nil conn, got: %v", err)
	}
}

func TestClient_ParseSQL_NotImplemented(t *testing.T) {
	c := &Client{}
	ctx := context.Background()

	_, err := c.ParseSQL(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for unimplemented ParseSQL")
	}
}

func TestClient_ValidateSQL_NotImplemented(t *testing.T) {
	c := &Client{}
	ctx := context.Background()

	_, err := c.ValidateSQL(ctx, "SELECT 1")
	if err == nil {
		t.Error("Expected error for unimplemented ValidateSQL")
	}
}
