package seatunnel

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	conn    *grpc.ClientConn
	address string
	timeout time.Duration
}

type Config struct {
	Address string
	Timeout time.Duration
}

type SyncTask struct {
	ID       string
	Name     string
	Source   string
	Target   string
	Status   string
	Progress int
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}

	conn, err := grpc.Dial(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(cfg.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to seatunnel service: %w", err)
	}

	return &Client{
		conn:    conn,
		address: cfg.Address,
		timeout: cfg.Timeout,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) SubmitTask(ctx context.Context, task *SyncTask) (string, error) {
	return "", fmt.Errorf("not implemented: SubmitTask")
}

func (c *Client) GetTaskStatus(ctx context.Context, taskID string) (*SyncTask, error) {
	return nil, fmt.Errorf("not implemented: GetTaskStatus")
}

func (c *Client) CancelTask(ctx context.Context, taskID string) error {
	return fmt.Errorf("not implemented: CancelTask")
}
