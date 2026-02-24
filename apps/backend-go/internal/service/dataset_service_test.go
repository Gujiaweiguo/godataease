package service

import (
	"context"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"dataease/backend/internal/domain/dataset"
	calcitev1 "dataease/backend/proto/calcite/v1"

	"google.golang.org/grpc"
)

func TestInferSQLVariableDeType(t *testing.T) {
	cases := []struct {
		name     string
		types    []string
		expected int
	}{
		{name: "datetime", types: []string{"DATETIME"}, expected: 1},
		{name: "double", types: []string{"DOUBLE"}, expected: 3},
		{name: "bigint", types: []string{"BIGINT"}, expected: 2},
		{name: "text", types: []string{"TEXT"}, expected: 0},
	}

	for _, tc := range cases {
		if actual := inferSQLVariableDeType(tc.types); actual != tc.expected {
			t.Fatalf("%s expected %d, got %d", tc.name, tc.expected, actual)
		}
	}
}

func TestValidatePreviewSQL(t *testing.T) {
	if err := validatePreviewSQL("SELECT * FROM core_dataset_group"); err != nil {
		t.Fatalf("expected select sql valid, got error: %v", err)
	}

	if err := validatePreviewSQL("INSERT INTO x VALUES (1)"); err == nil {
		t.Fatal("expected insert sql to be rejected")
	}

	if err := validatePreviewSQL("SELECT 1; SELECT 2"); err == nil {
		t.Fatal("expected multi statement sql to be rejected")
	}
}

func TestParseFilterFieldIDs(t *testing.T) {
	ids := parseFilterFieldIDs(" 1,2,2,abc,0,-3, 4 ")
	if len(ids) != 3 {
		t.Fatalf("expected 3 ids, got %d", len(ids))
	}
	if ids[0] != 1 || ids[1] != 2 || ids[2] != 4 {
		t.Fatalf("unexpected ids: %#v", ids)
	}
}

func TestExtractFilterValues(t *testing.T) {
	values := extractFilterValues([]interface{}{"  A  ", "", nil, "A", 100, " 100 "})
	if len(values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(values))
	}
	if values[0] != "A" || values[1] != "100" {
		t.Fatalf("unexpected values: %#v", values)
	}
}

func TestNormalizeEnumValueScientific(t *testing.T) {
	deType := 3
	normalized := normalizeEnumValue("1.23E+3", &deType)
	if normalized == "" {
		t.Fatal("expected non-empty normalized value")
	}
	if strings.ContainsAny(strings.ToUpper(normalized), "E") {
		t.Fatalf("expected non scientific notation, got %s", normalized)
	}
}

type mockCalciteValidateServer struct {
	calcitev1.UnimplementedCalciteServiceServer
	validateCalls int32
}

func (m *mockCalciteValidateServer) ParseSQL(context.Context, *calcitev1.ParseSQLRequest) (*calcitev1.ParseSQLResponse, error) {
	return &calcitev1.ParseSQLResponse{NormalizedSql: "SELECT 1"}, nil
}

func (m *mockCalciteValidateServer) ValidateSQL(context.Context, *calcitev1.ValidateSQLRequest) (*calcitev1.ValidateSQLResponse, error) {
	atomic.AddInt32(&m.validateCalls, 1)
	return &calcitev1.ValidateSQLResponse{Valid: false, Message: "invalid sql"}, nil
}

func startMockCalciteServer(t *testing.T, srv calcitev1.CalciteServiceServer) (string, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}
	grpcServer := grpc.NewServer()
	calcitev1.RegisterCalciteServiceServer(grpcServer, srv)
	go func() {
		_ = grpcServer.Serve(lis)
	}()

	cleanup := func() {
		grpcServer.Stop()
		_ = lis.Close()
	}

	return lis.Addr().String(), cleanup
}

func TestPreviewSQL_ValidateWithCalciteFirstWhenEnabled(t *testing.T) {
	mock := &mockCalciteValidateServer{}
	addr, cleanup := startMockCalciteServer(t, mock)
	defer cleanup()

	svc := NewDatasetService(nil)
	svc.SetCalciteConfig(addr, 2*time.Second, 0)

	_, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "SELECT 1"})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "sql validation failed") {
		t.Fatalf("expected validation failure, got: %v", err)
	}
	if atomic.LoadInt32(&mock.validateCalls) == 0 {
		t.Fatal("expected calcite validate to be called")
	}
}
