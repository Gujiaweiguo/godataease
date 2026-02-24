package calcite

import (
	"context"
	"fmt"
	"strings"
	"time"

	calcitev1 "dataease/backend/proto/calcite/v1"

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

	rpc        calcitev1.CalciteServiceClient
	parseFn    func(context.Context, *calcitev1.ParseSQLRequest, ...grpc.CallOption) (*calcitev1.ParseSQLResponse, error)
	validateFn func(context.Context, *calcitev1.ValidateSQLRequest, ...grpc.CallOption) (*calcitev1.ValidateSQLResponse, error)
}

type Config struct {
	Address    string
	Timeout    time.Duration
	MaxRetries int
}

func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("calcite config is required")
	}
	if strings.TrimSpace(cfg.Address) == "" {
		return nil, fmt.Errorf("calcite address is required")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries < 0 {
		cfg.MaxRetries = 0
	}

	dialCtx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		cfg.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to calcite service: %w", err)
	}

	rpc := calcitev1.NewCalciteServiceClient(conn)
	return &Client{
		conn:       conn,
		address:    cfg.Address,
		timeout:    cfg.Timeout,
		maxRetries: cfg.MaxRetries,
		rpc:        rpc,
		parseFn:    rpc.ParseSQL,
		validateFn: rpc.ValidateSQL,
	}, nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *Client) ParseSQL(ctx context.Context, sql string) (string, error) {
	text := strings.TrimSpace(sql)
	if text == "" {
		return "", fmt.Errorf("sql is required")
	}

	resp, err := c.callParseWithRetry(ctx, &calcitev1.ParseSQLRequest{Sql: text})
	if err != nil {
		return "", fmt.Errorf("calcite parse failed: %w", err)
	}
	if resp == nil {
		return "", fmt.Errorf("calcite parse response is empty")
	}

	normalized := strings.TrimSpace(resp.GetNormalizedSql())
	if normalized != "" {
		return normalized, nil
	}
	raw := strings.TrimSpace(resp.GetRawSql())
	if raw != "" {
		return raw, nil
	}
	return text, nil
}

func (c *Client) ValidateSQL(ctx context.Context, sql string) (bool, error) {
	text := strings.TrimSpace(sql)
	if text == "" {
		return false, fmt.Errorf("sql is required")
	}

	resp, err := c.callValidateWithRetry(ctx, &calcitev1.ValidateSQLRequest{Sql: text})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.InvalidArgument {
			return false, nil
		}
		return false, fmt.Errorf("calcite validate failed: %w", err)
	}
	if resp == nil {
		return false, fmt.Errorf("calcite validate response is empty")
	}
	if resp.GetValid() {
		return true, nil
	}
	if strings.TrimSpace(resp.GetMessage()) != "" {
		return false, nil
	}
	return false, nil
}

func (c *Client) callParseWithRetry(ctx context.Context, req *calcitev1.ParseSQLRequest) (*calcitev1.ParseSQLResponse, error) {
	attempts := c.maxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.callParseWithTimeout(ctx, req)
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

func (c *Client) callValidateWithRetry(ctx context.Context, req *calcitev1.ValidateSQLRequest) (*calcitev1.ValidateSQLResponse, error) {
	attempts := c.maxRetries + 1
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		resp, err := c.callValidateWithTimeout(ctx, req)
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

func (c *Client) callParseWithTimeout(ctx context.Context, req *calcitev1.ParseSQLRequest) (*calcitev1.ParseSQLResponse, error) {
	if c == nil || c.parseFn == nil {
		return nil, fmt.Errorf("calcite client not initialized")
	}
	callCtx, cancel := c.withTimeoutContext(ctx)
	defer cancel()
	return c.parseFn(callCtx, req)
}

func (c *Client) callValidateWithTimeout(ctx context.Context, req *calcitev1.ValidateSQLRequest) (*calcitev1.ValidateSQLResponse, error) {
	if c == nil || c.validateFn == nil {
		return nil, fmt.Errorf("calcite client not initialized")
	}
	callCtx, cancel := c.withTimeoutContext(ctx)
	defer cancel()
	return c.validateFn(callCtx, req)
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
