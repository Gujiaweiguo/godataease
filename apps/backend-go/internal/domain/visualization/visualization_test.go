package visualization

import (
	"testing"
)

func TestDataVisualizationInfo_TableName(t *testing.T) {
	v := DataVisualizationInfo{}
	if v.TableName() != "data_visualization_info" {
		t.Errorf("Expected table name 'data_visualization_info', got '%s'", v.TableName())
	}
}

func TestDataVisualizationInfo_Fields(t *testing.T) {
	pid := int64(0)
	orgID := int64(1)
	level := 1
	nodeType := "folder"
	dsType := "panel"
	canvasStyleData := `{"width":1920}`
	componentData := `[]`
	mobileLayout := false
	status := 1
	sort := 0
	createTime := int64(1700000000)
	createBy := "admin"
	updateTime := int64(1700001000)
	updateBy := "admin"
	deleteFlag := false
	deleteTime := int64(0)
	deleteBy := ""
	version := 1
	contentID := "content-1"
	checkVersion := "v1.0"

	v := DataVisualizationInfo{
		ID:              1,
		Name:            "My Dashboard",
		PID:             &pid,
		OrgID:           &orgID,
		Level:           &level,
		NodeType:        &nodeType,
		Type:            &dsType,
		CanvasStyleData: &canvasStyleData,
		ComponentData:   &componentData,
		MobileLayout:    &mobileLayout,
		Status:          &status,
		Sort:            &sort,
		CreateTime:      &createTime,
		CreateBy:        &createBy,
		UpdateTime:      &updateTime,
		UpdateBy:        &updateBy,
		DeleteFlag:      &deleteFlag,
		DeleteTime:      &deleteTime,
		DeleteBy:        &deleteBy,
		Version:         &version,
		ContentID:       &contentID,
		CheckVersion:    &checkVersion,
	}

	if v.ID != 1 {
		t.Errorf("Expected ID 1, got %d", v.ID)
	}
	if v.Name != "My Dashboard" {
		t.Errorf("Expected Name 'My Dashboard', got '%s'", v.Name)
	}
}

func TestSaveRequest_Fields(t *testing.T) {
	pid := int64(0)
	dsType := "panel"
	nodeType := "leaf"
	canvasStyleData := `{"width":1920}`
	componentData := `[]`
	mobileLayout := false

	req := SaveRequest{
		Name:            "New Dashboard",
		PID:             &pid,
		Type:            &dsType,
		NodeType:        &nodeType,
		CanvasStyleData: &canvasStyleData,
		ComponentData:   &componentData,
		MobileLayout:    &mobileLayout,
	}

	if req.Name != "New Dashboard" {
		t.Errorf("Expected Name 'New Dashboard', got '%s'", req.Name)
	}
}

func TestUpdateRequest_Fields(t *testing.T) {
	name := "Updated Dashboard"
	pid := int64(1)
	dsType := "panel"
	canvasStyleData := `{"width":1920}`
	componentData := `[]`
	mobileLayout := true
	status := 1

	req := UpdateRequest{
		ID:              1,
		Name:            &name,
		PID:             &pid,
		Type:            &dsType,
		CanvasStyleData: &canvasStyleData,
		ComponentData:   &componentData,
		MobileLayout:    &mobileLayout,
		Status:          &status,
	}

	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
	if *req.Name != "Updated Dashboard" {
		t.Errorf("Expected Name 'Updated Dashboard', got '%s'", *req.Name)
	}
}

func TestDetailRequest_Fields(t *testing.T) {
	req := DetailRequest{ID: 1}
	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
}

func TestListRequest_Fields(t *testing.T) {
	keyword := "dashboard"
	dsType := "panel"
	req := ListRequest{
		Keyword: &keyword,
		Type:    &dsType,
		Current: 1,
		Size:    10,
	}

	if *req.Keyword != "dashboard" {
		t.Errorf("Expected Keyword 'dashboard', got '%s'", *req.Keyword)
	}
	if req.Current != 1 {
		t.Errorf("Expected Current 1, got %d", req.Current)
	}
}

func TestListResponse_Fields(t *testing.T) {
	item := &DataVisualizationInfo{ID: 1, Name: "Dashboard 1"}
	resp := ListResponse{
		List:    []*DataVisualizationInfo{item},
		Total:   1,
		Current: 1,
		Size:    10,
	}

	if len(resp.List) != 1 {
		t.Errorf("Expected 1 item, got %d", len(resp.List))
	}
	if resp.Total != 1 {
		t.Errorf("Expected Total 1, got %d", resp.Total)
	}
}
