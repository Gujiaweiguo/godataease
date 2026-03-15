//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestRoleMenuServiceIntegration_GetRoleMenuAuth(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create role
	testRole := &role.SysRole{
		RoleName: "Test Role",
		RoleCode: "test_role_auth",
		Status:   1,
		CreateBy: strPtr("tester"),
	}
	err := roleRepo.Create(testRole)
	assert.NoError(t, err)

	// Create menus
	menu1 := &menu.CoreMenu{
		Name:     "Menu 1",
		Pid:      0,
		MenuSort: 1,
	}
	menu2 := &menu.CoreMenu{
		Name:     "Menu 2",
		Pid:      0,
		MenuSort: 2,
	}
	err = menuRepo.Create(menu1)
	assert.NoError(t, err)
	err = menuRepo.Create(menu2)
	assert.NoError(t, err)

	// Create role-menu associations
	err = roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{menu1.ID, menu2.ID})
	assert.NoError(t, err)

	// Get role menu auth
	auth, err := svc.GetRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)
	assert.Equal(t, testRole.RoleID, auth.RoleID)
	assert.Len(t, auth.MenuIDs, 2)
}

func TestRoleMenuServiceIntegration_GetRoleMenuAuth_InvalidRoleID(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	_, err := svc.GetRoleMenuAuth(0)
	assert.Equal(t, ErrInvalidRoleID, err)

	_, err = svc.GetRoleMenuAuth(-1)
	assert.Equal(t, ErrInvalidRoleID, err)
}

func TestRoleMenuServiceIntegration_GetRoleMenuAuth_RoleNotFound(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	_, err := svc.GetRoleMenuAuth(99999)
	assert.Equal(t, ErrRoleNotFound, err)
}

func TestRoleMenuServiceIntegration_SaveRoleMenuAuth(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create role
	testRole := &role.SysRole{
		RoleName: "Save Test Role",
		RoleCode: "save_test_role",
		Status:   1,
		CreateBy: strPtr("tester"),
	}
	err := roleRepo.Create(testRole)
	assert.NoError(t, err)

	// Create menus
	menu1 := &menu.CoreMenu{Name: "Save Menu 1", Pid: 0, MenuSort: 1}
	menu2 := &menu.CoreMenu{Name: "Save Menu 2", Pid: 0, MenuSort: 2}
	menuRepo.Create(menu1)
	menuRepo.Create(menu2)

	// Save role menu auth
	req := &SaveRoleMenuRequest{
		RoleID:  testRole.RoleID,
		MenuIDs: []int64{menu1.ID, menu2.ID},
	}
	err = svc.SaveRoleMenuAuth(req)
	assert.NoError(t, err)

	// Verify saved
	auth, err := svc.GetRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)
	assert.Len(t, auth.MenuIDs, 2)
}

func TestRoleMenuServiceIntegration_SaveRoleMenuAuth_EmptyMenuIDs(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create role
	testRole := &role.SysRole{
		RoleName: "Empty Menu Role",
		RoleCode: "empty_menu_role",
		Status:   1,
		CreateBy: strPtr("tester"),
	}
	err := roleRepo.Create(testRole)
	assert.NoError(t, err)

	// First add some menus
	menu1 := &menu.CoreMenu{Name: "Temp Menu", Pid: 0, MenuSort: 1}
	menuRepo.Create(menu1)
	roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{menu1.ID})

	// Save with empty menu IDs (should clear all)
	req := &SaveRoleMenuRequest{
		RoleID:  testRole.RoleID,
		MenuIDs: []int64{},
	}
	err = svc.SaveRoleMenuAuth(req)
	assert.NoError(t, err)

	// Verify cleared
	auth, err := svc.GetRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)
	assert.Len(t, auth.MenuIDs, 0)
}

func TestRoleMenuServiceIntegration_SaveRoleMenuAuth_InvalidMenuID(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create role
	testRole := &role.SysRole{
		RoleName: "Invalid Menu Role",
		RoleCode: "invalid_menu_role",
		Status:   1,
		CreateBy: strPtr("tester"),
	}
	err := roleRepo.Create(testRole)
	assert.NoError(t, err)

	// Try to save with non-existent menu ID
	req := &SaveRoleMenuRequest{
		RoleID:  testRole.RoleID,
		MenuIDs: []int64{99999},
	}
	err = svc.SaveRoleMenuAuth(req)
	assert.Equal(t, ErrMenuNotFound, err)
}

func TestRoleMenuServiceIntegration_GetAuthorizedMenuIDs(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create roles
	role1 := &role.SysRole{RoleName: "Auth Role 1", RoleCode: "auth_role_1", Status: 1, CreateBy: strPtr("tester")}
	role2 := &role.SysRole{RoleName: "Auth Role 2", RoleCode: "auth_role_2", Status: 1, CreateBy: strPtr("tester")}
	roleRepo.Create(role1)
	roleRepo.Create(role2)

	// Create menus
	menu1 := &menu.CoreMenu{Name: "Auth Menu 1", Pid: 0, MenuSort: 1}
	menu2 := &menu.CoreMenu{Name: "Auth Menu 2", Pid: 0, MenuSort: 2}
	menu3 := &menu.CoreMenu{Name: "Auth Menu 3", Pid: 0, MenuSort: 3}
	menuRepo.Create(menu1)
	menuRepo.Create(menu2)
	menuRepo.Create(menu3)

	// Role1 has menu1, menu2
	roleMenuRepo.SaveRoleMenus(role1.RoleID, []int64{menu1.ID, menu2.ID})
	// Role2 has menu2, menu3
	roleMenuRepo.SaveRoleMenus(role2.RoleID, []int64{menu2.ID, menu3.ID})

	// Get authorized menu IDs for both roles
	menuIDs, err := svc.GetAuthorizedMenuIDs([]int64{role1.RoleID, role2.RoleID})
	assert.NoError(t, err)
	assert.Len(t, menuIDs, 3) // Should have menu1, menu2, menu3 (unique)
}

func TestRoleMenuServiceIntegration_GetAuthorizedMenuIDs_Empty(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	menuIDs, err := svc.GetAuthorizedMenuIDs([]int64{})
	assert.NoError(t, err)
	assert.Len(t, menuIDs, 0)
}

func TestRoleMenuServiceIntegration_IsMenuAuthorized(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create role
	testRole := &role.SysRole{
		RoleName: "Auth Check Role",
		RoleCode: "auth_check_role",
		Status:   1,
		CreateBy: strPtr("tester"),
	}
	err := roleRepo.Create(testRole)
	assert.NoError(t, err)

	// Create menus
	authMenu := &menu.CoreMenu{Name: "Authorized Menu", Pid: 0, MenuSort: 1}
	unauthMenu := &menu.CoreMenu{Name: "Unauthorized Menu", Pid: 0, MenuSort: 2}
	menuRepo.Create(authMenu)
	menuRepo.Create(unauthMenu)

	// Grant only authMenu
	roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{authMenu.ID})

	// Check authorized
	authorized, err := svc.IsMenuAuthorized([]int64{testRole.RoleID}, authMenu.ID)
	assert.NoError(t, err)
	assert.True(t, authorized)

	// Check unauthorized
	authorized, err = svc.IsMenuAuthorized([]int64{testRole.RoleID}, unauthMenu.ID)
	assert.NoError(t, err)
	assert.False(t, authorized)
}

func TestRoleMenuServiceIntegration_IsMenuAuthorized_EmptyRoleIDs(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create menu
	testMenu := &menu.CoreMenu{Name: "Test Menu", Pid: 0, MenuSort: 1}
	menuRepo.Create(testMenu)

	// Check with empty roleIDs - should return false
	authorized, err := svc.IsMenuAuthorized([]int64{}, testMenu.ID)
	assert.NoError(t, err)
	assert.False(t, authorized)

	// Check with nil roleIDs - should return false
	authorized, err = svc.IsMenuAuthorized(nil, testMenu.ID)
	assert.NoError(t, err)
	assert.False(t, authorized)
}

func TestRoleMenuServiceIntegration_DeleteRoleMenuAuth(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create role
	testRole := &role.SysRole{
		RoleName: "Delete Auth Role",
		RoleCode: "delete_auth_role",
		Status:   1,
		CreateBy: strPtr("tester"),
	}
	err := roleRepo.Create(testRole)
	assert.NoError(t, err)

	// Create menu and assign
	testMenu := &menu.CoreMenu{Name: "Delete Test Menu", Pid: 0, MenuSort: 1}
	menuRepo.Create(testMenu)
	roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{testMenu.ID})

	// Verify assigned
	auth, _ := svc.GetRoleMenuAuth(testRole.RoleID)
	assert.Len(t, auth.MenuIDs, 1)

	// Delete auth
	err = svc.DeleteRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)

	// Verify deleted
	auth, err = svc.GetRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)
	assert.Len(t, auth.MenuIDs, 0)
}

func TestRoleMenuServiceIntegration_DeleteRoleMenuAuth_InvalidRoleID(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	svc := NewRoleMenuService(repository.NewRoleMenuRepository(testDB), repository.NewRoleRepository(testDB), repository.NewMenuRepository(testDB))
	assert.Equal(t, ErrInvalidRoleID, svc.DeleteRoleMenuAuth(0))
	assert.Equal(t, ErrInvalidRoleID, svc.DeleteRoleMenuAuth(-1))
}

func TestRoleMenuServiceIntegration_SaveRoleMenuAuth_RoleNotFound(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create menus
	menu1 := &menu.CoreMenu{Name: "NoRole Menu", Pid: 0, MenuSort: 1}
	menuRepo.Create(menu1)

	// Try to save without creating role
	req := &SaveRoleMenuRequest{
		RoleID:  99999,
		MenuIDs: []int64{menu1.ID},
	}
	err := svc.SaveRoleMenuAuth(req)
	assert.Equal(t, ErrRoleNotFound, err)
}

func TestRoleMenuServiceIntegration_SaveRoleMenuAuth_InvalidRoleID(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	req := &SaveRoleMenuRequest{
		RoleID:  0,
		MenuIDs: []int64{1},
	}
	err := svc.SaveRoleMenuAuth(req)
	assert.Equal(t, ErrInvalidRoleID, err)

	req.RoleID = -1
	err = svc.SaveRoleMenuAuth(req)
	assert.Equal(t, ErrInvalidRoleID, err)
}

func TestRoleMenuServiceIntegration_DeleteRoleMenuAuth_Success(t *testing.T) {
	cleanupTables(&role.RoleMenu{}, &role.SysRole{}, &menu.CoreMenu{})

	roleRepo := repository.NewRoleRepository(testDB)
	menuRepo := repository.NewMenuRepository(testDB)
	roleMenuRepo := repository.NewRoleMenuRepository(testDB)
	svc := NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo)

	// Create role and menus
	testRole := &role.SysRole{
		RoleName: "DeleteSuccess Role",
		RoleCode: "delete_success_role",
		Status:   1,
		CreateBy: strPtr("tester"),
	}
	err := roleRepo.Create(testRole)
	assert.NoError(t, err)

	menu1 := &menu.CoreMenu{Name: "DeleteSuccess Menu", Pid: 0, MenuSort: 1}
	menuRepo.Create(menu1)

	// Save role menu auth
	err = roleMenuRepo.SaveRoleMenus(testRole.RoleID, []int64{menu1.ID})
	assert.NoError(t, err)

	// Verify it exists
	auth, err := svc.GetRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)
	assert.Len(t, auth.MenuIDs, 1)

	// Delete role menu auth
	err = svc.DeleteRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)

	// Verify it's deleted
	auth, err = svc.GetRoleMenuAuth(testRole.RoleID)
	assert.NoError(t, err)
	assert.Len(t, auth.MenuIDs, 0)
}
