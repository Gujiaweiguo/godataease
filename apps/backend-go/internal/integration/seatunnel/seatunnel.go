package seatunnel

import (
	"context"
	"dataease/backend/internal/pkg/errno"
	"fmt"
	"strings"
	"time"

	seatunnelv1 "dataease/backend/proto/seatunnel/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	conn       *grpc.ClientConn
	address    string
	timeout    time.Duration
	maxRetries int

	rpc      seatunnelv1.SyncServiceClient
	submitFn func(context.Context, *seatunnelv1.SubmitTaskRequest, ...grpc.CallOption) (*seatunnelv1.SubmitTaskResponse, error)
	statusFn func(context.Context, *seatunnelv1.GetTaskStatusRequest, ...grpc.CallOption) (*seatunnelv1.GetTaskStatusResponse, error)
	cancelFn func(context.Context, *seatunnelv1.CancelTaskRequest, ...grpc.CallOption) (*seatunnelv1.CancelTaskResponse, error)
}

type Config struct {
	Address    string
	Timeout    time.Duration
	MaxRetries int
}

type SyncTask struct {
	ID       string
	Name     string
	Source   string
	Target   string
	Status   string
	Progress int
}

const (
	StatusPending   = "pending"
	StatusRunning   = "running"
	StatusSuccess   = "success"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
)

func NormalizeStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "created", "submitted", "queued", "waiting", "scheduled", "initializing", "initialized", "pending":
		return StatusPending
	case "running", "in_progress", "processing", "executing", "active":
		return StatusRunning
	case "success", "succeeded", "completed", "finished", "done":
		return StatusSuccess
	case "failed", "error", "errored", "timeout", "timed_out", "terminated":
		return StatusFailed
	case "cancelled", "canceled", "aborted", "abort", "stopped":
		return StatusCancelled
	default:
		return StatusPending
	}
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("seatunnel config is required")
	}
	if strings.TrimSpace(cfg.Address) == "" {
		return nil, fmt.Errorf("seatunnel address is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := grpc.DialContext( //nolint:staticcheck // grpc.NewClient migration requires larger refactor
		dialCtx,
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(), //nolint:staticcheck // grpc.NewClient migration requires larger refactor
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to seatunnel service: %w", err)
	}

	rpc := seatunnelv1.NewSyncServiceClient(conn)
	return &Client{
		conn:       conn,
		address:    cfg.Address,
		timeout:    cfg.Timeout,
		maxRetries: cfg.MaxRetries,
		rpc:        rpc,
		submitFn:   rpc.SubmitTask,
		statusFn:   rpc.GetTaskStatus,
		cancelFn:   rpc.CancelTask,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) SubmitTask(ctx context.Context, task *SyncTask) (string, error) {
	if task == nil {
		return "", fmt.Errorf("sync task is required")
	}
	if strings.TrimSpace(task.Name) == "" {
		return "", fmt.Errorf("sync task name is required")
	}

	resp, err := c.callSubmitWithRetry(ctx, &seatunnelv1.SubmitTaskRequest{Task: &seatunnelv1.SyncTask{
		Id:       task.ID,
		Name:     task.Name,
		Source:   task.Source,
		Target:   task.Target,
		Status:   NormalizeStatus(task.Status),
		Progress: int32(task.Progress),
	}})
	if err != nil {
		return "", fmt.Errorf("seatunnel submit task failed: %w", err)
	}
	if resp == nil || strings.TrimSpace(resp.GetTaskId()) == "" {
		return "", fmt.Errorf("seatunnel submit response missing taskId")
	}
	return resp.GetTaskId(), nil
}

func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*SyncTask, error) {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return nil, fmt.Errorf("task id is required")
	}

	resp, err := c.callStatusWithRetry(ctx, &seatunnelv1.GetTaskStatusRequest{TaskId: trimmedTaskID})
	if err != nil {
		return nil, fmt.Errorf("seatunnel get task status failed: %w", err)
	}
	if resp == nil || resp.GetTask() == nil {
		return nil, fmt.Errorf("seatunnel status response is empty")
	}

	item := resp.GetTask()
	return &SyncTask{
		ID:       item.GetId(),
		Name:     item.GetName(),
		Source:   item.GetSource(),
		Target:   item.GetTarget(),
		Status:   NormalizeStatus(item.GetStatus()),
		Progress: int(item.GetProgress()),
	}, nil
}

func (c *Client) CancelTask(ctx context.Context, taskID string) error {
	trimmedTaskID := strings.TrimSpace(taskID)
	if trimmedTaskID == "" {
		return fmt.Errorf("task id is required")
	}

	resp, err := c.callCancelWithRetry(ctx, &seatunnelv1.CancelTaskRequest{TaskId: trimmedTaskID})
	if err != nil {
		return fmt.Errorf("seatunnel cancel task failed: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("seatunnel cancel response is empty")
	}
	if !resp.GetSuccess() {
		if msg := strings.TrimSpace(resp.GetMessage()); msg != "" {
			return fmt.Errorf("seatunnel cancel task failed: %s", msg)
		}
		return fmt.Errorf("seatunnel cancel task failed")
	}
	return nil
}

func (c *Client) callSubmitWithRetry(ctx context.Context, req *seatunnelv1.SubmitTaskRequest) (*seatunnelv1.SubmitTaskResponse, error) {
	attempts := c.maxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.callSubmitWithTimeout(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetriable(err) || attempt == attempts-1 {
			return nil, err
		}
		time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
	}

	return nil, lastErr
}

func (c *Client) callStatusWithRetry(ctx context.Context, req *seatunnelv1.GetTaskStatusRequest) (*seatunnelv1.GetTaskStatusResponse, error) {
	attempts := c.maxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.callStatusWithTimeout(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetriable(err) || attempt == attempts-1 {
			return nil, err
		}
		time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
	}

	return nil, lastErr
}

func (c *Client) callCancelWithRetry(ctx context.Context, req *seatunnelv1.CancelTaskRequest) (*seatunnelv1.CancelTaskResponse, error) {
	attempts := c.maxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.callCancelWithTimeout(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetriable(err) || attempt == attempts-1 {
			return nil, err
		}
		time.Sleep(time.Duration(attempt+1) * 50 * time.Millisecond)
	}

	return nil, lastErr
}

func (c *Client) callSubmitWithTimeout(ctx context.Context, req *seatunnelv1.SubmitTaskRequest) (*seatunnelv1.SubmitTaskResponse, error) {
	if c == nil || c.submitFn == nil {
		return nil, fmt.Errorf(errno.ErrSeatunnelNotInitialized)
	}
	callCtx, cancel := c.withTimeoutContext(ctx)
	defer cancel()
	return c.submitFn(callCtx, req)
}

func (c *Client) callStatusWithTimeout(ctx context.Context, req *seatunnelv1.GetTaskStatusRequest) (*seatunnelv1.GetTaskStatusResponse, error) {
	if c == nil || c.statusFn == nil {
		return nil, fmt.Errorf(errno.ErrSeatunnelNotInitialized)
	}
	callCtx, cancel := c.withTimeoutContext(ctx)
	defer cancel()
	return c.statusFn(callCtx, req)
}

func (c *Client) callCancelWithTimeout(ctx context.Context, req *seatunnelv1.CancelTaskRequest) (*seatunnelv1.CancelTaskResponse, error) {
	if c == nil || c.cancelFn == nil {
		return nil, fmt.Errorf(errno.ErrSeatunnelNotInitialized)
	}
	callCtx, cancel := c.withTimeoutContext(ctx)
	defer cancel()
	return c.cancelFn(callCtx, req)
}

func (c *Client) withTimeoutContext(ctx context.Context) (context.Context, context.CancelFunc) {
	callCtx := ctx
	if callCtx == nil {
		callCtx = context.Background()
	}
	if _, ok := callCtx.Deadline(); !ok && c.timeout > 0 {
		return context.WithTimeout(callCtx, c.timeout)
	}
	return callCtx, func() {}
}

func isRetriable(err error) bool {
	if err == nil {
		return false
	}
	if err == context.DeadlineExceeded {
		return true
	}
	st, ok := status.FromError(err)
	if !ok {
		return false
	}
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}
