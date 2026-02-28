package seatunnel

import (
	"context"
	"testing"
	"time"

	seatunnelv1 "dataease/backend/proto/seatunnel/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{Address: "localhost:9091", Timeout: 60 * time.Second, MaxRetries: 1}
	if cfg.Address != "localhost:9091" {
		t.Fatalf("unexpected address: %s", cfg.Address)
	}
	if cfg.Timeout != 60*time.Second {
		t.Fatalf("unexpected timeout: %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 1 {
		t.Fatalf("unexpected retries: %d", cfg.MaxRetries)
	}
}

func TestClient_Close_Nil(t *testing.T) {
	c := &Client{conn: nil}
	if err := c.Close(); err != nil {
		t.Fatalf("close should not fail: %v", err)
	}
}

func TestClient_SubmitTaskEmptyName(t *testing.T) {
	c := &Client{}
	_, err := c.SubmitTask(context.Background(), &SyncTask{ID: "x"})
	if err == nil {
		t.Fatal("expected empty name error")
	}
}

func TestClient_SubmitTaskSuccess(t *testing.T) {
	c := &Client{
		timeout:    time.Second,
		maxRetries: 0,
		submitFn: func(_ context.Context, _ *seatunnelv1.SubmitTaskRequest, _ ...grpc.CallOption) (*seatunnelv1.SubmitTaskResponse, error) {
			return &seatunnelv1.SubmitTaskResponse{TaskId: "task-1"}, nil
		},
	}
	result, err := c.SubmitTask(context.Background(), &SyncTask{Name: "sync"})
	if err != nil {
		t.Fatalf("SubmitTask failed: %v", err)
	}
	if result != "task-1" {
		t.Fatalf("unexpected task id: %s", result)
	}
}

func TestClient_GetTaskStatusSuccess(t *testing.T) {
	c := &Client{
		timeout:    time.Second,
		maxRetries: 0,
		statusFn: func(_ context.Context, _ *seatunnelv1.GetTaskStatusRequest, _ ...grpc.CallOption) (*seatunnelv1.GetTaskStatusResponse, error) {
			return &seatunnelv1.GetTaskStatusResponse{Task: &seatunnelv1.SyncTask{Id: "task-1", Status: "running", Progress: 60}}, nil
		},
	}
	task, err := c.GetTaskStatus(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("GetTaskStatus failed: %v", err)
	}
	if task.ID != "task-1" || task.Status != StatusRunning || task.Progress != 60 {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestNormalizeStatus(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{input: "", expected: StatusPending},
		{input: "queued", expected: StatusPending},
		{input: "running", expected: StatusRunning},
		{input: "in_progress", expected: StatusRunning},
		{input: "completed", expected: StatusSuccess},
		{input: "finished", expected: StatusSuccess},
		{input: "error", expected: StatusFailed},
		{input: "timed_out", expected: StatusFailed},
		{input: "cancelled", expected: StatusCancelled},
		{input: "canceled", expected: StatusCancelled},
		{input: "unknown_status", expected: StatusPending},
	}

	for _, tc := range cases {
		if actual := NormalizeStatus(tc.input); actual != tc.expected {
			t.Fatalf("NormalizeStatus(%q) expected %q, got %q", tc.input, tc.expected, actual)
		}
	}
}

func TestClient_GetTaskStatusNormalizesStatus(t *testing.T) {
	c := &Client{
		timeout:    time.Second,
		maxRetries: 0,
		statusFn: func(_ context.Context, _ *seatunnelv1.GetTaskStatusRequest, _ ...grpc.CallOption) (*seatunnelv1.GetTaskStatusResponse, error) {
			return &seatunnelv1.GetTaskStatusResponse{Task: &seatunnelv1.SyncTask{Id: "task-1", Status: "completed", Progress: 100}}, nil
		},
	}
	task, err := c.GetTaskStatus(context.Background(), "task-1")
	if err != nil {
		t.Fatalf("GetTaskStatus failed: %v", err)
	}
	if task.Status != StatusSuccess {
		t.Fatalf("expected normalized success status, got %q", task.Status)
	}
}

func TestClient_CancelTaskSuccess(t *testing.T) {
	c := &Client{
		timeout:    time.Second,
		maxRetries: 0,
		cancelFn: func(_ context.Context, _ *seatunnelv1.CancelTaskRequest, _ ...grpc.CallOption) (*seatunnelv1.CancelTaskResponse, error) {
			return &seatunnelv1.CancelTaskResponse{Success: true}, nil
		},
	}
	if err := c.CancelTask(context.Background(), "task-1"); err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}
}

func TestClient_CancelTaskRetryTransientError(t *testing.T) {
	attempt := 0
	c := &Client{
		timeout:    time.Second,
		maxRetries: 1,
		cancelFn: func(_ context.Context, _ *seatunnelv1.CancelTaskRequest, _ ...grpc.CallOption) (*seatunnelv1.CancelTaskResponse, error) {
			attempt++
			if attempt == 1 {
				return nil, status.Error(codes.Unavailable, "temporary unavailable")
			}
			return &seatunnelv1.CancelTaskResponse{Success: true}, nil
		},
	}
	if err := c.CancelTask(context.Background(), "task-1"); err != nil {
		t.Fatalf("CancelTask failed: %v", err)
	}
	if attempt != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempt)
	}
}
