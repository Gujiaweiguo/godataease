package service

import (
	"context"
	"encoding/base64"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"dataease/backend/internal/domain/dataset"
	calcitev1 "dataease/backend/proto/calcite/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type mockCalciteValidateErrorServer struct {
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

func (m *mockCalciteValidateErrorServer) ParseSQL(context.Context, *calcitev1.ParseSQLRequest) (*calcitev1.ParseSQLResponse, error) {
	return &calcitev1.ParseSQLResponse{NormalizedSql: "SELECT 1"}, nil
}

func (m *mockCalciteValidateErrorServer) ValidateSQL(context.Context, *calcitev1.ValidateSQLRequest) (*calcitev1.ValidateSQLResponse, error) {
	atomic.AddInt32(&m.validateCalls, 1)
	return nil, status.Error(codes.Unavailable, "calcite unavailable")
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

func TestPreviewSQL_BlockExecutionWhenCalciteUnavailable(t *testing.T) {
	mock := &mockCalciteValidateErrorServer{}
	addr, cleanup := startMockCalciteServer(t, mock)
	defer cleanup()

	svc := NewDatasetService(nil)
	svc.SetCalciteConfig(addr, 2*time.Second, 0)

	_, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "SELECT 1"})
	if err == nil {
		t.Fatal("expected calcite error")
	}
	if !strings.Contains(err.Error(), "calcite validate sql failed") {
		t.Fatalf("expected calcite failure, got: %v", err)
	}
	if atomic.LoadInt32(&mock.validateCalls) == 0 {
		t.Fatal("expected calcite validate to be called")
	}
}

func TestNewDatasetServiceWithPermission(t *testing.T) {
	svc := NewDatasetServiceWithPermission(nil, &RowPermissionService{})
	if svc == nil {
		t.Fatal("expected service instance")
	}
	if svc.rowPermissionService == nil {
		t.Fatal("expected row permission service configured")
	}
}

func TestEnumAliasAndFieldIDFromAlias(t *testing.T) {
	alias := enumAlias(123)
	if alias != "f_123" {
		t.Fatalf("unexpected alias: %s", alias)
	}
	if got := enumFieldIDFromAlias(alias); got != 123 {
		t.Fatalf("expected 123, got %d", got)
	}
	if got := enumFieldIDFromAlias("bad_alias"); got != 0 {
		t.Fatalf("expected 0 for invalid alias, got %d", got)
	}
}

func TestBuildPreviewFieldsAndNormalizeRow(t *testing.T) {
	now := time.Date(2026, 3, 3, 8, 0, 0, 0, time.UTC)
	rows := []map[string]interface{}{
		{"b": []byte("abc"), "a": now, "c": "123"},
	}
	fields := buildPreviewFields(rows)
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
	if fields[0].OriginName != "a" || fields[0].DeType != 1 {
		t.Fatalf("unexpected first field: %+v", fields[0])
	}
	if fields[1].OriginName != "b" || fields[1].DeType != 0 {
		t.Fatalf("unexpected second field: %+v", fields[1])
	}
	if fields[2].OriginName != "c" || fields[2].DeType != 2 {
		t.Fatalf("unexpected third field: %+v", fields[2])
	}

	normalized := normalizePreviewRow(rows[0])
	if normalized["b"] != "abc" {
		t.Fatalf("expected []byte converted to string, got %#v", normalized["b"])
	}
	if normalized["a"] != "2026-03-03 08:00:00" {
		t.Fatalf("unexpected time format: %#v", normalized["a"])
	}

	empty := buildPreviewFields([]map[string]interface{}{})
	if len(empty) != 0 {
		t.Fatalf("expected empty fields for empty rows, got %d", len(empty))
	}
}

func TestInferPreviewDeTypeAndDateTimeText(t *testing.T) {
	if inferPreviewDeType(true) != 4 {
		t.Fatal("bool should infer to 4")
	}
	if inferPreviewDeType(int64(1)) != 2 {
		t.Fatal("int should infer to 2")
	}
	if inferPreviewDeType(1.23) != 3 {
		t.Fatal("float should infer to 3")
	}
	if inferPreviewDeType("2026-03-03") != 1 {
		t.Fatal("date text should infer to 1")
	}
	if inferPreviewDeType("45.67") != 3 {
		t.Fatal("float text should infer to 3")
	}
	if inferPreviewDeType("x") != 0 {
		t.Fatal("plain text should infer to 0")
	}

	if !isDateTimeText("2026-03-03T08:00:00Z") {
		t.Fatal("rfc3339 text should be datetime")
	}
	if !isDateTimeText("2026/03/03") {
		t.Fatal("slash date text should be datetime")
	}
	if isDateTimeText("not-a-date") {
		t.Fatal("invalid datetime text should be false")
	}
}

func TestPreviewSQL_EdgeCases(t *testing.T) {
	svc := NewDatasetService(nil)

	// Test nil request
	result, err := svc.PreviewSQL(nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result["sql"])

	// Test empty SQL
	result, err = svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: ""})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result["sql"])

	// Test whitespace only SQL
	result, err = svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "   "})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result["sql"])
}

func TestPreviewSQL_ValidateError(t *testing.T) {
	svc := NewDatasetService(nil)

	// Test invalid SQL (INSERT)
	_, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "INSERT INTO x VALUES (1)"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "select")

	// Test multi-statement SQL
	_, err = svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "SELECT 1; SELECT 2"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "single")
}

func TestPreviewSQL_Base64DecodedEmpty(t *testing.T) {
	svc := NewDatasetService(nil)

	// Test base64 encoded empty string
	encodedEmpty := base64.StdEncoding.EncodeToString([]byte("   "))
	result, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: encodedEmpty})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result["sql"])
}

func TestSetCalciteConfig_DefaultPreserve(t *testing.T) {
	svc := NewDatasetService(nil)
	defaultTimeout := svc.calciteTimeout
	defaultRetries := svc.calciteRetries

	svc.SetCalciteConfig(" 127.0.0.1:7001 ", 0, -1)

	assert.Equal(t, "127.0.0.1:7001", svc.calciteAddress)
	assert.Equal(t, defaultTimeout, svc.calciteTimeout)
	assert.Equal(t, defaultRetries, svc.calciteRetries)

	svc.SetCalciteConfig("127.0.0.1:7002", 3*time.Second, 2)
	assert.Equal(t, 3*time.Second, svc.calciteTimeout)
	assert.Equal(t, 2, svc.calciteRetries)
}

func TestDatasetService_EarlyValidation(t *testing.T) {
	svc := NewDatasetService(nil)

	_, err := svc.Save(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request is required")

	_, err = svc.Create(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "request is required")

	_, err = svc.Rename(0, "name")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")

	_, err = svc.Move(0, 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")

	_, err = svc.Move(10, 10)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be itself")

	err = svc.Delete(0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "id is required")
}

func TestDatasetService_NormalizedHelpers(t *testing.T) {
	assert.Equal(t, int64(0), normalizedDatasetPID(nil))
	neg := int64(-3)
	assert.Equal(t, int64(0), normalizedDatasetPID(&neg))
	pos := int64(7)
	assert.Equal(t, int64(7), normalizedDatasetPID(&pos))

	assert.Equal(t, "", normalizedDatasetNodeType(""))
	assert.Equal(t, dataset.NodeTypeFolder, normalizedDatasetNodeType(dataset.NodeTypeFolder))
	assert.Equal(t, dataset.NodeTypeDataset, normalizedDatasetNodeType(dataset.NodeTypeDataset))
	assert.Equal(t, dataset.NodeTypeDataset, normalizedDatasetNodeType("unknown"))
}
