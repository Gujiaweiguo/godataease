package repository

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/permission"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDatasetRepository_Gap2CreateAndFieldHelpers(t *testing.T) {
	repo, _ := setupDatasetRepositoryTest(t)

	tableName := "preview_rows"
	tableAlias := "Orders"
	require.NoError(t, repo.CreateTable(&dataset.CoreDatasetTable{ID: 91, Name: &tableAlias, DatasetGroupID: 88, PhysicalTable: &tableName}))

	extField := 2
	fieldName := "city_alias"
	originName := "city"
	typeName := "varchar"
	dsID := int64(301)
	chartID := int64(401)
	fields := []dataset.CoreDatasetTableField{
		{ID: 501, DatasourceID: &dsID, DatasetGroupID: 88, DatasetTableID: int64PtrDatasetGap2(91), ChartID: &chartID, OriginName: &originName, Name: &fieldName, Type: &typeName, ExtField: &extField},
		{ID: 502, DatasourceID: &dsID, DatasetGroupID: 88, DatasetTableID: int64PtrDatasetGap2(91), ChartID: &chartID, OriginName: strPtrDatasetRepo("[501]"), Name: strPtrDatasetRepo("derived"), Type: &typeName, ExtField: &extField},
	}
	require.NoError(t, repo.BatchCreateFields(fields))
	require.NoError(t, repo.BatchCreateFields(nil))

	chartFields, err := repo.ListFieldsByChartID(chartID)
	require.NoError(t, err)
	require.Len(t, chartFields, 2)

	chartFields, err = repo.ListFieldsByChartID(0)
	require.NoError(t, err)
	assert.Empty(t, chartFields)

	count, err := repo.CountDerivedFieldReferences(501)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	count, err = repo.CountDerivedFieldReferences(0)
	require.NoError(t, err)
	assert.Zero(t, count)

	byDsIDs, err := repo.ListFieldsByDsIds([]int64{dsID})
	require.NoError(t, err)
	require.Len(t, byDsIDs, 2)

	createField := &dataset.CoreDatasetTableField{ID: 503, DatasourceID: &dsID, DatasetGroupID: 88, OriginName: strPtrDatasetRepo("province"), Name: strPtrDatasetRepo("Province")}
	require.NoError(t, repo.CreateDatasetField(createField))

	updatedOrigin := "province_code"
	updatedName := "Province Code"
	updatedDataeaseName := "province_code_de"
	updatedShortName := "province"
	updatedGroupType := "quota"
	updatedType := "int"
	updatedDeType := 2
	updatedExtractType := 3
	updatedExtField := 4
	updatedChecked := false
	updatedParams := "{\"k\":1}"
	createField.OriginName = &updatedOrigin
	createField.Name = &updatedName
	createField.DataeaseName = &updatedDataeaseName
	createField.FieldShortName = &updatedShortName
	createField.GroupType = &updatedGroupType
	createField.Type = &updatedType
	createField.DeType = &updatedDeType
	createField.DeExtractType = &updatedExtractType
	createField.ExtField = &updatedExtField
	createField.Checked = &updatedChecked
	createField.Params = &updatedParams
	require.NoError(t, repo.UpdateDatasetField(createField))

	found, err := repo.GetFieldByID(503)
	require.NoError(t, err)
	assert.Equal(t, updatedOrigin, *found.OriginName)
	assert.Equal(t, updatedName, *found.Name)
	assert.Equal(t, updatedDataeaseName, *found.DataeaseName)
	assert.Equal(t, updatedShortName, *found.FieldShortName)
	assert.Equal(t, updatedGroupType, *found.GroupType)
	assert.Equal(t, updatedType, *found.Type)
	assert.Equal(t, updatedDeType, *found.DeType)
	assert.Equal(t, updatedExtractType, *found.DeExtractType)
	assert.Equal(t, updatedExtField, *found.ExtField)
	assert.Equal(t, updatedChecked, *found.Checked)
	assert.Equal(t, updatedParams, *found.Params)
}

func TestDatasetRepository_Gap2FolderAndReferenceQueries(t *testing.T) {
	repo, db := setupDatasetRepositoryTest(t)
	folderType := dataset.NodeTypeFolder
	datasetType := dataset.NodeTypeDataset
	deletedFlag := 1
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 601, Name: "Shared", NodeType: &folderType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 602, Name: "Shared", NodeType: &datasetType}).Error)
	require.NoError(t, db.Create(&dataset.CoreDatasetGroup{ID: 603, Name: "Shared", NodeType: &folderType, DelFlag: &deletedFlag}).Error)

	count, err := repo.CountFolderByNameAndPID("Shared", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	require.NoError(t, db.AutoMigrate(&permission.DataPermRow{}, &permission.DataPermColumn{}, &auto.VisualizationLinkageField{}, &auto.VisualizationLinkJumpInfo{}, &auto.VisualizationOuterParamsTargetViewInfo{}))
	require.NoError(t, db.Create(&permission.DataPermRow{ID: 701, DatasetGroupID: 601}).Error)
	require.NoError(t, db.Create(&permission.DataPermColumn{ID: 702, DatasetGroupID: 601}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkageField{ID: 703, SourceField: 801}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkageField{ID: 704, TargetField: 801}).Error)
	require.NoError(t, db.Create(&auto.VisualizationLinkJumpInfo{ID: 705, SourceFieldID: 801}).Error)
	require.NoError(t, db.Create(&auto.VisualizationOuterParamsTargetViewInfo{TargetID: "706", TargetFieldID: "801"}).Error)

	rows, err := repo.ListRowPermissionsByDatasetGroupID(601)
	require.NoError(t, err)
	require.Len(t, rows, 1)

	columns, err := repo.ListColumnPermissionsByDatasetGroupID(601)
	require.NoError(t, err)
	require.Len(t, columns, 1)

	linkageCount, err := repo.CountVisualizationLinkageFieldReferences(801)
	require.NoError(t, err)
	assert.Equal(t, int64(2), linkageCount)

	jumpCount, err := repo.CountVisualizationLinkJumpReferences(801)
	require.NoError(t, err)
	assert.Equal(t, int64(1), jumpCount)

	outerCount, err := repo.CountVisualizationOuterParamReferences(801)
	require.NoError(t, err)
	assert.Equal(t, int64(1), outerCount)

	chartViews, err := repo.ListChartViewsByDatasetGroupID(0)
	require.NoError(t, err)
	assert.Empty(t, chartViews)
}

func TestDatasetRepository_Gap2MissingTableFallbacksAndTreeValues(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&dataset.CoreDatasetGroup{}, &dataset.CoreDatasetTable{}, &dataset.CoreDatasetTableField{}))
	require.NoError(t, db.Exec("CREATE TABLE preview_rows (id INTEGER PRIMARY KEY, category TEXT, city TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO preview_rows (id, category, city) VALUES (1, 'A', 'Shanghai'), (2, 'A', 'Beijing'), (3, 'B', 'Shenzhen')").Error)

	repo := NewDatasetRepository(db)

	views, err := repo.ListChartViewsByDatasetGroupID(99)
	require.NoError(t, err)
	assert.Empty(t, views)

	rows, err := repo.ListRowPermissionsByDatasetGroupID(99)
	require.NoError(t, err)
	assert.Empty(t, rows)

	columns, err := repo.ListColumnPermissionsByDatasetGroupID(99)
	require.NoError(t, err)
	assert.Empty(t, columns)

	count, err := repo.CountVisualizationLinkageFieldReferences(88)
	require.NoError(t, err)
	assert.Zero(t, count)

	count, err = repo.CountVisualizationLinkJumpReferences(88)
	require.NoError(t, err)
	assert.Zero(t, count)

	count, err = repo.CountVisualizationOuterParamReferences(88)
	require.NoError(t, err)
	assert.Zero(t, count)

	treeRows, err := repo.QueryFieldTreeValues(
		"preview_rows",
		[]dataset.EnumObjectColumn{{Column: "category", Alias: "label"}, {Column: "city", Alias: "value"}},
		[]dataset.EnumFilterClause{{Column: "category", Values: []string{"A"}}},
		1,
	)
	require.NoError(t, err)
	require.Len(t, treeRows, 1)
	assert.Equal(t, "A", treeRows[0]["label"])

	treeRows, err = repo.QueryFieldTreeValues("preview_rows", nil, nil, 0)
	require.NoError(t, err)
	assert.Empty(t, treeRows)
}

func TestDatasetRepository_Gap2TreeBuilderAndMissingTableDetection(t *testing.T) {
	selectParts, orderParts, err := buildTreeSelectOrder([]dataset.EnumObjectColumn{{Column: "city", Alias: "label"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"`city` AS `label`"}, selectParts)
	assert.Equal(t, []string{"`city` ASC"}, orderParts)

	_, _, err = buildTreeSelectOrder([]dataset.EnumObjectColumn{{Column: "city", Alias: "   "}})
	require.Error(t, err)

	assert.True(t, isMissingTableErr(errors.New("no such table: data_perm_row")))
	assert.True(t, isMissingTableErr(errors.New("Table 'x' doesn't exist")))
	assert.False(t, isMissingTableErr(nil))
	assert.False(t, isMissingTableErr(errors.New("different error")))
}

func int64PtrDatasetGap2(v int64) *int64 { return &v }
