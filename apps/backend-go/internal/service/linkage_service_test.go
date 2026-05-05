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

const linkageServiceSQLiteDriverName = "sqlite3_linkage_service_test"

var linkageServiceSQLiteDriverOnce sync.Once

func TestLinkageService_SaveLinkage(t *testing.T) {
	t.Run("validates required fields", func(t *testing.T) {
		cases := []struct {
			name string
			req  *LinkageRequest
			want string
		}{
			{name: "missing source view id", req: &LinkageRequest{DvID: 1}, want: "sourceViewId is required"},
			{name: "missing dv id", req: &LinkageRequest{SourceViewID: 1}, want: "dvId is required"},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				svc, _ := setupLinkageServiceTest(t)
				err := svc.SaveLinkage(tc.req)
				require.EqualError(t, err, tc.want)
			})
		}
	})

	t.Run("recreates linkage rows and skips self target", func(t *testing.T) {
		svc, db := setupLinkageServiceTest(t)
		linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage (id, dv_id, source_view_id, target_view_id, update_time, update_people, linkage_active) VALUES (1, 1, 100, 999, 1, '', 1)`)
		linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage_field (id, linkage_id, source_field, target_field, update_time) VALUES (11, 1, 1, 2, 1)`)

		err := svc.SaveLinkage(&LinkageRequest{
			DvID:         1,
			SourceViewID: 100,
			LinkageInfo: []LinkageInfoDTO{
				{TargetViewID: 100, LinkageActive: true, LinkageFields: []LinkageFieldVO{{SourceField: 1, TargetField: 2}}},
				{TargetViewID: 200, LinkageActive: true, LinkageFields: []LinkageFieldVO{{SourceField: 10, TargetField: 20}}},
				{TargetViewID: 201, LinkageActive: false, LinkageFields: []LinkageFieldVO{{SourceField: 30, TargetField: 40}}},
			},
		})
		require.NoError(t, err)

		var linkageCount int64
		require.NoError(t, db.Table("snapshot_visualization_linkage").Where("dv_id = ? AND source_view_id = ?", 1, 100).Count(&linkageCount).Error)
		assert.Equal(t, int64(2), linkageCount)

		var fieldCount int64
		require.NoError(t, db.Table("snapshot_visualization_linkage_field").Count(&fieldCount).Error)
		assert.Equal(t, int64(1), fieldCount)

		var targets []int64
		require.NoError(t, db.Table("snapshot_visualization_linkage").Where("dv_id = ? AND source_view_id = ?", 1, 100).Order("target_view_id").Pluck("target_view_id", &targets).Error)
		assert.Equal(t, []int64{200, 201}, targets)
	})

	t.Run("returns delete error", func(t *testing.T) {
		svc := setupClosedLinkageServiceTest(t)
		err := svc.SaveLinkage(&LinkageRequest{DvID: 1, SourceViewID: 100})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "delete existing linkage")
	})

	t.Run("returns create linkage field error", func(t *testing.T) {
		svc, db := setupLinkageServiceTest(t)
		linkageServiceMustExec(t, db, `CREATE TRIGGER deny_linkage_field_insert BEFORE INSERT ON snapshot_visualization_linkage_field BEGIN SELECT RAISE(FAIL, 'deny linkage field'); END;`)

		err := svc.SaveLinkage(&LinkageRequest{
			DvID:         1,
			SourceViewID: 100,
			LinkageInfo: []LinkageInfoDTO{{
				TargetViewID:  200,
				LinkageActive: true,
				LinkageFields: []LinkageFieldVO{{SourceField: 10, TargetField: 20}},
			}},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "create linkage field")
	})
}

func TestLinkageService_GetViewLinkageGather(t *testing.T) {
	t.Run("returns empty when no target ids", func(t *testing.T) {
		svc, _ := setupLinkageServiceTest(t)
		result, err := svc.GetViewLinkageGather(&LinkageRequest{DvID: 1, SourceViewID: 100})
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("aggregates rows and dataset fields", func(t *testing.T) {
		svc, db := setupLinkageServiceTest(t)
		seedLinkageGatherData(t, db)

		result, err := svc.GetViewLinkageGather(&LinkageRequest{
			DvID:          1,
			SourceViewID:  100,
			TargetViewIds: []int64{200, 201},
			ResourceTable: "snapshot",
		})
		require.NoError(t, err)
		require.Len(t, result, 2)

		active := result["200"]
		assert.Equal(t, int64(200), active.TargetViewID)
		assert.Equal(t, "bar", active.TargetViewType)
		assert.True(t, active.LinkageActive)
		require.Len(t, active.TargetViewFields, 1)
		assert.Equal(t, int64(301), active.TargetViewFields[0].ID)
		assert.Equal(t, []LinkageFieldVO{{SourceField: 10, TargetField: 20}, {SourceField: 11, TargetField: 21}}, active.LinkageFields)

		inactive := result["201"]
		assert.Equal(t, int64(201), inactive.TargetViewID)
		assert.False(t, inactive.LinkageActive)
		assert.Empty(t, inactive.LinkageFields)
	})

	t.Run("array wrapper returns gathered items", func(t *testing.T) {
		svc, db := setupLinkageServiceTest(t)
		seedLinkageGatherData(t, db)

		result, err := svc.GetViewLinkageGatherArray(&LinkageRequest{
			DvID:          1,
			SourceViewID:  100,
			TargetViewIds: []int64{200},
			ResourceTable: "snapshot",
		})
		require.NoError(t, err)
		require.Len(t, result, 1)
		assert.Equal(t, int64(200), result[0].TargetViewID)
	})

	t.Run("returns repo error", func(t *testing.T) {
		svc := setupClosedLinkageServiceTest(t)
		result, err := svc.GetViewLinkageGather(&LinkageRequest{DvID: 1, SourceViewID: 100, TargetViewIds: []int64{200}})
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestLinkageService_GetVisualizationAllLinkageInfo(t *testing.T) {
	svc, db := setupLinkageServiceTest(t)
	seedLinkageInfoData(t, db)

	result, err := svc.GetVisualizationAllLinkageInfo(1, "snapshot")
	require.NoError(t, err)
	assert.Equal(t, map[string][]string{"100#10": {"200#20"}}, result)
}

func TestLinkageService_UpdateLinkageActive(t *testing.T) {
	t.Run("updates chart active flag and returns refreshed info", func(t *testing.T) {
		svc, db := setupLinkageServiceTest(t)
		seedLinkageInfoData(t, db)
		linkageServiceMustExec(t, db, `UPDATE snapshot_core_chart_view SET linkage_active = 0 WHERE id = 100`)

		result, err := svc.UpdateLinkageActive(&LinkageRequest{DvID: 1, SourceViewID: 100, ActiveStatus: true})
		require.NoError(t, err)
		assert.Equal(t, map[string][]string{"100#10": {"200#20"}}, result)

		var active bool
		require.NoError(t, db.Table("snapshot_core_chart_view").Select("linkage_active").Where("id = ?", 100).Scan(&active).Error)
		assert.True(t, active)
	})

	t.Run("returns update error", func(t *testing.T) {
		svc := setupClosedLinkageServiceTest(t)
		result, err := svc.UpdateLinkageActive(&LinkageRequest{DvID: 1, SourceViewID: 100, ActiveStatus: true})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "update linkage active")
	})
}

func TestLinkageService_RemoveLinkage(t *testing.T) {
	t.Run("deletes linkage and fields", func(t *testing.T) {
		svc, db := setupLinkageServiceTest(t)
		linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage (id, dv_id, source_view_id, target_view_id, update_time, update_people, linkage_active) VALUES (1, 1, 100, 200, 1, '', 1)`)
		linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage_field (id, linkage_id, source_field, target_field, update_time) VALUES (11, 1, 10, 20, 1)`)

		require.NoError(t, svc.RemoveLinkage(&LinkageRequest{DvID: 1, SourceViewID: 100}))

		var linkageCount int64
		require.NoError(t, db.Table("snapshot_visualization_linkage").Count(&linkageCount).Error)
		assert.Zero(t, linkageCount)

		var fieldCount int64
		require.NoError(t, db.Table("snapshot_visualization_linkage_field").Count(&fieldCount).Error)
		assert.Zero(t, fieldCount)
	})

	t.Run("returns repo error", func(t *testing.T) {
		svc := setupClosedLinkageServiceTest(t)
		err := svc.RemoveLinkage(&LinkageRequest{DvID: 1, SourceViewID: 100})
		require.Error(t, err)
	})
}

func setupLinkageServiceTest(t *testing.T) (*LinkageService, *gorm.DB) {
	t.Helper()
	registerLinkageServiceSQLiteDriver(t)
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Dialector{DriverName: linkageServiceSQLiteDriverName, DSN: dsn}, &gorm.Config{})
	require.NoError(t, err)

	linkageServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_linkage (id INTEGER PRIMARY KEY, dv_id INTEGER, source_view_id INTEGER, target_view_id INTEGER, update_time INTEGER, update_people TEXT, linkage_active BOOLEAN, ext1 TEXT, ext2 TEXT, copy_from INTEGER, copy_id INTEGER)`)
	linkageServiceMustExec(t, db, `CREATE TABLE snapshot_visualization_linkage_field (id INTEGER PRIMARY KEY, linkage_id INTEGER, source_field INTEGER, target_field INTEGER, update_time INTEGER, copy_from INTEGER, copy_id INTEGER)`)
	linkageServiceMustExec(t, db, `CREATE TABLE snapshot_core_chart_view (id INTEGER PRIMARY KEY, title TEXT, type TEXT, table_id INTEGER, linkage_active BOOLEAN, update_time INTEGER)`)
	linkageServiceMustExec(t, db, `CREATE TABLE core_dataset_table_field (id INTEGER PRIMARY KEY, dataset_group_id INTEGER, origin_name TEXT, name TEXT, de_type INTEGER)`)

	return NewLinkageService(repository.NewLinkageRepository(db)), db
}

func setupClosedLinkageServiceTest(t *testing.T) *LinkageService {
	t.Helper()
	svc, db := setupLinkageServiceTest(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return svc
}

func seedLinkageGatherData(t *testing.T, db *gorm.DB) {
	t.Helper()
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, type, table_id, linkage_active, update_time) VALUES (200, 'Sales', 'bar', 300, 1, 1)`)
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, type, table_id, linkage_active, update_time) VALUES (201, 'Profit', 'line', 301, 0, 1)`)
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage (id, dv_id, source_view_id, target_view_id, update_time, update_people, linkage_active) VALUES (1, 1, 100, 200, 1, '', 1)`)
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage_field (id, linkage_id, source_field, target_field, update_time) VALUES (11, 1, 10, 20, 1)`)
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage_field (id, linkage_id, source_field, target_field, update_time) VALUES (12, 1, 11, 21, 1)`)
	linkageServiceMustExec(t, db, `INSERT INTO core_dataset_table_field (id, dataset_group_id, origin_name, name, de_type) VALUES (301, 300, 'city', 'City', 2)`)
}

func seedLinkageInfoData(t *testing.T, db *gorm.DB) {
	t.Helper()
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_core_chart_view (id, title, type, table_id, linkage_active, update_time) VALUES (100, 'Source', 'bar', 300, 1, 1)`)
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage (id, dv_id, source_view_id, target_view_id, update_time, update_people, linkage_active) VALUES (1, 1, 100, 200, 1, '', 1)`)
	linkageServiceMustExec(t, db, `INSERT INTO snapshot_visualization_linkage_field (id, linkage_id, source_field, target_field, update_time) VALUES (11, 1, 10, 20, 1)`)
}

func registerLinkageServiceSQLiteDriver(t *testing.T) {
	t.Helper()
	linkageServiceSQLiteDriverOnce.Do(func() {
		sql.Register(linkageServiceSQLiteDriverName, &sqlite3.SQLiteDriver{ConnectHook: func(conn *sqlite3.SQLiteConn) error {
			return conn.RegisterFunc("CONCAT", func(args ...any) string {
				var builder strings.Builder
				for _, arg := range args {
					_, _ = fmt.Fprint(&builder, arg)
				}
				return builder.String()
			}, true)
		}})
	})
}

func linkageServiceMustExec(t *testing.T, db *gorm.DB, query string, args ...any) {
	t.Helper()
	require.NoError(t, db.Exec(query, args...).Error)
}
