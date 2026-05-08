package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"dataease/backend/internal/domain/auto"
	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/export"
	"dataease/backend/internal/domain/visualization"

	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== Shared helpers ====================

func round8BOpenDB(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	// Register MySQL-compatible LOCATE function for SQLite
	conn, err := sqlDB.Conn(context.Background())
	require.NoError(t, err)
	err = conn.Raw(func(d interface{}) error {
		sc, ok := d.(*sqlite3.SQLiteConn)
		if !ok {
			return nil
		}
		return sc.RegisterFunc("LOCATE", func(substr, str string) int {
			return strings.Index(str, substr) + 1
		}, true)
	})
	require.NoError(t, err)
	_ = conn.Close()

	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(models...))
	return db
}

func round8bPtrInt64(v int64) *int64 { return &v }
func round8bPtrStr(v string) *string { return &v }
func round8bPtrInt(v int) *int       { return &v }
func round8bPtrBool(v bool) *bool    { return &v }

// ==================== OuterParamsRepository (10 functions) ====================

func round8BOuterParamsDB(t *testing.T) *gorm.DB {
	return round8BOpenDB(t,
		&auto.SnapshotVisualizationOuterParam{},
		&auto.SnapshotVisualizationOuterParamsInfo{},
		&auto.SnapshotVisualizationOuterParamsTargetViewInfo{},
		&auto.SnapshotDataVisualizationInfo{},
		&auto.DataVisualizationInfo{},
		&auto.VisualizationOuterParam{},
		&auto.VisualizationOuterParamsInfo{},
		&auto.VisualizationOuterParamsTargetViewInfo{},
		&auto.CoreDatasetGroup{},
		&auto.CoreChartView{},
		&visualization.SnapshotCanvasChartView{},
		&visualization.CanvasChartView{},
		&dataset.CoreDatasetTableField{},
	)
}

func TestRound8B_OuterParams_New(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	require.NotNil(t, repo)
}

func TestRound8B_OuterParams_QueryWithVisualizationId_Found(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)

	require.NoError(t, db.Create(&auto.SnapshotDataVisualizationInfo{ID: "viz-1"}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParam{
		ParamsID: "p-1", VisualizationID: "viz-1", Checked: true,
	}).Error)

	row, err := repo.QueryWithVisualizationId("viz-1")
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, "viz-1", row.VisualizationID)
	assert.True(t, row.Checked)
	assert.Equal(t, "p-1", row.ParamsID)
}

func TestRound8B_OuterParams_QueryWithVisualizationId_NoParams(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)

	require.NoError(t, db.Create(&auto.SnapshotDataVisualizationInfo{ID: "viz-2"}).Error)

	row, err := repo.QueryWithVisualizationId("viz-2")
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.False(t, row.Checked)
	assert.Empty(t, row.ParamsID)
}

func TestRound8B_OuterParams_GetOuterParamsInfoSnapshot_Found(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)

	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParam{
		ParamsID: "sp-1", VisualizationID: "viz-10",
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParamsInfo{
		ParamsInfoID: "spi-1", ParamsID: "sp-1", ParamName: "p1", Checked: true, Required: true,
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParamsTargetViewInfo{
		TargetID: "t-1", ParamsInfoID: "spi-1", TargetViewID: "v100", TargetDsID: "ds1", TargetFieldID: "f1",
	}).Error)

	rows, err := repo.GetOuterParamsInfoSnapshot("viz-10")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "p1", rows[0].ParamName)
	assert.Equal(t, "v100", rows[0].TargetViewID)
}

func TestRound8B_OuterParams_GetOuterParamsInfoSnapshot_Empty(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	rows, err := repo.GetOuterParamsInfoSnapshot("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_OuterParams_GetOuterParamsInfoBase_Found(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)

	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParam{
		ParamsID: "bp-1", VisualizationID: "viz-20",
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParamsInfo{
		ParamsInfoID: "bpi-1", ParamsID: "bp-1", ParamName: "myParam",
	}).Error)

	rows, err := repo.GetOuterParamsInfoBase("viz-20")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "myParam", rows[0].ParamName)
	assert.Equal(t, "bpi-1", rows[0].ParamsInfoID)
}

func TestRound8B_OuterParams_GetOuterParamsInfoBase_Empty(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	rows, err := repo.GetOuterParamsInfoBase("missing")
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_OuterParams_DeleteOuterParamsCascadeSnapshot(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)

	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParam{
		ParamsID: "dp-1", VisualizationID: "viz-30",
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParamsInfo{
		ParamsInfoID: "dpi-1", ParamsID: "dp-1", ParamName: "x",
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationOuterParamsTargetViewInfo{
		TargetID: "dt-1", ParamsInfoID: "dpi-1",
	}).Error)

	require.NoError(t, repo.DeleteOuterParamsCascadeSnapshot("viz-30"))

	rows, err := repo.GetOuterParamsInfoSnapshot("viz-30")
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_OuterParams_DeleteOuterParamsCascadeSnapshot_Empty(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	require.NoError(t, repo.DeleteOuterParamsCascadeSnapshot("no-such-viz"))
}

func TestRound8B_OuterParams_CreateSnapshotOuterParams(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	p := &auto.SnapshotVisualizationOuterParam{
		ParamsID: "cp-1", VisualizationID: "viz-40", Checked: true,
	}
	require.NoError(t, repo.CreateSnapshotOuterParams(p))
}

func TestRound8B_OuterParams_CreateSnapshotOuterParamsInfo(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	info := &auto.SnapshotVisualizationOuterParamsInfo{
		ParamsInfoID: "cpi-1", ParamsID: "cp-1", ParamName: "param1", Required: true,
	}
	require.NoError(t, repo.CreateSnapshotOuterParamsInfo(info))
}

func TestRound8B_OuterParams_CreateSnapshotOuterParamsTargetViewInfo(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	tv := &auto.SnapshotVisualizationOuterParamsTargetViewInfo{
		TargetID: "ct-1", ParamsInfoID: "cpi-1", TargetViewID: "tv1", TargetFieldID: "tf1",
	}
	require.NoError(t, repo.CreateSnapshotOuterParamsTargetViewInfo(tv))
}

func TestRound8B_OuterParams_GetOuterParamsInfo_Found(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)

	require.NoError(t, db.Create(&auto.VisualizationOuterParam{
		ParamsID: "rp-1", VisualizationID: "viz-50", Checked: true,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationOuterParamsInfo{
		ParamsInfoID: "rpi-1", ParamsID: "rp-1", ParamName: "runtimeP", Checked: true, Required: false, DefaultValue: "dv",
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationOuterParamsTargetViewInfo{
		TargetID: "rt-1", ParamsInfoID: "rpi-1", TargetViewID: "rv1", TargetFieldID: "rf1",
	}).Error)

	rows, err := repo.GetOuterParamsInfo("viz-50")
	require.NoError(t, err)
	assert.Len(t, rows, 1)
}

func TestRound8B_OuterParams_GetOuterParamsInfo_Empty(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	rows, err := repo.GetOuterParamsInfo("nothing")
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_OuterParams_GetDatasetGroupsWithFields_Empty(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)
	rows, err := repo.GetDatasetGroupsWithFields("empty-viz")
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_OuterParams_GetDatasetGroupsWithFields_WithData(t *testing.T) {
	db := round8BOuterParamsDB(t)
	repo := NewOuterParamsRepository(db)

	require.NoError(t, db.Create(&auto.CoreDatasetGroup{ID: 100, Name: "ds1", NodeType: "dataset", Type: "sql", Mode: 0}).Error)
	require.NoError(t, db.Create(&visualization.SnapshotCanvasChartView{
		ID: 200, Title: round8bPtrStr("chart1"), SceneID: round8bPtrInt64(300), TableID: round8bPtrInt64(100), Type: round8bPtrStr("bar"),
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotDataVisualizationInfo{
		ID: "300", ComponentData: `["200"]`,
	}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{
		ID: 400, DatasetGroupID: 100, Name: round8bPtrStr("field1"), DeType: round8bPtrInt(0), OriginName: round8bPtrStr("orig1"),
	}).Error)

	rows, err := repo.GetDatasetGroupsWithFields("300")
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
	assert.Equal(t, int64(100), rows[0].ID)
	assert.Equal(t, "ds1", rows[0].Name)
}

// ==================== LinkJumpRepository (14 functions) ====================

func round8BLinkJumpDB(t *testing.T) *gorm.DB {
	return round8BOpenDB(t,
		&auto.SnapshotVisualizationLinkJump{},
		&auto.SnapshotVisualizationLinkJumpInfo{},
		&auto.SnapshotVisualizationLinkJumpTargetViewInfo{},
		&auto.VisualizationLinkJump{},
		&auto.VisualizationLinkJumpInfo{},
		&auto.VisualizationLinkJumpTargetViewInfo{},
		&visualization.SnapshotCanvasChartView{},
		&visualization.CanvasChartView{},
		&dataset.CoreDatasetTableField{},
		&auto.DataVisualizationInfo{},
		&auto.CoreDatasetGroup{},
		&auto.VisualizationOuterParam{},
		&auto.VisualizationOuterParamsInfo{},
	)
}

func TestRound8B_LinkJump_New(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	require.NotNil(t, repo)
}

func TestRound8B_LinkJump_GetTableFieldWithViewID_Found(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{
		ID: 50, DatasetGroupID: 10, Name: round8bPtrStr("fname"), OriginName: round8bPtrStr("oname"), DeType: round8bPtrInt(2),
	}).Error)
	require.NoError(t, db.Create(&auto.CoreChartView{
		ID: 100, Title: "cv1", SceneID: 200, TableID: 10, Type: "bar",
	}).Error)

	fields, err := repo.GetTableFieldWithViewID(100)
	require.NoError(t, err)
	assert.Len(t, fields, 1)
	assert.Equal(t, int64(50), fields[0].ID)
	assert.Equal(t, "fname", fields[0].Name)
}

func TestRound8B_LinkJump_GetTableFieldWithViewID_Empty(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	fields, err := repo.GetTableFieldWithViewID(999)
	require.NoError(t, err)
	assert.Empty(t, fields)
}

func TestRound8B_LinkJump_QueryWithViewId_Found(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&visualization.SnapshotCanvasChartView{
		ID: 10, Title: round8bPtrStr("v10"), SceneID: round8bPtrInt64(100), Type: round8bPtrStr("bar"),
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJump{
		ID: 1, SourceDvID: 100, SourceViewID: 10, LinkJumpInfo: "info", Checked: true,
	}).Error)

	row, err := repo.QueryWithViewId(100, 10)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.Equal(t, int64(10), row.SourceViewID)
	assert.Equal(t, int64(1), row.ID)
	assert.Equal(t, int64(100), row.SourceDvID)
	assert.True(t, row.Checked)
}

func TestRound8B_LinkJump_QueryWithViewId_NotFound(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&visualization.SnapshotCanvasChartView{
		ID: 20, Title: round8bPtrStr("v20"), SceneID: round8bPtrInt64(200), Type: round8bPtrStr("bar"),
	}).Error)

	row, err := repo.QueryWithViewId(200, 20)
	require.NoError(t, err)
	require.NotNil(t, row)
	assert.False(t, row.Checked)
}

func TestRound8B_LinkJump_GetLinkJumpInfo_Snapshot(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.CoreDatasetGroup{ID: 500, Name: "dsjump", NodeType: "dataset", Type: "sql", Mode: 0}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{
		ID: 600, DatasetGroupID: 500, Name: round8bPtrStr("jfield"), DeType: round8bPtrInt(0), OriginName: round8bPtrStr("jorig"),
	}).Error)
	require.NoError(t, db.Create(&visualization.SnapshotCanvasChartView{
		ID: 700, Title: round8bPtrStr("jchart"), SceneID: round8bPtrInt64(800), TableID: round8bPtrInt64(500), Type: round8bPtrStr("line"),
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJump{
		ID: 10, SourceDvID: 800, SourceViewID: 700, Checked: true,
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJumpInfo{
		ID: 20, LinkJumpID: 10, SourceFieldID: 600, Checked: true, LinkType: "inner",
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJumpTargetViewInfo{
		TargetID: 30, LinkJumpInfoID: 20, TargetViewID: "tv1", TargetFieldID: "tf1",
	}).Error)

	rows, err := repo.GetLinkJumpInfo(10, 700, true)
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
}

func TestRound8B_LinkJump_GetLinkJumpInfo_Core(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.CoreDatasetGroup{ID: 510, Name: "coreDS", NodeType: "dataset", Type: "sql", Mode: 0}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{
		ID: 610, DatasetGroupID: 510, Name: round8bPtrStr("coreField"), DeType: round8bPtrInt(1), OriginName: round8bPtrStr("coreOrig"),
	}).Error)
	require.NoError(t, db.Create(&auto.CoreChartView{
		ID: 710, Title: "coreChart", SceneID: 810, TableID: 510, Type: "pie",
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJump{
		ID: 11, SourceDvID: 810, SourceViewID: 710, Checked: true,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJumpInfo{
		ID: 21, LinkJumpID: 11, SourceFieldID: 610, Checked: true, LinkType: "inner",
	}).Error)

	rows, err := repo.GetLinkJumpInfo(11, 710, false)
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
}

func TestRound8B_LinkJump_GetLinkJumpInfo_Empty(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	rows, err := repo.GetLinkJumpInfo(999, 999, true)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_LinkJump_QueryWithDvId_Snapshot(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&visualization.SnapshotCanvasChartView{
		ID: 30, Title: round8bPtrStr("qd30"), SceneID: round8bPtrInt64(300),
		JumpActive: round8bPtrBool(true), Type: round8bPtrStr("bar"),
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJump{
		ID: 2, SourceDvID: 300, SourceViewID: 30, Checked: true,
	}).Error)

	rows, err := repo.QueryWithDvId(300, true)
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, int64(30), rows[0].SourceViewID)
}

func TestRound8B_LinkJump_QueryWithDvId_Core(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.CoreChartView{
		ID: 31, Title: "core31", SceneID: 301, Type: "bar", JumpActive: true,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJump{
		ID: 3, SourceDvID: 301, SourceViewID: 31, Checked: true,
	}).Error)

	rows, err := repo.QueryWithDvId(301, false)
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, int64(31), rows[0].SourceViewID)
}

func TestRound8B_LinkJump_QueryWithDvId_Empty(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	rows, err := repo.QueryWithDvId(999, true)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_LinkJump_DeleteJumpCascadeSnapshot(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJump{
		ID: 50, SourceDvID: 400, SourceViewID: 450,
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJumpInfo{
		ID: 60, LinkJumpID: 50, SourceFieldID: 100,
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJumpTargetViewInfo{
		TargetID: 70, LinkJumpInfoID: 60, TargetViewID: "tvx",
	}).Error)

	require.NoError(t, repo.DeleteJumpCascadeSnapshot(400, 450))

	rows, err := repo.QueryWithDvId(400, true)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_LinkJump_DeleteJumpCascadeSnapshot_Empty(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	require.NoError(t, repo.DeleteJumpCascadeSnapshot(999, 999))
}

func TestRound8B_LinkJump_CreateSnapshotJump(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	j := &auto.SnapshotVisualizationLinkJump{
		SourceDvID: 500, SourceViewID: 501, LinkJumpInfo: "jump-info", Checked: true,
	}
	require.NoError(t, repo.CreateSnapshotJump(j))
	assert.Positive(t, j.ID)
}

func TestRound8B_LinkJump_CreateSnapshotJumpInfo(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	info := &auto.SnapshotVisualizationLinkJumpInfo{
		LinkJumpID: 1, SourceFieldID: 100, LinkType: "inner", JumpType: "_blank", Checked: true,
	}
	require.NoError(t, repo.CreateSnapshotJumpInfo(info))
	assert.Positive(t, info.ID)
}

func TestRound8B_LinkJump_CreateSnapshotJumpTargetViewInfo(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	tv := &auto.SnapshotVisualizationLinkJumpTargetViewInfo{
		LinkJumpInfoID: 1, TargetViewID: "tv1", TargetFieldID: "tf1", SourceFieldActiveID: 100,
	}
	require.NoError(t, repo.CreateSnapshotJumpTargetViewInfo(tv))
	assert.Positive(t, tv.TargetID)
}

func TestRound8B_LinkJump_GetTargetVisualizationJumpInfo_Snapshot(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJump{
		ID: 80, SourceDvID: 600, SourceViewID: 601,
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJumpInfo{
		ID: 81, LinkJumpID: 80, SourceFieldID: 610, TargetDvID: 602, Checked: true,
	}).Error)
	require.NoError(t, db.Create(&auto.SnapshotVisualizationLinkJumpTargetViewInfo{
		TargetID: 82, LinkJumpInfoID: 81, SourceFieldActiveID: 610, TargetViewID: "603", TargetFieldID: "604",
	}).Error)

	rows, err := repo.GetTargetVisualizationJumpInfo(600, 601, 602, nil, true)
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
}

func TestRound8B_LinkJump_GetTargetVisualizationJumpInfo_Core(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.VisualizationLinkJump{
		ID: 83, SourceDvID: 620, SourceViewID: 621,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJumpInfo{
		ID: 84, LinkJumpID: 83, SourceFieldID: 630, TargetDvID: 622, Checked: true,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJumpTargetViewInfo{
		TargetID: 85, LinkJumpInfoID: 84, SourceFieldActiveID: 630, TargetViewID: "623", TargetFieldID: "624",
	}).Error)

	rows, err := repo.GetTargetVisualizationJumpInfo(620, 621, 622, nil, false)
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
}

func TestRound8B_LinkJump_GetTargetVisualizationJumpInfo_WithFieldFilter(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.VisualizationLinkJump{
		ID: 90, SourceDvID: 700, SourceViewID: 701,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJumpInfo{
		ID: 91, LinkJumpID: 90, SourceFieldID: 710, TargetDvID: 702, Checked: true,
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJumpTargetViewInfo{
		TargetID: 92, LinkJumpInfoID: 91, SourceFieldActiveID: 710, TargetViewID: "703", TargetFieldID: "704",
	}).Error)

	fieldID := int64(710)
	rows, err := repo.GetTargetVisualizationJumpInfo(700, 701, 702, &fieldID, false)
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
}

func TestRound8B_LinkJump_GetTargetVisualizationJumpInfo_Empty(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	rows, err := repo.GetTargetVisualizationJumpInfo(999, 999, 999, nil, true)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_LinkJump_GetViewTableDetails_Empty(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	rows, err := repo.GetViewTableDetails(999)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_LinkJump_GetViewTableDetails_WithData(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.CoreDatasetGroup{ID: 800, Name: "vds", NodeType: "dataset", Type: "sql", Mode: 0}).Error)
	require.NoError(t, db.Create(&auto.CoreChartView{
		ID: 801, Title: "vchart", SceneID: 802, TableID: 800, Type: "bar",
	}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{
		ID: 803, DatasetGroupID: 800, Name: round8bPtrStr("vf"), DeType: round8bPtrInt(0), OriginName: round8bPtrStr("vo"),
		Type: round8bPtrStr("varchar"),
	}).Error)
	require.NoError(t, db.Create(&auto.DataVisualizationInfo{
		ID: "802", ComponentData: `["801"]`,
	}).Error)

	rows, err := repo.GetViewTableDetails(802)
	require.NoError(t, err)
	assert.NotEmpty(t, rows)
	assert.Equal(t, int64(801), rows[0].ID)
}

func TestRound8B_LinkJump_GetOutParamsTargetWithDvID_Found(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.VisualizationOuterParam{
		ParamsID: "op-1", VisualizationID: "900",
	}).Error)
	require.NoError(t, db.Create(&auto.VisualizationOuterParamsInfo{
		ParamsInfoID: "opi-1", ParamsID: "op-1", ParamName: "outerP1",
	}).Error)

	rows, err := repo.GetOutParamsTargetWithDvID(900)
	require.NoError(t, err)
	assert.Len(t, rows, 1)
	assert.Equal(t, "outerP1", rows[0].Name)
	assert.Equal(t, "outerParams", rows[0].Type)
}

func TestRound8B_LinkJump_GetOutParamsTargetWithDvID_Empty(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	rows, err := repo.GetOutParamsTargetWithDvID(999)
	require.NoError(t, err)
	assert.Empty(t, rows)
}

func TestRound8B_LinkJump_GetComponentData_Found(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&auto.DataVisualizationInfo{
		ID: "910", ComponentData: `{"components":[]}`,
	}).Error)

	data, err := repo.GetComponentData(910)
	require.NoError(t, err)
	assert.Equal(t, `{"components":[]}`, data)
}

func TestRound8B_LinkJump_GetComponentData_NotFound(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)
	data, err := repo.GetComponentData(99999)
	require.NoError(t, err)
	assert.Empty(t, data)
}

func TestRound8B_LinkJump_UpdateChartJumpActive(t *testing.T) {
	db := round8BLinkJumpDB(t)
	repo := NewLinkJumpRepository(db)

	require.NoError(t, db.Create(&visualization.SnapshotCanvasChartView{
		ID: 950, Title: round8bPtrStr("activeChart"), Type: round8bPtrStr("bar"),
		JumpActive: round8bPtrBool(false),
	}).Error)

	require.NoError(t, repo.UpdateChartJumpActive(950, true))

	var view visualization.SnapshotCanvasChartView
	require.NoError(t, db.First(&view, 950).Error)
	assert.True(t, *view.JumpActive)
}

// ==================== ExportRepository (9 public + 3 private = 12 functions) ====================

func round8BExportDB(t *testing.T) *gorm.DB {
	return round8BOpenDB(t, &coreExportTask{})
}

func TestRound8B_Export_New(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)
	require.NotNil(t, repo)
}

func TestRound8B_Export_Create(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	task := &export.ExportTask{
		ID: "exp-1", UserID: 1, FileName: "test.xlsx", FileSize: 1024.5, FileSizeUnit: "KB",
		ExportFrom: 100, ExportStatus: "pending", ExportFromType: "chart",
		ExportTime: time.Now().UnixMilli(), ExportProgress: "0", ExportMachineName: "local",
	}
	require.NoError(t, repo.Create(task))
}

func TestRound8B_Export_GetByID_Success(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{
		ID: "exp-2", UserID: 2, FileName: "get.xlsx", ExportFrom: 200,
		ExportStatus: "completed", ExportFromType: "view", ExportTime: 1000,
	}))

	got, err := repo.GetByID("exp-2")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "exp-2", got.ID)
	assert.Equal(t, "get.xlsx", got.FileName)
	assert.Equal(t, int64(200), got.ExportFrom)
	assert.Equal(t, "completed", got.ExportStatus)
}

func TestRound8B_Export_GetByID_NotFound(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)
	_, err := repo.GetByID("no-such-id")
	assert.Error(t, err)
}

func TestRound8B_Export_List_All(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(&export.ExportTask{
			ID: "list-" + string(rune('a'+i)), UserID: 1, FileName: "f.xlsx",
			ExportStatus: "completed", ExportTime: int64(i), ExportFromType: "chart",
		}))
	}

	tasks, total, err := repo.List(1, 10, "all")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, tasks, 5)
}

func TestRound8B_Export_List_ByStatus(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{
		ID: "s1", UserID: 1, FileName: "f1", ExportStatus: "completed", ExportTime: 1, ExportFromType: "chart",
	}))
	require.NoError(t, repo.Create(&export.ExportTask{
		ID: "s2", UserID: 1, FileName: "f2", ExportStatus: "pending", ExportTime: 2, ExportFromType: "chart",
	}))
	require.NoError(t, repo.Create(&export.ExportTask{
		ID: "s3", UserID: 1, FileName: "f3", ExportStatus: "completed", ExportTime: 3, ExportFromType: "chart",
	}))

	tasks, total, err := repo.List(1, 10, "completed")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, tasks, 2)
}

func TestRound8B_Export_List_Empty(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)
	tasks, total, err := repo.List(1, 10, "")
	require.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Empty(t, tasks)
}

func TestRound8B_Export_List_Pagination(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(&export.ExportTask{
			ID: "page-" + string(rune('a'+i)), UserID: 1, FileName: "f.xlsx",
			ExportStatus: "completed", ExportTime: int64(5 - i), ExportFromType: "chart",
		}))
	}

	tasks, total, err := repo.List(2, 2, "all")
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, tasks, 2)
}

func TestRound8B_Export_UpdateStatus(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{
		ID: "upd-1", UserID: 1, FileName: "upd.xlsx", ExportStatus: "pending", ExportTime: 1, ExportFromType: "chart",
	}))

	require.NoError(t, repo.UpdateStatus("upd-1", "completed"))

	got, err := repo.GetByID("upd-1")
	require.NoError(t, err)
	assert.Equal(t, "completed", got.ExportStatus)
}

func TestRound8B_Export_UpdateStatus_NotFound(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)
	err := repo.UpdateStatus("ghost", "completed")
	assert.NoError(t, err)
}

func TestRound8B_Export_Delete(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{
		ID: "del-1", UserID: 1, FileName: "del.xlsx", ExportStatus: "pending", ExportTime: 1, ExportFromType: "chart",
	}))

	require.NoError(t, repo.Delete("del-1"))
	_, err := repo.GetByID("del-1")
	assert.Error(t, err)
}

func TestRound8B_Export_Delete_NotFound(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)
	err := repo.Delete("no-such")
	assert.NoError(t, err)
}

func TestRound8B_Export_DeleteBatch(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{ID: "b1", UserID: 1, ExportTime: 1, ExportFromType: "chart"}))
	require.NoError(t, repo.Create(&export.ExportTask{ID: "b2", UserID: 1, ExportTime: 2, ExportFromType: "chart"}))
	require.NoError(t, repo.Create(&export.ExportTask{ID: "b3", UserID: 1, ExportTime: 3, ExportFromType: "chart"}))

	require.NoError(t, repo.DeleteBatch([]string{"b1", "b3"}))
	_, err := repo.GetByID("b1")
	assert.Error(t, err)
	got, err := repo.GetByID("b2")
	require.NoError(t, err)
	assert.Equal(t, "b2", got.ID)
}

func TestRound8B_Export_DeleteBatch_Empty(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)
	require.NoError(t, repo.DeleteBatch([]string{}))
	require.NoError(t, repo.DeleteBatch(nil))
}

func TestRound8B_Export_DeleteAllByType_All(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{ID: "da1", UserID: 1, ExportTime: 1, ExportFromType: "chart"}))
	require.NoError(t, repo.Create(&export.ExportTask{ID: "da2", UserID: 1, ExportTime: 2, ExportFromType: "view"}))

	err := repo.DeleteAllByType("all")
	assert.Error(t, err)
}

func TestRound8B_Export_DeleteAllByType_SpecificType(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{ID: "dt1", UserID: 1, ExportTime: 1, ExportFromType: "chart"}))
	require.NoError(t, repo.Create(&export.ExportTask{ID: "dt2", UserID: 1, ExportTime: 2, ExportFromType: "view"}))

	require.NoError(t, repo.DeleteAllByType("chart"))
	tasks, total, err := repo.List(1, 10, "all")
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, tasks, 1)
	assert.Equal(t, "dt2", tasks[0].ID)
}

func TestRound8B_Export_CountByStatus(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)

	require.NoError(t, repo.Create(&export.ExportTask{ID: "cs1", UserID: 1, ExportStatus: "completed", ExportTime: 1, ExportFromType: "chart"}))
	require.NoError(t, repo.Create(&export.ExportTask{ID: "cs2", UserID: 1, ExportStatus: "completed", ExportTime: 2, ExportFromType: "chart"}))
	require.NoError(t, repo.Create(&export.ExportTask{ID: "cs3", UserID: 1, ExportStatus: "pending", ExportTime: 3, ExportFromType: "chart"}))

	counts, err := repo.CountByStatus()
	require.NoError(t, err)
	assert.Equal(t, int64(2), counts["completed"])
	assert.Equal(t, int64(1), counts["pending"])
}

func TestRound8B_Export_CountByStatus_Empty(t *testing.T) {
	db := round8BExportDB(t)
	repo := NewExportRepository(db)
	counts, err := repo.CountByStatus()
	require.NoError(t, err)
	assert.Empty(t, counts)
}

// ==================== DataFillingRepository (10 public + resolveLevel = 11 functions) ====================

func round8BDataFillingDB(t *testing.T) *gorm.DB {
	return round8BOpenDB(t, &datafillingdomain.DataFillingForm{})
}

func TestRound8B_DataFilling_New(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	require.NotNil(t, repo)
}

func TestRound8B_DataFilling_Create(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	form := &datafillingdomain.DataFillingForm{
		Name: "Form1", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder,
		CreateTime: time.Now().UnixMilli(),
	}
	require.NoError(t, repo.Create(ctx, form))
	assert.Positive(t, form.ID)
}

func TestRound8B_DataFilling_GetByID_Success(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	form := &datafillingdomain.DataFillingForm{
		Name: "GetMe", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder,
		CreateTime: 1000,
	}
	require.NoError(t, repo.Create(ctx, form))

	got, err := repo.GetByID(ctx, form.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "GetMe", got.Name)
	assert.Equal(t, datafillingdomain.NodeTypeFolder, got.NodeType)
}

func TestRound8B_DataFilling_GetByID_NotFound(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 999)
	assert.Error(t, err)
}

func TestRound8B_DataFilling_Update(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	form := &datafillingdomain.DataFillingForm{
		Name: "OldName", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder,
		CreateTime: 1000,
	}
	require.NoError(t, repo.Create(ctx, form))

	form.Name = "NewName"
	require.NoError(t, repo.Update(ctx, form))

	got, err := repo.GetByID(ctx, form.ID)
	require.NoError(t, err)
	assert.Equal(t, "NewName", got.Name)
}

func TestRound8B_DataFilling_DeleteByID(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	form := &datafillingdomain.DataFillingForm{
		Name: "DeleteMe", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeForm,
		CreateTime: 1000,
	}
	require.NoError(t, repo.Create(ctx, form))
	require.NoError(t, repo.DeleteByID(ctx, form.ID))

	_, err := repo.GetByID(ctx, form.ID)
	assert.Error(t, err)
}

func TestRound8B_DataFilling_DeleteByID_NotFound(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()
	err := repo.DeleteByID(ctx, 999)
	assert.NoError(t, err)
}

func TestRound8B_DataFilling_Rename(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	form := &datafillingdomain.DataFillingForm{
		Name: "Before", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder,
		CreateTime: 1000,
	}
	require.NoError(t, repo.Create(ctx, form))

	require.NoError(t, repo.Rename(ctx, form.ID, "After"))

	got, err := repo.GetByID(ctx, form.ID)
	require.NoError(t, err)
	assert.Equal(t, "After", got.Name)
}

func TestRound8B_DataFilling_Move_ToRoot(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	form := &datafillingdomain.DataFillingForm{
		Name: "MoveMe", PID: 10, Level: 2, NodeType: datafillingdomain.NodeTypeForm,
		CreateTime: 1000,
	}
	require.NoError(t, repo.Create(ctx, form))

	require.NoError(t, repo.Move(ctx, form.ID, 0))

	got, err := repo.GetByID(ctx, form.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), got.PID)
	assert.Equal(t, 0, got.Level)
}

func TestRound8B_DataFilling_Move_ToParent(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	parent := &datafillingdomain.DataFillingForm{
		Name: "ParentFolder", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder,
		CreateTime: 1000,
	}
	require.NoError(t, repo.Create(ctx, parent))

	child := &datafillingdomain.DataFillingForm{
		Name: "ChildForm", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeForm,
		CreateTime: 1000,
	}
	require.NoError(t, repo.Create(ctx, child))

	require.NoError(t, repo.Move(ctx, child.ID, parent.ID))

	got, err := repo.GetByID(ctx, child.ID)
	require.NoError(t, err)
	assert.Equal(t, parent.ID, got.PID)
	assert.Equal(t, 1, got.Level)
}

func TestRound8B_DataFilling_Move_NestedLevel(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	root := &datafillingdomain.DataFillingForm{
		Name: "Root", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 1,
	}
	require.NoError(t, repo.Create(ctx, root))

	sub := &datafillingdomain.DataFillingForm{
		Name: "Sub", PID: root.ID, Level: 1, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 2,
	}
	require.NoError(t, repo.Create(ctx, sub))

	leaf := &datafillingdomain.DataFillingForm{
		Name: "Leaf", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeForm, CreateTime: 3,
	}
	require.NoError(t, repo.Create(ctx, leaf))

	require.NoError(t, repo.Move(ctx, leaf.ID, sub.ID))

	got, err := repo.GetByID(ctx, leaf.ID)
	require.NoError(t, err)
	assert.Equal(t, sub.ID, got.PID)
	assert.Equal(t, 2, got.Level)
}

func TestRound8B_DataFilling_Move_NotFoundParent(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	form := &datafillingdomain.DataFillingForm{
		Name: "Orphan", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeForm, CreateTime: 1,
	}
	require.NoError(t, repo.Create(ctx, form))

	err := repo.Move(ctx, form.ID, 99999)
	assert.Error(t, err)
}

func TestRound8B_DataFilling_GetTree(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "Root1", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 1,
	}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "Child1", PID: 1, Level: 1, NodeType: datafillingdomain.NodeTypeForm, CreateTime: 2,
	}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "Root2", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 3,
	}))

	tree, err := repo.GetTree(ctx)
	require.NoError(t, err)
	require.Len(t, tree, 3)
	assert.Equal(t, "Root1", tree[0].Name)
	assert.Equal(t, "Root2", tree[1].Name)
	assert.Equal(t, "Child1", tree[2].Name)
}

func TestRound8B_DataFilling_GetTree_Empty(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	tree, err := repo.GetTree(ctx)
	require.NoError(t, err)
	assert.Empty(t, tree)
}

func TestRound8B_DataFilling_GetTree_FolderFirst(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "FormA", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeForm, CreateTime: 1,
	}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "FolderB", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 2,
	}))

	tree, err := repo.GetTree(ctx)
	require.NoError(t, err)
	require.Len(t, tree, 2)
	assert.Equal(t, "FolderB", tree[0].Name)
	assert.Equal(t, "FormA", tree[1].Name)
}

func TestRound8B_DataFilling_GetByPID(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "Parent", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 1,
	}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "Child1", PID: 1, Level: 1, NodeType: datafillingdomain.NodeTypeForm, CreateTime: 2,
	}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "Child2", PID: 1, Level: 1, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 3,
	}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "Other", PID: 99, Level: 1, NodeType: datafillingdomain.NodeTypeForm, CreateTime: 4,
	}))

	children, err := repo.GetByPID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, children, 2)
}

func TestRound8B_DataFilling_GetByPID_Empty(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	children, err := repo.GetByPID(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, children)
}

func TestRound8B_DataFilling_GetChildren(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "FolderC", PID: 0, Level: 0, NodeType: datafillingdomain.NodeTypeFolder, CreateTime: 1,
	}))
	require.NoError(t, repo.Create(ctx, &datafillingdomain.DataFillingForm{
		Name: "FormC1", PID: 1, Level: 1, NodeType: datafillingdomain.NodeTypeForm, CreateTime: 2,
	}))

	children, err := repo.GetChildren(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, children, 1)
	assert.Equal(t, "FormC1", children[0].Name)
}

func TestRound8B_DataFilling_GetChildren_Empty(t *testing.T) {
	db := round8BDataFillingDB(t)
	repo := NewDataFillingRepository(db)
	ctx := context.Background()

	children, err := repo.GetChildren(ctx, 999)
	require.NoError(t, err)
	assert.Empty(t, children)
}
