package engine

import (
	"testing"
)

func TestEngine_TableName(t *testing.T) {
	e := Engine{}
	if e.TableName() != "core_engine" {
		t.Errorf("Expected table name 'core_engine', got '%s'", e.TableName())
	}
}

func TestEngine_Fields(t *testing.T) {
	configuration := `{"host":"localhost"}`
	createBy := "admin"
	createTime := int64(1700000000)

	e := Engine{
		ID:            1,
		Name:          "MySQL Engine",
		Type:          "mysql",
		Configuration: &configuration,
		CreateBy:      &createBy,
		CreateTime:    &createTime,
	}

	if e.ID != 1 {
		t.Errorf("Expected ID 1, got %d", e.ID)
	}
	if e.Name != "MySQL Engine" {
		t.Errorf("Expected Name 'MySQL Engine', got '%s'", e.Name)
	}
	if e.Type != "mysql" {
		t.Errorf("Expected Type 'mysql', got '%s'", e.Type)
	}
}

func TestEngine_NilFields(t *testing.T) {
	e := Engine{
		ID:   1,
		Name: "Test Engine",
		Type: "test",
	}

	if e.Configuration != nil {
		t.Error("Expected Configuration to be nil")
	}
	if e.CreateBy != nil {
		t.Error("Expected CreateBy to be nil")
	}
}

func TestEngineDTO_Fields(t *testing.T) {
	configuration := `{"host":"localhost"}`
	dto := EngineDTO{
		ID:            1,
		Name:          "MySQL Engine",
		Type:          "mysql",
		Configuration: &configuration,
	}

	if dto.ID != 1 {
		t.Errorf("Expected ID 1, got %d", dto.ID)
	}
	if *dto.Configuration != `{"host":"localhost"}` {
		t.Errorf("Unexpected Configuration value")
	}
}

func TestValidateRequest_Fields(t *testing.T) {
	id := int64(1)
	engineType := "mysql"
	configuration := `{"host":"localhost"}`

	req := ValidateRequest{
		ID:            &id,
		Type:          &engineType,
		Configuration: &configuration,
	}

	if *req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", *req.ID)
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
