package calcite

import (
	"context"
	"testing"
	"time"

	calcitev1 "dataease/backend/proto/calcite/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{Address: "localhost:9090", Timeout: 30 * time.Second, MaxRetries: 1}
	if cfg.Address != "localhost:9090" {
		t.Fatalf("unexpected address: %s", cfg.Address)
	}
	if cfg.Timeout != 30*time.Second {
		t.Fatalf("unexpected timeout: %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 1 {
		t.Fatalf("unexpected max retries: %d", cfg.MaxRetries)
	}
}

func TestClient_Close_Nil(t *testing.T) {
	c := &Client{conn: nil}
	if err := c.Close(); err != nil {
		t.Fatalf("close should not fail: %v", err)
	}
}

func TestClient_ParseSQLEmpty(t *testing.T) {
	c := &Client{}
	_, err := c.ParseSQL(context.Background(), " ")
	if err == nil {
		t.Fatal("expected error for empty sql")
	}
}

func TestClient_ParseSQLSuccess(t *testing.T) {
	c := &Client{
		timeout:    time.Second,
		maxRetries: 0,
		parseFn: func(_ context.Context, _ *calcitev1.ParseSQLRequest, _ ...grpc.CallOption) (*calcitev1.ParseSQLResponse, error) {
			return &calcitev1.ParseSQLResponse{NormalizedSql: "SELECT 1"}, nil
		},
	}

	result, err := c.ParseSQL(context.Background(), "select 1")
	if err != nil {
		t.Fatalf("ParseSQL failed: %v", err)
	}
	if result != "SELECT 1" {
		t.Fatalf("unexpected parse result: %s", result)
	}
}

func TestClient_ValidateSQLSuccess(t *testing.T) {
	c := &Client{
		timeout:    time.Second,
		maxRetries: 0,
		validateFn: func(_ context.Context, _ *calcitev1.ValidateSQLRequest, _ ...grpc.CallOption) (*calcitev1.ValidateSQLResponse, error) {
			return &calcitev1.ValidateSQLResponse{Valid: true}, nil
		},
	}

	valid, err := c.ValidateSQL(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("ValidateSQL failed: %v", err)
	}
	if !valid {
		t.Fatal("expected sql to be valid")
	}
}

func TestClient_ValidateSQLInvalidArgument(t *testing.T) {
	c := &Client{
		timeout:    time.Second,
		maxRetries: 0,
		validateFn: func(_ context.Context, _ *calcitev1.ValidateSQLRequest, _ ...grpc.CallOption) (*calcitev1.ValidateSQLResponse, error) {
			return nil, status.Error(codes.InvalidArgument, "invalid sql")
		},
	}

	valid, err := c.ValidateSQL(context.Background(), "bad")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if valid {
		t.Fatal("expected invalid result")
	}
}

func TestClient_ValidateSQLRetryTransientError(t *testing.T) {
	attempt := 0
	c := &Client{
		timeout:    time.Second,
		maxRetries: 1,
		validateFn: func(_ context.Context, _ *calcitev1.ValidateSQLRequest, _ ...grpc.CallOption) (*calcitev1.ValidateSQLResponse, error) {
			attempt++
			if attempt == 1 {
				return nil, status.Error(codes.Unavailable, "temporary unavailable")
			}
			return &calcitev1.ValidateSQLResponse{Valid: true}, nil
		},
	}

	valid, err := c.ValidateSQL(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("ValidateSQL failed: %v", err)
	}
	if !valid {
		t.Fatal("expected valid sql")
	}
	if attempt != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempt)
	}
}
