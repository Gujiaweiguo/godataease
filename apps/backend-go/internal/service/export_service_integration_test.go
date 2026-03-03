//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/export"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestExportServiceIntegration_ExportTasks(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	resp := svc.ExportTasks()
	assert.NotNil(t, resp)
}

func TestExportServiceIntegration_Pager(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	req := &export.PagerRequest{
		GoPage:   1,
		PageSize: 10,
	}
	resp := svc.Pager(req)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.PageNum)
	assert.Equal(t, 10, resp.PageSize)
}

func TestExportServiceIntegration_Delete(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	// Delete non-existent task should not error
	err := svc.Delete("non-existent-id")
	assert.NoError(t, err)
}

func TestExportServiceIntegration_DeleteBatch(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	err := svc.DeleteBatch([]string{"id1", "id2"})
	assert.NoError(t, err)
}

func TestExportServiceIntegration_DeleteAll(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	err := svc.DeleteAll("panel")
	assert.NoError(t, err)
}

func TestExportServiceIntegration_GetByID(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	// Get non-existent task should return error
	_, err := svc.GetByID("non-existent-id")
	assert.Error(t, err)
}

func TestExportServiceIntegration_CheckAccess(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	// Test with nil task
	err := svc.CheckAccess(nil, 1, false)
	assert.Error(t, err)
	assert.Equal(t, ErrNotFound, err)

	// Test with non-owner user
	task := &export.ExportTask{
		ID:     "test-id",
		UserID: 100,
	}
	err = svc.CheckAccess(task, 200, false)
	assert.Error(t, err)
	assert.Equal(t, ErrUnauthorized, err)

	// Test with owner user
	err = svc.CheckAccess(task, 100, false)
	assert.NoError(t, err)

	// Test with admin user
	err = svc.CheckAccess(task, 200, true)
	assert.NoError(t, err)
}

func TestExportServiceIntegration_Retry(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	err := svc.Retry("test-id")
	assert.NoError(t, err)
}

func TestExportServiceIntegration_ExportLimit(t *testing.T) {
	repo := repository.NewExportRepository(testDB)
	svc := NewExportService(repo)

	limit := svc.ExportLimit()
	assert.NotNil(t, limit)
	assert.Equal(t, "10000", limit.Limit)
}
