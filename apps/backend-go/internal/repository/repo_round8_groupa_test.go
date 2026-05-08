package repository

import (
	"testing"

	"dataease/backend/internal/domain/menu"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/role"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ==================== Shared helpers ====================

func round8OpenDB(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(models...))
	return db
}

// round8RoleDB creates a DB with role tables AND the raw sys_user_role table.
func round8RoleDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := round8OpenDB(t, &role.SysRole{})
	// sys_user_role is used via raw .Table() calls, not a domain struct with AutoMigrate.
	require.NoError(t, db.Exec(
		"CREATE TABLE IF NOT EXISTS sys_user_role (user_id INTEGER, role_id INTEGER, org_id INTEGER)",
	).Error)
	return db
}

func round8RoleMenuDB(t *testing.T) *gorm.DB {
	db := round8OpenDB(t, &role.RoleMenu{})
	require.NoError(t, db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_menu_unique ON sys_role_menu (role_id, menu_id)").Error)
	return db
}

func round8OrgDB(t *testing.T) *gorm.DB {
	return round8OpenDB(t, &org.SysOrg{})
}

func round8MenuDB(t *testing.T) *gorm.DB {
	return round8OpenDB(t, &menu.CoreMenu{})
}

// ==================== RoleRepository (17 functions) ====================

func TestRound8A_Role_New(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NotNil(t, repo)
}

func TestRound8A_Role_DB(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	assert.Equal(t, db, repo.DB())
}

func TestRound8A_Role_Create(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	r := &role.SysRole{RoleName: "admin", RoleCode: "ADMIN", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(r))
	assert.Positive(t, r.RoleID)
}

func TestRound8A_Role_Update(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	r := &role.SysRole{RoleName: "old", RoleCode: "CODE1", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(r))
	r.RoleName = "new"
	require.NoError(t, repo.Update(r))
	got, err := repo.GetByID(r.RoleID)
	require.NoError(t, err)
	assert.Equal(t, "new", got.RoleName)
}

func TestRound8A_Role_Delete(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	r := &role.SysRole{RoleName: "del", RoleCode: "DEL", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(r))
	require.NoError(t, repo.Delete(r.RoleID))
	_, err := repo.GetByID(r.RoleID)
	assert.Error(t, err)
}

func TestRound8A_Role_Delete_NotFound(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	// Deleting a non-existent ID should not error with GORM
	require.NoError(t, repo.Delete(99999))
}

func TestRound8A_Role_GetByID(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	r := &role.SysRole{RoleName: "find", RoleCode: "FIND", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(r))
	got, err := repo.GetByID(r.RoleID)
	require.NoError(t, err)
	assert.Equal(t, "find", got.RoleName)
}

func TestRound8A_Role_GetByID_NotFound(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	_, err := repo.GetByID(99999)
	assert.Error(t, err)
}

func TestRound8A_Role_GetByRoleCode(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	r := &role.SysRole{RoleName: "coder", RoleCode: "UNIQUE_CODE", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(r))
	got, err := repo.GetByRoleCode("UNIQUE_CODE")
	require.NoError(t, err)
	assert.Equal(t, "coder", got.RoleName)
}

func TestRound8A_Role_GetByRoleCode_NotFound(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	_, err := repo.GetByRoleCode("MISSING")
	assert.Error(t, err)
}

func TestRound8A_Role_Query_NoKeyword(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "r1", RoleCode: "c1", Status: 1}))
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "r2", RoleCode: "c2", Status: 1}))
	roles, err := repo.Query("")
	require.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRound8A_Role_Query_WithKeyword(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "alpha_role", RoleCode: "ca", Status: 1}))
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "beta_role", RoleCode: "cb", Status: 1}))
	roles, err := repo.Query("alpha")
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, "alpha_role", roles[0].RoleName)
}

func TestRound8A_Role_Query_EmptyResult(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	roles, err := repo.Query("nonexistent")
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestRound8A_Role_CountByRoleCode(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "n1", RoleCode: "CNT_CODE", Status: role.StatusEnabled}))
	count, err := repo.CountByRoleCode("CNT_CODE")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRound8A_Role_CountByRoleCode_Zero(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	count, err := repo.CountByRoleCode("NO_SUCH_CODE")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRound8A_Role_CountUserRoles(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.BindUserRole(100, 1, 0))
	require.NoError(t, repo.BindUserRole(100, 2, 0))
	count, err := repo.CountUserRoles(100)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRound8A_Role_CountUserRoles_Zero(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	count, err := repo.CountUserRoles(999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRound8A_Role_GetUserRoleIDs(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.BindUserRole(200, 10, 0))
	require.NoError(t, repo.BindUserRole(200, 20, 0))
	ids, err := repo.GetUserRoleIDs(200)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestRound8A_Role_GetUserRoleIDs_Empty(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	ids, err := repo.GetUserRoleIDs(999)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestRound8A_Role_BindUserRole(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.BindUserRole(300, 5, 1))
	count, _ := repo.CountUserRoles(300)
	assert.Equal(t, int64(1), count)
}

func TestRound8A_Role_UnbindUserRole(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.BindUserRole(400, 7, 0))
	require.NoError(t, repo.UnbindUserRole(400, 7, 0))
	count, _ := repo.CountUserRoles(400)
	assert.Equal(t, int64(0), count)
}

func TestRound8A_Role_UnbindUserRole_WithOrgID(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.BindUserRole(401, 8, 10))
	require.NoError(t, repo.UnbindUserRole(401, 8, 10))
	count, _ := repo.CountUserRoles(401)
	assert.Equal(t, int64(0), count)
}

func TestRound8A_Role_CountUserRolesByOrg(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.BindUserRole(500, 1, 10))
	require.NoError(t, repo.BindUserRole(500, 2, 20))
	count, err := repo.CountUserRolesByOrg(500, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRound8A_Role_CountUserRolesByOrg_All(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.BindUserRole(501, 1, 10))
	require.NoError(t, repo.BindUserRole(501, 2, 20))
	// orgID=0 means no org filter → count all
	count, err := repo.CountUserRolesByOrg(501, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRound8A_Role_GetRolesByIDs(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	r1 := &role.SysRole{RoleName: "g1", RoleCode: "GC1", Status: role.StatusEnabled}
	r2 := &role.SysRole{RoleName: "g2", RoleCode: "GC2", Status: role.StatusEnabled}
	require.NoError(t, repo.Create(r1))
	require.NoError(t, repo.Create(r2))
	roles, err := repo.GetRolesByIDs([]int64{r1.RoleID, r2.RoleID})
	require.NoError(t, err)
	assert.Len(t, roles, 2)
}

func TestRound8A_Role_GetRolesByIDs_Empty(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	roles, err := repo.GetRolesByIDs([]int64{})
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestRound8A_Role_QueryByOrgID(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	// Create org-scoped role
	orgID := int64(10)
	rt := role.RoleTypeOrganization
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "org_role", RoleCode: "OR1", OrgID: &orgID, RoleType: &rt, Status: role.StatusEnabled}))
	// Bind a user-role for this org
	require.NoError(t, repo.BindUserRole(600, 1, orgID))
	roles, err := repo.QueryByOrgID(orgID, "")
	require.NoError(t, err)
	assert.NotEmpty(t, roles)
}

func TestRound8A_Role_QueryByOrgID_WithKeyword(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	orgID := int64(11)
	rt := role.RoleTypeOrganization
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "alpha_org", RoleCode: "OR2", OrgID: &orgID, RoleType: &rt, Status: role.StatusEnabled}))
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "beta_org", RoleCode: "OR3", OrgID: &orgID, RoleType: &rt, Status: role.StatusEnabled}))
	roles, err := repo.QueryByOrgID(orgID, "alpha")
	require.NoError(t, err)
	assert.Len(t, roles, 1)
	assert.Equal(t, "alpha_org", roles[0].RoleName)
}

func TestRound8A_Role_QueryWithPage(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	for i := 0; i < 5; i++ {
		require.NoError(t, repo.Create(&role.SysRole{RoleName: "page_role", RoleCode: "PR1", Status: role.StatusEnabled}))
	}
	roles, total, err := repo.QueryWithPage("", "", 1, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, roles, 3)
}

func TestRound8A_Role_QueryWithPage_Page2(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	for i := 0; i < 4; i++ {
		require.NoError(t, repo.Create(&role.SysRole{RoleName: "p2_role", RoleCode: "P2", Status: role.StatusEnabled}))
	}
	roles, total, err := repo.QueryWithPage("", "", 2, 3)
	require.NoError(t, err)
	assert.Equal(t, int64(4), total)
	assert.Len(t, roles, 1)
}

func TestRound8A_Role_QueryWithPage_WithKeyword(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "kw_alpha", RoleCode: "KWA", Status: role.StatusEnabled}))
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "kw_beta", RoleCode: "KWB", Status: role.StatusEnabled}))
	roles, total, err := repo.QueryWithPage("kw_alpha", "", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, roles, 1)
}

func TestRound8A_Role_QueryWithPage_WithRoleType(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	rt := role.RoleTypeSystem
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "sys", RoleCode: "SYS1", RoleType: &rt, Status: role.StatusEnabled}))
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "custom", RoleCode: "CUS1", Status: role.StatusEnabled}))
	roles, total, err := repo.QueryWithPage("", "system", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, roles, 1)
	assert.Equal(t, "sys", roles[0].RoleName)
}

func TestRound8A_Role_QueryWithPage_DefaultPaging(t *testing.T) {
	db := round8RoleDB(t)
	repo := NewRoleRepository(db)
	require.NoError(t, repo.Create(&role.SysRole{RoleName: "dp", RoleCode: "DP", Status: role.StatusEnabled}))
	roles, total, err := repo.QueryWithPage("", "", 0, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, roles, 1)
}

// ==================== RoleMenuRepository (11 functions) ====================

func TestRound8A_RoleMenu_New(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NotNil(t, repo)
}

func TestRound8A_RoleMenu_GetMenuIDsByRoleID(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 1, MenuID: 10}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 1, MenuID: 20}).Error)
	ids, err := repo.GetMenuIDsByRoleID(1)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestRound8A_RoleMenu_GetMenuIDsByRoleID_Empty(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	ids, err := repo.GetMenuIDsByRoleID(999)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestRound8A_RoleMenu_GetMenuIDsByRoleIDs(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 1, MenuID: 10}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 2, MenuID: 20}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 2, MenuID: 10}).Error)
	ids, err := repo.GetMenuIDsByRoleIDs([]int64{1, 2})
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestRound8A_RoleMenu_GetMenuIDsByRoleIDs_Empty(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	ids, err := repo.GetMenuIDsByRoleIDs([]int64{})
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestRound8A_RoleMenu_GetByRoleID(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 5, MenuID: 50}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 5, MenuID: 60}).Error)
	items, err := repo.GetByRoleID(5)
	require.NoError(t, err)
	assert.Len(t, items, 2)
}

func TestRound8A_RoleMenu_GetByRoleID_Empty(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	items, err := repo.GetByRoleID(999)
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestRound8A_RoleMenu_SaveRoleMenus(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, repo.SaveRoleMenus(10, []int64{100, 200, 300}))
	ids, err := repo.GetMenuIDsByRoleID(10)
	require.NoError(t, err)
	assert.Len(t, ids, 3)
}

func TestRound8A_RoleMenu_SaveRoleMenus_ClearAll(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, repo.SaveRoleMenus(11, []int64{101, 201}))
	require.NoError(t, repo.SaveRoleMenus(11, []int64{}))
	ids, err := repo.GetMenuIDsByRoleID(11)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestRound8A_RoleMenu_SaveRoleMenus_Replace(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, repo.SaveRoleMenus(12, []int64{111, 222}))
	require.NoError(t, repo.SaveRoleMenus(12, []int64{333}))
	ids, err := repo.GetMenuIDsByRoleID(12)
	require.NoError(t, err)
	assert.Len(t, ids, 1)
	assert.Equal(t, int64(333), ids[0])
}

func TestRound8A_RoleMenu_UpsertRoleMenus(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, repo.UpsertRoleMenus(20, []int64{400, 500}))
	ids, err := repo.GetMenuIDsByRoleID(20)
	require.NoError(t, err)
	assert.Len(t, ids, 2)
}

func TestRound8A_RoleMenu_UpsertRoleMenus_ClearOnEmpty(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, repo.UpsertRoleMenus(21, []int64{600}))
	require.NoError(t, repo.UpsertRoleMenus(21, []int64{}))
	ids, err := repo.GetMenuIDsByRoleID(21)
	require.NoError(t, err)
	assert.Empty(t, ids)
}

func TestRound8A_RoleMenu_DeleteByRoleID(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 30, MenuID: 700}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 30, MenuID: 800}).Error)
	require.NoError(t, repo.DeleteByRoleID(30))
	ids, _ := repo.GetMenuIDsByRoleID(30)
	assert.Empty(t, ids)
}

func TestRound8A_RoleMenu_DeleteByMenuID(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 40, MenuID: 900}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 41, MenuID: 900}).Error)
	require.NoError(t, repo.DeleteByMenuID(900))
	items, _ := repo.GetByRoleID(40)
	assert.Empty(t, items)
	items2, _ := repo.GetByRoleID(41)
	assert.Empty(t, items2)
}

func TestRound8A_RoleMenu_IsMenuAuthorizedForRole(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 50, MenuID: 1000}).Error)
	ok, err := repo.IsMenuAuthorizedForRole(50, 1000)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRound8A_RoleMenu_IsMenuAuthorizedForRole_False(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	ok, err := repo.IsMenuAuthorizedForRole(50, 9999)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRound8A_RoleMenu_IsMenuAuthorizedForRoles(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 60, MenuID: 1100}).Error)
	ok, err := repo.IsMenuAuthorizedForRoles([]int64{60, 61}, 1100)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRound8A_RoleMenu_IsMenuAuthorizedForRoles_Empty(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	ok, err := repo.IsMenuAuthorizedForRoles([]int64{}, 1100)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRound8A_RoleMenu_GetRoleMenuCount(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 70, MenuID: 1200}).Error)
	require.NoError(t, db.Create(&role.RoleMenu{RoleID: 70, MenuID: 1201}).Error)
	count, err := repo.GetRoleMenuCount(70)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)
}

func TestRound8A_RoleMenu_GetRoleMenuCount_Zero(t *testing.T) {
	db := round8RoleMenuDB(t)
	repo := NewRoleMenuRepository(db)
	count, err := repo.GetRoleMenuCount(999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// ==================== OrgRepository (13 functions) ====================

func TestRound8A_Org_New(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	require.NotNil(t, repo)
}

func TestRound8A_Org_DB(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	assert.Equal(t, db, repo.DB())
}

func TestRound8A_Org_Create(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	o := &org.SysOrg{OrgName: "root_org", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(o))
	assert.Positive(t, o.OrgID)
}

func TestRound8A_Org_Update(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	o := &org.SysOrg{OrgName: "old_name", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(o))
	o.OrgName = "new_name"
	require.NoError(t, repo.Update(o))
	got, err := repo.GetByID(o.OrgID)
	require.NoError(t, err)
	assert.Equal(t, "new_name", got.OrgName)
}

func TestRound8A_Org_Delete(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	o := &org.SysOrg{OrgName: "del_org", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(o))
	require.NoError(t, repo.Delete(o.OrgID))
	_, err := repo.GetByID(o.OrgID)
	assert.Error(t, err) // GetByID filters by del_flag=0
}

func TestRound8A_Org_GetByID(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	o := &org.SysOrg{OrgName: "find_org", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(o))
	got, err := repo.GetByID(o.OrgID)
	require.NoError(t, err)
	assert.Equal(t, "find_org", got.OrgName)
}

func TestRound8A_Org_GetByID_NotFound(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	_, err := repo.GetByID(99999)
	assert.Error(t, err)
}

func TestRound8A_Org_GetByName(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	o := &org.SysOrg{OrgName: "unique_org", ParentID: 0, Level: 1, Status: org.StatusEnabled, DelFlag: org.DelFlagNormal}
	require.NoError(t, repo.Create(o))
	got, err := repo.GetByName("unique_org")
	require.NoError(t, err)
	assert.Equal(t, o.OrgID, got.OrgID)
}

func TestRound8A_Org_GetByName_NotFound(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	_, err := repo.GetByName("missing_org")
	assert.Error(t, err)
}

func TestRound8A_Org_List(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	require.NoError(t, repo.Create(&org.SysOrg{OrgName: "o1", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}))
	require.NoError(t, repo.Create(&org.SysOrg{OrgName: "o2", ParentID: 0, Level: 2, Status: 1, DelFlag: 0}))
	orgs, err := repo.List()
	require.NoError(t, err)
	assert.Len(t, orgs, 2)
}

func TestRound8A_Org_List_Empty(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	orgs, err := repo.List()
	require.NoError(t, err)
	assert.Empty(t, orgs)
}

func TestRound8A_Org_ListByParentID(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	parent := &org.SysOrg{OrgName: "parent", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(parent))
	require.NoError(t, repo.Create(&org.SysOrg{OrgName: "child1", ParentID: parent.OrgID, Level: 2, Status: 1, DelFlag: 0}))
	require.NoError(t, repo.Create(&org.SysOrg{OrgName: "child2", ParentID: parent.OrgID, Level: 2, Status: 1, DelFlag: 0}))
	children, err := repo.ListByParentID(parent.OrgID)
	require.NoError(t, err)
	assert.Len(t, children, 2)
}

func TestRound8A_Org_ListByParentID_Empty(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	children, err := repo.ListByParentID(999)
	require.NoError(t, err)
	assert.Empty(t, children)
}

func TestRound8A_Org_CheckNameExists(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	require.NoError(t, repo.Create(&org.SysOrg{OrgName: "exist_name", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}))
	count, err := repo.CheckNameExists("exist_name", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRound8A_Org_CheckNameExists_ExcludeSelf(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	o := &org.SysOrg{OrgName: "self_check", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(o))
	count, err := repo.CheckNameExists("self_check", o.OrgID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRound8A_Org_CheckNameExists_NotFound(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	count, err := repo.CheckNameExists("no_such_name", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRound8A_Org_CountChildren(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	parent := &org.SysOrg{OrgName: "pc", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(parent))
	require.NoError(t, repo.Create(&org.SysOrg{OrgName: "cc1", ParentID: parent.OrgID, Level: 2, Status: 1, DelFlag: 0}))
	count, err := repo.CountChildren(parent.OrgID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRound8A_Org_CountChildren_Zero(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	count, err := repo.CountChildren(999)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRound8A_Org_GetByIDs(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	o1 := &org.SysOrg{OrgName: "bid1", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	o2 := &org.SysOrg{OrgName: "bid2", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(o1))
	require.NoError(t, repo.Create(o2))
	orgs, err := repo.GetByIDs([]int64{o1.OrgID, o2.OrgID})
	require.NoError(t, err)
	assert.Len(t, orgs, 2)
}

func TestRound8A_Org_GetByIDs_Empty(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	orgs, err := repo.GetByIDs([]int64{})
	require.NoError(t, err)
	assert.Empty(t, orgs)
}

func TestRound8A_Org_IsDescendant_True(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	root := &org.SysOrg{OrgName: "desc_root", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(root))
	child := &org.SysOrg{OrgName: "desc_child", ParentID: root.OrgID, Level: 2, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(child))
	grandchild := &org.SysOrg{OrgName: "desc_gc", ParentID: child.OrgID, Level: 3, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(grandchild))
	ok, err := repo.IsDescendant(root.OrgID, grandchild.OrgID)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRound8A_Org_IsDescendant_False(t *testing.T) {
	db := round8OrgDB(t)
	repo := NewOrgRepository(db)
	root := &org.SysOrg{OrgName: "desc_root", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(root))
	other := &org.SysOrg{OrgName: "desc_other", ParentID: 0, Level: 1, Status: 1, DelFlag: 0}
	require.NoError(t, repo.Create(other))
	ok, err := repo.IsDescendant(root.OrgID, other.OrgID)
	require.NoError(t, err)
	assert.False(t, ok)
}

// ==================== MenuRepository (11 functions) ====================

func TestRound8A_Menu_New(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	require.NotNil(t, repo)
}

func TestRound8A_Menu_GetAll(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	require.NoError(t, repo.Create(&menu.CoreMenu{Name: "m1", Path: "/m1", MenuSort: 1}))
	require.NoError(t, repo.Create(&menu.CoreMenu{Name: "m2", Path: "/m2", MenuSort: 2}))
	menus, err := repo.GetAll()
	require.NoError(t, err)
	assert.Len(t, menus, 2)
}

func TestRound8A_Menu_GetAll_Empty(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	menus, err := repo.GetAll()
	require.NoError(t, err)
	assert.Empty(t, menus)
}

func TestRound8A_Menu_GetByIDs(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m1 := &menu.CoreMenu{Name: "bid_m1", Path: "/bid1", MenuSort: 1}
	m2 := &menu.CoreMenu{Name: "bid_m2", Path: "/bid2", MenuSort: 2}
	require.NoError(t, repo.Create(m1))
	require.NoError(t, repo.Create(m2))
	menus, err := repo.GetByIDs([]int64{m1.ID, m2.ID})
	require.NoError(t, err)
	assert.Len(t, menus, 2)
}

func TestRound8A_Menu_GetByIDs_Empty(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	menus, err := repo.GetByIDs([]int64{})
	require.NoError(t, err)
	assert.Empty(t, menus)
}

func TestRound8A_Menu_GetByID(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m := &menu.CoreMenu{Name: "find_m", Path: "/find", MenuSort: 1}
	require.NoError(t, repo.Create(m))
	got, err := repo.GetByID(m.ID)
	require.NoError(t, err)
	assert.Equal(t, "find_m", got.Name)
}

func TestRound8A_Menu_GetByID_NotFound(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	_, err := repo.GetByID(99999)
	assert.Error(t, err)
}

func TestRound8A_Menu_GetByPath(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	require.NoError(t, repo.Create(&menu.CoreMenu{Name: "path_m", Path: "/unique/path", MenuSort: 1}))
	got, err := repo.GetByPath("/unique/path")
	require.NoError(t, err)
	assert.Equal(t, "path_m", got.Name)
}

func TestRound8A_Menu_GetByPath_NotFound(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	_, err := repo.GetByPath("/missing")
	assert.Error(t, err)
}

func TestRound8A_Menu_Create(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m := &menu.CoreMenu{Name: "create_m", Path: "/create", MenuSort: 1}
	require.NoError(t, repo.Create(m))
	assert.Positive(t, m.ID)
}

func TestRound8A_Menu_Update(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m := &menu.CoreMenu{Name: "old_m", Path: "/update", MenuSort: 1}
	require.NoError(t, repo.Create(m))
	m.Name = "updated_m"
	require.NoError(t, repo.Update(m))
	got, err := repo.GetByID(m.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated_m", got.Name)
}

func TestRound8A_Menu_Delete(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m := &menu.CoreMenu{Name: "del_m", Path: "/del", MenuSort: 1}
	require.NoError(t, repo.Create(m))
	require.NoError(t, repo.Delete(m.ID))
	_, err := repo.GetByID(m.ID)
	assert.Error(t, err)
}

func TestRound8A_Menu_Delete_NotFound(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	require.NoError(t, repo.Delete(99999))
}

func TestRound8A_Menu_HasChildren_True(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	parent := &menu.CoreMenu{Name: "parent", Path: "/parent", MenuSort: 1}
	require.NoError(t, repo.Create(parent))
	require.NoError(t, repo.Create(&menu.CoreMenu{Name: "child", Path: "/child", MenuSort: 1, Pid: parent.ID}))
	ok, err := repo.HasChildren(parent.ID)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRound8A_Menu_HasChildren_False(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m := &menu.CoreMenu{Name: "leaf", Path: "/leaf", MenuSort: 1}
	require.NoError(t, repo.Create(m))
	ok, err := repo.HasChildren(m.ID)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRound8A_Menu_UpdateSort(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m := &menu.CoreMenu{Name: "sort_m", Path: "/sort", MenuSort: 1}
	require.NoError(t, repo.Create(m))
	require.NoError(t, repo.UpdateSort(m.ID, 99))
	got, err := repo.GetByID(m.ID)
	require.NoError(t, err)
	assert.Equal(t, 99, got.MenuSort)
}

func TestRound8A_Menu_UpdateHidden(t *testing.T) {
	db := round8MenuDB(t)
	repo := NewMenuRepository(db)
	m := &menu.CoreMenu{Name: "hide_m", Path: "/hide", MenuSort: 1}
	require.NoError(t, repo.Create(m))
	require.NoError(t, repo.UpdateHidden(m.ID, true))
	got, err := repo.GetByID(m.ID)
	require.NoError(t, err)
	assert.True(t, got.Hidden)
}
