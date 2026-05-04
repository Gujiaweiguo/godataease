package service

import (
	"context"
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
