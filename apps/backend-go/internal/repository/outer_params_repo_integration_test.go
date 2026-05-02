//go:build integration
// +build integration

package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOuterParamsRepository_GetDatasetGroupsWithFields_AvoidsSubstringMatches(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	ensureOuterParamsDatasetGroupTables(t)
	cleanupTables("core_dataset_table_field", "snapshot_core_chart_view", "snapshot_data_visualization_info", "core_dataset_group")

	repo := NewOuterParamsRepository(testDB)

	require.NoError(t, testDB.Exec(`INSERT INTO core_dataset_group (id, name, pid, level, node_type, type, mode, create_by, create_time, update_by, last_update_time) VALUES (53104, 'group-a', 0, 0, 'dataset', 'dataset', 0, 'tester', 1, 'tester', 1)`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO core_dataset_group (id, name, pid, level, node_type, type, mode, create_by, create_time, update_by, last_update_time) VALUES (53114, 'group-b', 0, 0, 'dataset', 'dataset', 0, 'tester', 1, 'tester', 1)`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO snapshot_data_visualization_info (id, name, pid, org_id, level, node_type, type, canvas_style_data, component_data, mobile_layout, status, sort, create_time, create_by, update_time, update_by, remark, source, delete_flag, delete_time, delete_by, version, content_id, check_version) VALUES ('53101', 'dash-a', '0', '0', 0, 'panel', 'dashboard', '', '[{"id":"53012","component":"UserView"}]', 0, 1, 0, 1, 'tester', 1, 'tester', '', '', 0, 0, '', 3, '', '1')`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type) VALUES (53012, 'Exact View', 53101, 53104, 'bar')`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type) VALUES (3012, 'Substring Only View', 53101, 53114, 'bar')`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (53105, 53104, 'city_id', 'City ID', 2, 'int')`).Error)
	require.NoError(t, testDB.Exec(`INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (53115, 53114, 'country_id', 'Country ID', 2, 'int')`).Error)

	rows, err := repo.GetDatasetGroupsWithFields("53101")
	require.NoError(t, err)
	require.Len(t, rows, 1)

	var exactMatchCount int
	for _, row := range rows {
		assert.Equal(t, int64(53104), row.ID)
		assert.NotEqual(t, int64(53114), row.ID)
		if row.ChartID != nil {
			assert.Equal(t, int64(53012), *row.ChartID)
			assert.NotEqual(t, int64(3012), *row.ChartID)
			exactMatchCount++
		}
	}
	assert.GreaterOrEqual(t, exactMatchCount, 1)
}

func ensureOuterParamsDatasetGroupTables(t *testing.T) {
	t.Helper()
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS snapshot_core_chart_view (
		id BIGINT PRIMARY KEY,
		title VARCHAR(255),
		scene_id BIGINT,
		table_id BIGINT,
		type VARCHAR(64)
	)`).Error)
	require.NoError(t, testDB.Exec(`CREATE TABLE IF NOT EXISTS snapshot_data_visualization_info (
		id VARCHAR(64) PRIMARY KEY,
		name VARCHAR(255),
		pid VARCHAR(64),
		org_id VARCHAR(64),
		level INT,
		node_type VARCHAR(64),
		type VARCHAR(64),
		canvas_style_data TEXT,
		component_data LONGTEXT,
		mobile_layout INT,
		status INT,
		sort INT,
		create_time BIGINT,
		create_by VARCHAR(255),
		update_time BIGINT,
		update_by VARCHAR(255),
		remark TEXT,
		source VARCHAR(255),
		delete_flag TINYINT(1),
		delete_time BIGINT,
		delete_by VARCHAR(255),
		version INT,
		content_id VARCHAR(255),
		check_version VARCHAR(255)
	)`).Error)
	ensureColumnExists(t, "core_dataset_group", "mode", `ALTER TABLE core_dataset_group ADD COLUMN mode INT DEFAULT 0`)
	ensureColumnExists(t, "core_dataset_table_field", "origin_name", `ALTER TABLE core_dataset_table_field ADD COLUMN origin_name VARCHAR(255)`)
	ensureColumnExists(t, "core_dataset_table_field", "name", `ALTER TABLE core_dataset_table_field ADD COLUMN name VARCHAR(255)`)
	ensureColumnExists(t, "core_dataset_table_field", "de_type", `ALTER TABLE core_dataset_table_field ADD COLUMN de_type INT`)
	ensureColumnExists(t, "core_dataset_table_field", "type", `ALTER TABLE core_dataset_table_field ADD COLUMN type VARCHAR(64)`)
}

func ensureColumnExists(t *testing.T, tableName, columnName, ddl string) {
	t.Helper()
	var count int64
	require.NoError(t, testDB.Raw(`
		SELECT COUNT(*)
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, tableName, columnName).Scan(&count).Error)
	if count == 0 {
		require.NoError(t, testDB.Exec(ddl).Error)
	}
}
