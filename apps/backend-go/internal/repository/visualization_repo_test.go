package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/datasource"
	"dataease/backend/internal/domain/visualization"

	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const visualizationSQLiteDriver = "sqlite3_visualization_repo_test"

var registerVisualizationSQLiteDriverOnce sync.Once

type visualizationCoreOptRecent struct {
	UID        int64 `gorm:"column:uid"`
	ResourceID int64 `gorm:"column:resource_id"`
	Time       int64 `gorm:"column:time"`
}

func (visualizationCoreOptRecent) TableName() string { return "core_opt_recent" }

type visualizationCoreStore struct {
	ResourceID int64 `gorm:"column:resource_id"`
	UID        int64 `gorm:"column:uid"`
}

func (visualizationCoreStore) TableName() string { return "core_store" }

type visualizationCoreDatasourceTask struct {
	ID   int64 `gorm:"column:id;primaryKey"`
	DsID int64 `gorm:"column:ds_id"`
	Name string `gorm:"column:name"`
}

func (visualizationCoreDatasourceTask) TableName() string { return "core_datasource_task" }

type visualizationLinkage struct {
	ID            int64   `gorm:"column:id;primaryKey"`
	DvID          int64   `gorm:"column:dv_id"`
	SourceViewID  int64   `gorm:"column:source_view_id"`
	TargetViewID  int64   `gorm:"column:target_view_id"`
	UpdateTime    *int64  `gorm:"column:update_time"`
	UpdatePeople  *string `gorm:"column:update_people"`
	LinkageActive *bool   `gorm:"column:linkage_active"`
	Ext1          *string `gorm:"column:ext1"`
	Ext2          *string `gorm:"column:ext2"`
	CopyFrom      *int64  `gorm:"column:copy_from"`
	CopyID        *int64  `gorm:"column:copy_id"`
}

func (visualizationLinkage) TableName() string { return "visualization_linkage" }

type visualizationLinkageField struct {
	ID          int64   `gorm:"column:id;primaryKey"`
	LinkageID   int64   `gorm:"column:linkage_id"`
	SourceField *string `gorm:"column:source_field"`
	TargetField *string `gorm:"column:target_field"`
	UpdateTime  *int64  `gorm:"column:update_time"`
	CopyFrom    *int64  `gorm:"column:copy_from"`
	CopyID      *int64  `gorm:"column:copy_id"`
}

func (visualizationLinkageField) TableName() string { return "visualization_linkage_field" }

type visualizationLinkJump struct {
	ID           int64   `gorm:"column:id;primaryKey"`
	SourceDvID   int64   `gorm:"column:source_dv_id"`
	SourceViewID int64   `gorm:"column:source_view_id"`
	LinkJumpInfo *string `gorm:"column:link_jump_info"`
	Checked      *bool   `gorm:"column:checked"`
	CopyFrom     *int64  `gorm:"column:copy_from"`
	CopyID       *int64  `gorm:"column:copy_id"`
}

func (visualizationLinkJump) TableName() string { return "visualization_link_jump" }

type visualizationLinkJumpInfo struct {
	ID            int64   `gorm:"column:id;primaryKey"`
	LinkJumpID    int64   `gorm:"column:link_jump_id"`
	LinkType      *string `gorm:"column:link_type"`
	JumpType      *string `gorm:"column:jump_type"`
	TargetDvID    *int64  `gorm:"column:target_dv_id"`
	SourceFieldID *int64  `gorm:"column:source_field_id"`
	Content       *string `gorm:"column:content"`
	Checked       *bool   `gorm:"column:checked"`
	AttachParams  *string `gorm:"column:attach_params"`
	CopyFrom      *int64  `gorm:"column:copy_from"`
	CopyID        *int64  `gorm:"column:copy_id"`
}

func (visualizationLinkJumpInfo) TableName() string { return "visualization_link_jump_info" }

type visualizationLinkJumpTargetViewInfo struct {
	TargetID            int64  `gorm:"column:target_id;primaryKey"`
	LinkJumpInfoID      int64  `gorm:"column:link_jump_info_id"`
	SourceFieldActiveID *int64 `gorm:"column:source_field_active_id"`
	TargetViewID        *int64 `gorm:"column:target_view_id"`
	TargetFieldID       *int64 `gorm:"column:target_field_id"`
	CopyFrom            *int64 `gorm:"column:copy_from"`
	CopyID              *int64 `gorm:"column:copy_id"`
}

func (visualizationLinkJumpTargetViewInfo) TableName() string {
	return "visualization_link_jump_target_view_info"
}

func setupVisualizationRepositoryTest(t *testing.T) (*VisualizationRepository, *gorm.DB) {
	t.Helper()

	registerVisualizationSQLiteDriverOnce.Do(func() {
		sql.Register(visualizationSQLiteDriver, &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				return conn.RegisterFunc("concat", func(args ...interface{}) string {
					var builder strings.Builder
					for _, arg := range args {
						builder.WriteString(fmt.Sprint(arg))
					}
					return builder.String()
				}, true)
			},
		})
	})

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_").Replace(t.Name()))
	db, err := gorm.Open(sqlite.Dialector{DriverName: visualizationSQLiteDriver, DSN: dsn}, &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&visualization.DataVisualizationInfo{},
		&chart.CoreChartView{},
		&visualization.SnapshotCanvasChartView{},
		&dataset.CoreDatasetGroup{},
		&dataset.CoreDatasetTable{},
		&dataset.CoreDatasetTableField{},
		&datasource.CoreDatasource{},
		&visualizationCoreOptRecent{},
		&visualizationCoreStore{},
		&visualizationCoreDatasourceTask{},
		&visualizationLinkage{},
		&visualizationLinkageField{},
		&visualizationLinkJump{},
		&visualizationLinkJumpInfo{},
		&visualizationLinkJumpTargetViewInfo{},
	))

	return NewVisualizationRepository(db), db
}

func strPtrVisualizationRepo(v string) *string { return &v }
func intPtrVisualizationRepo(v int) *int       { return &v }
func int64PtrVisualizationRepo(v int64) *int64 { return &v }
func boolPtrVisualizationRepo(v bool) *bool    { return &v }

func newVisualizationInfo(id int64, name, dvType string, updateTime int64) *visualization.DataVisualizationInfo {
	nodeType := visualization.NodeTypeLeaf
	status := 1
	return &visualization.DataVisualizationInfo{
		ID:         id,
		Name:       name,
		NodeType:   &nodeType,
		Type:       &dvType,
		Status:     &status,
		UpdateTime: &updateTime,
	}
}

func newChartView(id, sceneID int64, title string) *chart.CoreChartView {
	typeName := "bar"
	render := "antv"
	chartType := "histogram"
	dataFrom := "dataset"
	refreshUnit := "minute"
	resultMode := "custom"
	resultCount := 10
	refreshTime := 5
	linkageActive := true
	jumpActive := true
	isPlugin := false
	return &chart.CoreChartView{
		ID:                id,
		Title:             &title,
		SceneID:           &sceneID,
		TableID:           int64PtrVisualizationRepo(300 + id),
		Type:              &typeName,
		Render:            &render,
		ResultCount:       &resultCount,
		ResultMode:        &resultMode,
		CustomAttr:        strPtrVisualizationRepo("{}"),
		CustomStyle:       strPtrVisualizationRepo("{}"),
		CustomFilter:      strPtrVisualizationRepo("[]"),
		ChartType:         &chartType,
		DataFrom:          &dataFrom,
		RefreshViewEnable: &linkageActive,
		RefreshUnit:       &refreshUnit,
		RefreshTime:       &refreshTime,
		LinkageActive:     &linkageActive,
		JumpActive:        &jumpActive,
		IsPlugin:          &isPlugin,
		FlowMapStartName:  strPtrVisualizationRepo("start"),
		FlowMapEndName:    strPtrVisualizationRepo("end"),
		ExtColor:          strPtrVisualizationRepo("blue"),
	}
}

func newSnapshotChartView(id, sceneID int64, title string) *visualization.SnapshotCanvasChartView {
	typeName := "bar"
	render := "antv"
	chartType := "histogram"
	dataFrom := "dataset"
	refreshUnit := "minute"
	resultMode := "custom"
	resultCount := 10
	refreshTime := 5
	linkageActive := true
	jumpActive := true
	isPlugin := false
	return &visualization.SnapshotCanvasChartView{
		ID:                id,
		Title:             &title,
		SceneID:           &sceneID,
		TableID:           int64PtrVisualizationRepo(300 + id),
		Type:              &typeName,
		Render:            &render,
		ResultCount:       &resultCount,
		ResultMode:        &resultMode,
		CustomAttr:        strPtrVisualizationRepo("{}"),
		CustomStyle:       strPtrVisualizationRepo("{}"),
		CustomFilter:      strPtrVisualizationRepo("[]"),
		ChartType:         &chartType,
		DataFrom:          &dataFrom,
		RefreshViewEnable: &linkageActive,
		RefreshUnit:       &refreshUnit,
		RefreshTime:       &refreshTime,
		LinkageActive:     &linkageActive,
		JumpActive:        &jumpActive,
		IsPlugin:          &isPlugin,
		FlowMapStartName:  strPtrVisualizationRepo("start"),
		FlowMapEndName:    strPtrVisualizationRepo("end"),
		ExtColor:          strPtrVisualizationRepo("blue"),
	}
}

func mapInt64Value(t *testing.T, row map[string]interface{}, key string) int64 {
	t.Helper()
	value, ok := row[key]
	require.True(t, ok, "missing key %s", key)
	switch v := value.(type) {
	case int64:
		return v
	case int:
		return int64(v)
	case int32:
		return int64(v)
	case float64:
		return int64(v)
	case []byte:
		var parsed int64
		_, err := fmt.Sscan(string(v), &parsed)
		require.NoError(t, err)
		return parsed
	default:
		t.Fatalf("unexpected type for %s: %T", key, value)
		return 0
	}
}

func TestVisualizationRepository_CRUDAndQueries(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)
	require.NotNil(t, repo)
	require.Same(t, db, repo.db)

	now := time.Now().UnixMilli()
	dashboardType := visualization.TypeDashboard
	panelType := "panel"

	item := newVisualizationInfo(0, "Main Dashboard", dashboardType, now)
	require.NoError(t, repo.Create(item))
	require.Positive(t, item.ID)

	found, err := repo.GetByID(item.ID)
	require.NoError(t, err)
	assert.Equal(t, "Main Dashboard", found.Name)

	item.Name = "Updated Dashboard"
	updatedAt := now + 100
	item.UpdateTime = &updatedAt
	require.NoError(t, repo.Update(item))

	found, err = repo.GetByID(item.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Dashboard", found.Name)

	snapshot := newSnapshotChartView(11, item.ID, "snapshot-view")
	require.NoError(t, repo.SaveSnapshotChartView(snapshot))

	views, err := repo.GetSnapshotChartViewsBySceneID(item.ID)
	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.Equal(t, int64(11), views[0].ID)

	older := newVisualizationInfo(101, "Alpha Dashboard", dashboardType, now-1000)
	middle := newVisualizationInfo(102, "Beta Panel", panelType, now-500)
	deleted := newVisualizationInfo(103, "Deleted Dashboard", dashboardType, now+1000)
	deleted.DeleteFlag = boolPtrVisualizationRepo(true)
	require.NoError(t, db.Create([]*visualization.DataVisualizationInfo{older, middle, deleted}).Error)

	queryList, total, err := repo.Query(&visualization.ListRequest{Current: 0, Size: 200})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, queryList, 3)
	assert.Equal(t, int64(102), queryList[1].ID)

	keyword := "alpha"
	typeFilter := dashboardType
	queryList, total, err = repo.Query(&visualization.ListRequest{Keyword: &keyword, Type: &typeFilter, Current: 1, Size: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, queryList, 1)
	assert.Equal(t, int64(101), queryList[0].ID)

	allByTypes, err := repo.ListAllByTypes([]string{dashboardType})
	require.NoError(t, err)
	require.Len(t, allByTypes, 2)

	orgID := int64(77)
	older.OrgID = &orgID
	middle.OrgID = &orgID
	require.NoError(t, db.Save(older).Error)
	require.NoError(t, db.Save(middle).Error)

	batchList, err := repo.ListAllByTypesBatch([]string{dashboardType, panelType}, 100, 1, &orgID)
	require.NoError(t, err)
	require.Len(t, batchList, 1)
	assert.Equal(t, int64(101), batchList[0].ID)

	count, err := repo.CountByNameAndPID("Alpha Dashboard", nil, nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	rootPID := int64(8)
	child := newVisualizationInfo(104, "Folder Child", dashboardType, now+50)
	child.PID = &rootPID
	require.NoError(t, db.Create(child).Error)

	count, err = repo.CountByNameAndPID("Folder Child", &rootPID, &child.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	require.NoError(t, repo.DeleteLogic(item.ID, "deleter"))
	_, err = repo.GetByID(item.ID)
	require.Error(t, err)
}

func TestVisualizationRepository_FindRecentAndRelations(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)

	datasetNodeType := dataset.NodeTypeDataset
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 201, Name: "Sales Dataset", NodeType: &datasetNodeType, CreateBy: "creatorA"}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 202, Name: "People Dataset", NodeType: &datasetNodeType, CreateBy: "creatorB"}).Error)

	datasourceStatus := "ready"
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 301, Name: "Orders Source", Type: "mysql", Status: &datasourceStatus, CreateBy: strPtrVisualizationRepo("creatorC")}).Error)
	require.NoError(t, db.Create(&datasource.CoreDatasource{ID: 302, Name: "Folder Source", Type: datasource.TypeFolder, CreateBy: strPtrVisualizationRepo("creatorD")}).Error)

	mobileLayout := true
	status := 1
	leaf := visualization.NodeTypeLeaf
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{
		ID:           401,
		Name:         "North Screen",
		NodeType:     &leaf,
		Type:         strPtrVisualizationRepo(visualization.TypeDataV),
		MobileLayout: &mobileLayout,
		Status:       &status,
		CreateBy:     strPtrVisualizationRepo("creatorE"),
	}).Error)
	require.NoError(t, db.Create(&visualization.DataVisualizationInfo{
		ID:       402,
		Name:     "South Panel",
		NodeType: &leaf,
		Type:     strPtrVisualizationRepo(visualization.TypeDashboard),
		Status:   &status,
		CreateBy: strPtrVisualizationRepo("creatorF"),
	}).Error)

	recentRows := []visualizationCoreOptRecent{
		{UID: 9001, ResourceID: 201, Time: 10},
		{UID: 9001, ResourceID: 301, Time: 20},
		{UID: 9001, ResourceID: 401, Time: 30},
		{UID: 9001, ResourceID: 402, Time: 40},
	}
	require.NoError(t, db.Create(&recentRows).Error)
	require.NoError(t, db.Create(&visualizationCoreStore{ResourceID: 401, UID: 9001}).Error)

	recent, err := repo.FindRecent(9001, nil)
	require.NoError(t, err)
	require.Len(t, recent, 4)
	assert.Equal(t, "South Panel", recent[0].Name)
	assert.Equal(t, "screen", recent[1].Type)
	assert.True(t, recent[1].Favorite)

	recent, err = repo.FindRecent(9001, &visualization.WorkbranchQueryRequest{Type: visualization.ResourceAliasDataset, Keyword: "Sales", Asc: true})
	require.NoError(t, err)
	require.Len(t, recent, 1)
	assert.Equal(t, "Sales Dataset", recent[0].Name)
	assert.Equal(t, "dataset", recent[0].Type)

	recent, err = repo.FindRecent(9001, &visualization.WorkbranchQueryRequest{Type: visualization.ResourceAliasDatasource})
	require.NoError(t, err)
	require.Len(t, recent, 1)
	assert.Equal(t, "Orders Source", recent[0].Name)

	fieldOrigin := "amount"
	fieldName := "Amount"
	require.NoError(t, db.Create(&dataset.CoreDatasetTable{ID: 501, DatasetGroupID: 201, DatasourceID: int64PtrVisualizationRepo(301), PhysicalTable: strPtrVisualizationRepo("orders")}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetTableField{ID: 601, DatasetGroupID: 201, DatasetTableID: int64PtrVisualizationRepo(501), DatasourceID: int64PtrVisualizationRepo(301), OriginName: &fieldOrigin, Name: &fieldName}).Error)
	require.NoError(t, db.Create(&visualizationCoreDatasourceTask{ID: 701, DsID: 301, Name: "sync task"}).Error)

	groups, err := repo.FindDatasetGroupsByIDs([]int64{201})
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, int64(201), mapInt64Value(t, groups[0], "id"))

	tables, err := repo.FindDatasetTablesByGroupIDs([]int64{201})
	require.NoError(t, err)
	require.Len(t, tables, 1)
	assert.Equal(t, int64(501), mapInt64Value(t, tables[0], "id"))

	fields, err := repo.FindDatasetTableFieldsByGroupIDs([]int64{201})
	require.NoError(t, err)
	require.Len(t, fields, 1)
	assert.Equal(t, int64(601), mapInt64Value(t, fields[0], "id"))

	datasources, err := repo.FindDatasourcesByGroupIDs([]int64{201})
	require.NoError(t, err)
	require.Len(t, datasources, 1)
	assert.Equal(t, int64(301), mapInt64Value(t, datasources[0], "id"))

	tasks, err := repo.FindDatasourceTasksByGroupIDs([]int64{201})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, int64(701), mapInt64Value(t, tasks[0], "id"))

	chartA := newChartView(801, 402, "chart-a")
	chartB := newChartView(802, 402, "chart-b")
	require.NoError(t, db.Create([]*chart.CoreChartView{chartA, chartB}).Error)

	chartViews, err := repo.GetChartViewsBySceneID(402)
	require.NoError(t, err)
	require.Len(t, chartViews, 2)

	chartMaps, err := repo.FindChartViewsByIDs([]int64{801, 802})
	require.NoError(t, err)
	require.Len(t, chartMaps, 2)

	emptyChartMaps, err := repo.FindChartViewsByIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyChartMaps)
	emptyGroups, err := repo.FindDatasetGroupsByIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyGroups)
	emptyTables, err := repo.FindDatasetTablesByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyTables)
	emptyFields, err := repo.FindDatasetTableFieldsByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyFields)
	emptyDatasources, err := repo.FindDatasourcesByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyDatasources)
	emptyTasks, err := repo.FindDatasourceTasksByGroupIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyTasks)
}

func TestVisualizationRepository_CopyAndLinkageFlows(t *testing.T) {
	repo, db := setupVisualizationRepositoryTest(t)

	sourceSceneID := int64(100)
	targetSceneID := int64(200)
	copyID := int64(1000)
	now := time.Now().UnixMilli()
	person := "tester"
	active := true
	linkType := "inner"
	jumpType := "panel"
	content := "jump-content"
	attachParams := "{}"
	jumpInfoText := "info"
	sourceField := "source_field"
	targetField := "target_field"

	require.NoError(t, db.Create([]*chart.CoreChartView{
		newChartView(1, sourceSceneID, "source-1"),
		newChartView(2, sourceSceneID, "source-2"),
	}).Error)
	require.NoError(t, repo.CopyChartViews(sourceSceneID, targetSceneID, copyID, ""))

	mapping, err := repo.GetCopiedChartViewMapping(copyID)
	require.NoError(t, err)
	assert.Equal(t, int64(1001), mapping[1])
	assert.Equal(t, int64(1002), mapping[2])

	viewCopies, err := repo.GetChartViewsBySceneID(targetSceneID)
	require.NoError(t, err)
	require.Len(t, viewCopies, 2)

	snapshotSourceSceneID := int64(300)
	snapshotTargetSceneID := int64(301)
	snapshotView := newSnapshotChartView(50, snapshotSourceSceneID, "snapshot-source")
	require.NoError(t, repo.SaveSnapshotChartView(snapshotView))
	require.NoError(t, repo.CopyChartViews(snapshotSourceSceneID, snapshotTargetSceneID, 2000, "snapshot"))

	snapshotCopies, err := repo.GetSnapshotChartViewsBySceneID(snapshotTargetSceneID)
	require.NoError(t, err)
	require.Len(t, snapshotCopies, 1)
	assert.Equal(t, int64(2050), snapshotCopies[0].ID)

	require.NoError(t, db.Create(&visualizationLinkage{ID: 10, DvID: sourceSceneID, SourceViewID: 1, TargetViewID: 2, UpdateTime: &now, UpdatePeople: &person, LinkageActive: &active}).Error)
	require.NoError(t, db.Create(&visualizationLinkageField{ID: 20, LinkageID: 10, SourceField: &sourceField, TargetField: &targetField, UpdateTime: &now}).Error)
	require.NoError(t, db.Create(&visualizationLinkJump{ID: 30, SourceDvID: sourceSceneID, SourceViewID: 1, LinkJumpInfo: &jumpInfoText, Checked: &active}).Error)
	require.NoError(t, db.Create(&visualizationLinkJumpInfo{ID: 40, LinkJumpID: 30, LinkType: &linkType, JumpType: &jumpType, TargetDvID: int64PtrVisualizationRepo(999), SourceFieldID: int64PtrVisualizationRepo(601), Content: &content, Checked: &active, AttachParams: &attachParams}).Error)
	require.NoError(t, db.Create(&visualizationLinkJumpTargetViewInfo{TargetID: 50, LinkJumpInfoID: 40, SourceFieldActiveID: int64PtrVisualizationRepo(1), TargetViewID: int64PtrVisualizationRepo(2), TargetFieldID: int64PtrVisualizationRepo(3)}).Error)

	linkages, err := repo.FindLinkagesByDvID(sourceSceneID)
	require.NoError(t, err)
	require.Len(t, linkages, 1)
	assert.Equal(t, int64(10), mapInt64Value(t, linkages[0], "id"))

	linkageFields, err := repo.FindLinkageFieldsByDvID(sourceSceneID)
	require.NoError(t, err)
	require.Len(t, linkageFields, 1)
	assert.Equal(t, int64(20), mapInt64Value(t, linkageFields[0], "id"))

	linkJumps, err := repo.FindLinkJumpsByDvID(sourceSceneID)
	require.NoError(t, err)
	require.Len(t, linkJumps, 1)
	assert.Equal(t, int64(30), mapInt64Value(t, linkJumps[0], "id"))

	jumpInfos, err := repo.FindLinkJumpInfosByDvID(sourceSceneID)
	require.NoError(t, err)
	require.Len(t, jumpInfos, 1)
	assert.Equal(t, int64(40), mapInt64Value(t, jumpInfos[0], "id"))

	jumpTargets, err := repo.FindLinkJumpTargetViewInfosByDvID(sourceSceneID)
	require.NoError(t, err)
	require.Len(t, jumpTargets, 1)
	assert.Equal(t, int64(50), mapInt64Value(t, jumpTargets[0], "target_id"))

	require.NoError(t, repo.CopyLinkages(copyID))
	require.NoError(t, repo.CopyLinkageFields(copyID))
	require.NoError(t, repo.CopyLinkJumps(copyID))
	require.NoError(t, repo.CopyLinkJumpInfos(copyID))
	require.NoError(t, repo.CopyLinkJumpTargetInfos(copyID))

	copiedLinkages, err := repo.FindLinkagesByDvID(targetSceneID)
	require.NoError(t, err)
	require.Len(t, copiedLinkages, 1)
	assert.Equal(t, int64(1010), mapInt64Value(t, copiedLinkages[0], "id"))
	assert.Equal(t, int64(1001), mapInt64Value(t, copiedLinkages[0], "source_view_id"))
	assert.Equal(t, int64(1002), mapInt64Value(t, copiedLinkages[0], "target_view_id"))

	linkageFields, err = repo.FindLinkageFieldsByDvID(targetSceneID)
	require.NoError(t, err)
	require.Len(t, linkageFields, 1)
	assert.Equal(t, int64(1020), mapInt64Value(t, linkageFields[0], "id"))

	linkJumps, err = repo.FindLinkJumpsByDvID(targetSceneID)
	require.NoError(t, err)
	require.Len(t, linkJumps, 1)
	assert.Equal(t, int64(1030), mapInt64Value(t, linkJumps[0], "id"))

	jumpInfos, err = repo.FindLinkJumpInfosByDvID(targetSceneID)
	require.NoError(t, err)
	require.Len(t, jumpInfos, 1)
	assert.Equal(t, int64(1040), mapInt64Value(t, jumpInfos[0], "id"))

	jumpTargets, err = repo.FindLinkJumpTargetViewInfosByDvID(targetSceneID)
	require.NoError(t, err)
	require.Len(t, jumpTargets, 1)
	assert.Equal(t, int64(1050), mapInt64Value(t, jumpTargets[0], "target_id"))
}
