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

// ---------------------------------------------------------------------------
// Pure-function tests (no DB needed)
// ---------------------------------------------------------------------------

func TestHasQuotedComponentID(t *testing.T) {
	assert.True(t, hasQuotedComponentID(`[{"id":"53012"}]`, 53012))
	assert.False(t, hasQuotedComponentID(`[{"id":"53012"}]`, 3012))
	assert.True(t, hasQuotedComponentID(`{"views":["53012"]}`, 53012))
}

func TestIsSnapshotTable(t *testing.T) {
	assert.True(t, isSnapshotTable("snapshot"))
	assert.False(t, isSnapshotTable("visualization"))
	assert.False(t, isSnapshotTable(""))
	assert.False(t, isSnapshotTable("Snapshot"))
}

func TestGenerateLinkJumpID(t *testing.T) {
	id1 := generateLinkJumpID()
	id2 := generateLinkJumpID()
	assert.NotEqual(t, id1, id2, "two generated IDs should differ")
	assert.NotZero(t, id1)
	assert.NotZero(t, id2)
}

func TestBuildLinkJumpInfoArray_Empty(t *testing.T) {
	result := buildLinkJumpInfoArray(nil)
	assert.Empty(t, result)

	result = buildLinkJumpInfoArray([]repository.LinkJumpInfoFlatRow{})
	assert.Empty(t, result)
}

func TestBuildLinkJumpInfoArray_Grouping(t *testing.T) {
	rows := []repository.LinkJumpInfoFlatRow{
		{
			InfoID:          101,
			LinkJumpID:      501,
			LinkType:        "inner",
			JumpType:        "_blank",
			WindowSize:      "middle",
			TargetDvID:      200,
			SourceFieldID:   301,
			Content:         "",
			Checked:         true,
			AttachParams:    false,
			SourceFieldName: "Province",
			SourceDeType:    0,
			TargetDvType:    "dashboard",
			TargetID:        1001,
			TargetViewID:    "400",
			TargetFieldID:   "401",
			TargetType:      "view",
		},
		{
			InfoID:          101,
			LinkJumpID:      501,
			LinkType:        "inner",
			JumpType:        "_blank",
			WindowSize:      "middle",
			TargetDvID:      200,
			SourceFieldID:   301,
			Content:         "",
			Checked:         true,
			AttachParams:    false,
			SourceFieldName: "Province",
			SourceDeType:    0,
			TargetDvType:    "dashboard",
			TargetID:        1002,
			TargetViewID:    "500",
			TargetFieldID:   "501",
			TargetType:      "filter",
		},
	}

	result := buildLinkJumpInfoArray(rows)
	require.Len(t, result, 1)

	info := result[0]
	assert.Equal(t, int64(101), info.ID)
	assert.Equal(t, int64(501), info.LinkJumpID)
	assert.Equal(t, "inner", info.LinkType)
	assert.Equal(t, "_blank", info.JumpType)
	assert.Equal(t, "middle", info.WindowSize)
	assert.Equal(t, int64(200), info.TargetDvID)
	assert.Equal(t, int64(301), info.SourceFieldID)
	assert.Equal(t, "Province", info.SourceFieldName)
	assert.Equal(t, 0, info.SourceDeType)
	assert.Equal(t, "dashboard", info.TargetDvType)
	require.Len(t, info.TargetViewInfoList, 2)
	assert.Equal(t, int64(1001), info.TargetViewInfoList[0].TargetID)
	assert.Equal(t, "view", info.TargetViewInfoList[0].TargetType)
	assert.Equal(t, int64(1002), info.TargetViewInfoList[1].TargetID)
	assert.Equal(t, "filter", info.TargetViewInfoList[1].TargetType)
}

func TestBuildLinkJumpInfoArray_TargetIDZero_Skipped(t *testing.T) {
	rows := []repository.LinkJumpInfoFlatRow{
		{
			InfoID:        201,
			LinkJumpID:    601,
			SourceFieldID: 301,
			Checked:       true,
			TargetDvType:  "dashboard",
			TargetID:      0,
		},
	}

	result := buildLinkJumpInfoArray(rows)
	require.Len(t, result, 1)
	assert.Empty(t, result[0].TargetViewInfoList)
}

func TestBuildLinkJumpInfoArray_MultipleInfoIDs(t *testing.T) {
	rows := []repository.LinkJumpInfoFlatRow{
		{InfoID: 10, LinkJumpID: 1, SourceFieldID: 100, TargetID: 0},
		{InfoID: 20, LinkJumpID: 1, SourceFieldID: 200, TargetID: 0},
	}

	result := buildLinkJumpInfoArray(rows)
	require.Len(t, result, 2)
	assert.Equal(t, int64(10), result[0].ID)
	assert.Equal(t, int64(20), result[1].ID)
}

// ---------------------------------------------------------------------------
// DB-backed service tests
// ---------------------------------------------------------------------------

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

// --- UpdateJumpSet validation errors ---

func TestLinkJumpService_UpdateJumpSet_MissingSourceDvID(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	err := svc.UpdateJumpSet(&LinkJumpDTO{SourceDvID: 0, SourceViewID: 100})
	assert.EqualError(t, err, "sourceDvId is required")
}

func TestLinkJumpService_UpdateJumpSet_MissingSourceViewID(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	err := svc.UpdateJumpSet(&LinkJumpDTO{SourceDvID: 100, SourceViewID: 0})
	assert.EqualError(t, err, "sourceViewId is required")
}

func TestLinkJumpService_UpdateJumpSet_Success(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	dto := &LinkJumpDTO{
		SourceDvID:   1001,
		SourceViewID: 2001,
		LinkJumpInfo: "test-jump-info",
		Checked:      true,
		LinkJumpInfoArray: []LinkJumpInfoDTO{
			{
				LinkType:      "outer",
				JumpType:      "_self",
				WindowSize:    "large",
				TargetDvID:    1002,
				SourceFieldID: 4001,
				Content:       "https://example.com",
				Checked:       true,
				AttachParams:  true,
				TargetViewInfoList: []LinkJumpTargetViewInfoDTO{
					{
						SourceFieldActiveID: 4001,
						TargetViewID:        "2002",
						TargetFieldID:       "4002",
						TargetType:          "filter",
					},
				},
			},
		},
	}

	err := svc.UpdateJumpSet(dto)
	require.NoError(t, err)
	assert.NotZero(t, dto.ID, "UpdateJumpSet should set dto.ID")

	var count int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump WHERE source_dv_id = ? AND source_view_id = ?`, 1001, 2001).Scan(&count).Error)
	assert.Equal(t, int64(1), count)

	var infoCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump_info`).Scan(&infoCount).Error)
	assert.Equal(t, int64(1), infoCount)

	var targetCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump_target_view_info`).Scan(&targetCount).Error)
	assert.Equal(t, int64(1), targetCount)
}

// --- QueryWithViewId no-config branch (row.ID == 0) ---

func TestLinkJumpService_QueryWithViewId_NoConfig(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	result, err := svc.QueryWithViewId(1002, 2002)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Zero(t, result.ID, "no jump config exists for target view")
	assert.NotNil(t, result.LinkJumpInfoArray, "should still return non-nil array even with no config")
}

func TestLinkJumpService_QueryWithViewId_WithConfig(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	seedLinkJumpServiceSnapshotJump(t, db, "test-info", 1)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	result, err := svc.QueryWithViewId(1001, 2001)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(5001), result.ID)
	assert.Equal(t, int64(1001), result.SourceDvID)
	assert.Equal(t, int64(2001), result.SourceViewID)
	assert.Equal(t, "test-info", result.LinkJumpInfo)
	assert.True(t, result.Checked)
	require.Len(t, result.LinkJumpInfoArray, 1)

	info := result.LinkJumpInfoArray[0]
	assert.Equal(t, int64(5001), info.LinkJumpID)
	assert.Equal(t, "inner", info.LinkType)
	assert.Equal(t, int64(1002), info.TargetDvID)
	assert.Equal(t, int64(4001), info.SourceFieldID)
	assert.Equal(t, "Province", info.SourceFieldName)
	assert.Equal(t, "dashboard", info.TargetDvType)
	require.Len(t, info.TargetViewInfoList, 1)
	assert.Equal(t, int64(7001), info.TargetViewInfoList[0].TargetID)
	assert.Equal(t, "2002", info.TargetViewInfoList[0].TargetViewID)
	assert.Equal(t, "4002", info.TargetViewInfoList[0].TargetFieldID)
}

// --- QueryVisualizationJumpInfo filter branches ---

func TestLinkJumpService_QueryVisualizationJumpInfo_NoActiveCharts(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	// Set jump_active=0 so QueryWithDvId returns no rows
	linkJumpServiceMustExec(t, db, `UPDATE snapshot_core_chart_view SET jump_active = 0 WHERE id = 2001`)
	seedLinkJumpServiceSnapshotJump(t, db, "info", 1)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	result, err := svc.QueryVisualizationJumpInfo(1001, "snapshot")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.BaseJumpInfoMap)
}

func TestLinkJumpService_QueryVisualizationJumpInfo_InnerJumpZeroTargetDvID(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump (id, source_dv_id, source_view_id, link_jump_info, checked, copy_from, copy_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		5001, 1001, 2001, "info", 1, 0, 0)
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump_info (id, link_jump_id, link_type, jump_type, target_dv_id, source_field_id, content, checked, attach_params, copy_from, copy_id, window_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		6001, 5001, "inner", "_blank", 0, 4001, "", 1, 0, 0, 0, "middle")

	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))
	result, err := svc.QueryVisualizationJumpInfo(1001, "snapshot")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.BaseJumpInfoMap, "inner link with target_dv_id=0 should be filtered")
}

func TestLinkJumpService_QueryVisualizationJumpInfo_UncheckedInfo(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump (id, source_dv_id, source_view_id, link_jump_info, checked, copy_from, copy_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		5001, 1001, 2001, "info", 1, 0, 0)
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump_info (id, link_jump_id, link_type, jump_type, target_dv_id, source_field_id, content, checked, attach_params, copy_from, copy_id, window_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		6001, 5001, "inner", "_blank", 1002, 4001, "", 0, 0, 0, 0, "middle")

	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))
	result, err := svc.QueryVisualizationJumpInfo(1001, "snapshot")
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.BaseJumpInfoMap, "unchecked info should be filtered")
}

func TestLinkJumpService_QueryVisualizationJumpInfo_ValidSnapshot(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	seedLinkJumpServiceSnapshotJump(t, db, "info", 1)

	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))
	result, err := svc.QueryVisualizationJumpInfo(1001, "snapshot")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.BaseJumpInfoMap, 1)

	key := fmt.Sprintf("%d#%d", 2001, 4001)
	info, ok := result.BaseJumpInfoMap[key]
	require.True(t, ok)
	assert.Equal(t, "inner", info.LinkType)
	assert.Equal(t, int64(1002), info.TargetDvID)
	require.Len(t, info.TargetViewInfoList, 1)
	assert.Equal(t, "2002", info.TargetViewInfoList[0].TargetViewID)
}

func TestLinkJumpService_QueryVisualizationJumpInfo_RuntimeTable(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceRuntimeBaseData(t, db)
	seedLinkJumpServiceRuntimeJump(t, db, "runtime-info", 1)

	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))
	result, err := svc.QueryVisualizationJumpInfo(1001, "visualization")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.BaseJumpInfoMap, 1)

	key := fmt.Sprintf("%d#%d", 2001, 4001)
	info, ok := result.BaseJumpInfoMap[key]
	require.True(t, ok)
	assert.Equal(t, int64(1002), info.TargetDvID)
}

// --- QueryTargetVisualizationJumpInfo empty-source filter ---

func TestLinkJumpService_QueryTargetVisualizationJumpInfo_EmptySourceFilter(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	seedLinkJumpServiceSnapshotJump(t, db, "info", 1)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	fieldID := int64(4001)
	result, err := svc.QueryTargetVisualizationJumpInfo(&LinkJumpRequest{
		SourceDvID:    1001,
		SourceViewID:  2001,
		TargetDvID:    1002,
		SourceFieldID: &fieldID,
		ResourceTable: "snapshot",
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result, 1)

	key := fmt.Sprintf("%d#%d#%d", 2001, 4001, 4001)
	require.Contains(t, result, key)
	assert.Equal(t, []string{fmt.Sprintf("%d#%d", 2002, 4002)}, result[key])
}

func TestLinkJumpService_QueryTargetVisualizationJumpInfo_NoMatch(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	seedLinkJumpServiceSnapshotJump(t, db, "info", 1)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	result, err := svc.QueryTargetVisualizationJumpInfo(&LinkJumpRequest{
		SourceDvID:    9999,
		SourceViewID:  9999,
		TargetDvID:    9999,
		SourceFieldID: nil,
		ResourceTable: "snapshot",
	})
	require.NoError(t, err)
	assert.Empty(t, result)
}

// --- UpdateJumpActive ---

func TestLinkJumpService_UpdateJumpActive(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	seedLinkJumpServiceSnapshotJump(t, db, "info", 1)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	result, err := svc.UpdateJumpActive(&LinkJumpRequest{
		SourceDvID:   1001,
		SourceViewID: 2001,
		ActiveStatus: false,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.BaseJumpInfoMap)

	var jumpActive int
	require.NoError(t, db.Raw(`SELECT jump_active FROM snapshot_core_chart_view WHERE id = ?`, 2001).Scan(&jumpActive).Error)
	assert.Equal(t, 0, jumpActive)
}

// --- RemoveJumpSet ---

func TestLinkJumpService_RemoveJumpSet(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	seedLinkJumpServiceSnapshotJump(t, db, "info", 1)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	err := svc.RemoveJumpSet(&LinkJumpDTO{SourceDvID: 1001, SourceViewID: 2001})
	require.NoError(t, err)

	var jumpCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump WHERE source_dv_id = ? AND source_view_id = ?`, 1001, 2001).Scan(&jumpCount).Error)
	assert.Zero(t, jumpCount)

	var infoCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump_info`).Scan(&infoCount).Error)
	assert.Zero(t, infoCount)

	var targetCount int64
	require.NoError(t, db.Raw(`SELECT COUNT(*) FROM snapshot_visualization_link_jump_target_view_info`).Scan(&targetCount).Error)
	assert.Zero(t, targetCount)
}

// --- GetTableFieldWithViewID ---

func TestLinkJumpService_GetTableFieldWithViewID(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	seedLinkJumpServiceSnapshotBaseData(t, db)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	fields, err := svc.GetTableFieldWithViewID(2001)
	require.NoError(t, err)
	require.Len(t, fields, 1)
	assert.Equal(t, int64(4001), fields[0].ID)
	assert.Equal(t, int64(3001), fields[0].DatasetTableID)
	assert.Equal(t, "province", fields[0].OriginName)
	assert.Equal(t, "Province", fields[0].Name)
}

func TestLinkJumpService_GetTableFieldWithViewID_NoView(t *testing.T) {
	db := setupLinkJumpServiceTestDB(t)
	svc := NewLinkJumpService(repository.NewLinkJumpRepository(db))

	fields, err := svc.GetTableFieldWithViewID(9999)
	require.NoError(t, err)
	assert.Empty(t, fields)
}

// ---------------------------------------------------------------------------
// DB setup helpers
// ---------------------------------------------------------------------------

func setupLinkJumpServiceTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	registerLinkJumpServiceSQLiteDriver(t)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Dialector{DriverName: linkJumpServiceSQLiteDriverName, DSN: dsn}, &gorm.Config{})
	require.NoError(t, err)

	// Core tables
	linkJumpServiceMustExec(t, db, `CREATE TABLE core_chart_view (
		id INTEGER PRIMARY KEY,
		title TEXT,
		scene_id INTEGER,
		table_id INTEGER,
		type TEXT,
		jump_active INTEGER
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

	// Snapshot tables (needed by QueryWithViewId, QueryVisualizationJumpInfo, UpdateJumpSet, etc.)
	linkJumpServiceMustExec(t, db, `CREATE TABLE snapshot_core_chart_view (
		id INTEGER PRIMARY KEY,
		title TEXT,
		scene_id INTEGER,
		table_id INTEGER,
		type TEXT,
		jump_active INTEGER,
		update_time INTEGER
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_link_jump (
		id INTEGER PRIMARY KEY,
		source_dv_id INTEGER,
		source_view_id INTEGER,
		link_jump_info TEXT,
		checked INTEGER,
		copy_from INTEGER,
		copy_id INTEGER
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_link_jump_info (
		id INTEGER PRIMARY KEY,
		link_jump_id INTEGER,
		link_type TEXT,
		jump_type TEXT,
		target_dv_id INTEGER,
		source_field_id INTEGER,
		content TEXT,
		checked INTEGER,
		attach_params INTEGER,
		copy_from INTEGER,
		copy_id INTEGER,
		window_size TEXT
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_link_jump_target_view_info (
		target_id INTEGER PRIMARY KEY,
		link_jump_info_id INTEGER,
		source_field_active_id INTEGER,
		target_view_id TEXT,
		target_field_id TEXT,
		copy_from INTEGER,
		copy_id INTEGER,
		target_type TEXT
	)`)

	// Runtime tables (needed by QueryVisualizationJumpInfo with resourceTable != "snapshot")
	linkJumpServiceMustExec(t, db, `CREATE TABLE visualization_link_jump (
		id INTEGER PRIMARY KEY,
		source_dv_id INTEGER,
		source_view_id INTEGER,
		link_jump_info TEXT,
		checked INTEGER,
		copy_from INTEGER,
		copy_id INTEGER
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE visualization_link_jump_info (
		id INTEGER PRIMARY KEY,
		link_jump_id INTEGER,
		link_type TEXT,
		jump_type TEXT,
		target_dv_id INTEGER,
		source_field_id INTEGER,
		content TEXT,
		checked INTEGER,
		attach_params INTEGER,
		copy_from INTEGER,
		copy_id INTEGER,
		window_size TEXT
	)`)
	linkJumpServiceMustExec(t, db, `CREATE TABLE visualization_link_jump_target_view_info (
		target_id INTEGER PRIMARY KEY,
		link_jump_info_id INTEGER,
		source_field_active_id INTEGER,
		target_view_id TEXT,
		target_field_id TEXT,
		copy_from INTEGER,
		copy_id INTEGER,
		target_type TEXT
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

// ---------------------------------------------------------------------------
// Seed helpers
// ---------------------------------------------------------------------------

func seedLinkJumpServiceViewTableData(t *testing.T, db *gorm.DB) {
	t.Helper()
	linkJumpServiceMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`, 53011, "dashboard", `[{"id":"53012","component":"UserView"}]`)
	linkJumpServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`, 53105, 53104, "city_id", "City ID", 2, "int")
	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (?, ?, ?, ?, ?)`, 53012, "Exact View", 53011, 53104, "bar")
	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type) VALUES (?, ?, ?, ?, ?)`, 3012, "Substring View", 53011, 53104, "bar")
}

// seedLinkJumpServiceSnapshotBaseData seeds the minimal snapshot data needed by most service tests.
func seedLinkJumpServiceSnapshotBaseData(t *testing.T, db *gorm.DB) {
	t.Helper()
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type, jump_active, update_time) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		2001, "Source Chart", 1001, 3001, "bar", 1, 1)
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type, jump_active, update_time) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		2002, "Target Chart", 1002, 3002, "table", 0, 1)

	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type, jump_active) VALUES (?, ?, ?, ?, ?, ?)`,
		2001, "Source Chart", 1001, 3001, "bar", 1)
	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type, jump_active) VALUES (?, ?, ?, ?, ?, ?)`,
		2002, "Target Chart", 1002, 3002, "table", 0)

	linkJumpServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`,
		4001, 3001, "province", "Province", 0, "STRING")
	linkJumpServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`,
		4002, 3002, "city", "City", 0, "STRING")

	linkJumpServiceMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`,
		1001, "dashboard", "[]")
	linkJumpServiceMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`,
		1002, "dashboard", `[{"id":"2002","component":"UserView"}]`)
}

// seedLinkJumpServiceSnapshotJump seeds a complete snapshot jump chain (jump → info → target).
func seedLinkJumpServiceSnapshotJump(t *testing.T, db *gorm.DB, linkJumpInfo string, checked int) {
	t.Helper()
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump (id, source_dv_id, source_view_id, link_jump_info, checked, copy_from, copy_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		5001, 1001, 2001, linkJumpInfo, checked, 0, 0)
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump_info (id, link_jump_id, link_type, jump_type, target_dv_id, source_field_id, content, checked, attach_params, copy_from, copy_id, window_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		6001, 5001, "inner", "_blank", 1002, 4001, "", 1, 1, 0, 0, "middle")
	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_visualization_link_jump_target_view_info (target_id, link_jump_info_id, source_field_active_id, target_view_id, target_field_id, copy_from, copy_id, target_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		7001, 6001, 4001, "2002", "4002", 0, 0, "view")
}

// seedLinkJumpServiceRuntimeBaseData seeds chart views for runtime (non-snapshot) tests.
func seedLinkJumpServiceRuntimeBaseData(t *testing.T, db *gorm.DB) {
	t.Helper()
	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type, jump_active) VALUES (?, ?, ?, ?, ?, ?)`,
		2001, "Source Chart", 1001, 3001, "bar", 1)
	linkJumpServiceMustExec(t, db, `INSERT INTO core_chart_view (id, title, scene_id, table_id, type, jump_active) VALUES (?, ?, ?, ?, ?, ?)`,
		2002, "Target Chart", 1002, 3002, "table", 0)

	linkJumpServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`,
		4001, 3001, "province", "Province", 0, "STRING")
	linkJumpServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type, type) VALUES (?, ?, ?, ?, ?, ?)`,
		4002, 3002, "city", "City", 0, "STRING")

	linkJumpServiceMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`,
		1001, "dashboard", "[]")
	linkJumpServiceMustExec(t, db, `INSERT INTO data_visualization_info (id, type, component_data) VALUES (?, ?, ?)`,
		1002, "dashboard", `[{"id":"2002","component":"UserView"}]`)

	linkJumpServiceMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type, jump_active, update_time) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		2001, "Source Chart", 1001, 3001, "bar", 1, 1)
}

// seedLinkJumpServiceRuntimeJump seeds a complete runtime jump chain.
func seedLinkJumpServiceRuntimeJump(t *testing.T, db *gorm.DB, linkJumpInfo string, checked int) {
	t.Helper()
	linkJumpServiceMustExec(t, db, `INSERT INTO visualization_link_jump (id, source_dv_id, source_view_id, link_jump_info, checked, copy_from, copy_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		5001, 1001, 2001, linkJumpInfo, checked, 0, 0)
	linkJumpServiceMustExec(t, db, `INSERT INTO visualization_link_jump_info (id, link_jump_id, link_type, jump_type, target_dv_id, source_field_id, content, checked, attach_params, copy_from, copy_id, window_size) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		6001, 5001, "inner", "_blank", 1002, 4001, "", 1, 1, 0, 0, "middle")
	linkJumpServiceMustExec(t, db, `INSERT INTO visualization_link_jump_target_view_info (target_id, link_jump_info_id, source_field_active_id, target_view_id, target_field_id, copy_from, copy_id, target_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		7001, 6001, 4001, "2002", "4002", 0, 0, "filter")
}

func linkJumpServiceMustExec(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	require.NoError(t, db.Exec(query, args...).Error)
}
