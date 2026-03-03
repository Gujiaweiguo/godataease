//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
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
