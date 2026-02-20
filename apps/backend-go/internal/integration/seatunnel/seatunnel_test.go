package seatunnel

import (
	"context"
	"testing"
	"time"
)

func TestConfig_Fields(t *testing.T) {
	cfg := &Config{
		Address: "localhost:9091",
		Timeout: 60 * time.Second,
	}

	if cfg.Address != "localhost:9091" {
		t.Errorf("Expected Address 'localhost:9091', got '%s'", cfg.Address)
	}
	if cfg.Timeout != 60*time.Second {
		t.Errorf("Expected Timeout 60s, got %v", cfg.Timeout)
	}
}

func TestSyncTask_Fields(t *testing.T) {
	task := &SyncTask{
		ID:       "task-1",
		Name:     "MySQL to PostgreSQL",
		Source:   "mysql://localhost:3306/db",
		Target:   "postgresql://localhost:5432/db",
		Status:   "running",
		Progress: 50,
	}

	if task.ID != "task-1" {
		t.Errorf("Expected ID 'task-1', got '%s'", task.ID)
	}
	if task.Name != "MySQL to PostgreSQL" {
		t.Errorf("Expected Name 'MySQL to PostgreSQL', got '%s'", task.Name)
	}
	if task.Progress != 50 {
		t.Errorf("Expected Progress 50, got %d", task.Progress)
	}
}

func TestClient_Close_Nil(t *testing.T) {
	c := &Client{conn: nil}
	err := c.Close()
	if err != nil {
		t.Errorf("Expected no error on close with nil conn, got: %v", err)
	}
}

func TestClient_SubmitTask_NotImplemented(t *testing.T) {
	c := &Client{}
	ctx := context.Background()
	task := &SyncTask{ID: "test"}

	_, err := c.SubmitTask(ctx, task)
	if err == nil {
		t.Error("Expected error for unimplemented SubmitTask")
	}
}

func TestClient_GetTaskStatus_NotImplemented(t *testing.T) {
	c := &Client{}
	ctx := context.Background()

	_, err := c.GetTaskStatus(ctx, "task-1")
	if err == nil {
		t.Error("Expected error for unimplemented GetTaskStatus")
	}
}

func TestClient_CancelTask_NotImplemented(t *testing.T) {
	c := &Client{}
	ctx := context.Background()

	err := c.CancelTask(ctx, "task-1")
	if err == nil {
		t.Error("Expected error for unimplemented CancelTask")
	}
}
