package areamap

import (
	"testing"
)

func TestArea_TableName(t *testing.T) {
	a := Area{}
	if a.TableName() != "area" {
		t.Errorf("Expected table name 'area', got '%s'", a.TableName())
	}
}

func TestArea_Fields(t *testing.T) {
	a := Area{
		ID:    "110000",
		Level: "province",
		Name:  "Beijing",
		Pid:   "100000",
	}

	if a.ID != "110000" {
		t.Errorf("Expected ID '110000', got '%s'", a.ID)
	}
	if a.Level != "province" {
		t.Errorf("Expected Level 'province', got '%s'", a.Level)
	}
	if a.Name != "Beijing" {
		t.Errorf("Expected Name 'Beijing', got '%s'", a.Name)
	}
	if a.Pid != "100000" {
		t.Errorf("Expected Pid '100000', got '%s'", a.Pid)
	}
}

func TestCoreAreaCustom_TableName(t *testing.T) {
	a := CoreAreaCustom{}
	if a.TableName() != "core_area_custom" {
		t.Errorf("Expected table name 'core_area_custom', got '%s'", a.TableName())
	}
}

func TestCoreAreaCustom_Fields(t *testing.T) {
	a := CoreAreaCustom{
		ID:    "custom-1",
		Level: "custom",
		Name:  "Custom Area",
		Pid:   "110000",
	}

	if a.ID != "custom-1" {
		t.Errorf("Expected ID 'custom-1', got '%s'", a.ID)
	}
	if a.Level != "custom" {
		t.Errorf("Expected Level 'custom', got '%s'", a.Level)
	}
}

func TestAreaNode_Fields(t *testing.T) {
	children := []*AreaNode{
		{ID: "110100", Level: "city", Name: "Beijing City", Pid: "110000"},
	}

	node := AreaNode{
		ID:       "110000",
		Level:    "province",
		Name:     "Beijing",
		Pid:      "100000",
		Custom:   false,
		Country:  "China",
		Children: children,
	}

	if node.ID != "110000" {
		t.Errorf("Expected ID '110000', got '%s'", node.ID)
	}
	if node.Custom {
		t.Error("Expected Custom to be false")
	}
	if len(node.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(node.Children))
	}
}

func TestAreaNode_NilChildren(t *testing.T) {
	node := AreaNode{
		ID:    "110000",
		Level: "province",
		Name:  "Beijing",
	}

	if node.Children != nil {
		t.Error("Expected Children to be nil")
	}
}
