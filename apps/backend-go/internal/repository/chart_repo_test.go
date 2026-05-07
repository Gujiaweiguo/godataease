package repository

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupChartRepositoryTest(t *testing.T) (*ChartRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&chart.CoreChartView{}))
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetTable{}))
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetTableField{}))
	require.NoError(t, db.Exec(`
		CREATE TABLE data_visualization_info (
			id INTEGER PRIMARY KEY,
			name TEXT,
			type TEXT,
			component_data TEXT
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE snapshot_core_chart_view (
			id INTEGER PRIMARY KEY,
			title TEXT,
			type TEXT,
			table_id INTEGER,
			scene_id INTEGER,
			x_axis TEXT,
			x_axis_ext TEXT,
			y_axis TEXT,
			y_axis_ext TEXT,
			ext_stack TEXT,
			ext_bubble TEXT,
			flow_map_start_name TEXT,
			flow_map_end_name TEXT,
			ext_color TEXT,
			ext_label TEXT,
			ext_tooltip TEXT
		)
	`).Error)
	require.NoError(t, db.Exec(`
		CREATE TABLE snapshot_data_visualization_info (
			id INTEGER PRIMARY KEY,
			name TEXT,
			type TEXT
		)
	`).Error)

	return NewChartRepository(db), db
}

func strPtrChartRepo(v string) *string { return &v }
func int64PtrChartRepo(v int64) *int64 { return &v }
func boolPtrChartRepo(v bool) *bool    { return &v }

func createChartField(t *testing.T, db *gorm.DB, field *dataset.CoreDatasetTableField) {
	t.Helper()
	require.NoError(t, db.Create(field).Error)
}

func createChartView(t *testing.T, db *gorm.DB, view *chart.CoreChartView) {
	t.Helper()
	require.NoError(t, db.Create(view).Error)
}

func createDatasetTable(t *testing.T, db *gorm.DB, table *dataset.CoreDatasetTable) {
	t.Helper()
	require.NoError(t, db.Create(table).Error)
}

func TestChartRepository_NewGetByIDAndUpdate(t *testing.T) {
	repo, db := setupChartRepositoryTest(t)
	require.NotNil(t, repo)
	require.Equal(t, db, repo.db)

	viewType := "bar"
	title := "Original Chart"
	view := &chart.CoreChartView{Title: &title, Type: &viewType}
	createChartView(t, db, view)

	found, err := repo.GetByID(view.ID)
	require.NoError(t, err)
	require.NotNil(t, found.Title)
	assert.Equal(t, title, *found.Title)

	updated := "Updated Chart"
	view.Title = &updated
	require.NoError(t, repo.Update(view))

	found, err = repo.GetByID(view.ID)
	require.NoError(t, err)
	require.NotNil(t, found.Title)
	assert.Equal(t, updated, *found.Title)

	_, err = repo.GetByID(99999)
	require.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestChartRepository_QueryViewOption(t *testing.T) {
	repo, db := setupChartRepositoryTest(t)
	barType := "bar"
	vQueryType := "VQuery"
	pieType := "pie"
	sceneID := int64(31)
	otherSceneID := int64(32)
	createChartView(t, db, &chart.CoreChartView{ID: 1, Title: strPtrChartRepo("Sales"), Type: &barType, SceneID: &sceneID})
	createChartView(t, db, &chart.CoreChartView{ID: 2, Title: strPtrChartRepo("Hidden"), Type: &vQueryType, SceneID: &sceneID})
	createChartView(t, db, &chart.CoreChartView{ID: 3, Title: strPtrChartRepo("Other"), Type: &pieType, SceneID: &otherSceneID})

	options, err := repo.QueryViewOption(sceneID)
	require.NoError(t, err)
	require.Len(t, options, 1)
	assert.Equal(t, int64(1), options[0].ID)
	assert.Equal(t, "Sales", options[0].Title)
	assert.Equal(t, "bar", options[0].Type)
}

func TestChartRepository_GetVisualizationComponentData(t *testing.T) {
	repo, db := setupChartRepositoryTest(t)
	require.NoError(t, db.Exec(`INSERT INTO data_visualization_info (id, name, type, component_data) VALUES (1, 'Panel', 'dashboard', '{"layout":1}')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO data_visualization_info (id, name, type, component_data) VALUES (2, 'Empty', 'dashboard', NULL)`).Error)

	componentData, err := repo.GetVisualizationComponentData(1)
	require.NoError(t, err)
	assert.Equal(t, `{"layout":1}`, componentData)

	componentData, err = repo.GetVisualizationComponentData(2)
	require.NoError(t, err)
	assert.Empty(t, componentData)

	componentData, err = repo.GetVisualizationComponentData(999)
	require.NoError(t, err)
	assert.Empty(t, componentData)
}

func TestChartRepository_QueryChartBaseInfo(t *testing.T) {
	repo, db := setupChartRepositoryTest(t)
	require.NoError(t, db.Exec(`INSERT INTO data_visualization_info (id, name, type, component_data) VALUES (100, 'Main Board', 'dashboard', NULL)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO snapshot_data_visualization_info (id, name, type) VALUES (200, 'Snapshot Board', 'panel')`).Error)

	tableID := int64(8)
	sceneID := int64(100)
	chartType := "bar"
	title := "Revenue"
	xAxis := `[{"id":"x1","name":"month"}]`
	yAxis := `[{"id":"y1","name":"amount"}]`
	invalidJSON := `[{`
	nullJSON := "null"
	createChartView(t, db, &chart.CoreChartView{
		ID:               10,
		Title:            &title,
		SceneID:          &sceneID,
		TableID:          &tableID,
		Type:             &chartType,
		XAxis:            &xAxis,
		XAxisExt:         &invalidJSON,
		YAxis:            &yAxis,
		YAxisExt:         &nullJSON,
		ExtStack:         strPtrChartRepo(""),
		ExtBubble:        strPtrChartRepo(`[{"id":"bubble"}]`),
		FlowMapStartName: strPtrChartRepo(`[{"id":"start"}]`),
		FlowMapEndName:   strPtrChartRepo(`[{"id":"end"}]`),
		ExtColor:         strPtrChartRepo(`[{"id":"color"}]`),
		ExtLabel:         strPtrChartRepo(`[{"id":"label"}]`),
		ExtTooltip:       strPtrChartRepo(`[{"id":"tooltip"}]`),
	})
	require.NoError(t, db.Exec(`
		INSERT INTO snapshot_core_chart_view
		(id, title, type, table_id, scene_id, x_axis, x_axis_ext, y_axis, y_axis_ext, ext_stack, ext_bubble, flow_map_start_name, flow_map_end_name, ext_color, ext_label, ext_tooltip)
		VALUES
		(20, 'Snapshot Revenue', 'line', 18, 200, '[{"id":"sx"}]', NULL, '[{"id":"sy"}]', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)
	`).Error)

	baseInfo, err := repo.QueryChartBaseInfo(10, "")
	require.NoError(t, err)
	require.NotNil(t, baseInfo)
	assert.Equal(t, int64(10), baseInfo.ChartID)
	assert.Equal(t, int64(100), baseInfo.ResourceID)
	assert.Equal(t, "bar", baseInfo.ChartType)
	assert.Equal(t, "Revenue", baseInfo.ChartName)
	assert.Equal(t, "dashboard", baseInfo.ResourceType)
	assert.Equal(t, "Main Board", baseInfo.ResourceName)
	require.NotNil(t, baseInfo.TableID)
	assert.Equal(t, tableID, *baseInfo.TableID)
	assert.Len(t, baseInfo.XAxis, 1)
	assert.Len(t, baseInfo.YAxis, 1)
	assert.Empty(t, baseInfo.XAxisExt)
	assert.Empty(t, baseInfo.YAxisExt)
	assert.Empty(t, baseInfo.ExtStack)
	assert.Len(t, baseInfo.ExtBubble, 1)
	assert.Len(t, baseInfo.FlowMapStartName, 1)
	assert.Len(t, baseInfo.FlowMapEndName, 1)
	assert.Len(t, baseInfo.ExtColor, 1)
	assert.Len(t, baseInfo.ExtLabel, 1)
	assert.Len(t, baseInfo.ExtTooltip, 1)

	snapshotInfo, err := repo.QueryChartBaseInfo(20, "snapshot")
	require.NoError(t, err)
	require.NotNil(t, snapshotInfo)
	assert.Equal(t, int64(20), snapshotInfo.ChartID)
	assert.Equal(t, int64(200), snapshotInfo.ResourceID)
	assert.Equal(t, "line", snapshotInfo.ChartType)
	assert.Equal(t, "Snapshot Revenue", snapshotInfo.ChartName)
	assert.Equal(t, "panel", snapshotInfo.ResourceType)
	assert.Equal(t, "Snapshot Board", snapshotInfo.ResourceName)
	assert.Len(t, snapshotInfo.XAxis, 1)
	assert.Len(t, snapshotInfo.YAxis, 1)

	missing, err := repo.QueryChartBaseInfo(999, "")
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestChartRepository_QueryRowsAndDatasetGroupLookup(t *testing.T) {
	repo, db := setupChartRepositoryTest(t)
	physicalName := "chart_rows"
	createDatasetTable(t, db, &dataset.CoreDatasetTable{ID: 51, DatasetGroupID: 901, PhysicalTable: &physicalName})
	createChartView(t, db, &chart.CoreChartView{ID: 61, TableID: int64PtrChartRepo(51)})
	require.NoError(t, db.Exec("CREATE TABLE chart_rows (id INTEGER PRIMARY KEY, category TEXT, amount INTEGER)").Error)
	require.NoError(t, db.Exec("INSERT INTO chart_rows (id, category, amount) VALUES (1, 'A', 10), (2, 'B', 20), (3, 'A', 30)").Error)

	rows, total, err := repo.QueryRows(61, 0)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, int64(3), total)

	rows, total, err = repo.QueryRowsWithFilter(61, "category, amount", "category = ?", []interface{}{"A"}, 1)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(2), total)
	assert.Equal(t, "A", rows[0]["category"])

	rows, total, err = repo.QueryRowsWithFilter(61, "", "", nil, 999)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, int64(3), total)

	groupID, err := repo.GetDatasetGroupIDByChartID(61)
	require.NoError(t, err)
	assert.Equal(t, int64(901), groupID)
}

func TestChartRepository_QueryRowsErrorPaths(t *testing.T) {
	t.Run("chart missing dataset binding", func(t *testing.T) {
		repo, db := setupChartRepositoryTest(t)
		createChartView(t, db, &chart.CoreChartView{ID: 70})

		rows, total, err := repo.QueryRows(70, 10)
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.Zero(t, total)
		assert.Equal(t, "chart does not bind dataset table", err.Error())

		_, err = repo.GetDatasetGroupIDByChartID(70)
		require.Error(t, err)
		assert.Equal(t, "chart does not bind dataset table", err.Error())
	})

	t.Run("invalid dataset table name", func(t *testing.T) {
		repo, db := setupChartRepositoryTest(t)
		badName := "bad-name;drop"
		createDatasetTable(t, db, &dataset.CoreDatasetTable{ID: 71, DatasetGroupID: 902, PhysicalTable: &badName})
		createChartView(t, db, &chart.CoreChartView{ID: 72, TableID: int64PtrChartRepo(71)})

		rows, total, err := repo.QueryRows(72, 10)
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.Zero(t, total)
		assert.Equal(t, "invalid dataset table name", err.Error())
	})

	t.Run("missing dynamic table returns error", func(t *testing.T) {
		repo, db := setupChartRepositoryTest(t)
		physicalName := "missing_rows"
		createDatasetTable(t, db, &dataset.CoreDatasetTable{ID: 73, DatasetGroupID: 903, PhysicalTable: &physicalName})
		createChartView(t, db, &chart.CoreChartView{ID: 74, TableID: int64PtrChartRepo(73)})

		rows, total, err := repo.QueryRowsWithFilter(74, "id", "", nil, 10)
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.Zero(t, total)
	})

	t.Run("missing chart returns lookup error", func(t *testing.T) {
		repo, _ := setupChartRepositoryTest(t)

		rows, total, err := repo.QueryRows(404, 10)
		require.Error(t, err)
		assert.Nil(t, rows)
		assert.Zero(t, total)
		assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
	})
}

func TestChartRepository_DatasetFieldOperations(t *testing.T) {
	repo, db := setupChartRepositoryTest(t)
	chartID := int64(88)
	checked := true
	nameOne := "name_one"
	nameTwo := "name_two"
	nameThree := "name_three"
	aliasOne := "Field One"
	aliasTwo := "Field Two"
	origin := "origin"
	groupType := "d"
	fieldType := "VARCHAR"
	deType := 0

	fieldOne := &dataset.CoreDatasetTableField{
		ID:             801,
		DatasetGroupID: 501,
		Name:           &nameOne,
		DataeaseName:   strPtrChartRepo("de_one"),
		FieldShortName: strPtrChartRepo("one"),
		OriginName:     &origin,
		GroupType:      &groupType,
		Type:           &fieldType,
		DeType:         &deType,
		Checked:        &checked,
	}
	fieldTwo := &dataset.CoreDatasetTableField{
		ID:             802,
		DatasetGroupID: 501,
		ChartID:        &chartID,
		Name:           &nameTwo,
		DataeaseName:   &aliasTwo,
		FieldShortName: strPtrChartRepo("two"),
		OriginName:     &origin,
		GroupType:      &groupType,
		Type:           &fieldType,
		DeType:         &deType,
	}
	fieldThree := &dataset.CoreDatasetTableField{
		ID:             803,
		DatasetGroupID: 501,
		Name:           &nameThree,
		DataeaseName:   &aliasOne,
		FieldShortName: strPtrChartRepo("three"),
		OriginName:     &origin,
		GroupType:      &groupType,
		Type:           &fieldType,
		DeType:         &deType,
		Checked:        boolPtrChartRepo(false),
	}

	require.NoError(t, repo.CreateDatasetField(fieldOne))
	createChartField(t, db, fieldTwo)
	createChartField(t, db, fieldThree)

	byGroup, err := repo.ListDatasetFieldsByGroup(501)
	require.NoError(t, err)
	require.Len(t, byGroup, 1)
	assert.Equal(t, fieldOne.ID, byGroup[0].ID)

	byChart, err := repo.ListDatasetFieldsByChart(chartID)
	require.NoError(t, err)
	require.Len(t, byChart, 1)
	assert.Equal(t, fieldTwo.ID, byChart[0].ID)

	found, err := repo.GetDatasetFieldByID(fieldOne.ID)
	require.NoError(t, err)
	require.NotNil(t, found.Name)
	assert.Equal(t, nameOne, *found.Name)

	count, err := repo.CountDatasetFieldName(501, nameOne)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	require.NoError(t, repo.UpdateDatasetFieldNames(fieldOne.ID, "renamed", "short"))
	found, err = repo.GetDatasetFieldByID(fieldOne.ID)
	require.NoError(t, err)
	require.NotNil(t, found.DataeaseName)
	require.NotNil(t, found.FieldShortName)
	assert.Equal(t, "renamed", *found.DataeaseName)
	assert.Equal(t, "short", *found.FieldShortName)

	require.NoError(t, repo.DeleteDatasetField(fieldOne.ID))
	_, err = repo.GetDatasetFieldByID(fieldOne.ID)
	require.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	require.NoError(t, repo.DeleteDatasetFieldsByChart(chartID))
	remainingChartFields, err := repo.ListDatasetFieldsByChart(chartID)
	require.NoError(t, err)
	assert.Empty(t, remainingChartFields)
}

func TestChartRepository_GetDatasetFieldByID_NotFound(t *testing.T) {
	repo, _ := setupChartRepositoryTest(t)
	_, err := repo.GetDatasetFieldByID(123456)
	require.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestPtrInt64(t *testing.T) {
	assert.Equal(t, int64(7), ptrInt64(nil, 7))
	assert.Equal(t, int64(9), ptrInt64(int64PtrChartRepo(9), 7))
}
