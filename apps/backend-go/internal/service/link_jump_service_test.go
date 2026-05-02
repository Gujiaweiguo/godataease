package service

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"

	"dataease/backend/internal/repository"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const linkJumpServiceSQLiteDriverName = "sqlite3_link_jump_service_test"

var linkJumpServiceSQLiteDriverOnce sync.Once

func TestHasQuotedComponentID(t *testing.T) {
	assert.True(t, hasQuotedComponentID(`[{"id":"53012"}]`, 53012))
	assert.False(t, hasQuotedComponentID(`[{"id":"53012"}]`, 3012))
	assert.True(t, hasQuotedComponentID(`{"views":["53012"]}`, 53012))
}

func TestLinkJumpService_ViewTableDetailList_AvoidsSubstringMatches(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceViewTableData(t, db)

	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	result, err := svc.ViewTableDetailList(53011)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, `[{"id":"53012","component":"UserView"}]`, result.ComponentData)
	require.Len(t, result.ComponentView, 1)
	assert.Equal(t, int64(53012), result.ComponentView[0].ID)
	assert.Equal(t, "Exact View", result.ComponentView[0].Title)
	require.Len(t, result.ComponentView[0].TableFields, 1)
	assert.Equal(t, int64(53105), result.ComponentView[0].TableFields[0].ID)
	assert.Empty(t, result.OutParamsJumpInfo)
}

func setupLinkJumpServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	registerLinkJumpServiceSQLiteDriver(t)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Dialector{DriverName: linkJumpServiceSQLiteDriverName, DSN: dsn}, &gorm.Config{})
	require.NoError(t, err)

	linkJumpServiceMustExec(t, db, `CREATE TABLE core_chart_view (
		id INTEGER PRIMARY KEY,
		title TEXT,
		scene_id INTEGER,
		table_id INTEGER,
		type TEXT
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE core_dataset_table_field (
		id INTEGER PRIMARY KEY,
		dataset_group_id INTEGER,
		origin_name TEXT,
		name TEXT,
		de_type INTEGER,
		type TEXT
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE data_visualization_info (
		id INTEGER PRIMARY KEY,
		type TEXT,
		component_data TEXT
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE visualization_outer_params (
		id INTEGER PRIMARY KEY,
		params_id INTEGER,
		visualization_id INTEGER,
		target_view_id INTEGER,
		target_field_id INTEGER,
		copy_from INTEGER,
		copy_id INTEGER
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE visualization_outer_params_info (
		id INTEGER PRIMARY KEY,
		params_info_id INTEGER,
		params_id INTEGER,
		param_name TEXT,
		source_view_id INTEGER,
		source_field_id INTEGER,
		copy_from INTEGER,
		copy_id INTEGER
	)`)

	return db
}

func registerLinkJumpServiceSQLiteDriver(t *testing.T) {
	t.Helper()
	linkJumpServiceSQLiteDriverOnce.Do(func() {
		sql.Register(linkJumpServiceSQLiteDriverName, &sqlite3.SQLiteDriver{ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			if err := conn.RegisterFunc("CONCAT", func(args ...any) string {
				var builder strings.Builder
				for _, arg := range args {
					_, _ = fmt.Fprint(&builder, arg)
				}
				return builder.String()
			}, true); err != nil {
				return err
			}

			return conn.RegisterFunc("LOCATE", func(substr any, str any) int {
				index := strings.Index(fmt.Sprint(str), fmt.Sprint(substr))
				if index < 0 {
					return 0
				}
				return index + 1
			}, true)
		}})
	})
}

func seedLinkJumpServiceViewTableData(t *testing.T, db *gorm.DB) {
	t.Helper()
	linkJumpServiceMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`, 53011, "dashboard", `[{"id":"53012","component":"UserView"}]`)
	linkJumpServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`, 53105, 53104, "city_id", "City ID", 2, "int")
	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (?, ?, ?, ?, ?)`, 53012, "Exact View", 53011, 53104, "bar")
	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (?, ?, ?, ?, ?)`, 3012, "Substring View", 53011, 53104, "bar")
}

func linkJumpServiceMustExec(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	require.NoError(t, db.Exec(query, args...).Error)
}
