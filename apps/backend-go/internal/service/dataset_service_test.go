package service

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"
	calcitev1 "dataease/backend/proto/calcite/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDatasetServiceRepoTest(t *testing.T) (*DatasetService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}, &datasource.CoreDatasource{}))

	repo := repository.NewDatasetRepository(db)
	return NewDatasetService(repo), db
}

type stubPreviewExecutor struct {
	rows []map[string]interface{}
	err  error

	called bool
	rawSQL string
	limit  int
	closed bool
}

func (s *stubPreviewExecutor) PreviewSQL(_ context.Context, rawSQL string, limit int) ([]map[string]interface{}, error) {
	s.called = true
	s.rawSQL = rawSQL
	s.limit = limit
	if s.err != nil {
		return nil, s.err
	}
	return s.rows, nil
}

func (s *stubPreviewExecutor) Close() error {
	s.closed = true
	return nil
}

type datasetAdminChecker struct{ isAdmin bool }

func (c datasetAdminChecker) IsAdmin(int64) bool { return c.isAdmin }

func TestInferSQLVariableDeType(t *testing.T) {
	cases := []struct {
		name     string
		types    []string
		expected int
	}{
		{name: "empty", types: []string{}, expected: 0},
		{name: "blank", types: []string{"   "}, expected: 0},
		{name: "datetime", types: []string{"DATETIME"}, expected: 1},
		{name: "timestamp", types: []string{"timestamp"}, expected: 1},
		{name: "date", types: []string{"date"}, expected: 1},
		{name: "year", types: []string{"year"}, expected: 1},
		{name: "double", types: []string{"DOUBLE"}, expected: 3},
		{name: "decimal", types: []string{"decimal(10,2)"}, expected: 3},
		{name: "numeric", types: []string{"numeric"}, expected: 3},
		{name: "bigint", types: []string{"BIGINT"}, expected: 2},
		{name: "smallint", types: []string{"smallint"}, expected: 2},
		{name: "boolean", types: []string{"boolean"}, expected: 4},
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

	if err := validatePreviewSQL("SELECT 1 UNION SELECT 2"); err == nil {
		t.Fatal("expected union sql to be rejected")
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

type mockCalciteValidateSuccessServer struct {
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

func (m *mockCalciteValidateSuccessServer) ParseSQL(context.Context, *calcitev1.ParseSQLRequest) (*calcitev1.ParseSQLResponse, error) {
	return &calcitev1.ParseSQLResponse{NormalizedSql: "SELECT 1"}, nil
}

func (m *mockCalciteValidateSuccessServer) ValidateSQL(context.Context, *calcitev1.ValidateSQLRequest) (*calcitev1.ValidateSQLResponse, error) {
	atomic.AddInt32(&m.validateCalls, 1)
	return &calcitev1.ValidateSQLResponse{Valid: true}, nil
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

func TestPreviewSQL_ExternalDatasourceUnsupported(t *testing.T) {
	svc, db := setupDatasetServiceRepoTest(t)
	datasourceRepo := repository.NewDatasourceRepository(db)
	svc.SetDatasourceRepository(datasourceRepo)
	config := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "db.local", Port: 5432, Database: "analytics", Username: "pg", Password: "secret"})
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 42, Name: "pg-ds", Type: "pg", Configuration: &config}).Error)

	_, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{
		SQL:          base64.StdEncoding.EncodeToString([]byte("SELECT 1")),
		DatasourceID: 42,
	}, 9)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrPreviewSQLExternalDatasourceUnsupported)
}

func TestBuildMySQLPreviewConfig(t *testing.T) {
	cfg, err := buildMySQLPreviewConfig(&datasource.ConnectionConfig{
		Host:        "db.local",
		Port:        3306,
		Database:    "analytics",
		Username:    "root",
		Password:    "secret",
		ExtraParams: "useSSL=false",
	})
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "root", cfg.User)
	assert.Equal(t, "secret", cfg.Passwd)
	assert.Equal(t, "tcp", cfg.Net)
	assert.Equal(t, "db.local:3306", cfg.Addr)
	assert.Equal(t, "analytics", cfg.DBName)
	assert.Equal(t, "utf8mb4", cfg.Params["charset"])
	assert.Equal(t, "True", cfg.Params["parseTime"])
	assert.Equal(t, "false", cfg.Params["useSSL"])

	_, err = buildMySQLPreviewConfig(&datasource.ConnectionConfig{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "datasource host is required")
}

func TestIsDirectPreviewRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *dataset.SQLPreviewRequest
		want bool
	}{
		{name: "nil request", req: nil, want: false},
		{name: "empty request", req: &dataset.SQLPreviewRequest{}, want: false},
		{name: "zero datasource id", req: &dataset.SQLPreviewRequest{DatasourceID: 0}, want: false},
		{name: "negative datasource id", req: &dataset.SQLPreviewRequest{DatasourceID: -1}, want: false},
		{name: "positive datasource id", req: &dataset.SQLPreviewRequest{DatasourceID: 1}, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isDirectPreviewRequest(tt.req))
		})
	}
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

func TestDatasetService_CalciteHelpersDirect(t *testing.T) {
	t.Run("validate with calcite disabled returns nil", func(t *testing.T) {
		svc := NewDatasetService(nil)
		require.NoError(t, svc.validateWithCalciteIfEnabled("SELECT 1"))
	})

	t.Run("ensure calcite client creates and reuses client", func(t *testing.T) {
		mock := &mockCalciteValidateSuccessServer{}
		addr, cleanup := startMockCalciteServer(t, mock)
		defer cleanup()

		svc := NewDatasetService(nil)
		svc.SetCalciteConfig(addr, 0, -1)

		client1, err := svc.ensureCalciteClient()
		require.NoError(t, err)
		require.NotNil(t, client1)

		client2, err := svc.ensureCalciteClient()
		require.NoError(t, err)
		assert.Same(t, client1, client2)

		require.NoError(t, svc.validateWithCalciteIfEnabled("SELECT 1"))
		assert.Greater(t, atomic.LoadInt32(&mock.validateCalls), int32(0))
	})

	t.Run("ensure calcite client returns constructor error for blank address", func(t *testing.T) {
		svc := NewDatasetService(nil)
		svc.calciteAddress = "   "

		client, err := svc.ensureCalciteClient()
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "calcite address is required")
	})
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

func TestPreviewSQL_CalciteClientCaching(t *testing.T) {
	mock := &mockCalciteValidateServer{}
	addr, cleanup := startMockCalciteServer(t, mock)
	defer cleanup()

	svc := NewDatasetService(nil)
	svc.SetCalciteConfig(addr, 2*time.Second, 0)

	// First call - should create client
	_, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "SELECT 1"})
	assert.Error(t, err) // validation fails in mock

	// Second call - should use cached client
	initialCalls := atomic.LoadInt32(&mock.validateCalls)
	_, err = svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: "SELECT 2"})
	assert.Error(t, err)
	secondCalls := atomic.LoadInt32(&mock.validateCalls)

	// Both calls should have hit the server
	assert.Greater(t, secondCalls, initialCalls)
}

func TestPreviewSQL_CalciteEmptyAddress(t *testing.T) {
	svc := NewDatasetService(nil)
	// Don't set calcite config - should skip calcite validation

	// Test with empty SQL - should return empty result without calling calcite or repo
	result, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: ""})
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDatasetService_RepoBackedWrappers(t *testing.T) {
	t.Run("set resource permission service stores dependency", func(t *testing.T) {
		svc := NewDatasetService(nil)
		permSvc := &ResourcePermissionService{}
		svc.SetResourcePermissionService(permSvc)
		assert.Same(t, permSvc, svc.resourcePermService)
	})

	t.Run("tree get group and fields use repository data", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		folderType := dataset.NodeTypeFolder
		datasetType := dataset.NodeTypeDataset
		level1 := 1
		level2 := 2
		rootPID := int64(0)
		childPID := int64(1)
		fieldName := "amount"

		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, Level: &level1, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "Sales", PID: &childPID, Level: &level2, NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 11, DatasetGroupID: 2, Name: &fieldName}).Error)

		tree, err := svc.Tree(&dataset.TreeRequest{})
		require.NoError(t, err)
		require.Len(t, tree, 1)
		assert.Equal(t, int64(1), tree[0].ID)
		require.Len(t, tree[0].Children, 1)
		assert.Equal(t, int64(2), tree[0].Children[0].ID)

		keyword := "Sale"
		tree, err = svc.Tree(&dataset.TreeRequest{Keyword: &keyword})
		require.NoError(t, err)
		require.Len(t, tree, 0)

		group, err := svc.GetGroupByID(2)
		require.NoError(t, err)
		assert.Equal(t, "Sales", group.Name)

		fields, err := svc.Fields(&dataset.FieldsRequest{DatasetGroupID: 2})
		require.NoError(t, err)
		require.Len(t, fields, 1)
		require.NotNil(t, fields[0].Name)
		assert.Equal(t, "amount", *fields[0].Name)
	})

	t.Run("preview wrappers read dataset table rows and permission helper short-circuits", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		physicalTable := "orders_preview"
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 21, DatasetGroupID: 5, PhysicalTable: &physicalTable}).Error)
		require.NoError(t, db.Exec("CREATE TABLE orders_preview (name TEXT, amount INTEGER)").Error)
		require.NoError(t, db.Exec("INSERT INTO orders_preview (name, amount) VALUES ('alice', 10), ('bob', 20)").Error)

		preview, err := svc.Preview(&dataset.PreviewRequest{DatasetGroupID: 5, Limit: 1})
		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Equal(t, int64(2), preview.Total)
		require.Len(t, preview.Rows, 1)
		assert.Contains(t, preview.Columns, "amount")
		assert.Contains(t, preview.Columns, "name")

		previewWithPerm, err := svc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: 5, Limit: 1}, 7)
		require.NoError(t, err)
		require.NotNil(t, previewWithPerm)
		assert.Equal(t, int64(2), previewWithPerm.Total)
		require.Len(t, previewWithPerm.Rows, 1)

		require.NoError(t, svc.ensureDatasourceDependenciesViewable(5, 7))
		require.NoError(t, svc.ensureDatasourceDependenciesViewable(0, 7))
		require.NoError(t, svc.ensureDatasourceDependenciesViewable(5, 0))
	})
}

func setupDatasetEnumFixture(t *testing.T) (*DatasetService, *gorm.DB) {
	t.Helper()

	svc, db := setupDatasetServiceRepoTest(t)
	rootPID := int64(0)
	rootLevel := 0
	childLevel := 1
	rootType := dataset.NodeTypeFolder
	childType := dataset.NodeTypeDataset
	ordersTable := "orders_enum"
	otherTable := "other_enum"
	sqlVariables := `[{"variableName":"regionVar","type":["VARCHAR"],"params":["north"]},{"variableName":"amountVar","type":["DOUBLE"],"params":[1]}]`
	blankSQLVariables := `[{"variableName":"   ","type":["VARCHAR"]}]`
	invalidSQLVariables := `not-json`
	queryName := "status_label"
	queryOrigin := "status"
	queryType := 0
	queryDatasetID := int64(101)
	displayOrigin := "status_name"
	displayType := 0
	displayDatasetID := int64(101)
	sortOrigin := "sort_rank"
	sortType := 2
	sortDatasetID := int64(101)
	filterOrigin := "region"
	filterType := 0
	filterDatasetID := int64(101)
	crossOrigin := "category"
	crossType := 0
	crossDatasetID := int64(102)
	floatOrigin := "amount_text"
	floatType := 3
	blankName := "fallback_name"
	blankDatasetID := int64(0)
	blankGroupID := int64(2)

	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, Level: &rootLevel, NodeType: &rootType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "Sales", PID: int64PtrDataset(1), Level: &childLevel, NodeType: &childType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 3, Name: "Loop", PID: int64PtrDataset(3), Level: &childLevel, NodeType: &childType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 4, Name: "Orphan", PID: int64PtrDataset(99), Level: &childLevel, NodeType: &childType}).Error)

	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 101, DatasetGroupID: 2, PhysicalTable: &ordersTable, SQLVariables: &sqlVariables}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 102, DatasetGroupID: 2, PhysicalTable: &otherTable, SQLVariables: &blankSQLVariables}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 103, DatasetGroupID: 4, PhysicalTable: &ordersTable, SQLVariables: &invalidSQLVariables}).Error)

	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 11, DatasetGroupID: 2, DatasetTableID: &queryDatasetID, Name: &queryName, OriginName: &queryOrigin, DeType: &queryType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 12, DatasetGroupID: 2, DatasetTableID: &displayDatasetID, OriginName: &displayOrigin, DeType: &displayType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 13, DatasetGroupID: 2, DatasetTableID: &sortDatasetID, OriginName: &sortOrigin, DeType: &sortType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 14, DatasetGroupID: 2, DatasetTableID: &filterDatasetID, OriginName: &filterOrigin, DeType: &filterType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 15, DatasetGroupID: 2, DatasetTableID: &crossDatasetID, OriginName: &crossOrigin, DeType: &crossType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 16, DatasetGroupID: 2, OriginName: strPtrDataset(""), Name: &blankName, DeType: &filterType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 17, DatasetGroupID: 2, DatasetTableID: &queryDatasetID, OriginName: &floatOrigin, DeType: &floatType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 18, DatasetGroupID: blankGroupID, DatasetTableID: &blankDatasetID, OriginName: strPtrDataset("  "), Name: strPtrDataset("   "), DataeaseName: strPtrDataset("  ")}).Error)

	require.NoError(t, db.Exec("CREATE TABLE orders_enum (status TEXT, status_name TEXT, sort_rank INTEGER, region TEXT, amount_text TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO orders_enum (status, status_name, sort_rank, region, amount_text) VALUES ('A', 'Alpha', 2, 'North', '1.23E+3'), ('A', 'Alpha', 2, 'North', '1.23E+3'), ('B', 'Beta', 1, 'North', '2.00E+2'), ('C', 'Gamma', 3, 'South', '3.50E+1'), ('', 'Blank', 4, 'North', NULL)").Error)
	require.NoError(t, db.Exec("CREATE TABLE other_enum (category TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO other_enum (category) VALUES ('X'), ('Y')").Error)

	return svc, db
}

func TestDatasetService_SQLParamsAndEnumHelpers(t *testing.T) {
	t.Run("dataset full name walks hierarchy and stops on loops or missing parents", func(t *testing.T) {
		svc, _ := setupDatasetEnumFixture(t)

		fullName, err := svc.datasetFullName(2)
		require.NoError(t, err)
		assert.Equal(t, "Root/Sales", fullName)

		fullName, err = svc.datasetFullName(3)
		require.NoError(t, err)
		assert.Equal(t, "Loop", fullName)

		fullName, err = svc.datasetFullName(4)
		require.NoError(t, err)
		assert.Equal(t, "Orphan", fullName)

		fullName, err = svc.datasetFullName(0)
		require.NoError(t, err)
		assert.Equal(t, "", fullName)
	})

	t.Run("get sql params skips invalid inputs and parses valid variables", func(t *testing.T) {
		svc, _ := setupDatasetEnumFixture(t)

		params, err := svc.GetSQLParams(nil)
		require.NoError(t, err)
		assert.Empty(t, params)

		params, err = svc.GetSQLParams([]int64{-1, 2, 4, 999})
		require.NoError(t, err)
		require.Len(t, params, 2)
		assert.Equal(t, "regionVar", params[0].VariableName)
		assert.Equal(t, "Root/Sales", params[0].DatasetFullName)
		assert.Equal(t, 0, params[0].DeType)
		assert.Equal(t, "amountVar", params[1].VariableName)
		assert.Equal(t, 3, params[1].DeType)
	})

	t.Run("resolve enum field target uses origin dataease name and table fallbacks", func(t *testing.T) {
		svc, _ := setupDatasetEnumFixture(t)

		field, tableName, columnName, err := svc.resolveEnumFieldTarget(11)
		require.NoError(t, err)
		assert.Equal(t, int64(11), field.ID)
		assert.Equal(t, "orders_enum", tableName)
		assert.Equal(t, "status", columnName)

		field, tableName, columnName, err = svc.resolveEnumFieldTarget(16)
		require.NoError(t, err)
		assert.Equal(t, int64(16), field.ID)
		assert.Equal(t, "orders_enum", tableName)
		assert.Equal(t, "fallback_name", columnName)

		field, tableName, columnName, err = svc.resolveEnumFieldTarget(18)
		require.Error(t, err)
		assert.Nil(t, field)
		assert.Equal(t, "", tableName)
		assert.Equal(t, "", columnName)
		assert.Contains(t, err.Error(), "origin name is required")
	})

	t.Run("build enum filter clauses keeps same-table valid in filters only", func(t *testing.T) {
		svc, _ := setupDatasetEnumFixture(t)

		clauses, err := svc.buildEnumFilterClauses([]dataset.EnumFilter{{Operator: "eq", FieldID: "14", Value: []interface{}{"North"}}, {Operator: "in", FieldID: "14,999,15", Value: []interface{}{"North", "North", ""}}, {Operator: "in", FieldID: "", Value: []interface{}{"North"}}, {Operator: "in", FieldID: "14", Value: []interface{}{}}}, "orders_enum")
		require.NoError(t, err)
		require.Len(t, clauses, 1)
		assert.Equal(t, "region", clauses[0].Column)
		assert.Equal(t, []string{"North"}, clauses[0].Values)
	})
}

func TestDatasetService_EnumQueries(t *testing.T) {
	t.Run("get field enum deduplicates values and applies filters", func(t *testing.T) {
		svc, _ := setupDatasetEnumFixture(t)

		values, err := svc.GetFieldEnum(nil)
		require.NoError(t, err)
		assert.Empty(t, values)

		values, err = svc.GetFieldEnum(&dataset.MultFieldValuesRequest{FieldIDs: []int64{11, 11, 17, 999, -1}, Filter: []dataset.EnumFilter{{FieldID: "14", Operator: "in", Value: []interface{}{"North"}}}})
		require.NoError(t, err)
		assert.Equal(t, []string{"A", "B", "1230", "200"}, values)

		wrapped, err := svc.GetFieldEnumDs(11)
		require.NoError(t, err)
		assert.Equal(t, []string{"A", "B", "C"}, wrapped)

		wrapped, err = svc.GetFieldEnumDs(0)
		require.NoError(t, err)
		assert.Empty(t, wrapped)
	})

	t.Run("get field enum obj handles fallbacks search sort and dedupe", func(t *testing.T) {
		svc, _ := setupDatasetEnumFixture(t)

		items, err := svc.GetFieldEnumObj(nil)
		require.NoError(t, err)
		assert.Empty(t, items)

		items, err = svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: 999})
		require.NoError(t, err)
		assert.Empty(t, items)

		items, err = svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: 11, DisplayID: 15, SortID: 15, Sort: "DESC", SearchText: "a", Filter: []dataset.EnumFilter{{FieldID: "14", Operator: "in", Value: []interface{}{"North"}}}})
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, map[string]interface{}{"11": "A"}, items[0])

		items, err = svc.GetFieldEnumObj(&dataset.EnumValueRequest{QueryID: 11, DisplayID: 12, SortID: 13, Sort: "ASC", Filter: []dataset.EnumFilter{{FieldID: "14", Operator: "in", Value: []interface{}{"North"}}}, ResultMode: 1})
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, map[string]interface{}{"11": "B", "12": "Beta"}, items[0])
		assert.Equal(t, map[string]interface{}{"11": "A", "12": "Alpha"}, items[1])
	})
}

func TestDatasetService_GroupMutations(t *testing.T) {
	t.Run("create save rename move and delete exercise sqlite-backed paths", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		rootLevel := 0
		folderType := dataset.NodeTypeFolder
		datasetType := dataset.NodeTypeDataset
		dashboardType := "dashboard"
		chartTableName := "table_for_chart"
		require.NoError(t, db.Exec("CREATE TABLE core_chart_view (id INTEGER PRIMARY KEY AUTOINCREMENT, table_id INTEGER)").Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, Level: &rootLevel, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "ChildA", PID: int64PtrDataset(1), Level: intPtrDataset(1), NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 3, Name: "ChildB", PID: int64PtrDataset(1), Level: intPtrDataset(1), NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 4, Name: "Grand", PID: int64PtrDataset(2), Level: intPtrDataset(2), NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 201, DatasetGroupID: 2, PhysicalTable: &chartTableName}).Error)
		require.NoError(t, db.Exec("INSERT INTO core_chart_view (table_id) VALUES (201)").Error)

		created, err := svc.Create(&dataset.WriteRequest{Name: "CreatedRoot", NodeType: "", Type: &dashboardType})
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, "CreatedRoot", created.Name)
		require.NotNil(t, created.PID)
		assert.Equal(t, int64(0), *created.PID)
		require.NotNil(t, created.NodeType)
		assert.Equal(t, dataset.NodeTypeFolder, *created.NodeType)

		createdChild, err := svc.Create(&dataset.WriteRequest{Name: "CreatedChild", PID: int64PtrDataset(1), NodeType: dataset.NodeTypeDataset})
		require.NoError(t, err)
		require.NotNil(t, createdChild.Level)
		assert.Equal(t, 1, *createdChild.Level)

		_, err = svc.Create(&dataset.WriteRequest{Name: "ChildA", PID: int64PtrDataset(1)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		saved, err := svc.Save(&dataset.WriteRequest{ID: 2, Name: "  ", PID: int64PtrDataset(0), NodeType: "unknown", Type: &dashboardType})
		require.NoError(t, err)
		assert.Equal(t, "ChildA", saved.Name)
		require.NotNil(t, saved.PID)
		assert.Equal(t, int64(0), *saved.PID)
		require.NotNil(t, saved.NodeType)
		assert.Equal(t, dataset.NodeTypeDataset, *saved.NodeType)
		require.NotNil(t, saved.Type)
		assert.Equal(t, "dashboard", *saved.Type)

		_, err = svc.Save(&dataset.WriteRequest{ID: 999, Name: "missing"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dataset not found")

		_, err = svc.Save(&dataset.WriteRequest{ID: 3, Name: "ChildA", PID: int64PtrDataset(0)})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		renamed, err := svc.Rename(3, "ChildB-Renamed")
		require.NoError(t, err)
		assert.Equal(t, "ChildB-Renamed", renamed.Name)

		_, err = svc.Rename(3, "CreatedChild")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		_, err = svc.Rename(999, "missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dataset not found")

		_, err = svc.Move(3, 999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "destination folder not found")

		_, err = svc.Move(2, 4)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "child of current dataset")

		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 50, Name: "ChildB-Renamed", PID: &rootPID, Level: &rootLevel, NodeType: &datasetType}).Error)

		_, err = svc.Move(3, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")

		_, err = svc.Rename(50, "RootUnique")
		require.NoError(t, err)

		moved, err := svc.Move(3, 0)
		require.NoError(t, err)
		require.NotNil(t, moved.PID)
		assert.Equal(t, int64(0), *moved.PID)

		hasRelation, err := svc.PerDelete(2)
		require.NoError(t, err)
		assert.True(t, hasRelation)

		hasRelation, err = svc.PerDelete(3)
		require.NoError(t, err)
		assert.False(t, hasRelation)

		err = svc.Delete(2)
		require.NoError(t, err)

		deletedChild, err := svc.GetGroupByID(2)
		require.Error(t, err)
		assert.Nil(t, deletedChild)

		deletedGrand, err := svc.GetGroupByID(4)
		require.Error(t, err)
		assert.Nil(t, deletedGrand)
	})
}

func TestDatasetService_PreviewPermissionAndBackfillBranches(t *testing.T) {
	t.Run("preview handles empty table and permission preview default limit", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		emptyTable := "empty_preview"
		permTable := "perm_preview"
		datasourceID := int64(100)

		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 301, DatasetGroupID: 30, PhysicalTable: &emptyTable}).Error)
		require.NoError(t, db.Exec("CREATE TABLE empty_preview (name TEXT, amount INTEGER)").Error)

		preview, err := svc.Preview(&dataset.PreviewRequest{DatasetGroupID: 30, Limit: 0})
		require.NoError(t, err)
		require.NotNil(t, preview)
		assert.Empty(t, preview.Columns)
		assert.Empty(t, preview.Rows)
		assert.Equal(t, int64(0), preview.Total)

		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 302, DatasetGroupID: 31, PhysicalTable: &permTable, DatasourceID: &datasourceID}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 303, DatasetGroupID: 31, PhysicalTable: &permTable, DatasourceID: &datasourceID}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 304, DatasetGroupID: 31, PhysicalTable: &permTable, DatasourceID: int64PtrDataset(0)}).Error)
		require.NoError(t, db.Exec("CREATE TABLE perm_preview (name TEXT, amount INTEGER)").Error)
		require.NoError(t, db.Exec("INSERT INTO perm_preview (name, amount) VALUES ('alice', 10), ('bob', 20)").Error)

		svc.resourcePermService = &ResourcePermissionService{adminChecker: datasetAdminChecker{isAdmin: true}}
		resp, err := svc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: 31, Limit: 0}, 9)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(2), resp.Total)
		require.Len(t, resp.Rows, 2)
		assert.Contains(t, resp.Columns, "amount")
		assert.Contains(t, resp.Columns, "name")

		require.NoError(t, svc.ensureDatasourceDependenciesViewable(31, 9))
	})

	t.Run("permission dependency denial and lightweight guards", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		denyTable := "deny_preview"
		datasourceID := int64(200)

		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 401, DatasetGroupID: 40, PhysicalTable: &denyTable, DatasourceID: &datasourceID}).Error)
		require.NoError(t, db.Exec("CREATE TABLE deny_preview (name TEXT)").Error)
		require.NoError(t, db.Exec("INSERT INTO deny_preview (name) VALUES ('x')").Error)

		svc.resourcePermService = &ResourcePermissionService{}
		err := svc.ensureDatasourceDependenciesViewable(40, 9)
		require.ErrorIs(t, err, ErrDatasetDatasourcePermissionDenied)

		resp, err := svc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: 40, Limit: 1}, 9)
		require.ErrorIs(t, err, ErrDatasetDatasourcePermissionDenied)
		assert.Nil(t, resp)

		hasRelation, err := svc.PerDelete(0)
		require.Error(t, err)
		assert.False(t, hasRelation)

		svc.resourcePermService = &ResourcePermissionService{}
		require.NoError(t, svc.applyInheritedPermissionsOnCreate(1, "dataset", 0))

		svc = NewDatasetService(nil)
		report, err := svc.BackfillGovernedResources()
		require.Error(t, err)
		assert.Nil(t, report)
		assert.Contains(t, err.Error(), "dataset repository not initialized")

		svc, _ = setupDatasetServiceRepoTest(t)
		report, err = svc.BackfillGovernedResourcesWithOptions(nil)
		require.Error(t, err)
		assert.Nil(t, report)
		assert.Contains(t, err.Error(), "resource permission service not initialized")
	})

	t.Run("preview with permission applies row permission select and where clauses", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		permTable := "row_perm_preview"
		datasetTableID := int64(501)
		createBy := "tester"
		now := time.Now()

		require.NoError(t, db.AutoMigrate(&permission.DataPermRow{}, &permission.DataPermColumn{}))
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: datasetTableID, DatasetGroupID: 32, PhysicalTable: &permTable}).Error)
		require.NoError(t, db.Exec("CREATE TABLE row_perm_preview (`11` TEXT, `12` TEXT)").Error)
		require.NoError(t, db.Exec("INSERT INTO row_perm_preview (`11`, `12`) VALUES ('A', 'hide-a'), ('B', 'hide-b')").Error)

		expr := `{"logic":"and","items":[{"fieldId":11,"filterType":"logic","term":"eq","value":"A"}]}`
		require.NoError(t, db.Create(&permission.DataPermRow{DatasetID: 32, DatasetGroupID: 32, AuthTargetType: permission.AuthTargetTypeUser, AuthTargetID: 9, ExpressionTree: expr, Status: 1, CreateBy: &createBy, CreateTime: &now}).Error)
		require.NoError(t, db.Create(&permission.DataPermColumn{DatasetID: 32, DatasetGroupID: 32, FieldName: "11", PermType: "show", Status: 1, CreateBy: &createBy, CreateTime: &now}).Error)
		require.NoError(t, db.Create(&permission.DataPermColumn{DatasetID: 32, DatasetGroupID: 32, FieldName: "12", PermType: "disable", Status: 1, CreateBy: &createBy, CreateTime: &now}).Error)

		rowPermRepo := repository.NewRowPermissionRepository(db)
		colPermRepo := repository.NewColumnPermissionRepository(db)
		svc.rowPermissionService = NewRowPermissionService(rowPermRepo, colPermRepo, nil, nil)

		resp, err := svc.PreviewWithPermission(&dataset.PreviewRequest{DatasetGroupID: 32, Limit: 0}, 9)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(2), resp.Total)
		require.Len(t, resp.Rows, 1)
		assert.Equal(t, []string{"11"}, resp.Columns)
		assert.Equal(t, "A", resp.Rows[0]["11"])
		_, hiddenExists := resp.Rows[0]["12"]
		assert.False(t, hiddenExists)
	})
}

func TestDatasetService_CreateRollbackAndRecursiveHelpers(t *testing.T) {
	t.Run("create rolls back when inherited permissions fail and backfill repo nil is guarded", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		rootLevel := 0
		folderType := dataset.NodeTypeFolder
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, Level: &rootLevel, NodeType: &folderType}).Error)

		svc.resourcePermService = &ResourcePermissionService{}
		created, err := svc.Create(&dataset.WriteRequest{Name: "NeedsPerms", PID: int64PtrDataset(1), NodeType: dataset.NodeTypeDataset})
		require.Error(t, err)
		assert.Nil(t, created)
		assert.Contains(t, err.Error(), "repository not initialized")

		var deleted dataset.CoreDatasetGroup
		require.NoError(t, db.Unscoped().Where("name = ?", "NeedsPerms").First(&deleted).Error)
		require.NotNil(t, deleted.DelFlag)
		assert.Equal(t, 1, *deleted.DelFlag)

		svc = NewDatasetService(nil)
		report, err := svc.BackfillGovernedResourcesWithOptions(nil)
		require.Error(t, err)
		assert.Nil(t, report)
		assert.Contains(t, err.Error(), "dataset repository not initialized")
	})

	t.Run("recursive helpers traverse deep trees", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		folderType := dataset.NodeTypeFolder
		datasetType := dataset.NodeTypeDataset
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "A", PID: int64PtrDataset(1), NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 3, Name: "B", PID: int64PtrDataset(2), NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 4, Name: "C", PID: int64PtrDataset(3), NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 5, Name: "Sibling", PID: int64PtrDataset(1), NodeType: &datasetType}).Error)

		descendant, err := svc.isDescendant(1, 4)
		require.NoError(t, err)
		assert.True(t, descendant)

		descendant, err = svc.isDescendant(2, 5)
		require.NoError(t, err)
		assert.False(t, descendant)

		require.NoError(t, svc.deleteRecursive(2))
		_, err = svc.GetGroupByID(2)
		require.Error(t, err)
		_, err = svc.GetGroupByID(3)
		require.Error(t, err)
		_, err = svc.GetGroupByID(4)
		require.Error(t, err)

		sibling, err := svc.GetGroupByID(5)
		require.NoError(t, err)
		assert.Equal(t, "Sibling", sibling.Name)
	})
}

func TestDatasetService_GroupMutationExtraBranches(t *testing.T) {
	t.Run("move returns source not found and succeeds into non-root folder", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		level0 := 0
		level1 := 1
		folderType := dataset.NodeTypeFolder
		datasetType := dataset.NodeTypeDataset

		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, Level: &level0, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "Source", PID: &rootPID, Level: &level1, NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 3, Name: "Dest", PID: &rootPID, Level: &level1, NodeType: &datasetType}).Error)

		_, err := svc.Move(999, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dataset not found")

		moved, err := svc.Move(2, 3)
		require.NoError(t, err)
		require.NotNil(t, moved.PID)
		assert.Equal(t, int64(3), *moved.PID)
	})

	t.Run("move propagates update failure", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		level0 := 0
		level1 := 1
		folderType := dataset.NodeTypeFolder
		datasetType := dataset.NodeTypeDataset

		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, Level: &level0, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "SourceErr", PID: &rootPID, Level: &level1, NodeType: &datasetType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 3, Name: "DestErr", PID: &rootPID, Level: &level1, NodeType: &datasetType}).Error)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_dataset_move_update BEFORE UPDATE ON core_dataset_group BEGIN SELECT RAISE(FAIL, 'deny dataset move update'); END;").Error)

		moved, err := svc.Move(2, 3)
		require.Error(t, err)
		assert.Nil(t, moved)
		assert.Contains(t, err.Error(), "deny dataset move update")
	})

	t.Run("create handles missing parent and nil parent level fallback", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		folderType := dataset.NodeTypeFolder

		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "RootNilLevel", PID: &rootPID, Level: nil, NodeType: &folderType}).Error)

		_, err := svc.Create(&dataset.WriteRequest{Name: "MissingParent", PID: int64PtrDataset(999), NodeType: dataset.NodeTypeDataset})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "destination folder not found")

		created, err := svc.Create(&dataset.WriteRequest{Name: "NilLevelChild", PID: int64PtrDataset(1), NodeType: dataset.NodeTypeDataset})
		require.NoError(t, err)
		require.NotNil(t, created.Level)
		assert.Equal(t, 1, *created.Level)
	})
}

func TestDatasetService_SaveDelegationAndPreviewSQLExecution(t *testing.T) {
	t.Run("save delegates to create for new records", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		folderType := dataset.NodeTypeFolder
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, NodeType: &folderType}).Error)

		created, err := svc.Save(&dataset.WriteRequest{Name: "SavedNew", PID: int64PtrDataset(1), NodeType: dataset.NodeTypeDataset})
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.Equal(t, "SavedNew", created.Name)
		require.NotNil(t, created.PID)
		assert.Equal(t, int64(1), *created.PID)

		created, err = svc.Save(&dataset.WriteRequest{Name: "MissingParent", PID: int64PtrDataset(999), NodeType: dataset.NodeTypeDataset})
		require.Error(t, err)
		assert.Nil(t, created)
		assert.Contains(t, err.Error(), "destination folder not found")
	})

	t.Run("save propagates update failure for existing record", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		folderType := dataset.NodeTypeFolder
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Existing", PID: &rootPID, NodeType: &folderType}).Error)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_dataset_save_update BEFORE UPDATE ON core_dataset_group BEGIN SELECT RAISE(FAIL, 'deny dataset save update'); END;").Error)

		updated, err := svc.Save(&dataset.WriteRequest{ID: 1, Name: "ExistingUpdated", PID: int64PtrDataset(0), NodeType: dataset.NodeTypeFolder})
		require.Error(t, err)
		assert.Nil(t, updated)
		assert.Contains(t, err.Error(), "deny dataset save update")
	})

	t.Run("preview sql executes query and returns normalized payload", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TABLE preview_sql_rows (name TEXT, amount INTEGER)").Error)
		require.NoError(t, db.Exec("INSERT INTO preview_sql_rows (name, amount) VALUES ('alice', 10), ('bob', 20)").Error)

		rawSQL := "SELECT name, amount FROM preview_sql_rows ORDER BY amount DESC"
		result, err := svc.PreviewSQL(&dataset.SQLPreviewRequest{SQL: base64.StdEncoding.EncodeToString([]byte(rawSQL))})
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(rawSQL)), result["sql"])

		previewData, ok := result["data"].(dataset.SQLPreviewData)
		require.True(t, ok)
		require.Len(t, previewData.Fields, 2)
		assert.Equal(t, "amount", previewData.Fields[0].OriginName)
		assert.Equal(t, 2, previewData.Fields[0].DeType)
		assert.Equal(t, "name", previewData.Fields[1].OriginName)
		assert.Equal(t, 0, previewData.Fields[1].DeType)
		require.Len(t, previewData.Data, 2)
		assert.Equal(t, "bob", previewData.Data[0]["name"])
		assert.Equal(t, int64(20), previewData.Data[0]["amount"])
		assert.Equal(t, "alice", previewData.Data[1]["name"])
	})

	t.Run("preview sql rejects datasource-aware direct preview explicitly", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		datasourceRepo := repository.NewDatasourceRepository(db)
		svc.SetDatasourceRepository(datasourceRepo)
		config := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "db.local", Port: 5432, Database: "analytics", Username: "pg", Password: "secret"})
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 88, Name: "pg-ds", Type: "pg", Configuration: &config}).Error)

		_, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{
			SQL:                base64.StdEncoding.EncodeToString([]byte("SELECT 1")),
			DatasourceID:       88,
			SQLVariableDetails: `[{"variableName":"region"}]`,
		}, 9)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPreviewSQLExternalDatasourceUnsupported)
	})

	t.Run("preview sql accepts sqlVariableDetails without changing local routing", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TABLE preview_sql_variables (name TEXT, amount INTEGER)").Error)
		require.NoError(t, db.Exec("INSERT INTO preview_sql_variables (name, amount) VALUES ('alice', 10)").Error)

		result, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{
			SQL:                base64.StdEncoding.EncodeToString([]byte("SELECT name, amount FROM preview_sql_variables")),
			SQLVariableDetails: `[{"variableName":"city","defaultValue":"shanghai"}]`,
		}, 0)
		require.NoError(t, err)
		previewData, ok := result["data"].(dataset.SQLPreviewData)
		require.True(t, ok)
		require.Len(t, previewData.Data, 1)
		assert.Equal(t, "alice", previewData.Data[0]["name"])
	})

	t.Run("preview sql routes mysql datasource to executor factory", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		datasourceRepo := repository.NewDatasourceRepository(db)
		svc.SetDatasourceRepository(datasourceRepo)
		config := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "mysql.local", Port: 3306, Database: "analytics", Username: "root", Password: "secret"})
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 99, Name: "mysql-ds", Type: "mysql", Configuration: &config}).Error)

		executor := &stubPreviewExecutor{rows: []map[string]interface{}{{"name": "alice", "amount": int64(10)}}}
		svc.SetPreviewExecutorFactory(func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (PreviewExecutor, error) {
			require.Equal(t, int64(99), ds.ID)
			require.Equal(t, "mysql", ds.Type)
			require.Equal(t, "root", cfg.Username)
			require.Equal(t, "secret", cfg.Password)
			return executor, nil
		})

		result, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{
			SQL:                base64.StdEncoding.EncodeToString([]byte("SELECT amount, name FROM orders")),
			DatasourceID:       99,
			SQLVariableDetails: `[{"variableName":"limit"}]`,
		}, 9)
		require.NoError(t, err)
		assert.True(t, executor.called)
		assert.True(t, executor.closed)
		assert.Equal(t, "SELECT amount, name FROM orders", executor.rawSQL)
		assert.Equal(t, 100, executor.limit)

		previewData, ok := result["data"].(dataset.SQLPreviewData)
		require.True(t, ok)
		require.Len(t, previewData.Fields, 2)
		require.Len(t, previewData.Data, 1)
		assert.Equal(t, "alice", previewData.Data[0]["name"])
	})

	t.Run("preview sql denies direct preview without datasource view permission", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		datasourceRepo := repository.NewDatasourceRepository(db)
		svc.SetDatasourceRepository(datasourceRepo)
		svc.resourcePermService = &ResourcePermissionService{}
		config := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "mysql.local", Port: 3306, Database: "analytics", Username: "root", Password: "secret"})
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 109, Name: "mysql-ds", Type: "mysql", Configuration: &config}).Error)

		_, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{SQL: base64.StdEncoding.EncodeToString([]byte("SELECT 1")), DatasourceID: 109, SQLVariableDetails: `[{"variableName":"tenant"}]`}, 9)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDatasetDatasourcePermissionDenied)
	})

	t.Run("preview sql returns timeout from executor", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		datasourceRepo := repository.NewDatasourceRepository(db)
		svc.SetDatasourceRepository(datasourceRepo)
		config := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "mysql.local", Port: 3306, Database: "analytics", Username: "root", Password: "secret"})
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 110, Name: "mysql-ds", Type: "mysql", Configuration: &config}).Error)

		executor := &stubPreviewExecutor{err: ErrPreviewSQLTimeout}
		svc.SetPreviewExecutorFactory(func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (PreviewExecutor, error) {
			return executor, nil
		})

		_, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{SQL: base64.StdEncoding.EncodeToString([]byte("SELECT 1")), DatasourceID: 110}, 9)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPreviewSQLTimeout)
		assert.True(t, executor.closed)
	})

	t.Run("preview sql returns result too large from executor", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		datasourceRepo := repository.NewDatasourceRepository(db)
		svc.SetDatasourceRepository(datasourceRepo)
		config := encodeDatasourceConfig(t, &datasource.ConnectionConfig{Host: "mysql.local", Port: 3306, Database: "analytics", Username: "root", Password: "secret"})
		require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 111, Name: "mysql-ds", Type: "mysql", Configuration: &config}).Error)

		executor := &stubPreviewExecutor{err: ErrPreviewSQLResultTooLarge}
		svc.SetPreviewExecutorFactory(func(ds *datasource.CoreDatasource, cfg *datasource.ConnectionConfig) (PreviewExecutor, error) {
			return executor, nil
		})

		_, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{SQL: base64.StdEncoding.EncodeToString([]byte("SELECT 1")), DatasourceID: 111}, 9)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPreviewSQLResultTooLarge)
		assert.True(t, executor.closed)
	})

	t.Run("local preview sql applies result too large guard", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TABLE preview_sql_large_guard (c1 TEXT, c2 TEXT, c3 TEXT, c4 TEXT, c5 TEXT, c6 TEXT, c7 TEXT, c8 TEXT, c9 TEXT, c10 TEXT, c11 TEXT, c12 TEXT, c13 TEXT, c14 TEXT, c15 TEXT, c16 TEXT, c17 TEXT, c18 TEXT, c19 TEXT, c20 TEXT, c21 TEXT, c22 TEXT, c23 TEXT, c24 TEXT, c25 TEXT, c26 TEXT, c27 TEXT, c28 TEXT, c29 TEXT, c30 TEXT, c31 TEXT, c32 TEXT, c33 TEXT, c34 TEXT, c35 TEXT, c36 TEXT, c37 TEXT, c38 TEXT, c39 TEXT, c40 TEXT, c41 TEXT, c42 TEXT, c43 TEXT, c44 TEXT, c45 TEXT, c46 TEXT, c47 TEXT, c48 TEXT, c49 TEXT, c50 TEXT, c51 TEXT)").Error)
		for i := 0; i < 500; i++ {
			require.NoError(t, db.Exec("INSERT INTO preview_sql_large_guard (c1,c2,c3,c4,c5,c6,c7,c8,c9,c10,c11,c12,c13,c14,c15,c16,c17,c18,c19,c20,c21,c22,c23,c24,c25,c26,c27,c28,c29,c30,c31,c32,c33,c34,c35,c36,c37,c38,c39,c40,c41,c42,c43,c44,c45,c46,c47,c48,c49,c50,c51) VALUES ('1','2','3','4','5','6','7','8','9','10','11','12','13','14','15','16','17','18','19','20','21','22','23','24','25','26','27','28','29','30','31','32','33','34','35','36','37','38','39','40','41','42','43','44','45','46','47','48','49','50','51')").Error)
		}

		_, err := svc.PreviewSQLWithUser(&dataset.SQLPreviewRequest{SQL: base64.StdEncoding.EncodeToString([]byte("SELECT * FROM preview_sql_large_guard"))}, 0)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPreviewSQLResultTooLarge)
	})
}

func TestDatasetService_DeleteRecursiveErrorAndBackfillExecution(t *testing.T) {
	t.Run("deleteRecursive propagates soft delete failure", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		folderType := dataset.NodeTypeFolder
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "Child", PID: int64PtrDataset(1), NodeType: &folderType}).Error)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_dataset_soft_delete BEFORE UPDATE ON core_dataset_group BEGIN SELECT RAISE(FAIL, 'deny dataset soft delete'); END;").Error)

		err := svc.deleteRecursive(1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deny dataset soft delete")
	})

	t.Run("backfill executes inherited and skipped paths", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		childPID := int64(1)
		orphanParent := int64(99)
		folderType := dataset.NodeTypeFolder
		require.NoError(t, db.AutoMigrate(&permission.SysResource{}, &permission.SysResourcePerm{}))
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 1, Name: "Root", PID: &rootPID, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 2, Name: "Child", PID: &childPID, NodeType: &folderType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 3, Name: "OrphanChild", PID: &orphanParent, NodeType: &folderType}).Error)

		resourceRepo := repository.NewResourcePermissionRepository(db)
		svc.resourcePermService = NewResourcePermissionService(resourceRepo, nil)
		require.NoError(t, resourceRepo.RegisterResource(1, "Root", permission.ResourceTypeDataset, &rootPID))
		require.NoError(t, resourceRepo.ReplaceResourcePermissions(1, permission.ResourceTypeDataset, []int64{101, 101}))

		report, err := svc.BackfillGovernedResourcesWithOptions(&GovernanceBackfillOptions{AfterID: -5, Limit: 0})
		require.NoError(t, err)
		require.NotNil(t, report)
		assert.Equal(t, permission.ResourceTypeDataset, report.ResourceType)
		assert.Equal(t, int64(0), report.AfterID)
		assert.Equal(t, DefaultGovernanceBackfillLimit, report.Limit)
		assert.Equal(t, 3, report.Scanned)
		assert.Equal(t, 1, report.Governed)
		assert.Equal(t, 2, report.Skipped)
		assert.Equal(t, int64(3), report.NextAfterID)
		assert.Equal(t, []int64{2}, report.ResourceIDs)
		require.Len(t, report.SkippedItems, 2)
		assert.Equal(t, GovernanceBackfillSkipReasonMissingParent, report.SkippedItems[0].Reason)
		assert.Equal(t, GovernanceBackfillSkipReasonParentNotGoverned, report.SkippedItems[1].Reason)

		permIDs, exists, err := resourceRepo.GetResourcePermissionIDs(2, permission.ResourceTypeDataset)
		require.NoError(t, err)
		assert.True(t, exists)
		assert.Equal(t, []int64{101}, permIDs)
	})
}

func TestDatasetService_ExportDataset(t *testing.T) {
	t.Run("success creates export task", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		repoStub := &exportRepoStub{}
		svc.SetExportRepository(repoStub)

		rootPID := int64(0)
		nodeType := dataset.NodeTypeDataset
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 11, Name: "Sales Dataset", PID: &rootPID, NodeType: &nodeType}).Error)

		resp, err := svc.ExportDataset(&dataset.ExportDatasetRequest{ID: 11, ViewName: "Sales Export"}, 77)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.TaskID)
		assert.Equal(t, "PENDING", resp.Status)
		assert.Equal(t, int64(11), resp.ExportFrom)
		assert.Equal(t, permission.ResourceTypeDataset, resp.ExportFromType)
		assert.Equal(t, "Sales Export", resp.ExportFromName)

		require.NotNil(t, repoStub.createTask)
		assert.Equal(t, resp.TaskID, repoStub.createTask.ID)
		assert.Equal(t, int64(77), repoStub.createTask.UserID)
		assert.Equal(t, int64(11), repoStub.createTask.ExportFrom)
		assert.Equal(t, "PENDING", repoStub.createTask.ExportStatus)
		assert.Equal(t, permission.ResourceTypeDataset, repoStub.createTask.ExportFromType)
		assert.Equal(t, "Sales Export", repoStub.createTask.ExportFromName)
		assert.Contains(t, repoStub.createTask.FileName, "Sales_Export")
	})

	t.Run("compat dataset id resolves to canonical export source", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		repoStub := &exportRepoStub{}
		svc.SetExportRepository(repoStub)

		rootPID := int64(0)
		nodeType := dataset.NodeTypeDataset
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 101, Name: "Compat Dataset", PID: &rootPID, NodeType: &nodeType}).Error)

		resp, err := svc.ExportDataset(&dataset.ExportDatasetRequest{ID: 200, ViewName: "Compat Export"}, 88)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(101), resp.ExportFrom)
		require.NotNil(t, repoStub.createTask)
		assert.Equal(t, int64(101), repoStub.createTask.ExportFrom)
		assert.Equal(t, permission.ResourceTypeDataset, repoStub.createTask.ExportFromType)
	})

	t.Run("validates request and dependencies", func(t *testing.T) {
		svc, _ := setupDatasetServiceRepoTest(t)

		_, err := svc.ExportDataset(nil, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "export request is required")

		_, err = svc.ExportDataset(&dataset.ExportDatasetRequest{ID: 0}, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dataset id is required")

		_, err = svc.ExportDataset(&dataset.ExportDatasetRequest{ID: 1}, 1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "export repository not initialized")
	})

	t.Run("returns dataset not found and create failure", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)

		notFoundStub := &exportRepoStub{}
		svc.SetExportRepository(notFoundStub)
		_, err := svc.ExportDataset(&dataset.ExportDatasetRequest{ID: 999}, 2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dataset not found")

		rootPID := int64(0)
		nodeType := dataset.NodeTypeDataset
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 22, Name: "Fail Dataset", PID: &rootPID, NodeType: &nodeType}).Error)

		createErrStub := &exportRepoStub{createErr: assert.AnError}
		svc.SetExportRepository(createErrStub)
		_, err = svc.ExportDataset(&dataset.ExportDatasetRequest{ID: 22, ViewName: "Fail Export"}, 2)
		require.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestDatasetService_DeleteFieldOperations(t *testing.T) {
	t.Run("delete field validates and succeeds for chart-scoped field", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		chartID := int64(333)
		name := "calc_field"
		origin := "calc_field"
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 701, DatasetGroupID: 10, ChartID: &chartID, Name: &name, OriginName: &origin}).Error)

		err := svc.DeleteField(701)
		require.NoError(t, err)

		_, err = svc.repo.GetFieldByID(701)
		require.Error(t, err)
	})

	t.Run("delete field rejects invalid states", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)

		err := svc.DeleteField(0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "field id is required")

		err = svc.DeleteField(999)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dataset field not found")

		name := "plain_field"
		origin := "plain_field"
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 702, DatasetGroupID: 10, Name: &name, OriginName: &origin}).Error)
		err = svc.DeleteField(702)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "dataset field is not chart-scoped")
	})

	t.Run("delete field blocks chart and row permission dependencies", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.AutoMigrate(&auto.CoreChartView{}, &permission.DataPermRow{}, &dataset.CoreDatasetTable{}))

		chartID := int64(446)
		tableName := "orders"
		tableType := "db"
		fieldName := "sales"
		require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 900, DatasetGroupID: 10, Name: &tableName, PhysicalTable: &tableName, Type: &tableType}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 705, DatasetGroupID: 10, ChartID: &chartID, Name: &fieldName, OriginName: &fieldName, DataeaseName: &fieldName}).Error)
		require.NoError(t, db.Create(&auto.CoreChartView{ID: 3001, TableID: 900, XAxis: `[{"id":705,"name":"sales"}]`}).Error)
		require.NoError(t, db.Create(&permission.DataPermRow{DatasetID: 10, DatasetGroupID: 10, ExpressionTree: `{"logic":"and","items":[{"fieldId":705}]}`}).Error)

		err := svc.DeleteField(705)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDatasetFieldDependencyBlocked)
		assert.Contains(t, err.Error(), "chart views")
		assert.Contains(t, err.Error(), "row permissions")
	})

	t.Run("delete field blocks derived and configuration dependencies", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.AutoMigrate(&permission.DataPermColumn{}, &auto.VisualizationLinkageField{}, &auto.VisualizationLinkJumpInfo{}, &auto.VisualizationOuterParamsTargetViewInfo{}))

		chartID := int64(447)
		extField := 2
		fieldName := "profit"
		origin := fmt.Sprintf("[%d]", 706)
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 706, DatasetGroupID: 10, ChartID: &chartID, Name: &fieldName, OriginName: &fieldName, DataeaseName: &fieldName, FieldShortName: &fieldName}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 707, DatasetGroupID: 10, ChartID: &chartID, Name: &fieldName, OriginName: &origin, ExtField: &extField}).Error)
		require.NoError(t, db.Create(&permission.DataPermColumn{DatasetID: 10, DatasetGroupID: 10, FieldName: fieldName, PermType: permission.PermTypeDisable, Status: 1}).Error)
		require.NoError(t, db.Create(&auto.VisualizationLinkageField{ID: 1, SourceField: 706, TargetField: 999}).Error)
		require.NoError(t, db.Create(&auto.VisualizationLinkJumpInfo{ID: 1, SourceFieldID: 706}).Error)
		require.NoError(t, db.Create(&auto.VisualizationOuterParamsTargetViewInfo{TargetID: "1", TargetFieldID: "706"}).Error)

		err := svc.DeleteField(706)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDatasetFieldDependencyBlocked)
		assert.Contains(t, err.Error(), "derived fields")
		assert.Contains(t, err.Error(), "column permissions")
		assert.Contains(t, err.Error(), "visualization linkage")
		assert.Contains(t, err.Error(), "visualization jumps")
		assert.Contains(t, err.Error(), "outer parameter bindings")
	})

	t.Run("delete fields by chart validates and scopes", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		chartID := int64(444)
		otherChartID := int64(445)
		name := "field"
		origin := "field"
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 703, DatasetGroupID: 10, ChartID: &chartID, Name: &name, OriginName: &origin}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 704, DatasetGroupID: 10, ChartID: &otherChartID, Name: &name, OriginName: &origin}).Error)

		err := svc.DeleteFieldByChart(0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "chart id is required")

		err = svc.DeleteFieldByChart(chartID)
		require.NoError(t, err)

		_, err = svc.repo.GetFieldByID(703)
		require.Error(t, err)
		field, err := svc.repo.GetFieldByID(704)
		require.NoError(t, err)
		require.NotNil(t, field)
		assert.Equal(t, int64(704), field.ID)
	})

	t.Run("delete by chart blocks when any field has dependencies", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.AutoMigrate(&auto.VisualizationLinkJumpInfo{}))

		chartID := int64(448)
		name := "field"
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 708, DatasetGroupID: 10, ChartID: &chartID, Name: &name, OriginName: &name}).Error)
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 709, DatasetGroupID: 10, ChartID: &chartID, Name: &name, OriginName: &name}).Error)
		require.NoError(t, db.Create(&auto.VisualizationLinkJumpInfo{ID: 2, SourceFieldID: 708}).Error)

		err := svc.DeleteFieldByChart(chartID)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrDatasetFieldDependencyBlocked)

		field, getErr := svc.repo.GetFieldByID(708)
		require.NoError(t, getErr)
		require.NotNil(t, field)
	})
}

func TestDatasetService_GetGroupByID_CompatFallback(t *testing.T) {
	t.Run("falls back to nearest id for legacy compatibility ids", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		rootPID := int64(0)
		nodeType := dataset.NodeTypeDataset
		require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 101, Name: "CompatTarget", PID: &rootPID, NodeType: &nodeType}).Error)

		group, err := svc.GetGroupByID(200)
		require.NoError(t, err)
		require.NotNil(t, group)
		assert.Equal(t, int64(101), group.ID)
	})

	t.Run("keeps not found semantics for non-compat ids", func(t *testing.T) {
		svc, _ := setupDatasetServiceRepoTest(t)
		_, err := svc.GetGroupByID(199)
		require.Error(t, err)
		assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	})
}

func TestDatasetService_SaveFieldOperations(t *testing.T) {
	t.Run("creates field when id is zero", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		field := &dataset.CoreDatasetTableField{
			DatasetGroupID: 41,
			Name:           strPtrDataset("order_amount"),
			Type:           strPtrDataset("int"),
			OriginName:     strPtrDataset("amount"),
		}

		result, err := svc.SaveField(field)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.NotZero(t, result.ID)

		stored, err := svc.repo.GetFieldByID(result.ID)
		require.NoError(t, err)
		require.NotNil(t, stored.Name)
		require.NotNil(t, stored.Type)
		assert.Equal(t, "order_amount", *stored.Name)
		assert.Equal(t, "int", *stored.Type)
		assert.Equal(t, int64(41), stored.DatasetGroupID)

		var count int64
		require.NoError(t, db.Model(&dataset.CoreDatasetTableField{}).Where("id = ?", result.ID).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("updates field when id exists", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.Create(&dataset.CoreDatasetTableField{
			ID:             901,
			DatasetGroupID: 52,
			Name:           strPtrDataset("profit"),
			Type:           strPtrDataset("decimal"),
			OriginName:     strPtrDataset("profit"),
		}).Error)

		updated := &dataset.CoreDatasetTableField{
			ID:             901,
			DatasetGroupID: 52,
			Name:           strPtrDataset("profit_v2"),
			Type:           strPtrDataset("string"),
			OriginName:     strPtrDataset("profit_name"),
		}

		result, err := svc.SaveField(updated)
		require.NoError(t, err)
		require.NotNil(t, result)

		stored, err := svc.repo.GetFieldByID(901)
		require.NoError(t, err)
		require.NotNil(t, stored.Name)
		require.NotNil(t, stored.Type)
		require.NotNil(t, stored.OriginName)
		assert.Equal(t, "profit_v2", *stored.Name)
		assert.Equal(t, "string", *stored.Type)
		assert.Equal(t, "profit_name", *stored.OriginName)
		assert.Equal(t, int64(52), stored.DatasetGroupID)
	})

	t.Run("update of missing field id is a no-op but returns payload", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		missing := &dataset.CoreDatasetTableField{
			ID:             999,
			DatasetGroupID: 77,
			Name:           strPtrDataset("ghost_field"),
			Type:           strPtrDataset("string"),
		}

		result, err := svc.SaveField(missing)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, int64(999), result.ID)

		var count int64
		require.NoError(t, db.Model(&dataset.CoreDatasetTableField{}).Where("id = ?", 999).Count(&count).Error)
		assert.Equal(t, int64(0), count)
	})

	t.Run("validates required fields", func(t *testing.T) {
		svc, _ := setupDatasetServiceRepoTest(t)
		tests := []struct {
			name    string
			field   *dataset.CoreDatasetTableField
			wantErr string
		}{
			{name: "missing name", field: &dataset.CoreDatasetTableField{DatasetGroupID: 1, Type: strPtrDataset("int")}, wantErr: "field name is required"},
			{name: "missing dataset group", field: &dataset.CoreDatasetTableField{Name: strPtrDataset("field1"), Type: strPtrDataset("int")}, wantErr: "datasetGroupId is required"},
			{name: "missing type", field: &dataset.CoreDatasetTableField{DatasetGroupID: 1, Name: strPtrDataset("field1")}, wantErr: "field type is required"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result, err := svc.SaveField(tt.field)
				require.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.wantErr)
			})
		}
	})
}

func TestDatasetService_ResolveUserName(t *testing.T) {
	t.Run("falls back for nil repo and invalid ids", func(t *testing.T) {
		svc := NewDatasetService(nil)
		assert.Equal(t, "", svc.ResolveUserName(""))
		assert.Equal(t, "101", svc.ResolveUserName("101"))
		assert.Equal(t, "not-a-number", svc.ResolveUserName("not-a-number"))
	})

	t.Run("returns nickname then username then raw id", func(t *testing.T) {
		svc, db := setupDatasetServiceRepoTest(t)
		require.NoError(t, db.AutoMigrate(&user.SysUser{}))
		require.NoError(t, db.Create(&user.SysUser{UserID: 101, Username: "alice_login", NickName: "Alice", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Create(&user.SysUser{UserID: 102, Username: "bob", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		svc.SetUserRepository(repository.NewUserRepository(db))

		assert.Equal(t, "Alice", svc.ResolveUserName("101"))
		assert.Equal(t, "bob", svc.ResolveUserName("102"))
		assert.Equal(t, "999", svc.ResolveUserName("999"))
	})
}

func TestDatasetService_GetFieldFunctions(t *testing.T) {
	svc := NewDatasetService(nil)
	functions := svc.GetFieldFunctions()
	require.Len(t, functions, 5)
	assert.Equal(t, "聚合函数", functions[0].Name)
	assert.Equal(t, "日期函数", functions[1].Name)
	assert.Equal(t, "字符串函数", functions[2].Name)
	assert.Equal(t, "数学函数", functions[3].Name)
	assert.Equal(t, "条件函数", functions[4].Name)

	assert.Contains(t, functions[0].Functions, FunctionDef{Name: "SUM", Hint: "SUM(field)"})
	assert.Contains(t, functions[1].Functions, FunctionDef{Name: "DATE_FORMAT", Hint: "DATE_FORMAT(date, format)"})
	assert.Contains(t, functions[2].Functions, FunctionDef{Name: "CONCAT", Hint: "CONCAT(str1, str2, ...)"})
	assert.Contains(t, functions[3].Functions, FunctionDef{Name: "ABS", Hint: "ABS(x)"})
	assert.Contains(t, functions[4].Functions, FunctionDef{Name: "IF", Hint: "IF(cond, true_val, false_val)"})
}

func TestDatasetService_ListFieldsByDsIds(t *testing.T) {
	svc, db := setupDatasetServiceRepoTest(t)
	dsAID := int64(11)
	dsBID := int64(22)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 1001, DatasourceID: &dsAID, DatasetGroupID: 81, Name: strPtrDataset("field_a1"), Type: strPtrDataset("string")}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 1002, DatasourceID: &dsAID, DatasetGroupID: 81, Name: strPtrDataset("field_a2"), Type: strPtrDataset("int")}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 1003, DatasourceID: &dsBID, DatasetGroupID: 82, Name: strPtrDataset("field_b1"), Type: strPtrDataset("string")}).Error)

	fields, err := svc.ListFieldsByDsIds([]int64{11})
	require.NoError(t, err)
	require.Len(t, fields, 2)
	assert.Equal(t, int64(11), *fields[0].DatasourceID)
	assert.Equal(t, int64(11), *fields[1].DatasourceID)

	empty, err := svc.ListFieldsByDsIds([]int64{})
	require.NoError(t, err)
	assert.Len(t, empty, 0)

	nilFields, err := svc.ListFieldsByDsIds(nil)
	require.NoError(t, err)
	assert.Len(t, nilFields, 0)
}

func int64PtrDataset(v int64) *int64 { return &v }

func intPtrDataset(v int) *int { return &v }

func strPtrDataset(v string) *string { return &v }
