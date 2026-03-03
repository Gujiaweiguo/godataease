//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestMenuServiceIntegration_Query(t *testing.T) {
	cleanupTables(&menu.CoreMenu{})

	repo := repository.NewMenuRepository(testDB)
	svc := NewMenuService(repo)

	// Create test menus
	menus := []*menu.CoreMenu{
		{ID: 100, Name: "System", Path: "/system", Pid: 0, Type: 0, MenuSort: 1, Hidden: false},
		{ID: 101, Name: "User", Path: "/system/user", Pid: 100, Type: 0, MenuSort: 1, Hidden: false},
		{ID: 102, Name: "Role", Path: "/system/role", Pid: 100, Type: 0, MenuSort: 2, Hidden: false},
		{ID: 103, Name: "Org", Path: "/system/org", Pid: 100, Type: 0, MenuSort: 3, Hidden: false},
		{ID: 200, Name: "Data", Path: "/data", Pid: 0, Type: 0, MenuSort: 2, Hidden: false},
	}

	for _, m := range menus {
		err := repo.Create(m)
		assert.NoError(t, err)
	}

	// Query all menus
	result, err := svc.Query()
	assert.NoError(t, err)
	assert.Len(t, result, 2) // 2 root menus

	// Verify tree structure - System should have 3 children
	var systemRoot *menu.MenuVO
	for _, r := range result {
		if r.ID == 100 {
			systemRoot = r
			break
		}
	}
	assert.NotNil(t, systemRoot)
	assert.Len(t, systemRoot.Children, 3)
}

func TestMenuServiceIntegration_CRUD(t *testing.T) {
	cleanupTables(&menu.CoreMenu{})

	repo := repository.NewMenuRepository(testDB)
	svc := NewMenuService(repo)

	// Create
	m := &menu.CoreMenu{
		Name:     "TestMenu",
		Path:     "/test",
		Pid:      0,
		Type:     0,
		MenuSort: 1,
		Hidden:   false,
	}
	err := svc.Create(m)
	assert.NoError(t, err)

	// GetByID
	found, err := svc.GetByID(m.ID)
	assert.NoError(t, err)
	assert.Equal(t, "TestMenu", found.Name)

	// Update
	found.Name = "UpdatedMenu"
	err = svc.Update(found)
	assert.NoError(t, err)

	updated, err := svc.GetByID(m.ID)
	assert.NoError(t, err)
	assert.Equal(t, "UpdatedMenu", updated.Name)

	// UpdateSort
	err = svc.UpdateSort(m.ID, 10)
	assert.NoError(t, err)

	// UpdateHidden
	err = svc.UpdateHidden(m.ID, true)
	assert.NoError(t, err)

	// Delete
	err = svc.Delete(m.ID)
	assert.NoError(t, err)

	_, err = svc.GetByID(m.ID)
	assert.Error(t, err)
}

func TestMenuServiceIntegration_DeleteWithChildren(t *testing.T) {
	cleanupTables(&menu.CoreMenu{})

	repo := repository.NewMenuRepository(testDB)
	svc := NewMenuService(repo)

	// Create parent
	parent := &menu.CoreMenu{
		Name:     "Parent",
		Path:     "/parent",
		Pid:      0,
		Type:     0,
		MenuSort: 1,
	}
	err := svc.Create(parent)
	assert.NoError(t, err)

	// Create child
	child := &menu.CoreMenu{
		Name:     "Child",
		Path:     "/parent/child",
		Pid:      parent.ID,
		Type:     0,
		MenuSort: 1,
	}
	err = svc.Create(child)
	assert.NoError(t, err)

	// Try to delete parent (should fail)
	err = svc.Delete(parent.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrMenuHasChildren, err)
}

func TestMenuServiceIntegration_QueryByRoleIDs(t *testing.T) {
	cleanupTables(&menu.CoreMenu{}, &role.RoleMenu{})

	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewMenuServiceWithRoleFilter(menuRepo, roleMenuRepo)

	// Create menus
	menus := []*menu.CoreMenu{
		{ID: 300, Name: "Root1", Path: "/root1", Pid: 0, Type: 0, MenuSort: 1},
		{ID: 301, Name: "Child1", Path: "/root1/child1", Pid: 300, Type: 0, MenuSort: 1},
		{ID: 302, Name: "Root2", Path: "/root2", Pid: 0, Type: 0, MenuSort: 2},
	}
	for _, m := range menus {
		err := menuRepo.Create(m)
		assert.NoError(t, err)
	}

	// Assign menus to role 2
	err := roleMenuRepo.SaveRoleMenus(2, []int64{300, 301})
	assert.NoError(t, err)

	// Query by role IDs (non-admin)
	result, err := svc.QueryByRoleIDs([]int64{2})
	assert.NoError(t, err)
	assert.Len(t, result, 1) // Only Root1 with Child1
	assert.Equal(t, int64(300), result[0].ID)
	assert.Len(t, result[0].Children, 1)

	// Query with admin role (should get all)
	adminResult, err := svc.QueryByRoleIDs([]int64{1})
	assert.NoError(t, err)
	assert.Len(t, adminResult, 2)

	// Query with empty role IDs
	emptyResult, err := svc.QueryByRoleIDs([]int64{})
	assert.NoError(t, err)
	assert.Len(t, emptyResult, 0)
}

func TestMenuServiceIntegration_GetAuthorizedMenuIDs(t *testing.T) {
	cleanupTables(&menu.CoreMenu{}, &role.RoleMenu{})

	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewMenuServiceWithRoleFilter(menuRepo, roleMenuRepo)

	// Create menus
	menus := []*menu.CoreMenu{
		{ID: 400, Name: "Menu1", Path: "/menu1", Pid: 0, Type: 0, MenuSort: 1},
		{ID: 401, Name: "Menu2", Path: "/menu2", Pid: 0, Type: 0, MenuSort: 2},
	}
	for _, m := range menus {
		err := menuRepo.Create(m)
		assert.NoError(t, err)
	}

	// Assign menu1 to role 3
	err := roleMenuRepo.SaveRoleMenus(3, []int64{400})
	assert.NoError(t, err)

	// Get authorized menu IDs for role 3
	ids, err := svc.GetAuthorizedMenuIDs([]int64{3})
	assert.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Contains(t, ids, int64(400))

	// Get authorized menu IDs for admin
	adminIDs, err := svc.GetAuthorizedMenuIDs([]int64{1})
	assert.NoError(t, err)
	assert.Len(t, adminIDs, 2)
}

func TestMenuServiceIntegration_ShouldUseDynamicMenu(t *testing.T) {
	repo := repository.NewMenuRepository(testDB)
	svc := NewMenuService(repo)

	// Should always return true
	assert.True(t, svc.ShouldUseDynamicMenu())
}

func TestMenuServiceIntegration_IsAdminRole(t *testing.T) {
	repo := repository.NewMenuRepository(testDB)
	svc := NewMenuService(repo)

	// Admin role
	assert.True(t, svc.isAdminRole([]int64{1}))
	assert.True(t, svc.isAdminRole([]int64{2, 1}))

	// Non-admin roles
	assert.False(t, svc.isAdminRole([]int64{2, 3}))
	assert.False(t, svc.isAdminRole([]int64{}))
}
