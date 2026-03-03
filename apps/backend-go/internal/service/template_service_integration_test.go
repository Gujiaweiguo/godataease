//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/template"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestTemplateServiceIntegration_CreateTemplate(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	req := &template.TemplateCreateRequest{
		Name:          "Test Template",
		Pid:           0,
		DvType:        "dashboard",
		NodeType:      "folder",
		Snapshot:      "snapshot-data",
		TemplateType:  "self",
		TemplateStyle: "dark",
		TemplateData:  "{}",
	}

	tpl, err := svc.CreateTemplate(req, "creator1")
	assert.NoError(t, err)
	assert.NotNil(t, tpl)
	assert.Greater(t, tpl.ID, int64(0))
	assert.Equal(t, "Test Template", tpl.Name)
	assert.Equal(t, "creator1", tpl.CreateBy)
	assert.Equal(t, 0, tpl.UseCount)
}

func TestTemplateServiceIntegration_GetTemplate(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	// Create template first
	createReq := &template.TemplateCreateRequest{
		Name:     "Get Test",
		Pid:      0,
		DvType:   "dashboard",
		NodeType: "leaf",
	}
	created, err := svc.CreateTemplate(createReq, "tester")
	assert.NoError(t, err)

	// Get template
	tpl, err := svc.GetTemplate(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Get Test", tpl.Name)
	assert.Equal(t, "tester", tpl.CreateBy)
}

func TestTemplateServiceIntegration_GetTemplate_NotFound(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	_, err := svc.GetTemplate(99999)
	assert.Error(t, err)
}

func TestTemplateServiceIntegration_ListTemplates(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	// Create multiple templates
	for i := 1; i <= 5; i++ {
		_, _ = svc.CreateTemplate(&template.TemplateCreateRequest{
			Name:     "Template",
			Pid:      0,
			DvType:   "dashboard",
			NodeType: "leaf",
		}, "tester")
	}

	// List templates
	req := &template.TemplateListRequest{
		Pid:    "0",
		DvType: "dashboard",
	}
	resp, err := svc.ListTemplates(req)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, resp.Total, int64(5))
}

func TestTemplateServiceIntegration_ListTemplates_Empty(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	req := &template.TemplateListRequest{
		Pid:    "0",
		DvType: "dashboard",
	}
	resp, err := svc.ListTemplates(req)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), resp.Total)
	assert.Len(t, resp.List, 0)
}

func TestTemplateServiceIntegration_UpdateTemplate(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	// Create template
	created, _ := svc.CreateTemplate(&template.TemplateCreateRequest{
		Name:     "Old Name",
		Pid:      0,
		DvType:   "dashboard",
		NodeType: "leaf",
	}, "tester")

	// Update template
	updateReq := &template.TemplateUpdateRequest{
		ID:            created.ID,
		Name:          "New Name",
		Snapshot:      "new-snapshot",
		TemplateStyle: "light",
		TemplateData:  "{\"updated\":true}",
	}
	updated, err := svc.UpdateTemplate(updateReq)
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "new-snapshot", updated.Snapshot)
	assert.Equal(t, "light", updated.TemplateStyle)
}

func TestTemplateServiceIntegration_UpdateTemplate_NotFound(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	updateReq := &template.TemplateUpdateRequest{
		ID:   99999,
		Name: "New Name",
	}
	_, err := svc.UpdateTemplate(updateReq)
	assert.Error(t, err)
}

func TestTemplateServiceIntegration_DeleteTemplate(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	// Create template
	created, _ := svc.CreateTemplate(&template.TemplateCreateRequest{
		Name:     "To Delete",
		Pid:      0,
		DvType:   "dashboard",
		NodeType: "leaf",
	}, "tester")

	// Delete template
	err := svc.DeleteTemplate(created.ID)
	assert.NoError(t, err)

	// Verify deleted
	_, err = svc.GetTemplate(created.ID)
	assert.Error(t, err)
}

func TestTemplateServiceIntegration_DeleteTemplate_NotFound(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	// Delete non-existent template (should not error)
	err := svc.DeleteTemplate(99999)
	assert.NoError(t, err)
}

func TestTemplateServiceIntegration_IncrementUseCount(t *testing.T) {
	cleanupTables(&template.Template{})

	repo := repository.NewTemplateRepository(testDB)
	svc := NewTemplateService(repo)

	// Create template
	created, _ := svc.CreateTemplate(&template.TemplateCreateRequest{
		Name:     "Use Count Test",
		Pid:      0,
		DvType:   "dashboard",
		NodeType: "leaf",
	}, "tester")

	// Increment use count
	err := svc.IncrementUseCount(created.ID)
	assert.NoError(t, err)

	// Verify use count incremented
	tpl, err := svc.GetTemplate(created.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, tpl.UseCount)
}
