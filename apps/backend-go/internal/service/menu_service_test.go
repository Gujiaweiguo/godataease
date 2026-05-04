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

func setupMenuServiceTest(t *testing.T) (*MenuService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&menu.CoreMenu{}, &role.RoleMenu{}))

	menuRepo := repository.NewMenuRepository(db)
	roleMenuRepo := repository.NewRoleMenuRepository(db)
	return NewMenuServiceWithRoleFilter(menuRepo, roleMenuRepo), db
}

func TestMenuService_QueryAndTreeBuilders(t *testing.T) {
	svc, db := setupMenuServiceTest(t)
	actionCfg := menu.JSON{"method": "post", "confirm": true}
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 10, Pid: 0, Type: 0, Name: "Root", Path: "/root", MenuSort: 1, Icon: "home"}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 11, Pid: 10, Type: 0, Name: "Child", Path: "/root/child", MenuSort: 1, ActionConfig: actionCfg}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 12, Pid: 11, Type: 1, Name: "Button Child", Path: "/root/child/button", MenuSort: 1}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 13, Pid: 0, Type: 1, Name: "Hidden Root Button", Path: "/button", MenuSort: 2}).Error)

	result, err := svc.Query()
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(10), result[0].ID)
	assert.Equal(t, "/root", result[0].Path)
	require.Len(t, result[0].Children, 1)
	assert.Equal(t, "root/child", result[0].Children[0].Path)
	assert.Empty(t, result[0].Children[0].Children)
	require.NotNil(t, result[0].Children[0].ActionConfig)
	assert.Equal(t, "post", result[0].Children[0].ActionConfig["method"])
	require.NotNil(t, result[0].Meta)
	assert.Equal(t, "Root", result[0].Meta.Title)

	assert.True(t, svc.ShouldUseDynamicMenu())
	assert.True(t, svc.isAdminRole([]int64{2, 1}))
	assert.False(t, svc.isAdminRole([]int64{2, 3}))

	t.Run("orders roots by menu sort", func(t *testing.T) {
		svc, db := setupMenuServiceTest(t)
		require.NoError(t, db.Create(&menu.CoreMenu{ID: 14, Pid: 0, Type: 0, Name: "Late", Path: "/late", MenuSort: 20}).Error)
		require.NoError(t, db.Create(&menu.CoreMenu{ID: 15, Pid: 0, Type: 0, Name: "Early", Path: "/early", MenuSort: 1}).Error)
		require.NoError(t, db.Create(&menu.CoreMenu{ID: 16, Pid: 0, Type: 0, Name: "Middle", Path: "/middle", MenuSort: 10}).Error)

		result, err := svc.Query()
		require.NoError(t, err)
		require.Len(t, result, 3)
		assert.Equal(t, []int64{15, 16, 14}, []int64{result[0].ID, result[1].ID, result[2].ID})
	})
}

func TestMenuService_QueryByRoleIDsAndAuthorizedIDs(t *testing.T) {
	svc, db := setupMenuServiceTest(t)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 20, Pid: 0, Type: 0, Name: "Root 1", Path: "/root1", MenuSort: 1}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 21, Pid: 20, Type: 0, Name: "Child 1", Path: "/root1/child1", MenuSort: 2}).Error)
	require.NoError(t, db.Create(&menu.CoreMenu{ID: 22, Pid: 0, Type: 0, Name: "Root 2", Path: "/root2", MenuSort: 3}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 2, MenuID: 20}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 2, MenuID: 21}).Error)

	result, err := svc.QueryByRoleIDs([]int64{2})
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, int64(20), result[0].ID)
	require.Len(t, result[0].Children, 1)

	adminResult, err := svc.QueryByRoleIDs([]int64{1})
	require.NoError(t, err)
	require.Len(t, adminResult, 2)

	emptySvc := NewMenuServiceWithRoleFilter(repository.NewMenuRepository(db), nil)
	emptyResult, err := emptySvc.QueryByRoleIDs([]int64{2})
	require.NoError(t, err)
	assert.Empty(t, emptyResult)

	ids, err := svc.GetAuthorizedMenuIDs([]int64{2})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{20, 21}, ids)

	adminIDs, err := svc.GetAuthorizedMenuIDs([]int64{1})
	require.NoError(t, err)
	assert.ElementsMatch(t, []int64{20, 21, 22}, adminIDs)

	ids, err = emptySvc.GetAuthorizedMenuIDs([]int64{2})
	require.NoError(t, err)
	assert.Empty(t, ids)

	emptyResult, err = svc.QueryByRoleIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyResult)

	emptyIDs, err := svc.GetAuthorizedMenuIDs(nil)
	require.NoError(t, err)
	assert.Empty(t, emptyIDs)

	t.Run("filters out button only root", func(t *testing.T) {
		svc, db := setupMenuServiceTest(t)
		require.NoError(t, db.Create(&menu.CoreMenu{ID: 23, Pid: 0, Type: 1, Name: "Button Root", Path: "/button-root", MenuSort: 1}).Error)
		require.NoError(t, db.Create(&role.RoleMenu{RoleID: 2, MenuID: 23}).Error)

		result, err := svc.QueryByRoleIDs([]int64{2})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestMenuService_CRUDAndDeleteGuards(t *testing.T) {
	svc, db := setupMenuServiceTest(t)
	require.NoError(t, svc.Create(&menu.CoreMenu{ID: 30, Pid: 0, Type: 0, Name: "Parent", Path: "/parent", MenuSort: 1}))
	require.NoError(t, svc.Create(&menu.CoreMenu{ID: 31, Pid: 30, Type: 0, Name: "Child", Path: "/parent/child", MenuSort: 2}))
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 9, MenuID: 31}).Error)

	found, err := svc.GetByID(30)
	require.NoError(t, err)
	assert.Equal(t, "Parent", found.Name)

	found.Name = "Parent Updated"
	require.NoError(t, svc.Update(found))
	require.NoError(t, svc.UpdateSort(30, 99))
	require.NoError(t, svc.UpdateHidden(30, true))

	updated, err := svc.GetByID(30)
	require.NoError(t, err)
	assert.Equal(t, "Parent Updated", updated.Name)
	assert.Equal(t, 99, updated.MenuSort)
	assert.True(t, updated.Hidden)

	byPath, err := svc.GetByPath("/parent")
	require.NoError(t, err)
	assert.Equal(t, int64(30), byPath.ID)

	err = svc.Delete(30)
	require.ErrorIs(t, err, ErrMenuHasChildren)

	err = svc.Delete(31)
	require.NoError(t, err)

	_, err = svc.GetByID(31)
	require.Error(t, err)

	var count int64
	require.NoError(t, db.Model(&role.RoleMenu{}).Where("menu_id = ?", 31).Count(&count).Error)
	assert.Zero(t, count)

	err = svc.Delete(30)
	require.NoError(t, err)

	t.Run("delete without role menu rows still succeeds", func(t *testing.T) {
		svc, _ := setupMenuServiceTest(t)
		require.NoError(t, svc.Create(&menu.CoreMenu{ID: 32, Pid: 0, Type: 0, Name: "Leaf", Path: "/leaf", MenuSort: 1}))

		err := svc.Delete(32)
		require.NoError(t, err)

		_, err = svc.GetByID(32)
		require.Error(t, err)
	})
}

func TestMenuService_ErrorPaths(t *testing.T) {
	t.Run("query propagates repository error", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&menu.CoreMenu{}))
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		svc := NewMenuService(repository.NewMenuRepository(db))
		result, err := svc.Query()
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("query by role ids propagates role menu lookup error", func(t *testing.T) {
		menuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, menuDB.AutoMigrate(&menu.CoreMenu{}))

		roleMenuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, roleMenuDB.AutoMigrate(&role.RoleMenu{}))
		sqlDB, err := roleMenuDB.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		svc := NewMenuServiceWithRoleFilter(repository.NewMenuRepository(menuDB), repository.NewRoleMenuRepository(roleMenuDB))
		result, err := svc.QueryByRoleIDs([]int64{2})
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("query by role ids propagates menu lookup error", func(t *testing.T) {
		menuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, menuDB.AutoMigrate(&menu.CoreMenu{}))
		menuSQLDB, err := menuDB.DB()
		require.NoError(t, err)
		require.NoError(t, menuSQLDB.Close())

		roleMenuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, roleMenuDB.AutoMigrate(&role.RoleMenu{}))
		require.NoError(t, roleMenuDB.Create(&role.RoleMenu{RoleID: 2, MenuID: 40}).Error)

		svc := NewMenuServiceWithRoleFilter(repository.NewMenuRepository(menuDB), repository.NewRoleMenuRepository(roleMenuDB))
		result, err := svc.QueryByRoleIDs([]int64{2})
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("get authorized menu ids propagates admin repo error", func(t *testing.T) {
		menuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, menuDB.AutoMigrate(&menu.CoreMenu{}))
		menuSQLDB, err := menuDB.DB()
		require.NoError(t, err)
		require.NoError(t, menuSQLDB.Close())

		roleMenuDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, roleMenuDB.AutoMigrate(&role.RoleMenu{}))

		svc := NewMenuServiceWithRoleFilter(repository.NewMenuRepository(menuDB), repository.NewRoleMenuRepository(roleMenuDB))
		ids, err := svc.GetAuthorizedMenuIDs([]int64{1})
		require.Error(t, err)
		assert.Nil(t, ids)
	})

	t.Run("delete returns role menu cleanup error after deleting menu", func(t *testing.T) {
		svc, db := setupMenuServiceTest(t)
		require.NoError(t, svc.Create(&menu.CoreMenu{ID: 50, Pid: 0, Type: 0, Name: "Delete Target", Path: "/delete-target", MenuSort: 1}))
		require.NoError(t, db.Create(&role.RoleMenu{RoleID: 8, MenuID: 50}).Error)
		require.NoError(t, db.Exec(`
			CREATE TRIGGER sys_role_menu_delete_fail
			BEFORE DELETE ON sys_role_menu
			BEGIN
				SELECT RAISE(FAIL, 'forced role_menu delete failure');
			END;
		`).Error)

		err := svc.Delete(50)
		require.Error(t, err)

		_, getErr := svc.GetByID(50)
		assert.Error(t, getErr)

		var count int64
		require.NoError(t, db.Model(&role.RoleMenu{}).Where("menu_id = ?", 50).Count(&count).Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("delete propagates has children lookup error", func(t *testing.T) {
		svc, db := setupMenuServiceTest(t)
		require.NoError(t, db.Exec("DROP TABLE core_menu").Error)

		err := svc.Delete(60)
		require.Error(t, err)
	})
}
