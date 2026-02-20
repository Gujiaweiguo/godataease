package trace

import (
	"context"
	"testing"
)

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		Enabled:  true,
		Endpoint: "localhost:4317",
		Service:  "dataease-backend",
	}

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if cfg.Endpoint != "localhost:4317" {
		t.Errorf("Expected Endpoint 'localhost:4317', got '%s'", cfg.Endpoint)
	}
	if cfg.Service != "dataease-backend" {
		t.Errorf("Expected Service 'dataease-backend', got '%s'", cfg.Service)
	}
}

func TestInit_Disabled(t *testing.T) {
	cfg := &Config{
		Enabled:  false,
		Endpoint: "",
		Service:  "test",
	}

	shutdown, err := Init(cfg)
	if err != nil {
		t.Errorf("Init with disabled config should not fail: %v", err)
	}
	if shutdown == nil {
		t.Error("Expected shutdown function to be returned")
	}

	_ = shutdown(context.Background())
}

func TestInit_EmptyEndpoint(t *testing.T) {
	cfg := &Config{
		Enabled:  true,
		Endpoint: "",
		Service:  "test",
	}

	shutdown, err := Init(cfg)
	if err != nil {
		t.Errorf("Init with empty endpoint should not fail: %v", err)
	}
	if shutdown == nil {
		t.Error("Expected shutdown function to be returned")
	}
}

func TestStartSpan(t *testing.T) {
	ctx := context.Background()
	newCtx, end := StartSpan(ctx, "test-operation")

	if newCtx == nil {
		t.Error("Expected new context to be returned")
	}
	if end == nil {
		t.Error("Expected end function to be returned")
	}

	end()
}
