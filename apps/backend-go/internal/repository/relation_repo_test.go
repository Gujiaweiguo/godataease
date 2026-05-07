package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRelationRepositoryTest(t *testing.T) (*RelationRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE core_datasource (
			id INTEGER PRIMARY KEY,
			name TEXT,
			create_by TEXT,
			update_time INTEGER
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE core_dataset_table (
			id INTEGER PRIMARY KEY,
			name TEXT,
			datasource_id INTEGER,
			dataset_group_id INTEGER,
			create_by TEXT,
			update_time INTEGER
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE core_chart_view (
			id INTEGER PRIMARY KEY,
			title TEXT,
			table_id INTEGER,
			scene_id INTEGER,
			create_by TEXT,
			update_time INTEGER
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE data_visualization_info (
			id INTEGER PRIMARY KEY,
			name TEXT,
			create_by TEXT,
			update_time INTEGER
		)
	`).Error)

	return NewRelationRepository(db), db
}

func TestRelationRepository_GetDatasourceRelations(t *testing.T) {
	repo, db := setupRelationRepositoryTest(t)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_table (id, name, datasource_id, dataset_group_id, create_by, update_time) VALUES
		(10, 'dataset-a', 1, 101, 'alice', 1710000000000),
		(11, 'dataset-b', 1, 101, 'bob', 1710000001000)
	`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_chart_view (id, title, table_id, scene_id, create_by, update_time) VALUES
		(20, 'chart-a', 10, 30, 'charlie', 1710000002000),
		(21, 'chart-b', 11, NULL, 'dora', 1710000003000)
	`).Error)
	require.NoError(t, db.Exec(`INSERT INTO data_visualization_info (id, name, create_by, update_time) VALUES
		(30, 'dashboard-a', 'eric', 1710000004000)
	`).Error)

	rows, err := repo.GetDatasourceRelations(1)
	require.NoError(t, err)
	require.Len(t, rows, 2)

	require.NotNil(t, rows[0].DatasetID)
	assert.Equal(t, int64(10), *rows[0].DatasetID)
	require.NotNil(t, rows[0].ChartID)
	assert.Equal(t, int64(20), *rows[0].ChartID)
	require.NotNil(t, rows[0].DashboardID)
	assert.Equal(t, int64(30), *rows[0].DashboardID)
	assert.Equal(t, "dashboard-a", *rows[0].DashboardName)

	require.NotNil(t, rows[1].DatasetID)
	assert.Equal(t, int64(11), *rows[1].DatasetID)
	require.NotNil(t, rows[1].ChartID)
	assert.Equal(t, int64(21), *rows[1].ChartID)
	assert.Nil(t, rows[1].DashboardID)
}

func TestRelationRepository_GetDatasetRelations_Empty(t *testing.T) {
	repo, _ := setupRelationRepositoryTest(t)

	rows, err := repo.GetDatasetRelations(999)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRelationRepository_GetPanelRelations(t *testing.T) {
	repo, db := setupRelationRepositoryTest(t)
	require.NoError(t, db.Exec(`INSERT INTO core_datasource (id, name, create_by, update_time) VALUES
		(1, 'datasource-a', 'alice', 1710000000000)
	`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_dataset_table (id, name, datasource_id, dataset_group_id, create_by, update_time) VALUES
		(10, 'dataset-a', 1, 101, 'bob', 1710000001000)
	`).Error)
	require.NoError(t, db.Exec(`INSERT INTO core_chart_view (id, title, table_id, scene_id, create_by, update_time) VALUES
		(20, 'chart-a', 10, 30, 'charlie', 1710000002000)
	`).Error)

	rows, err := repo.GetPanelRelations(30)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	require.NotNil(t, rows[0].ChartID)
	assert.Equal(t, int64(20), *rows[0].ChartID)
	require.NotNil(t, rows[0].DatasetID)
	assert.Equal(t, int64(10), *rows[0].DatasetID)
	require.NotNil(t, rows[0].DatasourceID)
	assert.Equal(t, int64(1), *rows[0].DatasourceID)
	assert.Equal(t, "datasource-a", *rows[0].DatasourceName)
}
