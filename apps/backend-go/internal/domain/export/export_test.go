package export

import (
	"testing"
)

func TestExportTask_Fields(t *testing.T) {
	task := ExportTask{
		ID:                "task-1",
		UserID:            1,
		FileName:          "report.xlsx",
		FileSize:          1024.5,
		FileSizeUnit:      "KB",
		ExportFrom:        100,
		ExportStatus:      "SUCCESS",
		Msg:               "Export completed",
		ExportFromType:    "dashboard",
		ExportTime:        1700000000,
		ExportProgress:    "100%",
		ExportMachineName: "node-1",
		ExportFromName:    "Sales Dashboard",
		OrgName:           "Default Org",
	}

	if task.ID != "task-1" {
		t.Errorf("Expected ID 'task-1', got '%s'", task.ID)
	}
	if task.FileName != "report.xlsx" {
		t.Errorf("Expected FileName 'report.xlsx', got '%s'", task.FileName)
	}
	if task.ExportStatus != "SUCCESS" {
		t.Errorf("Expected ExportStatus 'SUCCESS', got '%s'", task.ExportStatus)
	}
}

func TestPagerRequest_Fields(t *testing.T) {
	req := PagerRequest{
		GoPage:   1,
		PageSize: 10,
		Status:   "SUCCESS",
	}

	if req.GoPage != 1 {
		t.Errorf("Expected GoPage 1, got %d", req.GoPage)
	}
	if req.PageSize != 10 {
		t.Errorf("Expected PageSize 10, got %d", req.PageSize)
	}
}

func TestPagerResponse_Fields(t *testing.T) {
	tasks := []ExportTask{
		{ID: "task-1", FileName: "file1.xlsx"},
		{ID: "task-2", FileName: "file2.xlsx"},
	}

	resp := PagerResponse{
		List:     tasks,
		Total:    2,
		PageNum:  1,
		PageSize: 10,
	}

	if len(resp.List) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(resp.List))
	}
	if resp.Total != 2 {
		t.Errorf("Expected Total 2, got %d", resp.Total)
	}
}

func TestDeleteRequest_Fields(t *testing.T) {
	req := DeleteRequest{
		IDs: []string{"task-1", "task-2"},
	}

	if len(req.IDs) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(req.IDs))
	}
}

func TestDeleteAllRequest_Fields(t *testing.T) {
	req := DeleteAllRequest{Type: "SUCCESS"}
	if req.Type != "SUCCESS" {
		t.Errorf("Expected Type 'SUCCESS', got '%s'", req.Type)
	}
}

func TestDownloadRequest_Fields(t *testing.T) {
	req := DownloadRequest{ID: "task-1"}
	if req.ID != "task-1" {
		t.Errorf("Expected ID 'task-1', got '%s'", req.ID)
	}
}

func TestDownloadResponse_Fields(t *testing.T) {
	resp := DownloadResponse{URL: "https://example.com/download/task-1"}
	if resp.URL != "https://example.com/download/task-1" {
		t.Errorf("Expected URL 'https://example.com/download/task-1', got '%s'", resp.URL)
	}
}

func TestRetryRequest_Fields(t *testing.T) {
	req := RetryRequest{ID: "task-1"}
	if req.ID != "task-1" {
		t.Errorf("Expected ID 'task-1', got '%s'", req.ID)
	}
}

func TestExportLimitResponse_Fields(t *testing.T) {
	resp := ExportLimitResponse{Limit: "100"}
	if resp.Limit != "100" {
		t.Errorf("Expected Limit '100', got '%s'", resp.Limit)
	}
}
