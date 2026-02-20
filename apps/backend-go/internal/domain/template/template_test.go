package template

import (
	"testing"
	"time"
)

func TestTemplate_Fields(t *testing.T) {
	now := time.Now()
	tpl := Template{
		ID:            1,
		Name:          "test-template",
		Pid:           0,
		Level:         1,
		DvType:        "dashboard",
		NodeType:      "folder",
		CreateBy:      "admin",
		CreateTime:    &now,
		Snapshot:      "base64snapshot",
		TemplateType:  "system",
		TemplateStyle: "{}",
		TemplateData:  "{}",
		DynamicData:   "{}",
		AppData:       "{}",
		UseCount:      10,
		Version:       1,
	}

	if tpl.ID != 1 {
		t.Errorf("Expected ID 1, got %d", tpl.ID)
	}
	if tpl.Name != "test-template" {
		t.Errorf("Expected Name 'test-template', got '%s'", tpl.Name)
	}
	if tpl.DvType != "dashboard" {
		t.Errorf("Expected DvType 'dashboard', got '%s'", tpl.DvType)
	}
	if tpl.UseCount != 10 {
		t.Errorf("Expected UseCount 10, got %d", tpl.UseCount)
	}
}

func TestTemplateListRequest_Fields(t *testing.T) {
	req := TemplateListRequest{
		Pid:        "0",
		DvType:     "dashboard",
		WithBlobs:  "true",
		Keyword:    "test",
		CategoryID: "1",
		Sort:       "name",
	}

	if req.Pid != "0" {
		t.Errorf("Expected Pid '0', got '%s'", req.Pid)
	}
	if req.DvType != "dashboard" {
		t.Errorf("Expected DvType 'dashboard', got '%s'", req.DvType)
	}
	if req.Keyword != "test" {
		t.Errorf("Expected Keyword 'test', got '%s'", req.Keyword)
	}
}

func TestTemplateListResponse_Fields(t *testing.T) {
	resp := TemplateListResponse{
		List:  []Template{{ID: 1, Name: "test"}},
		Total: 1,
	}

	if len(resp.List) != 1 {
		t.Errorf("Expected List length 1, got %d", len(resp.List))
	}
	if resp.Total != 1 {
		t.Errorf("Expected Total 1, got %d", resp.Total)
	}
}

func TestTemplateCreateRequest_Fields(t *testing.T) {
	req := TemplateCreateRequest{
		Name:          "new-template",
		Pid:           0,
		DvType:        "dashboard",
		NodeType:      "leaf",
		Snapshot:      "snapshot",
		TemplateType:  "user",
		TemplateStyle: "{}",
		TemplateData:  "{}",
		DynamicData:   "{}",
		AppData:       "{}",
	}

	if req.Name != "new-template" {
		t.Errorf("Expected Name 'new-template', got '%s'", req.Name)
	}
	if req.DvType != "dashboard" {
		t.Errorf("Expected DvType 'dashboard', got '%s'", req.DvType)
	}
}

func TestTemplateUpdateRequest_Fields(t *testing.T) {
	req := TemplateUpdateRequest{
		ID:            1,
		Name:          "updated-template",
		Snapshot:      "new-snapshot",
		TemplateStyle: "{}",
		TemplateData:  "{}",
		DynamicData:   "{}",
		AppData:       "{}",
	}

	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
	if req.Name != "updated-template" {
		t.Errorf("Expected Name 'updated-template', got '%s'", req.Name)
	}
}

func TestTemplateDeleteRequest_Fields(t *testing.T) {
	req := TemplateDeleteRequest{
		ID:         1,
		CategoryID: 2,
	}

	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
	if req.CategoryID != 2 {
		t.Errorf("Expected CategoryID 2, got %d", req.CategoryID)
	}
}

func TestTemplate_CreateTimeNil(t *testing.T) {
	tpl := Template{
		ID:         1,
		Name:       "test",
		CreateTime: nil,
	}

	if tpl.CreateTime != nil {
		t.Error("Expected CreateTime to be nil")
	}
}
