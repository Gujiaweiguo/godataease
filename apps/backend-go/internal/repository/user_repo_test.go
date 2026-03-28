package repository

import (
	"fmt"
	"testing"
	"time"

	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserRepositoryTest(t *testing.T) (*UserRepository, *UserRoleRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{}, &role.SysRole{}))
	return NewUserRepository(db), NewUserRoleRepository(db), db
}

func TestUserRepository_Query_Unit(t *testing.T) {
	t.Run("caps page size at 100 and orders by create time desc", func(t *testing.T) {
		repo, _, db := setupUserRepositoryTest(t)
		base := time.Now()
		for i := range 105 {
			u := &user.SysUser{
				Username:   fmt.Sprintf("bulk-%03d", i),
				NickName:   fmt.Sprintf("Bulk %03d", i),
				Status:     user.StatusEnabled,
				DelFlag:    user.DelFlagNormal,
				CreateTime: base.Add(time.Duration(i) * time.Minute),
			}
			require.NoError(t, db.Create(u).Error)
		}

		rows, total, err := repo.Query(&user.UserQueryRequest{Current: 1, Size: 200})
		require.NoError(t, err)
		assert.Equal(t, int64(105), total)
		require.Len(t, rows, 100)
		assert.Equal(t, "bulk-104", rows[0].Username)
		assert.Equal(t, "bulk-005", rows[99].Username)
	})

	t.Run("filters by keyword status and org", func(t *testing.T) {
		repo, userRoleRepo, db := setupUserRepositoryTest(t)
		matchingEmail := "alice@example.com"
		require.NoError(t, db.Create(&user.SysUser{UserID: 1, Username: "alice", NickName: "Alice", Email: &matchingEmail, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Create(&user.SysUser{UserID: 2, Username: "alicia", NickName: "Other Alice", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Create(&user.SysUser{UserID: 3, Username: "bob", NickName: "Bob", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 1, RoleID: 1, OrgID: 7}))
		require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 2, RoleID: 1, OrgID: 8}))
		require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 3, RoleID: 1, OrgID: 8}))

		keyword := "ali"
		status := user.StatusEnabled
		orgID := int64(7)
		rows, total, err := repo.Query(&user.UserQueryRequest{Keyword: &keyword, Status: &status, OrgID: &orgID, Current: 0, Size: 0})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		require.Len(t, rows, 1)
		assert.Equal(t, int64(1), rows[0].UserID)
	})
}

func TestUserRepository_CRUDAndLookupHelpers(t *testing.T) {
	repo, _, db := setupUserRepositoryTest(t)
	mail := "lookup@example.com"
	u := &user.SysUser{Username: "lookup-user", NickName: "Lookup User", Email: &mail, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal, CreateTime: time.Now()}
	require.NoError(t, repo.Create(u))
	require.Positive(t, u.UserID)

	loaded, err := repo.GetByID(u.UserID)
	require.NoError(t, err)
	assert.Equal(t, "lookup-user", loaded.Username)

	loaded, err = repo.GetByUsername("lookup-user")
	require.NoError(t, err)
	assert.Equal(t, u.UserID, loaded.UserID)

	u.NickName = "Updated Nick"
	require.NoError(t, repo.Update(u))
	loaded, err = repo.GetByID(u.UserID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Nick", loaded.NickName)

	count, err := repo.CountByUsername("lookup-user")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	deletedUser := &user.SysUser{Username: "deleted-user", Status: user.StatusEnabled, DelFlag: user.DelFlagDeleted}
	require.NoError(t, db.Create(deletedUser).Error)
	count, err = repo.CountByUsername("deleted-user")
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	active2 := &user.SysUser{Username: "listed-user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}
	require.NoError(t, repo.Create(active2))
	users, err := repo.ListUsersByIds([]int64{u.UserID, active2.UserID, deletedUser.UserID})
	require.NoError(t, err)
	require.Len(t, users, 2)

	require.NoError(t, repo.Delete(u.UserID))
	_, err = repo.GetByID(u.UserID)
	require.Error(t, err)
	_, err = repo.GetByUsername("lookup-user")
	require.Error(t, err)
}

func TestUserRepository_CheckEmailExists_Unit(t *testing.T) {
	repo, _, db := setupUserRepositoryTest(t)
	keepEmail := "exists@example.com"
	deletedEmail := "deleted@example.com"
	require.NoError(t, db.Create(&user.SysUser{UserID: 1, Username: "exists", Email: &keepEmail, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, db.Create(&user.SysUser{UserID: 2, Username: "deleted", Email: &deletedEmail, Status: user.StatusEnabled, DelFlag: user.DelFlagDeleted}).Error)

	exists, err := repo.CheckEmailExists("exists@example.com", 0)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.CheckEmailExists("exists@example.com", 1)
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = repo.CheckEmailExists("deleted@example.com", 0)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepository_SearchExternalUser(t *testing.T) {
	repo, userRoleRepo, db := setupUserRepositoryTest(t)
	mailA := "alice@example.com"
	mailB := "bob@example.com"
	mailC := "carol@example.com"
	require.NoError(t, db.Create(&user.SysUser{UserID: 11, Username: "alice", NickName: "Alice", Email: &mailA, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, db.Create(&user.SysUser{UserID: 12, Username: "bob", NickName: "Bob", Email: &mailB, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, db.Create(&user.SysUser{UserID: 13, Username: "carol", NickName: "Carol", Email: &mailC, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 12, RoleID: 1, OrgID: 9}))

	rows, err := repo.SearchExternalUser("   ", 9)
	require.NoError(t, err)
	assert.Empty(t, rows)

	rows, err = repo.SearchExternalUser("11", 9)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(11), rows[0].UserID)

	rows, err = repo.SearchExternalUser("bob@example.com", 9)
	require.NoError(t, err)
	assert.Empty(t, rows)

	rows, err = repo.SearchExternalUser("carol@example.com", 9)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(13), rows[0].UserID)
}

func TestUserRoleRepository_RelationsAndCounts(t *testing.T) {
	userRepo, userRoleRepo, db := setupUserRepositoryTest(t)
	require.NoError(t, db.Create(&role.SysRole{RoleID: 21, RoleName: "Enabled Role", RoleCode: "enabled", Status: role.StatusEnabled}).Error)
	require.NoError(t, db.Create(&role.SysRole{RoleID: 22, RoleName: "Disabled Role", RoleCode: "disabled", Status: 2}).Error)
	require.NoError(t, db.Create(&user.SysUser{UserID: 31, Username: "u1", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	require.NoError(t, db.Create(&user.SysUser{UserID: 32, Username: "u2", Status: user.StatusEnabled, DelFlag: user.DelFlagDeleted}).Error)

	created, err := userRoleRepo.CreateIfMissing(&user.SysUserRole{UserID: 31, RoleID: 21, OrgID: 7})
	require.NoError(t, err)
	assert.True(t, created)

	created, err = userRoleRepo.CreateIfMissing(&user.SysUserRole{UserID: 31, RoleID: 21, OrgID: 7})
	require.NoError(t, err)
	assert.False(t, created)

	exists, err := userRoleRepo.Exists(31, 21, 7)
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 31, RoleID: 22, OrgID: 7}))
	require.NoError(t, userRoleRepo.Create(&user.SysUserRole{UserID: 32, RoleID: 21, OrgID: 7}))

	roleIDs, err := userRoleRepo.GetRoleIDsByUserID(31)
	require.NoError(t, err)
	assert.Equal(t, []int64{21}, roleIDs)

	count, err := userRepo.CountByOrgID(7)
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	inOrg, err := userRoleRepo.IsUserInOrg(31, 7)
	require.NoError(t, err)
	assert.True(t, inOrg)

	roles, err := userRoleRepo.GetByUserID(31)
	require.NoError(t, err)
	require.Len(t, roles, 2)

	require.NoError(t, userRoleRepo.DeleteByUserID(31))
	roles, err = userRoleRepo.GetByUserID(31)
	require.NoError(t, err)
	assert.Empty(t, roles)
}

func TestUserPermRepository_Wrappers(t *testing.T) {
	_, _, db := setupUserRepositoryTest(t)
	permRepo := NewUserPermRepository(db)
	require.NotNil(t, permRepo)

	require.NoError(t, permRepo.Create(&user.SysUserPerm{UserID: 41, OrgID: 7, PermID: 100, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))
	require.NoError(t, permRepo.Create(&user.SysUserPerm{UserID: 41, OrgID: 7, PermID: 101, Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}))

	perms, err := permRepo.GetByUserID(41)
	require.NoError(t, err)
	require.Len(t, perms, 2)

	require.NoError(t, permRepo.DeleteByUserID(41))
	perms, err = permRepo.GetByUserID(41)
	require.NoError(t, err)
	assert.Empty(t, perms)
}
