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

const outerParamsServiceSQLiteDriverName = "sqlite3_outer_params_service_test"

var outerParamsServiceSQLiteDriverOnce sync.Once

func TestOuterParamsService_QueryWithVisualizationId(t *testing.T) {
	t.Run("returns nested config when snapshot data exists", func(t *testing.T) {
		svc, db := setupOuterParamsServiceTest(t)
		seedSnapshotVisualization(t, db, "dv-1", `["101"]`)
		outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_visualization_outer_params (params_id, visualization_id, checked, remark) VALUES ('params-1', 'dv-1', 1, 'remark')`)
		outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_visualization_outer_params_info (params_info_id, params_id, param_name, checked, required, default_value, enabled_default) VALUES ('info-1', 'params-1', 'region', 1, 1, 'CN', 1)`)
		outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_visualization_outer_params_target_view_info (target_id, params_info_id, target_view_id, target_ds_id, target_field_id) VALUES ('target-1', 'info-1', '101', 'ds-1', 'field-1')`)

		result, err := svc.QueryWithVisualizationId("dv-1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "params-1", result.ParamsID)
		assert.Equal(t, "dv-1", result.VisualizationID)
		assert.True(t, result.Checked)
		require.Len(t, result.OuterParamsInfoArray, 1)
		assert.Equal(t, "region", result.OuterParamsInfoArray[0].ParamName)
		require.Len(t, result.OuterParamsInfoArray[0].TargetViewInfoList, 1)
		assert.Equal(t, "101", result.OuterParamsInfoArray[0].TargetViewInfoList[0].TargetViewID)
	})

	t.Run("returns empty info when unchecked and no params id", func(t *testing.T) {
		svc, db := setupOuterParamsServiceTest(t)
		seedSnapshotVisualization(t, db, "dv-2", `[]`)

		result, err := svc.QueryWithVisualizationId("dv-2")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "dv-2", result.VisualizationID)
		assert.False(t, result.Checked)
		assert.Empty(t, result.ParamsID)
		assert.Empty(t, result.OuterParamsInfoArray)
	})

	t.Run("returns repo error", func(t *testing.T) {
		svc := setupClosedOuterParamsServiceTest(t)
		result, err := svc.QueryWithVisualizationId("dv-1")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestOuterParamsService_UpdateOuterParamsSet(t *testing.T) {
	t.Run("validates visualization id", func(t *testing.T) {
		svc, _ := setupOuterParamsServiceTest(t)
		err := svc.UpdateOuterParamsSet(&OuterParamsDTO{})
		require.EqualError(t, err, "visualizationId is required")
	})

	t.Run("recreates snapshot data and preserves params info id by name", func(t *testing.T) {
		svc, db := setupOuterParamsServiceTest(t)
		seedSnapshotVisualization(t, db, "dv-1", `[]`)
		outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_visualization_outer_params (params_id, visualization_id, checked, remark) VALUES ('old-params', 'dv-1', 1, 'old')`)
		outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_visualization_outer_params_info (params_info_id, params_id, param_name, checked, required, default_value, enabled_default) VALUES ('keep-info', 'old-params', 'region', 1, 0, '', 0)`)
		outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_visualization_outer_params_target_view_info (target_id, params_info_id, target_view_id, target_ds_id, target_field_id) VALUES ('old-target', 'keep-info', 'old-view', 'old-ds', 'old-field')`)

		dto := &OuterParamsDTO{
			VisualizationID: "dv-1",
			Checked:         true,
			Remark:          "new remark",
			OuterParamsInfoArray: []OuterParamsInfoDTO{{
				ParamName:      "region",
				Checked:        true,
				Required:       true,
				DefaultValue:   "CN",
				EnabledDefault: true,
				TargetViewInfoList: []OuterParamsTargetViewInfoDTO{{
					TargetViewID:  "101",
					TargetDsID:    "ds-1",
					TargetFieldID: "field-1",
				}},
			}},
		}

		require.NoError(t, svc.UpdateOuterParamsSet(dto))
		assert.NotEmpty(t, dto.ParamsID)

		var paramsRows []struct {
			ParamsID string
			Remark   string
			Checked  bool
		}
		require.NoError(t, db.Table("snapshot_visualization_outer_params").Where("visualization_id = ?", "dv-1").Find(&paramsRows).Error)
		require.Len(t, paramsRows, 1)
		assert.Equal(t, dto.ParamsID, paramsRows[0].ParamsID)
		assert.Equal(t, "new remark", paramsRows[0].Remark)
		assert.True(t, paramsRows[0].Checked)

		var infoRows []struct {
			ParamsInfoID   string
			ParamsID       string
			ParamName      string
			Required       bool
			DefaultValue   string
			EnabledDefault bool
		}
		require.NoError(t, db.Table("snapshot_visualization_outer_params_info").Find(&infoRows).Error)
		require.Len(t, infoRows, 1)
		assert.Equal(t, "keep-info", infoRows[0].ParamsInfoID)
		assert.Equal(t, dto.ParamsID, infoRows[0].ParamsID)
		assert.Equal(t, "region", infoRows[0].ParamName)
		assert.True(t, infoRows[0].Required)
		assert.Equal(t, "CN", infoRows[0].DefaultValue)
		assert.True(t, infoRows[0].EnabledDefault)

		var targetCount int64
		require.NoError(t, db.Table("snapshot_visualization_outer_params_target_view_info").Count(&targetCount).Error)
		assert.Equal(t, int64(1), targetCount)
	})

	t.Run("returns delete error", func(t *testing.T) {
		svc := setupClosedOuterParamsServiceTest(t)
		err := svc.UpdateOuterParamsSet(&OuterParamsDTO{VisualizationID: "dv-1"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete existing outer params")
	})

	t.Run("returns target creation error", func(t *testing.T) {
		svc, db := setupOuterParamsServiceTest(t)
		seedSnapshotVisualization(t, db, "dv-1", `[]`)
		outerParamsServiceMustExec(t, db, `CREATE TRIGGER deny_outer_target_insert BEFORE INSERT ON snapshot_visualization_outer_params_target_view_info BEGIN SELECT RAISE(FAIL, 'deny target'); END;`)

		err := svc.UpdateOuterParamsSet(&OuterParamsDTO{
			VisualizationID: "dv-1",
			OuterParamsInfoArray: []OuterParamsInfoDTO{{
				ParamName: "region",
				TargetViewInfoList: []OuterParamsTargetViewInfoDTO{{
					TargetViewID:  "101",
					TargetDsID:    "ds-1",
					TargetFieldID: "field-1",
				}},
			}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create outer params target view info")
	})
}

func TestOuterParamsService_GetOuterParamsInfo(t *testing.T) {
	t.Run("returns empty maps when no runtime params exist", func(t *testing.T) {
		svc, db := setupOuterParamsServiceTest(t)
		outerParamsServiceMustExec(t, db, `INSERT INTO visualization_outer_params (params_id, visualization_id, checked, remark, copy_from, copy_id) VALUES ('params-1', 'dv-1', 0, 'remark', '', '')`)

		result, err := svc.GetOuterParamsInfo("dv-1")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result.OuterParamsInfoMap)
		assert.Empty(t, result.OuterParamsInfoBaseMap)
	})

	t.Run("returns repo error", func(t *testing.T) {
		svc := setupClosedOuterParamsServiceTest(t)
		result, err := svc.GetOuterParamsInfo("dv-1")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestOuterParamsService_QueryDsWithVisualizationId(t *testing.T) {
	t.Run("builds grouped dataset response and deduplicates chart views", func(t *testing.T) {
		svc, db := setupOuterParamsServiceTest(t)
		seedSnapshotVisualization(t, db, "dv-1", `[{"id":"101"}]`)
		outerParamsServiceMustExec(t, db, `INSERT INTO core_dataset_group (id, name, pid, level, node_type, type, mode) VALUES (10, 'Sales Dataset', 0, 1, 'group', 'db', 1)`)
		outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, scene_id, table_id, type) VALUES (101, 'Sales View', 'dv-1', 10, 'bar')`)
		outerParamsServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, name, de_type, origin_name) VALUES (1001, 10, 'City', 2, 'city')`)
		outerParamsServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, name, de_type, origin_name) VALUES (1002, 10, 'Amount', 3, 'amount')`)

		result, err := svc.QueryDsWithVisualizationId("dv-1")
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, int64(10), result[0].ID)
		assert.Equal(t, "Sales Dataset", result[0].Name)
		require.Len(t, result[0].DatasetFields, 2)
		require.Len(t, result[0].DatasetViews, 1)
		assert.Equal(t, int64(101), result[0].DatasetViews[0].ChartID)
		assert.Equal(t, "Sales View", result[0].DatasetViews[0].ChartName)
	})

	t.Run("returns repo error", func(t *testing.T) {
		svc := setupClosedOuterParamsServiceTest(t)
		result, err := svc.QueryDsWithVisualizationId("dv-1")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func setupOuterParamsServiceTest(t *testing.T) (*OuterParamsService, *gorm.DB) {
	t.Helper()
	registerOuterParamsServiceSQLiteDriver(t)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Dialector{DriverName: outerParamsServiceSQLiteDriverName, DSN: dsn}, &gorm.Config{})
	require.NoError(t, err)

	outerParamsServiceMustExec(t, db, `CREATE TABLE snapshot_data_visualization_info (id TEXT PRIMARY KEY, component_data TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_outer_params (params_id TEXT PRIMARY KEY, visualization_id TEXT, checked BOOLEAN, remark TEXT, copy_from TEXT, copy_id TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_outer_params_info (params_info_id TEXT PRIMARY KEY, params_id TEXT, param_name TEXT, checked BOOLEAN, required BOOLEAN, default_value TEXT, enabled_default BOOLEAN, copy_from TEXT, copy_id TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_outer_params_target_view_info (target_id TEXT PRIMARY KEY, params_info_id TEXT, target_view_id TEXT, target_ds_id TEXT, target_field_id TEXT, copy_from TEXT, copy_id TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE visualization_outer_params (params_id TEXT PRIMARY KEY, visualization_id TEXT, checked BOOLEAN, remark TEXT, copy_from TEXT, copy_id TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE visualization_outer_params_info (params_info_id TEXT PRIMARY KEY, params_id TEXT, param_name TEXT, checked BOOLEAN, required BOOLEAN, default_value TEXT, enabled_default BOOLEAN, copy_from TEXT, copy_id TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE visualization_outer_params_target_view_info (target_id TEXT PRIMARY KEY, params_info_id TEXT, target_view_id TEXT, target_ds_id TEXT, target_field_id TEXT, copy_from TEXT, copy_id TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE core_dataset_group (id INTEGER PRIMARY KEY, name TEXT, pid INTEGER, level INTEGER, node_type TEXT, type TEXT, mode INTEGER)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE snapshot_core_chart_view (id INTEGER PRIMARY KEY, title TEXT, scene_id TEXT, table_id INTEGER, type TEXT)`)
	outerParamsServiceMustExec(t, db, `CREATE TABLE core_dataset_table_field (id INTEGER PRIMARY KEY, dataset_group_id INTEGER, name TEXT, de_type INTEGER, origin_name TEXT)`)

	return NewOuterParamsService(repository.NewOuterParamsRepository(db)), db
}

func setupClosedOuterParamsServiceTest(t *testing.T) *OuterParamsService {
	t.Helper()
	svc, db := setupOuterParamsServiceTest(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return svc
}

func seedSnapshotVisualization(t *testing.T, db *gorm.DB, visualizationID string, componentData string) {
	t.Helper()
	outerParamsServiceMustExec(t, db, `INSERT INTO snapshot_data_visualization_info (id, component_data) VALUES (?, ?)`, visualizationID, componentData)
}

func registerOuterParamsServiceSQLiteDriver(t *testing.T) {
	t.Helper()
	outerParamsServiceSQLiteDriverOnce.Do(func() {
		sql.Register(outerParamsServiceSQLiteDriverName, &sqlite3.SQLiteDriver{ConnectHook: func(conn *sqlite3.SQLiteConn) error {
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

func outerParamsServiceMustExec(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	require.NoError(t, db.Exec(query, args...).Error)
}
