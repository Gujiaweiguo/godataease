package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPreviewExecutors_LocalAndMySQLNilGuards(t *testing.T) {
	rows, err := (*localPreviewExecutor)(nil).PreviewSQL(context.Background(), "SELECT 1", 10)
	require.Error(t, err)
	assert.Nil(t, rows)
	assert.Contains(t, err.Error(), "dataset repository is unavailable")

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	repo := repository.NewDatasetRepository(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rows, err = (&localPreviewExecutor{repo: repo}).PreviewSQL(ctx, "SELECT 1", 10)
	require.ErrorIs(t, err, ErrPreviewSQLTimeout)
	assert.Nil(t, rows)
	assert.NoError(t, (&localPreviewExecutor{repo: repo}).Close())

	rows, err = (&mysqlPreviewExecutor{}).PreviewSQL(context.Background(), "SELECT 1", 10)
	require.Error(t, err)
	assert.Nil(t, rows)
	assert.Contains(t, err.Error(), "mysql preview executor is unavailable")
	assert.NoError(t, (&mysqlPreviewExecutor{}).Close())
}

func TestPreviewExecutorHelpers(t *testing.T) {
	config, err := buildMySQLPreviewConfig(&datasource.ConnectionConfig{
		Host:        "db.local",
		Port:        3306,
		Database:    "analytics",
		Username:    "root",
		Password:    "secret",
		ExtraParams: "?tls=skip-verify&charset=latin1&badparam",
	})
	require.NoError(t, err)
	assert.Equal(t, "root", config.User)
	assert.Equal(t, "secret", config.Passwd)
	assert.Equal(t, "db.local:3306", config.Addr)
	assert.Equal(t, "analytics", config.DBName)
	assert.Equal(t, "skip-verify", config.Params["tls"])
	assert.Equal(t, "latin1", config.Params["charset"])

	_, err = buildMySQLPreviewConfig(nil)
	require.Error(t, err)
	_, err = buildMySQLPreviewConfig(&datasource.ConnectionConfig{Port: 3306, Database: "db", Username: "root"})
	require.Error(t, err)
	_, err = buildMySQLPreviewConfig(&datasource.ConnectionConfig{Host: "db.local", Database: "db", Username: "root"})
	require.Error(t, err)
	_, err = buildMySQLPreviewConfig(&datasource.ConnectionConfig{Host: "db.local", Port: 3306, Username: "root"})
	require.Error(t, err)
	_, err = buildMySQLPreviewConfig(&datasource.ConnectionConfig{Host: "db.local", Port: 3306, Database: "db"})
	require.Error(t, err)

	_, err = defaultPreviewExecutorFactory(nil, nil)
	require.Error(t, err)
	_, err = defaultPreviewExecutorFactory(&datasource.CoreDatasource{Type: "postgres"}, &datasource.ConnectionConfig{})
	require.ErrorIs(t, err, ErrPreviewSQLExternalDatasourceUnsupported)
	_, err = defaultPreviewExecutorFactory(&datasource.CoreDatasource{Type: "mysql"}, &datasource.ConnectionConfig{})
	require.Error(t, err)

	assert.Equal(t, 3, previewCellCount([]map[string]interface{}{{"a": 1, "b": 2}, {"c": 3}}))
	assert.Zero(t, previewCellCount(nil))
}

// createCTETable creates a SQLite table with colCount columns and rowCount rows
// using a recursive CTE for efficient bulk data generation.
func createCTETable(t *testing.T, db *gorm.DB, colCount, rowCount int) string {
	t.Helper()
	tableName := fmt.Sprintf("t%d_%d", colCount, rowCount)
	colSelects := make([]string, colCount)
	for i := 0; i < colCount; i++ {
		colSelects[i] = fmt.Sprintf("x AS c%d", i)
	}
	sql := fmt.Sprintf(
		"CREATE TABLE %s AS WITH RECURSIVE cnt(x) AS (SELECT 1 UNION ALL SELECT x+1 FROM cnt WHERE x < %d) SELECT %s FROM cnt",
		tableName, rowCount, strings.Join(colSelects, ", "),
	)
	require.NoError(t, db.Exec(sql).Error)
	return tableName
}

// ---------------------------------------------------------------------------
// localPreviewExecutor: success path
// ---------------------------------------------------------------------------

func TestLocalPreviewExecutor_SuccessPath(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	tableName := createCTETable(t, db, 3, 5)

	repo := repository.NewDatasetRepository(db)
	executor := &localPreviewExecutor{repo: repo}

	rows, err := executor.PreviewSQL(context.Background(), fmt.Sprintf("SELECT * FROM %s", tableName), 10)
	require.NoError(t, err)
	assert.Len(t, rows, 5)
	for _, row := range rows {
		assert.Len(t, row, 3)
	}
	assert.NoError(t, executor.Close())
}

// ---------------------------------------------------------------------------
// localPreviewExecutor: limit normalization
// ---------------------------------------------------------------------------

func TestLocalPreviewExecutor_LimitNormalization(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// 600 rows, 1 column — enough to observe all limit thresholds
	tableName := createCTETable(t, db, 1, 600)
	repo := repository.NewDatasetRepository(db)
	executor := &localPreviewExecutor{repo: repo}
	q := fmt.Sprintf("SELECT * FROM %s", tableName)

	// limit=0 → default (100)
	rows, err := executor.PreviewSQL(context.Background(), q, 0)
	require.NoError(t, err)
	assert.Len(t, rows, 100)

	// limit=-5 → default (100)
	rows, err = executor.PreviewSQL(context.Background(), q, -5)
	require.NoError(t, err)
	assert.Len(t, rows, 100)

	// limit=50 → used as-is
	rows, err = executor.PreviewSQL(context.Background(), q, 50)
	require.NoError(t, err)
	assert.Len(t, rows, 50)

	// limit=600 → capped at 500
	rows, err = executor.PreviewSQL(context.Background(), q, 600)
	require.NoError(t, err)
	assert.Len(t, rows, 500)

	// limit=1000 → capped at 500
	rows, err = executor.PreviewSQL(context.Background(), q, 1000)
	require.NoError(t, err)
	assert.Len(t, rows, 500)
}

// ---------------------------------------------------------------------------
// localPreviewExecutor: repo error propagation on active context
// ---------------------------------------------------------------------------

func TestLocalPreviewExecutor_RepoErrorPropagation(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	repo := repository.NewDatasetRepository(db)
	executor := &localPreviewExecutor{repo: repo}

	// Active context + bad SQL → raw error, NOT wrapped as timeout
	rows, err := executor.PreviewSQL(context.Background(), "SELECT * FROM nonexistent_xyz", 10)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrPreviewSQLTimeout)
	assert.Nil(t, rows)
}

// ---------------------------------------------------------------------------
// localPreviewExecutor: ErrPreviewSQLResultTooLarge branch
// ---------------------------------------------------------------------------

func TestLocalPreviewExecutor_ResultTooLarge(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// 51 cols × 100 rows = 5100 cells > previewResultMaxCells (5000)
	tableName := createCTETable(t, db, 51, 100)

	repo := repository.NewDatasetRepository(db)
	executor := &localPreviewExecutor{repo: repo}

	rows, err := executor.PreviewSQL(context.Background(), fmt.Sprintf("SELECT * FROM %s", tableName), 100)
	require.ErrorIs(t, err, ErrPreviewSQLResultTooLarge)
	assert.Nil(t, rows)
}

// ---------------------------------------------------------------------------
// mysqlPreviewExecutor: timeout branch (pre-cancelled context, SQLite-backed db)
// ---------------------------------------------------------------------------

func TestMySQLPreviewExecutor_TimeoutBranch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	executor := &mysqlPreviewExecutor{db: db}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rows, err := executor.PreviewSQL(ctx, "SELECT 1", 10)
	require.ErrorIs(t, err, ErrPreviewSQLTimeout)
	assert.Nil(t, rows)
	assert.NoError(t, executor.Close())
}

// ---------------------------------------------------------------------------
// mysqlPreviewExecutor: error branch (bad SQL, active context)
// ---------------------------------------------------------------------------

func TestMySQLPreviewExecutor_ErrorBranch(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	executor := &mysqlPreviewExecutor{db: db}

	rows, err := executor.PreviewSQL(context.Background(), "SELECT * FROM nonexistent_abc", 10)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrPreviewSQLTimeout)
	assert.Nil(t, rows)
	assert.NoError(t, executor.Close())
}

// ---------------------------------------------------------------------------
// mysqlPreviewExecutor: ErrPreviewSQLResultTooLarge branch
// ---------------------------------------------------------------------------

func TestMySQLPreviewExecutor_ResultTooLarge(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// 51 cols × 100 rows = 5100 cells > 5000
	tableName := createCTETable(t, db, 51, 100)

	executor := &mysqlPreviewExecutor{db: db}

	rows, err := executor.PreviewSQL(context.Background(), fmt.Sprintf("SELECT * FROM %s", tableName), 100)
	require.ErrorIs(t, err, ErrPreviewSQLResultTooLarge)
	assert.Nil(t, rows)
	assert.NoError(t, executor.Close())
}

// ---------------------------------------------------------------------------
// buildMySQLPreviewConfig: extra params edge cases
// ---------------------------------------------------------------------------

func TestBuildMySQLPreviewConfig_EdgeCases(t *testing.T) {
	base := &datasource.ConnectionConfig{
		Host: "db.local", Port: 3306, Database: "mydb", Username: "root", Password: "pw",
	}

	// No extra params → defaults remain
	cfg := *base
	config, err := buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "utf8mb4", config.Params["charset"])
	assert.Equal(t, "True", config.Params["parseTime"])
	assert.Equal(t, "Local", config.Params["loc"])

	// Extra params without leading "?"
	cfg = *base
	cfg.ExtraParams = "tls=skip-verify&timeout=30"
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "skip-verify", config.Params["tls"])
	assert.Equal(t, "30", config.Params["timeout"])

	// Only "?" → no additional params
	cfg = *base
	cfg.ExtraParams = "?"
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "utf8mb4", config.Params["charset"])

	// Malformed: key without "="
	cfg = *base
	cfg.ExtraParams = "?justkey&valid=yes"
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "yes", config.Params["valid"])
	_, hasKey := config.Params["justkey"]
	assert.False(t, hasKey, "param without '=' should be skipped")

	// Empty key: "=novalue"
	cfg = *base
	cfg.ExtraParams = "?=novalue&ok=1"
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "1", config.Params["ok"])
	_, hasEmpty := config.Params[""]
	assert.False(t, hasEmpty, "empty key should be skipped")

	// Value containing "="
	cfg = *base
	cfg.ExtraParams = "?conn=str=ing"
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "str=ing", config.Params["conn"])

	// Empty value
	cfg = *base
	cfg.ExtraParams = "?blank="
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "", config.Params["blank"])

	// Multiple consecutive "&"
	cfg = *base
	cfg.ExtraParams = "?&&a=1&&b=2&&"
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "1", config.Params["a"])
	assert.Equal(t, "2", config.Params["b"])

	// Whitespace around key/value
	cfg = *base
	cfg.ExtraParams = "  ?  key = value  "
	config, err = buildMySQLPreviewConfig(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "value", config.Params["key"])
}

// ---------------------------------------------------------------------------
// previewCellCount: additional coverage
// ---------------------------------------------------------------------------

func TestPreviewCellCount_Additional(t *testing.T) {
	// Empty slice
	assert.Zero(t, previewCellCount([]map[string]interface{}{}))
	// Single empty row
	assert.Zero(t, previewCellCount([]map[string]interface{}{{}}))
	// Mixed row sizes
	assert.Equal(t, 6, previewCellCount([]map[string]interface{}{
		{"a": 1, "b": 2},
		{"x": 10},
		{"p": 1, "q": 2, "r": 3},
	}))
}
