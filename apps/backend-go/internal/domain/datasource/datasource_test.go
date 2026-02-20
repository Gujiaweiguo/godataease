package datasource

import (
	"testing"
)

func TestStatusConstants(t *testing.T) {
	if StatusSuccess != "Success" {
		t.Errorf("Expected StatusSuccess 'Success', got '%s'", StatusSuccess)
	}
	if StatusError != "Error" {
		t.Errorf("Expected StatusError 'Error', got '%s'", StatusError)
	}
}

func TestTypeConstants(t *testing.T) {
	if TypeFolder != "folder" {
		t.Errorf("Expected TypeFolder 'folder', got '%s'", TypeFolder)
	}
	if TypeExcel != "Excel" {
		t.Errorf("Expected TypeExcel 'Excel', got '%s'", TypeExcel)
	}
}

func TestCoreDatasource_TableName(t *testing.T) {
	ds := CoreDatasource{}
	if ds.TableName() != "core_datasource" {
		t.Errorf("Expected table name 'core_datasource', got '%s'", ds.TableName())
	}
}

func TestCoreDatasource_Fields(t *testing.T) {
	pid := int64(0)
	description := "MySQL production database"
	editType := "edit"
	configuration := `{"host":"localhost","port":3306}`
	status := "Success"
	qrtzInstance := "instance-1"
	taskStatus := "Running"
	enableDataFill := true
	createTime := int64(1700000000)
	updateTime := int64(1700001000)
	updateBy := int64(1)
	createBy := "admin"
	delFlag := 0

	ds := CoreDatasource{
		ID:             1,
		PID:            &pid,
		Name:           "MySQL DB",
		Description:    &description,
		Type:           "mysql",
		EditType:       &editType,
		Configuration:  &configuration,
		Status:         &status,
		QrtzInstance:   &qrtzInstance,
		TaskStatus:     &taskStatus,
		EnableDataFill: &enableDataFill,
		CreateTime:     &createTime,
		UpdateTime:     &updateTime,
		UpdateBy:       &updateBy,
		CreateBy:       &createBy,
		DelFlag:        &delFlag,
	}

	if ds.ID != 1 {
		t.Errorf("Expected ID 1, got %d", ds.ID)
	}
	if ds.Name != "MySQL DB" {
		t.Errorf("Expected Name 'MySQL DB', got '%s'", ds.Name)
	}
	if ds.Type != "mysql" {
		t.Errorf("Expected Type 'mysql', got '%s'", ds.Type)
	}
	if !*ds.EnableDataFill {
		t.Error("Expected EnableDataFill to be true")
	}
}

func TestCoreDatasource_NilFields(t *testing.T) {
	ds := CoreDatasource{
		ID:   1,
		Name: "Test",
		Type: "mysql",
	}

	if ds.PID != nil {
		t.Error("Expected PID to be nil")
	}
	if ds.Description != nil {
		t.Error("Expected Description to be nil")
	}
	if ds.Configuration != nil {
		t.Error("Expected Configuration to be nil")
	}
}

func TestListRequest_Fields(t *testing.T) {
	keyword := "mysql"
	req := ListRequest{
		Keyword: &keyword,
		Current: 1,
		Size:    10,
	}

	if *req.Keyword != "mysql" {
		t.Errorf("Expected Keyword 'mysql', got '%s'", *req.Keyword)
	}
	if req.Current != 1 {
		t.Errorf("Expected Current 1, got %d", req.Current)
	}
	if req.Size != 10 {
		t.Errorf("Expected Size 10, got %d", req.Size)
	}
}

func TestListRequest_NilKeyword(t *testing.T) {
	req := ListRequest{
		Current: 1,
		Size:    20,
	}

	if req.Keyword != nil {
		t.Error("Expected Keyword to be nil")
	}
}

func TestListResponse_Fields(t *testing.T) {
	ds1 := &CoreDatasource{ID: 1, Name: "DB1", Type: "mysql"}
	ds2 := &CoreDatasource{ID: 2, Name: "DB2", Type: "postgresql"}

	resp := ListResponse{
		List:    []*CoreDatasource{ds1, ds2},
		Total:   2,
		Current: 1,
		Size:    10,
	}

	if len(resp.List) != 2 {
		t.Errorf("Expected 2 items, got %d", len(resp.List))
	}
	if resp.Total != 2 {
		t.Errorf("Expected Total 2, got %d", resp.Total)
	}
}

func TestValidateRequest_Fields(t *testing.T) {
	datasourceID := int64(1)
	dsType := "mysql"
	configuration := `{"host":"localhost"}`

	req := ValidateRequest{
		DatasourceID:  &datasourceID,
		Type:          &dsType,
		Configuration: &configuration,
	}

	if *req.DatasourceID != 1 {
		t.Errorf("Expected DatasourceID 1, got %d", *req.DatasourceID)
	}
	if *req.Type != "mysql" {
		t.Errorf("Expected Type 'mysql', got '%s'", *req.Type)
	}
}

func TestValidateResponse_Fields(t *testing.T) {
	resp := ValidateResponse{
		Status:  "Success",
		Message: "Connection successful",
	}

	if resp.Status != "Success" {
		t.Errorf("Expected Status 'Success', got '%s'", resp.Status)
	}
	if resp.Message != "Connection successful" {
		t.Errorf("Expected Message 'Connection successful', got '%s'", resp.Message)
	}
}

func TestTableRequest_Fields(t *testing.T) {
	req := TableRequest{
		DatasourceID: 1,
		TableName:    "users",
		Limit:        100,
	}

	if req.DatasourceID != 1 {
		t.Errorf("Expected DatasourceID 1, got %d", req.DatasourceID)
	}
	if req.TableName != "users" {
		t.Errorf("Expected TableName 'users', got '%s'", req.TableName)
	}
}

func TestTableInfo_Fields(t *testing.T) {
	info := TableInfo{
		ID:           1,
		DatasourceID: 1,
		Name:         "Users Table",
		TableName:    "users",
		Type:         "table",
		Status:       "Success",
		LastUpdate:   1700000000,
	}

	if info.ID != 1 {
		t.Errorf("Expected ID 1, got %d", info.ID)
	}
	if info.TableName != "users" {
		t.Errorf("Expected TableName 'users', got '%s'", info.TableName)
	}
}

func TestTableField_Fields(t *testing.T) {
	field := TableField{
		OriginName: "user_id",
		Name:       "User ID",
		Type:       "BIGINT",
		DeType:     2,
	}

	if field.OriginName != "user_id" {
		t.Errorf("Expected OriginName 'user_id', got '%s'", field.OriginName)
	}
	if field.DeType != 2 {
		t.Errorf("Expected DeType 2, got %d", field.DeType)
	}
}

func TestPreviewDataResponse_Fields(t *testing.T) {
	resp := PreviewDataResponse{
		Fields: []TableField{
			{OriginName: "id", Name: "ID", Type: "BIGINT", DeType: 2},
		},
		Data: []map[string]interface{}{
			{"id": 1},
			{"id": 2},
		},
		Total: 2,
	}

	if len(resp.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(resp.Fields))
	}
	if len(resp.Data) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(resp.Data))
	}
	if resp.Total != 2 {
		t.Errorf("Expected Total 2, got %d", resp.Total)
	}
}

func TestWriteRequest_Fields(t *testing.T) {
	pid := int64(0)
	description := "Test database"
	editType := "edit"
	configuration := `{"host":"localhost"}`
	enableDataFill := false

	req := WriteRequest{
		ID:             0,
		PID:            &pid,
		Name:           "New DB",
		Description:    &description,
		Type:           "mysql",
		NodeType:       "datasource",
		EditType:       &editType,
		Configuration:  &configuration,
		EnableDataFill: &enableDataFill,
	}

	if req.Name != "New DB" {
		t.Errorf("Expected Name 'New DB', got '%s'", req.Name)
	}
	if req.Type != "mysql" {
		t.Errorf("Expected Type 'mysql', got '%s'", req.Type)
	}
	if req.NodeType != "datasource" {
		t.Errorf("Expected NodeType 'datasource', got '%s'", req.NodeType)
	}
}

func TestConnectionConfig_Fields(t *testing.T) {
	cfg := ConnectionConfig{
		Host:     "localhost",
		Port:     3306,
		JDBCUrl:  "jdbc:mysql://localhost:3306/test",
		Database: "test",
		Schema:   "public",
	}

	if cfg.Host != "localhost" {
		t.Errorf("Expected Host 'localhost', got '%s'", cfg.Host)
	}
	if cfg.Port != 3306 {
		t.Errorf("Expected Port 3306, got %d", cfg.Port)
	}
	if cfg.Database != "test" {
		t.Errorf("Expected Database 'test', got '%s'", cfg.Database)
	}
}
