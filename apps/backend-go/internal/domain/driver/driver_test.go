package driver

import (
	"testing"
)

func TestDriver_TableName(t *testing.T) {
	d := Driver{}
	if d.TableName() != "de_driver" {
		t.Errorf("Expected table name 'de_driver', got '%s'", d.TableName())
	}
}

func TestDriver_Fields(t *testing.T) {
	typeDesc := "MySQL Driver"
	desc := "Official MySQL JDBC driver"

	d := Driver{
		ID:       1,
		Name:     "MySQL",
		Type:     "mysql",
		TypeDesc: &typeDesc,
		Desc:     &desc,
	}

	if d.ID != 1 {
		t.Errorf("Expected ID 1, got %d", d.ID)
	}
	if d.Name != "MySQL" {
		t.Errorf("Expected Name 'MySQL', got '%s'", d.Name)
	}
	if *d.TypeDesc != "MySQL Driver" {
		t.Errorf("Expected TypeDesc 'MySQL Driver', got '%s'", *d.TypeDesc)
	}
}

func TestDriver_NilFields(t *testing.T) {
	d := Driver{
		ID:   1,
		Name: "Test Driver",
		Type: "test",
	}

	if d.TypeDesc != nil {
		t.Error("Expected TypeDesc to be nil")
	}
	if d.Desc != nil {
		t.Error("Expected Desc to be nil")
	}
}

func TestDriverDTO_Fields(t *testing.T) {
	typeDesc := "PostgreSQL Driver"
	desc := "Official PostgreSQL JDBC driver"

	dto := DriverDTO{
		ID:       1,
		Name:     "PostgreSQL",
		Type:     "postgresql",
		TypeDesc: &typeDesc,
		Desc:     &desc,
	}

	if dto.ID != 1 {
		t.Errorf("Expected ID 1, got %d", dto.ID)
	}
	if dto.Type != "postgresql" {
		t.Errorf("Expected Type 'postgresql', got '%s'", dto.Type)
	}
}

func TestDriverJar_TableName(t *testing.T) {
	dj := DriverJar{}
	if dj.TableName() != "de_driver_jar" {
		t.Errorf("Expected table name 'de_driver_jar', got '%s'", dj.TableName())
	}
}

func TestDriverJar_Fields(t *testing.T) {
	version := "8.0.33"
	createBy := "admin"
	createTime := int64(1700000000)

	dj := DriverJar{
		ID:         1,
		DriverID:   10,
		FileName:   "mysql-connector-java-8.0.33.jar",
		FilePath:   "/drivers/mysql/",
		Version:    &version,
		CreateBy:   &createBy,
		CreateTime: &createTime,
	}

	if dj.ID != 1 {
		t.Errorf("Expected ID 1, got %d", dj.ID)
	}
	if dj.DriverID != 10 {
		t.Errorf("Expected DriverID 10, got %d", dj.DriverID)
	}
	if *dj.Version != "8.0.33" {
		t.Errorf("Expected Version '8.0.33', got '%s'", *dj.Version)
	}
}

func TestDriverJarDTO_Fields(t *testing.T) {
	version := "42.6.0"

	dto := DriverJarDTO{
		ID:       1,
		DriverID: 10,
		FileName: "postgresql-42.6.0.jar",
		FilePath: "/drivers/postgresql/",
		Version:  &version,
	}

	if dto.ID != 1 {
		t.Errorf("Expected ID 1, got %d", dto.ID)
	}
	if *dto.Version != "42.6.0" {
		t.Errorf("Expected Version '42.6.0', got '%s'", *dto.Version)
	}
}
