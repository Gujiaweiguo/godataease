//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"
	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkJumpRepository_GetTargetVisualizationJumpInfo(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewLinkJumpRepository(testDB)
	cleanupTables(
		auto.TableNameSnapshotVisualizationLinkJumpTargetViewInfo,
		auto.TableNameSnapshotVisualizationLinkJumpInfo,
		auto.TableNameSnapshotVisualizationLinkJump,
		auto.TableNameVisualizationLinkJumpTargetViewInfo,
		auto.TableNameVisualizationLinkJumpInfo,
		auto.TableNameVisualizationLinkJump,
	)

	const (
		coreJumpID           = int64(51001)
		coreJumpInfoID       = int64(51002)
		coreTargetInfoID     = int64(51003)
		coreSourceDvID       = int64(51004)
		coreSourceViewID     = int64(51005)
		coreTargetDvID       = int64(51006)
		coreSourceFieldID    = int64(51007)
		coreSourceActiveID   = int64(51008)
		snapshotJumpID       = int64(52001)
		snapshotJumpInfoID   = int64(52002)
		snapshotTargetInfoID = int64(52003)
		snapshotSourceDvID   = int64(52004)
		snapshotSourceViewID = int64(52005)
		snapshotTargetDvID   = int64(52006)
		snapshotSourceField  = int64(52007)
		snapshotActiveField  = int64(52008)
	)

	require.NoError(t, testDB.Create(&auto.VisualizationLinkJump{
		ID:           coreJumpID,
		SourceDvID:   coreSourceDvID,
		SourceViewID: coreSourceViewID,
		Checked:      true,
	}).Error)
	require.NoError(t, testDB.Create(&auto.VisualizationLinkJumpInfo{
		ID:            coreJumpInfoID,
		LinkJumpID:    coreJumpID,
		LinkType:      "inner",
		JumpType:      "_blank",
		TargetDvID:    coreTargetDvID,
		SourceFieldID: coreSourceFieldID,
		Checked:       true,
		WindowSize:    "middle",
	}).Error)
	require.NoError(t, testDB.Create(&auto.VisualizationLinkJumpTargetViewInfo{
		TargetID:            coreTargetInfoID,
		LinkJumpInfoID:      coreJumpInfoID,
		SourceFieldActiveID: coreSourceActiveID,
		TargetViewID:        "999",
		TargetFieldID:       "888",
		TargetType:          "view",
	}).Error)

	require.NoError(t, testDB.Create(&auto.SnapshotVisualizationLinkJump{
		ID:           snapshotJumpID,
		SourceDvID:   snapshotSourceDvID,
		SourceViewID: snapshotSourceViewID,
		Checked:      true,
	}).Error)
	require.NoError(t, testDB.Create(&auto.SnapshotVisualizationLinkJumpInfo{
		ID:            snapshotJumpInfoID,
		LinkJumpID:    snapshotJumpID,
		LinkType:      "inner",
		JumpType:      "_self",
		TargetDvID:    snapshotTargetDvID,
		SourceFieldID: snapshotSourceField,
		Checked:       true,
		WindowSize:    "middle",
	}).Error)
	require.NoError(t, testDB.Create(&auto.SnapshotVisualizationLinkJumpTargetViewInfo{
		TargetID:            snapshotTargetInfoID,
		LinkJumpInfoID:      snapshotJumpInfoID,
		SourceFieldActiveID: snapshotActiveField,
		TargetViewID:        "777",
		TargetFieldID:       "666",
		TargetType:          "view",
	}).Error)

	t.Run("core tables support mysql concat and field filter", func(t *testing.T) {
		sourceFieldID := coreSourceFieldID
		rows, err := repo.GetTargetVisualizationJumpInfo(coreSourceDvID, coreSourceViewID, coreTargetDvID, nil, false)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "51005#51008#51007", rows[0].SourceInfo)
		assert.Equal(t, "999#888", rows[0].TargetInfo)

		rows, err = repo.GetTargetVisualizationJumpInfo(coreSourceDvID, coreSourceViewID, coreTargetDvID, &sourceFieldID, false)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "51005#51008#51007", rows[0].SourceInfo)

		missingFieldID := sourceFieldID + 999
		rows, err = repo.GetTargetVisualizationJumpInfo(coreSourceDvID, coreSourceViewID, coreTargetDvID, &missingFieldID, false)
		require.NoError(t, err)
		assert.Empty(t, rows)
	})

	t.Run("snapshot tables support mysql concat", func(t *testing.T) {
		rows, err := repo.GetTargetVisualizationJumpInfo(snapshotSourceDvID, snapshotSourceViewID, snapshotTargetDvID, nil, true)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, "52005#52008#52007", rows[0].SourceInfo)
		assert.Equal(t, "777#666", rows[0].TargetInfo)
	})
}

func TestLinkJumpRepository_GetViewTableDetails(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewLinkJumpRepository(testDB)
	cleanupTables("core_chart_view", "core_dataset_table_field", "data_visualization_info")

	const (
		dvID          = int64(53001)
		matchedViewID = int64(53002)
		skippedViewID = int64(53003)
		datasetGroup  = int64(53004)
		fieldID       = int64(53005)
	)

	componentData := `{"views":["53002"],"activeView":"53002"}`
	dvType := "dashboard"
	viewType := "bar"
	originName := "order_id"
	fieldName := "Order ID"
	fieldType := "int"
	deType := 2

	require.NoError(t, testDB.Create(&visualization.DataVisualizationInfo{
		ID:            dvID,
		Name:          "jump dashboard",
		Type:          &dvType,
		ComponentData: &componentData,
	}).Error)
	require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{
		ID:             fieldID,
		DatasetGroupID: datasetGroup,
		OriginName:     &originName,
		Name:           &fieldName,
		Type:           &fieldType,
		DeType:         &deType,
	}).Error)
	require.NoError(t, testDB.Create(&chart.CoreChartView{
		ID:      matchedViewID,
		Title:   strPtrLinkJump("Matched View"),
		SceneID: int64PtrLinkJump(dvID),
		TableID: int64PtrLinkJump(datasetGroup),
		Type:    &viewType,
	}).Error)
	require.NoError(t, testDB.Create(&chart.CoreChartView{
		ID:      skippedViewID,
		Title:   strPtrLinkJump("Skipped View"),
		SceneID: int64PtrLinkJump(dvID),
		TableID: int64PtrLinkJump(datasetGroup),
		Type:    &viewType,
	}).Error)

	rows, err := repo.GetViewTableDetails(dvID)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, matchedViewID, rows[0].ID)
	assert.Equal(t, "Matched View", rows[0].Title)
	assert.Equal(t, dvID, rows[0].DvID)
	assert.Equal(t, fieldID, *rows[0].FieldID)
	assert.Equal(t, originName, rows[0].OriginName)
	assert.Equal(t, fieldName, rows[0].FieldName)
	assert.Equal(t, fieldType, rows[0].FieldType)
	assert.Equal(t, "2", rows[0].DeType)

	t.Run("quoted locate does not match substring IDs", func(t *testing.T) {
		const (
			substringDvID        = int64(53101)
			exactViewID          = int64(53012)
			substringOnlyViewID  = int64(3012)
			substringDatasetID   = int64(53104)
			substringFieldID     = int64(53105)
		)

		substringComponentData := `[{"id":"53012","component":"UserView"}]`
		substringOriginName := "city_id"
		substringFieldName := "City ID"
		substringFieldType := "int"
		substringDeType := 2

		require.NoError(t, testDB.Create(&visualization.DataVisualizationInfo{
			ID:            substringDvID,
			Name:          "substring dashboard",
			Type:          &dvType,
			ComponentData: &substringComponentData,
		}).Error)
		require.NoError(t, testDB.Create(&dataset.CoreDatasetTableField{
			ID:             substringFieldID,
			DatasetGroupID: substringDatasetID,
			OriginName:     &substringOriginName,
			Name:           &substringFieldName,
			Type:           &substringFieldType,
			DeType:         &substringDeType,
		}).Error)
		require.NoError(t, testDB.Create(&chart.CoreChartView{
			ID:      exactViewID,
			Title:   strPtrLinkJump("Exact View"),
			SceneID: int64PtrLinkJump(substringDvID),
			TableID: int64PtrLinkJump(substringDatasetID),
			Type:    &viewType,
		}).Error)
		require.NoError(t, testDB.Create(&chart.CoreChartView{
			ID:      substringOnlyViewID,
			Title:   strPtrLinkJump("Substring View"),
			SceneID: int64PtrLinkJump(substringDvID),
			TableID: int64PtrLinkJump(substringDatasetID),
			Type:    &viewType,
		}).Error)

		rows, err := repo.GetViewTableDetails(substringDvID)
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, exactViewID, rows[0].ID)
		assert.NotEqual(t, substringOnlyViewID, rows[0].ID)
	})
}

func int64PtrLinkJump(v int64) *int64 { return &v }
func strPtrLinkJump(v string) *string { return &v }
