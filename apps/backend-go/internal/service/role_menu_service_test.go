package service

import (
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRoleMenuServiceTest(t *testing.T) (*RoleMenuService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&role.SysRole{}, &menu.CoreMenu{}, &role.RoleMenu{}))

	roleRepo := repository.NewRoleRepository(db)
	menuRepo := repository.NewMenuRepository(db)
	roleMenuRepo := repository.NewRoleMenuRepository(db)
	return NewRoleMenuService(roleMenuRepo, roleRepo, menuRepo), db
}

func setupRoleMenuServiceSplitDBTest(t *testing.T) (*RoleMenuService, *gorm.DB, *gorm.DB, *gorm.DB) {
	t.Helper()

	roleDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, roleDB.AutoMigrate(&role.SysRole{}))

	menuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, menuDB.AutoMigrate(&menu.CoreMenu{}))

	roleMenuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, roleMenuDB.AutoMigrate(&role.RoleMenu{}))

	return NewRoleMenuService(
		repository.NewRoleMenuRepository(roleMenuDB),
		repository.NewRoleRepository(roleDB),
		repository.NewMenuRepository(menuDB),
	), roleDB, menuDB, roleMenuDB
}

func closeGormDB(t *testing.T, db *gorm.DB) {
	t.Helper()

	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
}

func TestRoleMenuService_GuardsAndRoleNotFound(t *testing.T) {
	svc, _ := setupRoleMenuServiceTest(t)

	auth, err := svc.GetRoleMenuAuth(0)
	require.ErrorIs(t, err, ErrInvalidRoleID)
	assert.Nil(t, auth)

	err = svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: 0})
	require.ErrorIs(t, err, ErrInvalidRoleID)

	ids, err := svc.GetAuthorizedMenuIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, ids)

	ok, err := svc.IsMenuAuthorized(nil, 1)
	require.NoError(t, err)
	assert.False(t, ok)

	err = svc.DeleteRoleMenuAuth(0)
	require.ErrorIs(t, err, ErrInvalidRoleID)

	auth, err = svc.GetRoleMenuAuth(999)
	require.ErrorIs(t, err, ErrRoleNotFound)
	assert.Nil(t, auth)

	err = svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: 999, MenuIDs: []int64{1}})
	require.ErrorIs(t, err, ErrRoleNotFound)
}

func TestRoleMenuService_SaveAndGetRoleMenuAuth(t *testing.T) {
	svc, db := setupRoleMenuServiceTest(t)
	require.NoError(t, db.Create(&role.SysRole{RoleName: "Role A", RoleCode: "role-a", Status: 1}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 10, Name: "Menu A", Pid: 0, MenuSort: 1}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 11, Name: "Menu B", Pid: 0, MenuSort: 2}).Error)

	var savedRole role.SysRole
	require.NoError(t, db.Where("role_code = ?", "role-a").First(&savedRole).Error)

	err := svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: savedRole.RoleID, MenuIDs: []int64{10, 11}})
	require.NoError(t, err)

	auth, err := svc.GetRoleMenuAuth(savedRole.RoleID)
	require.NoError(t, err)
	require.NotNil(t, auth)
	assert.Equal(t, savedRole.RoleID, auth.RoleID)
	assert.ElementsMatch(t, []int64{10, 11}, auth.MenuIDs)

	err = svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: savedRole.RoleID, MenuIDs: []int64{999}})
	require.ErrorIs(t, err, ErrMenuNotFound)

	err = svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: savedRole.RoleID, MenuIDs: []int64{}})
	require.NoError(t, err)
	auth, err = svc.GetRoleMenuAuth(savedRole.RoleID)
	require.NoError(t, err)
	assert.Empty(t, auth.MenuIDs)
}

func TestRoleMenuService_AuthorizationHelpersAndDelete(t *testing.T) {
	svc, db := setupRoleMenuServiceTest(t)
	require.NoError(t, db.Create(&role.SysRole{RoleName: "Role 1", RoleCode: "role-1", Status: 1}).Error)
	require.NoError(t, db.Create(&role.SysRole{RoleName: "Role 2", RoleCode: "role-2", Status: 1}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 20, Name: "Menu 1", Pid: 0, MenuSort: 1}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 21, Name: "Menu 2", Pid: 0, MenuSort: 2}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 22, Name: "Menu 3", Pid: 0, MenuSort: 3}).Error)

	var roles []role.SysRole
	require.NoError(t, db.Order("role_id asc").Find(&roles).Error)
	require.Len(t, roles, 2)

	require.NoError(t, svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: roles[0].RoleID, MenuIDs: []int64{20, 21}}))
	require.NoError(t, svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: roles[1].RoleID, MenuIDs: []int64{21, 22}}))

	ids, err := svc.GetAuthorizedMenuIDs([]int64{roles[0].RoleID, roles[1].RoleID})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{20, 21, 22}, ids)

	ok, err := svc.IsMenuAuthorized([]int64{roles[0].RoleID}, 20)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = svc.IsMenuAuthorized([]int64{roles[0].RoleID}, 22)
	require.NoError(t, err)
	assert.False(t, ok)

	err = svc.DeleteRoleMenuAuth(roles[0].RoleID)
	require.NoError(t, err)
	auth, err := svc.GetRoleMenuAuth(roles[0].RoleID)
	require.NoError(t, err)
	assert.Empty(t, auth.MenuIDs)
}

func TestRoleMenuService_RepoErrorPropagation(t *testing.T) {
	t.Run("get role menu auth returns role-menu repo error", func(t *testing.T) {
		svc, roleDB, _, roleMenuDB := setupRoleMenuServiceSplitDBTest(t)
		require.NoError(t, roleDB.Create(&role.SysRole{RoleName: "Role A", RoleCode: "role-a", Status: 1}).Error)
		var savedRole role.SysRole
		require.NoError(t, roleDB.Where("role_code = ?", "role-a").First(&savedRole).Error)
		closeGormDB(t, roleMenuDB)

		auth, err := svc.GetRoleMenuAuth(savedRole.RoleID)
		require.Error(t, err)
		assert.Nil(t, auth)
	})

	t.Run("save role menu auth returns menu repo error", func(t *testing.T) {
		svc, roleDB, menuDB, _ := setupRoleMenuServiceSplitDBTest(t)
		require.NoError(t, roleDB.Create(&role.SysRole{RoleName: "Role B", RoleCode: "role-b", Status: 1}).Error)
		var savedRole role.SysRole
		require.NoError(t, roleDB.Where("role_code = ?", "role-b").First(&savedRole).Error)
		closeGormDB(t, menuDB)

		err := svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: savedRole.RoleID, MenuIDs: []int64{1}})
		require.Error(t, err)
	})

	t.Run("save and delete role menu auth return role-menu repo errors", func(t *testing.T) {
		svc, roleDB, menuDB, roleMenuDB := setupRoleMenuServiceSplitDBTest(t)
		require.NoError(t, roleDB.Create(&role.SysRole{RoleName: "Role C", RoleCode: "role-c", Status: 1}).Error)
		require.NoError(t, menuDB.Create(&menu.CoreMenu{ID: 30, Name: "Menu 30", Pid: 0, MenuSort: 1}).Error)
		var savedRole role.SysRole
		require.NoError(t, roleDB.Where("role_code = ?", "role-c").First(&savedRole).Error)
		closeGormDB(t, roleMenuDB)

		err := svc.SaveRoleMenuAuth(&SaveRoleMenuRequest{RoleID: savedRole.RoleID, MenuIDs: []int64{30}})
		require.Error(t, err)

		_, err = svc.GetAuthorizedMenuIDs([]int64{savedRole.RoleID})
		require.Error(t, err)

		_, err = svc.IsMenuAuthorized([]int64{savedRole.RoleID}, 30)
		require.Error(t, err)

		err = svc.DeleteRoleMenuAuth(savedRole.RoleID)
		require.Error(t, err)
	})
}
