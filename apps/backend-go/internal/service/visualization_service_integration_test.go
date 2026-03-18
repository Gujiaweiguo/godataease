//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVisualizationServiceIntegration_Save(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Test save panel
	nodeType := "panel"
	req := &visualization.SaveRequest{
		Name:            "Test Dashboard",
		NodeType:        &nodeType,
		CanvasStyleData: strPtr("{\"style\":\"dark\"}"),
		ComponentData:   strPtr("{\"components\":[]}"),
	}

	id, err := svc.Save(req, "tester")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify saved
	detail, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "Test Dashboard", detail.Name)
}

func TestVisualizationServiceIntegration_Save_Folder(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Test save folder
	nodeType := "folder"
	req := &visualization.SaveRequest{
		Name:     "Test Folder",
		NodeType: &nodeType,
	}

	id, err := svc.Save(req, "tester")
	assert.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Verify folder status is 1 (active)
	detail, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "Test Folder", detail.Name)
	if detail.Status != nil {
		assert.Equal(t, 1, *detail.Status)
	}
}

func TestVisualizationServiceIntegration_Copy(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	nodeType := "panel"
	visType := "dashboard"
	mobileLayout := true
	createReq := &visualization.SaveRequest{
		Name:            "Copy Source",
		NodeType:        &nodeType,
		Type:            &visType,
		CanvasStyleData: strPtr("{\"style\":\"source\"}"),
		ComponentData:   strPtr("{\"components\":[1]}"),
		MobileLayout:    &mobileLayout,
		ContentID:       strPtr("content-1"),
		CheckVersion:    strPtr("ver-1"),
	}
	sourceID, err := svc.Save(createReq, "creator")
	require.NoError(t, err)

	newPID := int64(100)
	copyID, err := svc.Copy(&visualization.CopyRequest{
		ID:           sourceID,
		Name:         "Copy Target",
		PID:          &newPID,
		Type:         &visType,
		NodeType:     &nodeType,
		MobileLayout: &mobileLayout,
	}, "copier")
	require.NoError(t, err)
	assert.NotEqual(t, sourceID, copyID)

	copied, err := repo.GetByID(copyID)
	require.NoError(t, err)
	assert.Equal(t, "Copy Target", copied.Name)
	assert.Equal(t, newPID, *copied.PID)
	assert.Equal(t, "dashboard", *copied.Type)
	assert.Equal(t, "panel", *copied.NodeType)
	assert.Equal(t, "{\"style\":\"source\"}", *copied.CanvasStyleData)
	assert.Equal(t, "{\"components\":[1]}", *copied.ComponentData)
	assert.Equal(t, "content-1", *copied.ContentID)
	assert.Equal(t, "ver-1", *copied.CheckVersion)
	assert.Equal(t, "copier", *copied.CreateBy)
}

func TestVisualizationServiceIntegration_Update(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:     "Original Name",
		NodeType: &nodeType,
	}
	id, err := svc.Save(createReq, "creator")
	assert.NoError(t, err)

	// Update
	updateReq := &visualization.UpdateRequest{
		ID:   id,
		Name: strPtr("Updated Name"),
	}
	err = svc.Update(updateReq, "updater")
	assert.NoError(t, err)

	// Verify updated
	detail, err := repo.GetByID(id)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", detail.Name)
}

func TestVisualizationServiceIntegration_Update_NotFound(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	updateReq := &visualization.UpdateRequest{
		ID:   99999,
		Name: strPtr("Updated Name"),
	}
	err := svc.Update(updateReq, "updater")
	assert.Error(t, err)
}

func TestVisualizationServiceIntegration_Detail(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:     "Detail Test",
		NodeType: &nodeType,
	}
	id, err := svc.Save(createReq, "tester")
	assert.NoError(t, err)

	// Get detail
	req := &visualization.DetailRequest{ID: id}
	detail, err := svc.Detail(req)
	assert.NoError(t, err)
	assert.Equal(t, "Detail Test", detail.Name)
	assert.Equal(t, id, detail.ID)
}

func TestVisualizationServiceIntegration_Detail_Completeness(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	nodeType := "panel"
	visType := "dashboard"
	mobileLayout := true
	createReq := &visualization.SaveRequest{
		Name:            "Detail Complete",
		NodeType:        &nodeType,
		Type:            &visType,
		CanvasStyleData: strPtr("{\"theme\":\"dark\"}"),
		ComponentData:   strPtr("{\"views\":[1,2]}"),
		MobileLayout:    &mobileLayout,
		ContentID:       strPtr("content-complete"),
		CheckVersion:    strPtr("v-check-1"),
	}
	id, err := svc.Save(createReq, "tester")
	assert.NoError(t, err)

	detail, err := svc.Detail(&visualization.DetailRequest{ID: id})
	assert.NoError(t, err)
	assert.Equal(t, "Detail Complete", detail.Name)
	if assert.NotNil(t, detail.NodeType) {
		assert.Equal(t, "panel", *detail.NodeType)
	}
	if assert.NotNil(t, detail.Type) {
		assert.Equal(t, "dashboard", *detail.Type)
	}
	if assert.NotNil(t, detail.CanvasStyleData) {
		assert.Equal(t, "{\"theme\":\"dark\"}", *detail.CanvasStyleData)
	}
	if assert.NotNil(t, detail.ComponentData) {
		assert.Equal(t, "{\"views\":[1,2]}", *detail.ComponentData)
	}
	if assert.NotNil(t, detail.MobileLayout) {
		assert.True(t, *detail.MobileLayout)
	}
	if assert.NotNil(t, detail.ContentID) {
		assert.Equal(t, "content-complete", *detail.ContentID)
	}
	if assert.NotNil(t, detail.CheckVersion) {
		assert.Equal(t, "v-check-1", *detail.CheckVersion)
	}
}

func TestVisualizationServiceIntegration_Detail_NotFound(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	req := &visualization.DetailRequest{ID: 99999}
	_, err := svc.Detail(req)
	assert.Error(t, err)
}

func TestVisualizationServiceIntegration_List(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create multiple items
	nodeType := "panel"
	for i := 0; i < 5; i++ {
		_, err := svc.Save(&visualization.SaveRequest{
			Name:     "List Test " + string(rune('A'+i)),
			NodeType: &nodeType,
		}, "tester")
		assert.NoError(t, err)
	}

	// List with pagination
	req := &visualization.ListRequest{
		Current: 1,
		Size:    10,
	}
	resp, err := svc.List(req)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, resp.Total, int64(5))
	assert.Len(t, resp.List, 5)
}

func TestVisualizationServiceIntegration_List_WithKeyword(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create items
	nodeType := "panel"
	svc.Save(&visualization.SaveRequest{Name: "Alpha Dashboard", NodeType: &nodeType}, "tester")
	svc.Save(&visualization.SaveRequest{Name: "Beta Dashboard", NodeType: &nodeType}, "tester")
	svc.Save(&visualization.SaveRequest{Name: "Gamma Report", NodeType: &nodeType}, "tester")

	// Search with keyword
	keyword := "Dashboard"
	req := &visualization.ListRequest{
		Keyword: &keyword,
		Current: 1,
		Size:    10,
	}
	resp, err := svc.List(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), resp.Total)
}

func TestVisualizationServiceIntegration_DeleteLogic(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:     "To Delete",
		NodeType: &nodeType,
	}
	id, err := svc.Save(createReq, "tester")
	assert.NoError(t, err)

	// Delete (logic)
	err = svc.DeleteLogic(id, "deleter")
	assert.NoError(t, err)

	// Verify deleted (should not be found)
	_, err = repo.GetByID(id)
	assert.Error(t, err) // Should be record not found
}

func TestVisualizationServiceIntegration_DeleteLogic_NotFound(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Delete non-existent should not error (just updates 0 rows)
	err := svc.DeleteLogic(99999, "deleter")
	assert.NoError(t, err)
}

func TestVisualizationServiceIntegration_Update_AllFields(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create first
	nodeType := "panel"
	createReq := &visualization.SaveRequest{
		Name:            "Original Name",
		NodeType:        &nodeType,
		CanvasStyleData: strPtr("{\"style\":\"original\"}"),
		ComponentData:   strPtr("{\"components\":[]}"),
	}
	id, err := svc.Save(createReq, "creator")
	require.NoError(t, err)

	// Update all fields
	newPID := int64(0)
	newType := "dashboard"
	newStatus := 0
	mobileLayout := true
	updateReq := &visualization.UpdateRequest{
		ID:              id,
		Name:            strPtr("Updated Name"),
		PID:             &newPID,
		Type:            &newType,
		CanvasStyleData: strPtr("{\"style\":\"updated\"}"),
		ComponentData:   strPtr("{\"components\":[1,2,3]}"),
		MobileLayout:    &mobileLayout,
		Status:          &newStatus,
	}
	err = svc.Update(updateReq, "updater")
	require.NoError(t, err)

	// Verify updated
	detail, err := repo.GetByID(id)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", detail.Name)
	require.NotNil(t, detail.Type)
	assert.Equal(t, newType, *detail.Type)
	require.NotNil(t, detail.Status)
	assert.Equal(t, newStatus, *detail.Status)
}

func TestVisualizationServiceIntegration_List_WithPaging(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create multiple items
	nodeType := "panel"
	for i := 0; i < 15; i++ {
		_, err := svc.Save(&visualization.SaveRequest{
			Name:     "Dashboard " + string(rune('A'+i)),
			NodeType: &nodeType,
		}, "tester")
		require.NoError(t, err)
	}

	// Test pagination - page 1
	current := 1
	size := 10
	resp, err := svc.List(&visualization.ListRequest{Current: current, Size: size})
	require.NoError(t, err)
	assert.Equal(t, int64(15), resp.Total)
	assert.Equal(t, 10, len(resp.List))
	assert.Equal(t, 1, resp.Current)
	assert.Equal(t, 10, resp.Size)

	// Test pagination - page 2
	current = 2
	resp, err = svc.List(&visualization.ListRequest{Current: current, Size: size})
	require.NoError(t, err)
	assert.Equal(t, 5, len(resp.List))
}

func TestVisualizationServiceIntegration_List_EdgeCases(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	// Create some items
	nodeType := "panel"
	for i := 0; i < 5; i++ {
		_, err := svc.Save(&visualization.SaveRequest{
			Name:     "Dashboard " + string(rune('A'+i)),
			NodeType: &nodeType,
		}, "tester")
		require.NoError(t, err)
	}

	t.Run("list with zero current", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: 0, Size: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Current) // Should default to 1
	})

	t.Run("list with negative current", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: -1, Size: 10})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Current) // Should default to 1
	})

	t.Run("list with zero size", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: 1, Size: 0})
		require.NoError(t, err)
		assert.Equal(t, 10, resp.Size) // Should default to 10
	})

	t.Run("list with negative size", func(t *testing.T) {
		resp, err := svc.List(&visualization.ListRequest{Current: 1, Size: -5})
		require.NoError(t, err)
		assert.Equal(t, 10, resp.Size) // Should default to 10
	})
}

func TestVisualizationServiceIntegration_InteractiveTree(t *testing.T) {
	cleanupTables(&visualization.DataVisualizationInfo{})

	repo := repository.NewVisualizationRepository(testDB)
	svc := NewVisualizationService(repo)

	folderType := "folder"
	panelType := "panel"
	dashboardType := "dashboard"
	dataVType := "dataV"
	mobileLayout := true

	dashboardFolderID, err := svc.Save(&visualization.SaveRequest{
		Name:     "Dashboard Root",
		NodeType: &folderType,
		Type:     &dashboardType,
	}, "tester")
	require.NoError(t, err)

	dashboardPanelID, err := svc.Save(&visualization.SaveRequest{
		Name:         "Dashboard Child",
		PID:          &dashboardFolderID,
		NodeType:     &panelType,
		Type:         &dashboardType,
		MobileLayout: &mobileLayout,
	}, "tester")
	require.NoError(t, err)

	dataVID, err := svc.Save(&visualization.SaveRequest{
		Name:     "Big Screen A",
		NodeType: &panelType,
		Type:     &dataVType,
	}, "tester")
	require.NoError(t, err)

	published := 1
	err = svc.Update(&visualization.UpdateRequest{ID: dashboardPanelID, Status: &published}, "tester")
	require.NoError(t, err)
	err = svc.Update(&visualization.UpdateRequest{ID: dataVID, Status: &published}, "tester")
	require.NoError(t, err)

	deleted := true
	err = repo.Create(&visualization.DataVisualizationInfo{Name: "Deleted Dashboard", NodeType: &panelType, Type: &dashboardType, DeleteFlag: &deleted})
	require.NoError(t, err)

	dashboardItems, err := svc.InteractiveTree("dashboard")
	require.NoError(t, err)
	assert.Len(t, dashboardItems, 2)
	for _, item := range dashboardItems {
		if assert.NotNil(t, item.Type) {
			assert.Equal(t, "dashboard", *item.Type)
		}
		assert.NotEqual(t, "Deleted Dashboard", item.Name)
	}

	dataVItems, err := svc.InteractiveTree("dataV")
	require.NoError(t, err)
	assert.Len(t, dataVItems, 1)
	assert.Equal(t, "Big Screen A", dataVItems[0].Name)
	if assert.NotNil(t, dataVItems[0].Type) {
		assert.Equal(t, "dataV", *dataVItems[0].Type)
	}
}
