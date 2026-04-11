package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/export"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type exportRepoStub struct {
	createTask        *export.ExportTask
	createErr         error
	countResp         map[string]int64
	countErr          error
	listResp          []export.ExportTask
	listTotal         int64
	listErr           error
	listPage          int
	listPageSize      int
	listStatus        string
	getResp           *export.ExportTask
	getErr            error
	getID             string
	updateStatusID    string
	updateStatusValue string
	updateStatusErr   error
	deleteID          string
	deleteErr         error
	deleteBatchIDs    []string
	deleteBatchErr    error
	deleteAllType     string
	deleteAllErr      error
}

func (s *exportRepoStub) Create(task *export.ExportTask) error {
	s.createTask = task
	return s.createErr
}

func (s *exportRepoStub) GetByID(id string) (*export.ExportTask, error) {
	s.getID = id
	return s.getResp, s.getErr
}

func (s *exportRepoStub) List(page, pageSize int, status string) ([]export.ExportTask, int64, error) {
	s.listPage = page
	s.listPageSize = pageSize
	s.listStatus = status
	return s.listResp, s.listTotal, s.listErr
}

func (s *exportRepoStub) UpdateStatus(id string, status string) error {
	s.updateStatusID = id
	s.updateStatusValue = status
	return s.updateStatusErr
}

func (s *exportRepoStub) Delete(id string) error {
	s.deleteID = id
	return s.deleteErr
}

func (s *exportRepoStub) DeleteBatch(ids []string) error {
	s.deleteBatchIDs = ids
	return s.deleteBatchErr
}

func (s *exportRepoStub) DeleteAllByType(exportFromType string) error {
	s.deleteAllType = exportFromType
	return s.deleteAllErr
}

func (s *exportRepoStub) CountByStatus() (map[string]int64, error) {
	if s.countErr != nil {
		return nil, s.countErr
	}
	return s.countResp, nil
}

func TestExportService_ExportTasks_Error(t *testing.T) {
	svc := NewExportService(&exportRepoStub{countErr: errors.New("count failed")})

	resp := svc.ExportTasks()
	require.NotNil(t, resp)
	assert.Len(t, resp, 0)
}

func TestExportService_ExportTasks_Success(t *testing.T) {
	expected := map[string]int64{"SUCCESS": 3, "FAILED": 1}
	svc := NewExportService(&exportRepoStub{countResp: expected})

	resp := svc.ExportTasks()
	assert.Len(t, resp, len(expected))
	assert.Equal(t, int64(3), resp["SUCCESS"])
	assert.Equal(t, int64(1), resp["FAILED"])
}

func TestExportService_Pager(t *testing.T) {
	t.Run("defaults invalid pagination", func(t *testing.T) {
		repo := &exportRepoStub{listResp: []export.ExportTask{}}
		svc := NewExportService(repo)

		resp := svc.Pager(&export.PagerRequest{GoPage: 0, PageSize: 0, Status: "FAILED"})
		require.NotNil(t, resp)
		assert.Equal(t, 1, resp.PageNum)
		assert.Equal(t, 10, resp.PageSize)
		assert.Equal(t, 1, repo.listPage)
		assert.Equal(t, 10, repo.listPageSize)
		assert.Equal(t, "FAILED", repo.listStatus)
	})

	t.Run("repo error returns empty page", func(t *testing.T) {
		repo := &exportRepoStub{listErr: errors.New("list failed")}
		svc := NewExportService(repo)

		resp := svc.Pager(&export.PagerRequest{GoPage: -1, PageSize: -2})
		require.NotNil(t, resp)
		assert.Empty(t, resp.List)
		assert.Zero(t, resp.Total)
		assert.Equal(t, 1, resp.PageNum)
		assert.Equal(t, 10, resp.PageSize)
	})

	t.Run("success returns repo data", func(t *testing.T) {
		repo := &exportRepoStub{listResp: []export.ExportTask{{ID: "task-1", UserID: 7}}, listTotal: 1}
		svc := NewExportService(repo)

		resp := svc.Pager(&export.PagerRequest{GoPage: 2, PageSize: 20, Status: "SUCCESS"})
		require.NotNil(t, resp)
		assert.Len(t, resp.List, 1)
		assert.Equal(t, "task-1", resp.List[0].ID)
		assert.Equal(t, int64(1), resp.Total)
		assert.Equal(t, 2, resp.PageNum)
		assert.Equal(t, 20, resp.PageSize)
	})
}

func TestExportService_CheckAccess(t *testing.T) {
	t.Run("nil task returns not found", func(t *testing.T) {
		svc := NewExportService(&exportRepoStub{})
		assert.ErrorIs(t, svc.CheckAccess(nil, 1, false), ErrNotFound)
	})

	t.Run("owner allowed", func(t *testing.T) {
		svc := NewExportService(&exportRepoStub{})
		require.NoError(t, svc.CheckAccess(&export.ExportTask{ID: "task-1", UserID: 100}, 100, false))
	})

	t.Run("admin allowed", func(t *testing.T) {
		svc := NewExportService(&exportRepoStub{})
		require.NoError(t, svc.CheckAccess(&export.ExportTask{ID: "task-1", UserID: 100}, 200, true))
	})

	t.Run("other user denied", func(t *testing.T) {
		svc := NewExportService(&exportRepoStub{})
		assert.ErrorIs(t, svc.CheckAccess(&export.ExportTask{ID: "task-1", UserID: 100}, 200, false), ErrUnauthorized)
	})
}

func TestExportService_Retry(t *testing.T) {
	repo := &exportRepoStub{}
	svc := NewExportService(repo)

	require.NoError(t, svc.Retry("task-9"))
	assert.Equal(t, "task-9", repo.updateStatusID)
	assert.Equal(t, "PENDING", repo.updateStatusValue)
}

func TestExportService_DeleteDelegations(t *testing.T) {
	t.Run("delete delegates", func(t *testing.T) {
		repo := &exportRepoStub{}
		svc := NewExportService(repo)

		require.NoError(t, svc.Delete("task-1"))
		assert.Equal(t, "task-1", repo.deleteID)
	})

	t.Run("delete batch delegates", func(t *testing.T) {
		repo := &exportRepoStub{}
		svc := NewExportService(repo)
		ids := []string{"a", "b"}

		require.NoError(t, svc.DeleteBatch(ids))
		assert.Equal(t, ids, repo.deleteBatchIDs)
	})

	t.Run("delete all delegates", func(t *testing.T) {
		repo := &exportRepoStub{}
		svc := NewExportService(repo)

		require.NoError(t, svc.DeleteAll("panel"))
		assert.Equal(t, "panel", repo.deleteAllType)
	})
}

func TestExportService_GetByID(t *testing.T) {
	repo := &exportRepoStub{getResp: &export.ExportTask{ID: "task-2", UserID: 8}}
	svc := NewExportService(repo)

	resp, err := svc.GetByID("task-2")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "task-2", repo.getID)
	assert.Equal(t, int64(8), resp.UserID)
}

func TestExportService_ExportLimit(t *testing.T) {
	svc := NewExportService(&exportRepoStub{})

	resp := svc.ExportLimit()
	require.NotNil(t, resp)
	assert.Equal(t, "10000", resp.Limit)
}
