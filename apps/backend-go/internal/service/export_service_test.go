package service

import (
	"errors"
	"testing"

	"dataease/backend/internal/domain/export"
)

type exportRepoStub struct {
	countResp map[string]int64
	countErr  error
}

func (s *exportRepoStub) Create(task *export.ExportTask) error {
	return nil
}

func (s *exportRepoStub) GetByID(id string) (*export.ExportTask, error) {
	return nil, nil
}

func (s *exportRepoStub) List(page, pageSize int, status string) ([]export.ExportTask, int64, error) {
	return nil, 0, nil
}

func (s *exportRepoStub) UpdateStatus(id string, status string) error {
	return nil
}

func (s *exportRepoStub) Delete(id string) error {
	return nil
}

func (s *exportRepoStub) DeleteBatch(ids []string) error {
	return nil
}

func (s *exportRepoStub) DeleteAllByType(exportFromType string) error {
	return nil
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
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if len(resp) != 0 {
		t.Fatalf("expected empty map on error, got %d entries", len(resp))
	}
}

func TestExportService_ExportTasks_Success(t *testing.T) {
	expected := map[string]int64{"SUCCESS": 3, "FAILED": 1}
	svc := NewExportService(&exportRepoStub{countResp: expected})

	resp := svc.ExportTasks()
	if len(resp) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(resp))
	}
	if resp["SUCCESS"] != 3 || resp["FAILED"] != 1 {
		t.Fatalf("unexpected response: %#v", resp)
	}
}
