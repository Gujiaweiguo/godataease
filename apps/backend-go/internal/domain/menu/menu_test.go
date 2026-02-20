package menu

import (
	"testing"
)

func TestCoreMenu_TableName(t *testing.T) {
	menu := CoreMenu{}
	if menu.TableName() != "core_menu" {
		t.Errorf("Expected table name 'core_menu', got '%s'", menu.TableName())
	}
}

func TestCoreMenu_Fields(t *testing.T) {
	menu := CoreMenu{
		ID:        1,
		Pid:       0,
		Type:      1,
		Name:      "Dashboard",
		Component: "views/dashboard/index",
		MenuSort:  1,
		Icon:      "dashboard",
		Path:      "/dashboard",
		Hidden:    false,
		InLayout:  true,
		Auth:      true,
	}

	if menu.ID != 1 {
		t.Errorf("Expected ID 1, got %d", menu.ID)
	}
	if menu.Pid != 0 {
		t.Errorf("Expected Pid 0, got %d", menu.Pid)
	}
	if menu.Name != "Dashboard" {
		t.Errorf("Expected Name 'Dashboard', got '%s'", menu.Name)
	}
	if menu.Component != "views/dashboard/index" {
		t.Errorf("Expected Component 'views/dashboard/index', got '%s'", menu.Component)
	}
	if menu.MenuSort != 1 {
		t.Errorf("Expected MenuSort 1, got %d", menu.MenuSort)
	}
	if menu.Hidden {
		t.Error("Expected Hidden to be false")
	}
	if !menu.InLayout {
		t.Error("Expected InLayout to be true")
	}
	if !menu.Auth {
		t.Error("Expected Auth to be true")
	}
}

func TestCoreMenu_DefaultValues(t *testing.T) {
	menu := CoreMenu{}

	if menu.Hidden != false {
		t.Error("Expected Hidden default to be false")
	}
	if menu.InLayout != false {
		t.Error("Expected InLayout default to be false")
	}
	if menu.Auth != false {
		t.Error("Expected Auth default to be false")
	}
}

func TestMenuMeta_Fields(t *testing.T) {
	meta := MenuMeta{
		Title: "Dashboard",
		Icon:  "dashboard",
	}

	if meta.Title != "Dashboard" {
		t.Errorf("Expected Title 'Dashboard', got '%s'", meta.Title)
	}
	if meta.Icon != "dashboard" {
		t.Errorf("Expected Icon 'dashboard', got '%s'", meta.Icon)
	}
}

func TestMenuVO_Fields(t *testing.T) {
	meta := &MenuMeta{Title: "Dashboard", Icon: "dashboard"}
	child := &MenuVO{Path: "/child", Name: "Child"}

	vo := MenuVO{
		ID:        1,
		Path:      "/dashboard",
		Component: "views/dashboard/index",
		Hidden:    false,
		IsPlugin:  false,
		Name:      "Dashboard",
		InLayout:  true,
		Redirect:  "/dashboard/home",
		Meta:      meta,
		Children:  []*MenuVO{child},
	}

	if vo.ID != 1 {
		t.Errorf("Expected ID 1, got %d", vo.ID)
	}
	if vo.Path != "/dashboard" {
		t.Errorf("Expected Path '/dashboard', got '%s'", vo.Path)
	}
	if vo.Redirect != "/dashboard/home" {
		t.Errorf("Expected Redirect '/dashboard/home', got '%s'", vo.Redirect)
	}
	if vo.Meta == nil {
		t.Fatal("Expected Meta to be non-nil")
	}
	if vo.Meta.Title != "Dashboard" {
		t.Errorf("Expected Meta.Title 'Dashboard', got '%s'", vo.Meta.Title)
	}
	if len(vo.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(vo.Children))
	}
}

func TestMenuVO_NilFields(t *testing.T) {
	vo := MenuVO{
		Path: "/dashboard",
		Name: "Dashboard",
	}

	if vo.Meta != nil {
		t.Error("Expected Meta to be nil")
	}
	if vo.Children != nil {
		t.Error("Expected Children to be nil")
	}
	if vo.Redirect != "" {
		t.Errorf("Expected Redirect to be empty, got '%s'", vo.Redirect)
	}
}

func TestMenuVO_EmptyChildren(t *testing.T) {
	vo := MenuVO{
		Path:     "/dashboard",
		Name:     "Dashboard",
		Children: []*MenuVO{},
	}

	if len(vo.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(vo.Children))
	}
}

func TestMenuVO_NestedChildren(t *testing.T) {
	grandchild := &MenuVO{Path: "/grandchild", Name: "Grandchild"}
	child := &MenuVO{
		Path:     "/child",
		Name:     "Child",
		Children: []*MenuVO{grandchild},
	}
	parent := &MenuVO{
		Path:     "/parent",
		Name:     "Parent",
		Children: []*MenuVO{child},
	}

	if len(parent.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(parent.Children))
	}
	if len(parent.Children[0].Children) != 1 {
		t.Fatalf("Expected 1 grandchild, got %d", len(parent.Children[0].Children))
	}
	if parent.Children[0].Children[0].Name != "Grandchild" {
		t.Errorf("Expected grandchild name 'Grandchild', got '%s'", parent.Children[0].Children[0].Name)
	}
}
