package service

import (
	"errors"

	"dataease/backend/internal/domain/export"
	"dataease/backend/internal/repository"
)

// Export status transitions:
// PENDING -> RUNNING -> SUCCESS
//
//	\-> FAILED
//
// Retry: FAILED -> PENDING (allows re-execution)
var (
	ErrUnauthorized = errors.New("无权访问该导出任务")
	ErrNotFound     = errors.New("导出任务不存在")
)

type ExportService struct {
	repo *repository.ExportRepository
}

func NewExportService(repo *repository.ExportRepository) *ExportService {
	return &ExportService{repo: repo}
}

func (s *ExportService) ExportTasks() export.ExportTasksResponse {
	counts, err := s.repo.CountByStatus()
	if err != nil {
		return make(export.ExportTasksResponse)
	}
	return counts
}

func (s *ExportService) Pager(req *export.PagerRequest) *export.PagerResponse {
	page := req.GoPage
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	list, total, err := s.repo.List(page, pageSize, req.Status)
	if err != nil {
		return &export.PagerResponse{
			List:     []export.ExportTask{},
			Total:    0,
			PageNum:  page,
			PageSize: pageSize,
		}
	}

	return &export.PagerResponse{
		List:     list,
		Total:    total,
		PageNum:  page,
		PageSize: pageSize,
	}
}

func (s *ExportService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *ExportService) DeleteBatch(ids []string) error {
	return s.repo.DeleteBatch(ids)
}

func (s *ExportService) DeleteAll(exportFromType string) error {
	return s.repo.DeleteAllByType(exportFromType)
}

func (s *ExportService) GetByID(id string) (*export.ExportTask, error) {
	return s.repo.GetByID(id)
}

func (s *ExportService) CheckAccess(task *export.ExportTask, userID int64, isAdmin bool) error {
	if task == nil {
		return ErrNotFound
	}
	if task.UserID != userID && !isAdmin {
		return ErrUnauthorized
	}
	return nil
}

func (s *ExportService) Retry(id string) error {
	return s.repo.UpdateStatus(id, "PENDING")
}

func (s *ExportService) ExportLimit() *export.ExportLimitResponse {
	return &export.ExportLimitResponse{Limit: "10000"}
}
