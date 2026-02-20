//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/menu"
)

func TestMenuRepository_GetAll(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewMenuRepository(testDB)
	cleanupTables("core_menu")

	m1 := &menu.CoreMenu{
		Pid:       0,
		Type:      0,
		Name:      "Dashboard",
		Component: "views/dashboard/index",
		MenuSort:  1,
		Icon:      "dashboard",
		Path:      "/dashboard",
		Hidden:    false,
		InLayout:  true,
		Auth:      true,
	}

	m2 := &menu.CoreMenu{
		Pid:       0,
		Type:      0,
		Name:      "Datasource",
		Component: "views/datasource/index",
		MenuSort:  2,
		Icon:      "datasource",
		Path:      "/datasource",
		Hidden:    false,
		InLayout:  true,
		Auth:      true,
	}

	if err := testDB.Create(m1).Error; err != nil {
		t.Fatalf("Failed to create menu 1: %v", err)
	}
	if err := testDB.Create(m2).Error; err != nil {
		t.Fatalf("Failed to create menu 2: %v", err)
	}

	menus, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(menus) != 2 {
		t.Errorf("Expected 2 menus, got %d", len(menus))
	}

	if menus[0].MenuSort > menus[1].MenuSort {
		t.Error("Expected menus to be sorted by menu_sort ASC")
	}
}

func TestMenuRepository_GetAll_Empty(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewMenuRepository(testDB)
	cleanupTables("core_menu")

	menus, err := repo.GetAll()
	if err != nil {
		t.Fatalf("GetAll failed: %v", err)
	}

	if len(menus) != 0 {
		t.Errorf("Expected 0 menus, got %d", len(menus))
	}
}
