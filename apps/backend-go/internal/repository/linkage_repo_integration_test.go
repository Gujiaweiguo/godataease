//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkageRepository_CRUD(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	ensureLinkageRepositoryTables(t)
	cleanupTables("snapshot_visualization_linkage_field", "snapshot_visualization_linkage")

	repo := NewLinkageRepository(testDB)
	linkage := &auto.SnapshotVisualizationLinkage{
		DvID:          6101,
		SourceViewID:  6102,
		TargetViewID:  6103,
		UpdateTime:    1710004001,
		UpdatePeople:  "tester",
		LinkageActive: true,
		Ext1:          "ext-1",
		Ext2:          "ext-2",
		CopyFrom:      1,
		CopyID:        2,
	}
	require.NoError(t, repo.CreateLinkage(linkage))
	assert.NotZero(t, linkage.ID)

	field := &auto.SnapshotVisualizationLinkageField{
		LinkageID:   linkage.ID,
		SourceField: 7101,
		TargetField: 7102,
		UpdateTime:  1710004002,
		CopyFrom:    3,
		CopyID:      4,
	}
	require.NoError(t, repo.CreateLinkageField(field))
	assert.NotZero(t, field.ID)

	var linkageCount int64
	require.NoError(t, testDB.Model(&auto.SnapshotVisualizationLinkage{}).Where("id = ?", linkage.ID).Count(&linkageCount).Error)
	assert.Equal(t, int64(1), linkageCount)

	var fieldCount int64
	require.NoError(t, testDB.Model(&auto.SnapshotVisualizationLinkageField{}).Where("id = ?", field.ID).Count(&fieldCount).Error)
	assert.Equal(t, int64(1), fieldCount)

	require.NoError(t, repo.DeleteLinkageAndFields(linkage.DvID, linkage.SourceViewID))

	require.NoError(t, testDB.Model(&auto.SnapshotVisualizationLinkage{}).Where("id = ?", linkage.ID).Count(&linkageCount).Error)
	assert.Equal(t, int64(0), linkageCount)
	require.NoError(t, testDB.Model(&auto.SnapshotVisualizationLinkageField{}).Where("id = ?", field.ID).Count(&fieldCount).Error)
	assert.Equal(t, int64(0), fieldCount)
}

func TestLinkageRepository_GetViewLinkageGather(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	ensureLinkageRepositoryTables(t)
	cleanupTables("snapshot_visualization_linkage_field", "snapshot_visualization_linkage", "snapshot_core_chart_view")

	repo := NewLinkageRepository(testDB)
	dvID := int64(6201)
	sourceViewID := int64(6202)
	targetViewID := int64(6203)
	vqueryViewID := int64(6204)
	sourceFieldID := int64(7201)
	targetFieldID := int64(7202)

	require.NoError(t, testDB.Create(&visualization.SnapshotCanvasChartView{
		ID:            sourceViewID,
		Title:         strPtrLinkage("Source View"),
		SceneID:       int64PtrLinkage(dvID),
		TableID:       int64PtrLinkage(8201),
		Type:          strPtrLinkage("bar"),
		LinkageActive: boolPtrLinkage(true),
		UpdateTime:    int64PtrLinkage(1710004101),
	}).Error)
	require.NoError(t, testDB.Create(&visualization.SnapshotCanvasChartView{
		ID:            targetViewID,
		Title:         strPtrLinkage("Target View"),
		SceneID:       int64PtrLinkage(dvID),
		TableID:       int64PtrLinkage(8202),
		Type:          strPtrLinkage("table"),
		LinkageActive: boolPtrLinkage(true),
		UpdateTime:    int64PtrLinkage(1710004102),
	}).Error)
	require.NoError(t, testDB.Create(&visualization.SnapshotCanvasChartView{
		ID:            vqueryViewID,
		Title:         strPtrLinkage("Ignored VQuery"),
		SceneID:       int64PtrLinkage(dvID),
		TableID:       int64PtrLinkage(8203),
		Type:          strPtrLinkage("VQuery"),
		LinkageActive: boolPtrLinkage(true),
		UpdateTime:    int64PtrLinkage(1710004103),
	}).Error)

	linkage := &auto.SnapshotVisualizationLinkage{ID: 6205, DvID: dvID, SourceViewID: sourceViewID, TargetViewID: targetViewID, LinkageActive: true, UpdateTime: 1710004104}
	require.NoError(t, repo.CreateLinkage(linkage))
	field := &auto.SnapshotVisualizationLinkageField{ID: 6206, LinkageID: linkage.ID, SourceField: sourceFieldID, TargetField: targetFieldID, UpdateTime: 1710004105}
	require.NoError(t, repo.CreateLinkageField(field))

	empty, err := repo.GetViewLinkageGather(dvID, sourceViewID, nil, true)
	require.NoError(t, err)
	assert.Nil(t, empty)

	rows, err := repo.GetViewLinkageGather(dvID, sourceViewID, []int64{targetViewID, vqueryViewID}, true)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, targetViewID, rows[0].TargetViewID)
	assert.Equal(t, "table", rows[0].TargetViewType)
	assert.Equal(t, int64(8202), rows[0].TableID)
	assert.Equal(t, "Target View", rows[0].TargetViewName)
	assert.Equal(t, sourceViewID, rows[0].SourceViewID)
	assert.True(t, rows[0].LinkageActive)
	if assert.NotNil(t, rows[0].SourceField) {
		assert.Equal(t, sourceFieldID, *rows[0].SourceField)
	}
	if assert.NotNil(t, rows[0].TargetField) {
		assert.Equal(t, targetFieldID, *rows[0].TargetField)
	}
}

func TestLinkageRepository_InfoFieldsAndUpdate(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	ensureLinkageRepositoryTables(t)
	cleanupTables("snapshot_visualization_linkage_field", "snapshot_visualization_linkage", "snapshot_core_chart_view", "core_dataset_table_field")

	repo := NewLinkageRepository(testDB)
	dvID := int64(6301)
	sourceViewID := int64(6302)
	targetViewID := int64(6303)
	datasetGroupID := int64(8301)
	sourceFieldID := int64(7301)
	targetFieldID := int64(7302)
	deType := 2
	originName := "order_id"
	fieldName := "Order ID"
	fieldType := "int"

	require.NoError(t, testDB.Create(&visualization.SnapshotCanvasChartView{
		ID:            sourceViewID,
		Title:         strPtrLinkage("Source Active View"),
		SceneID:       int64PtrLinkage(dvID),
		TableID:       int64PtrLinkage(datasetGroupID),
		Type:          strPtrLinkage("bar"),
		LinkageActive: boolPtrLinkage(true),
		UpdateTime:    int64PtrLinkage(1710004201),
	}).Error)
	require.NoError(t, testDB.Create(&visualization.SnapshotCanvasChartView{
		ID:            targetViewID,
		Title:         strPtrLinkage("Target Active View"),
		SceneID:       int64PtrLinkage(dvID),
		TableID:       int64PtrLinkage(datasetGroupID),
		Type:          strPtrLinkage("line"),
		LinkageActive: boolPtrLinkage(false),
		UpdateTime:    int64PtrLinkage(1710004202),
	}).Error)
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{
		ID:             sourceFieldID,
		DatasetGroupID: datasetGroupID,
		OriginName:     &originName,
		Name:           &fieldName,
		Type:           &fieldType,
		DeType:         &deType,
	}).Error)
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{
		ID:             targetFieldID,
		DatasetGroupID: datasetGroupID,
		OriginName:     strPtrLinkage("customer_id"),
		Name:           strPtrLinkage("Customer ID"),
		Type:           &fieldType,
		DeType:         &deType,
	}).Error)

	activeLinkage := &auto.SnapshotVisualizationLinkage{ID: 6304, DvID: dvID, SourceViewID: sourceViewID, TargetViewID: targetViewID, LinkageActive: true, UpdateTime: 1710004203}
	require.NoError(t, repo.CreateLinkage(activeLinkage))
	require.NoError(t, repo.CreateLinkageField(&auto.SnapshotVisualizationLinkageField{ID: 6305, LinkageID: activeLinkage.ID, SourceField: sourceFieldID, TargetField: targetFieldID, UpdateTime: 1710004204}))

	info, err := repo.GetAllLinkageInfo(dvID, true)
	require.NoError(t, err)
	require.Contains(t, info, "6302#7301")
	assert.Equal(t, []string{"6303#7302"}, info["6302#7301"])

	fields, err := repo.GetDatasetFieldsByGroupID(datasetGroupID)
	require.NoError(t, err)
	require.Len(t, fields, 2)
	fieldMap := make(map[int64]DatasetFieldDTO, len(fields))
	for _, field := range fields {
		fieldMap[field.ID] = field
	}
	require.Contains(t, fieldMap, sourceFieldID)
	require.Contains(t, fieldMap, targetFieldID)
	assert.Equal(t, datasetGroupID, fieldMap[sourceFieldID].DatasetTableID)
	assert.Equal(t, originName, fieldMap[sourceFieldID].OriginName)
	assert.Equal(t, fieldName, fieldMap[sourceFieldID].Name)
	assert.Equal(t, deType, fieldMap[sourceFieldID].DeType)

	require.NoError(t, repo.UpdateChartLinkageActive(targetViewID, true))

	var updated visualization.SnapshotCanvasChartView
	require.NoError(t, testDB.First(&updated, "id = ?", targetViewID).Error)
	if assert.NotNil(t, updated.LinkageActive) {
		assert.True(t, *updated.LinkageActive)
	}
	if assert.NotNil(t, updated.UpdateTime) {
		assert.NotZero(t, *updated.UpdateTime)
	}
}

func ensureLinkageRepositoryTables(t *testing.T) {
	t.Helper()
	require.NoError(t, testDB.AutoMigrate(&visualization.SnapshotCanvasChartView{}))
	ensureColumnExists(t, "core_dataset_table_field", "origin_name", `ALTER TABLE core_dataset_table_field ADD COLUMN origin_name VARCHAR(255)`)
	ensureColumnExists(t, "core_dataset_table_field", "name", `ALTER TABLE core_dataset_table_field ADD COLUMN name VARCHAR(255)`)
	ensureColumnExists(t, "core_dataset_table_field", "de_type", `ALTER TABLE core_dataset_table_field ADD COLUMN de_type INT`)
	ensureColumnExists(t, "core_dataset_table_field", "type", `ALTER TABLE core_dataset_table_field ADD COLUMN type VARCHAR(64)`)
}

func int64PtrLinkage(v int64) *int64 { return &v }
func strPtrLinkage(v string) *string { return &v }
func boolPtrLinkage(v bool) *bool    { return &v }
