package calcite

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

func NewClient(cfg *Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	conn, err := grpc.Dial(
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(cfg.Timeout),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to calcite service: %w", err)
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

func (c *Client) ParseSQL(ctx context.Context, sql string) (string, error) {
	return "", fmt.Errorf("not implemented: ParseSQL")
}

func (c *Client) ValidateSQL(ctx context.Context, sql string) (bool, error) {
	return false, fmt.Errorf("not implemented: ValidateSQL")
}
