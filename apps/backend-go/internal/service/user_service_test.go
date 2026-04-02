package service

import (
	"path/filepath"
	"testing"
	"time"

	"dataease/backend/internal/domain/audit"
	domainorg "dataease/backend/internal/domain/org"
	domainrole "dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupUserServiceRepoTest(t *testing.T) (*UserService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.SysUser{}))

	userRepo := repository.NewUserRepository(db)
	return NewUserService(userRepo, nil, nil), db
}

func setupUserServiceOrgBindingTest(t *testing.T) (*UserService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.SysUser{}, &domainorg.SysOrg{}, &domainrole.SysRole{}, &user.SysUserRole{}))

	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	orgRepo := repository.NewOrgRepository(db)

	svc := NewUserService(userRepo, userRoleRepo, nil)
	svc.SetRoleRepository(roleRepo)
	svc.SetOrgRepository(orgRepo)

	return svc, db
}

func setupUserServiceFullRepoTest(t *testing.T) (*UserService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.SysUser{}, &user.SysUserRole{}, &user.SysUserPerm{}))

	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	userPermRepo := repository.NewUserPermRepository(db)

	return NewUserService(userRepo, userRoleRepo, userPermRepo), db
}

func setupUserServiceWithAuditTest(t *testing.T) (*UserService, *gorm.DB) {
	t.Helper()

	dbFile := filepath.Join(t.TempDir(), "test.db")
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&user.SysUser{}, &audit.AuditLog{}))

	userRepo := repository.NewUserRepository(db)
	auditLogRepo := repository.NewAuditLogRepository(db)
	auditSvc := NewAuditService(auditLogRepo, repository.NewLoginFailureRepository(db), repository.NewAuditLogDetailRepository(db))

	svc := NewUserService(userRepo, nil, nil)
	svc.SetAuditService(auditSvc)

	return svc, db
}

func TestUserService_ResolveDefaultPassword(t *testing.T) {
	svc := &UserService{}

	t.Setenv(DefaultPasswordEnvName, "")
	assert.Equal(t, FallbackDefaultPwd, svc.ResolveDefaultPassword())

	t.Setenv(DefaultPasswordEnvName, "custom-password")
	assert.Equal(t, "custom-password", svc.ResolveDefaultPassword())
}

func TestUserServiceHelpers_NormalizeLanguage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "zh cn", input: "zh-cn", want: defaultLanguageZhCN},
		{name: "zh underscore", input: "zh_CN", want: defaultLanguageZhCN},
		{name: "zh short uppercase", input: "ZH", want: defaultLanguageZhCN},
		{name: "tw locale", input: "zh-TW", want: "tw"},
		{name: "tw short", input: "tw", want: "tw"},
		{name: "en short", input: "en", want: "en"},
		{name: "en us", input: "en-US", want: "en"},
		{name: "empty fallback", input: "", want: defaultLanguageZhCN},
		{name: "unknown fallback", input: "fr", want: defaultLanguageZhCN},
		{name: "trim fallback", input: " ja ", want: defaultLanguageZhCN},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, normalizeLanguage(tt.input))
		})
	}
}

func TestUserServiceHelpers_RequestedOrgID(t *testing.T) {
	five := int64(5)
	three := int64(3)
	zero := int64(0)
	negative := int64(-1)

	tests := []struct {
		name           string
		orgID          *int64
		organizationID *int64
		want           int64
		ok             bool
	}{
		{name: "prefers organization id", orgID: &three, organizationID: &five, want: 5, ok: true},
		{name: "falls back to org id", orgID: &three, organizationID: nil, want: 3, ok: true},
		{name: "ignores zero organization id", orgID: &three, organizationID: &zero, want: 3, ok: true},
		{name: "rejects negative org id", orgID: &negative, organizationID: nil, want: 0, ok: false},
		{name: "none provided", orgID: nil, organizationID: nil, want: 0, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := requestedOrgID(tt.orgID, tt.organizationID)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUserServiceHelpers_PtrUserStr(t *testing.T) {
	value := ptrUserStr("hello")
	require.NotNil(t, value)
	assert.Equal(t, "hello", *value)
}

func TestUserService_SetAuditService(t *testing.T) {
	svc, db := setupUserServiceWithAuditTest(t)

	require.NotNil(t, svc.auditSvc)

	auditLogRepo := repository.NewAuditLogRepository(db)
	replacement := NewAuditService(auditLogRepo, repository.NewLoginFailureRepository(db), repository.NewAuditLogDetailRepository(db))
	svc.SetAuditService(replacement)

	assert.Same(t, replacement, svc.auditSvc)
}

func TestUserService_UpdateUserStatus(t *testing.T) {
	t.Run("returns not found when user missing", func(t *testing.T) {
		svc, _ := setupUserServiceRepoTest(t)

		err := svc.UpdateUserStatus(999, user.StatusDisabled)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("persists new status", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{UserID: 301, Username: "status-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		var existing user.SysUser
		require.NoError(t, db.Where("username = ?", "status-user").First(&existing).Error)
		err := svc.UpdateUserStatus(existing.UserID, user.StatusDisabled)
		require.NoError(t, err)
		var updated user.SysUser
		require.NoError(t, db.First(&updated, "user_id = ?", existing.UserID).Error)
		assert.Equal(t, user.StatusDisabled, updated.Status)
		assert.NotNil(t, updated.UpdateTime)
	})

	t.Run("returns wrapped update error", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{UserID: 302, Username: "status-update-error", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		var existing user.SysUser
		require.NoError(t, db.Where("username = ?", "status-update-error").First(&existing).Error)

		require.NoError(t, db.Exec("CREATE TRIGGER deny_user_status_update BEFORE UPDATE ON sys_user BEGIN SELECT RAISE(FAIL, 'deny user update'); END;").Error)

		err := svc.UpdateUserStatus(existing.UserID, user.StatusDisabled)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update user status")
	})
}

func TestUserService_EnsureUserInOrg(t *testing.T) {
	t.Run("returns nil when user belongs to org", func(t *testing.T) {
		svc, db := setupUserServiceFullRepoTest(t)
		require.NoError(t, db.Create(&user.SysUserRole{UserID: 301, RoleID: 1, OrgID: 7}).Error)

		err := svc.EnsureUserInOrg(301, 7)
		require.NoError(t, err)
	})

	t.Run("returns error when user does not belong to org", func(t *testing.T) {
		svc, db := setupUserServiceFullRepoTest(t)
		require.NoError(t, db.Create(&user.SysUserRole{UserID: 302, RoleID: 1, OrgID: 8}).Error)

		err := svc.EnsureUserInOrg(302, 7)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotInCurrentOrg)
	})

	t.Run("rejects missing org context", func(t *testing.T) {
		svc, _ := setupUserServiceFullRepoTest(t)

		err := svc.EnsureUserInOrg(302, 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "org id is required")
	})
}

func TestUserService_SwitchLanguage(t *testing.T) {
	t.Run("returns not found when user missing", func(t *testing.T) {
		svc, _ := setupUserServiceRepoTest(t)

		err := svc.SwitchLanguage(999, "en")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("normalizes and persists language", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{Username: "lang-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		var existing user.SysUser
		require.NoError(t, db.Where("username = ?", "lang-user").First(&existing).Error)

		err := svc.SwitchLanguage(existing.UserID, "en-US")
		require.NoError(t, err)

		var updated user.SysUser
		require.NoError(t, db.First(&updated, "user_id = ?", existing.UserID).Error)
		require.NotNil(t, updated.Language)
		assert.Equal(t, "en", *updated.Language)
		assert.NotNil(t, updated.UpdateTime)
	})

	t.Run("returns wrapped update error and normalizes fallback language", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{Username: "lang-update-error", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		var existing user.SysUser
		require.NoError(t, db.Where("username = ?", "lang-update-error").First(&existing).Error)

		require.NoError(t, svc.SwitchLanguage(existing.UserID, "unknown-lang"))
		var updated user.SysUser
		require.NoError(t, db.First(&updated, "user_id = ?", existing.UserID).Error)
		require.NotNil(t, updated.Language)
		assert.Equal(t, "zh-CN", *updated.Language)

		require.NoError(t, db.Exec("CREATE TRIGGER deny_user_language_update BEFORE UPDATE ON sys_user BEGIN SELECT RAISE(FAIL, 'deny language update'); END;").Error)

		err := svc.SwitchLanguage(existing.UserID, "tw")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to switch user language")
	})
}

func TestUserService_RecordPasswordResetAudit(t *testing.T) {
	t.Run("no-op when audit service is nil", func(t *testing.T) {
		svc, _ := setupUserServiceRepoTest(t)

		assert.NotPanics(t, func() {
			svc.recordPasswordResetAudit(&user.SysUser{Username: "alice"}, 1, 2, "admin", "127.0.0.1", audit.StatusSuccess, "")
		})
	})

	t.Run("writes success audit log asynchronously", func(t *testing.T) {
		svc, db := setupUserServiceWithAuditTest(t)
		targetUser := &user.SysUser{Username: "alice"}

		svc.recordPasswordResetAudit(targetUser, 10, 20, "admin", "10.0.0.1", audit.StatusSuccess, "")

		require.Eventually(t, func() bool {
			var count int64
			if err := db.Model(&audit.AuditLog{}).Count(&count).Error; err != nil {
				return false
			}
			return count == 1
		}, time.Second, 20*time.Millisecond)

		var log audit.AuditLog
		require.NoError(t, db.First(&log).Error)
		require.NotNil(t, log.UserID)
		assert.Equal(t, int64(20), *log.UserID)
		require.NotNil(t, log.Username)
		assert.Equal(t, "admin", *log.Username)
		assert.Equal(t, audit.ActionTypeUserAction, log.ActionType)
		assert.Equal(t, "重置密码", log.ActionName)
		require.NotNil(t, log.ResourceType)
		assert.Equal(t, string(audit.ResourceTypeUser), *log.ResourceType)
		require.NotNil(t, log.ResourceID)
		assert.Equal(t, int64(10), *log.ResourceID)
		require.NotNil(t, log.ResourceName)
		assert.Equal(t, "alice", *log.ResourceName)
		assert.Equal(t, audit.OperationUpdate, log.Operation)
		require.NotNil(t, log.IPAddress)
		assert.Equal(t, "10.0.0.1", *log.IPAddress)
		assert.Equal(t, audit.StatusSuccess, log.Status)
		require.NotNil(t, log.BeforeValue)
		assert.Equal(t, "[REDACTED]", *log.BeforeValue)
		require.NotNil(t, log.AfterValue)
		assert.Equal(t, "password reset to default policy", *log.AfterValue)
		assert.Nil(t, log.FailureReason)
	})

	t.Run("writes failed audit log with failure reason", func(t *testing.T) {
		svc, db := setupUserServiceWithAuditTest(t)

		svc.recordPasswordResetAudit(nil, 11, 21, "operator", "10.0.0.2", audit.StatusFailed, "user not found")

		require.Eventually(t, func() bool {
			var count int64
			if err := db.Model(&audit.AuditLog{}).Count(&count).Error; err != nil {
				return false
			}
			return count == 1
		}, time.Second, 20*time.Millisecond)

		var log audit.AuditLog
		require.NoError(t, db.First(&log).Error)
		assert.Equal(t, audit.StatusFailed, log.Status)
		require.NotNil(t, log.FailureReason)
		assert.Equal(t, "user not found", *log.FailureReason)
		assert.Nil(t, log.ResourceName)
		assert.Nil(t, log.BeforeValue)
		assert.Nil(t, log.AfterValue)
	})
}

func TestUserService_ResetPasswordWithAudit_PersistsAuditLog(t *testing.T) {
	svc, db := setupUserServiceWithAuditTest(t)
	require.NoError(t, db.Create(&user.SysUser{UserID: 401, Username: "audit-user", Password: "old-secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
	var existing user.SysUser
	require.NoError(t, db.Where("username = ?", "audit-user").First(&existing).Error)
	err := svc.ResetPasswordWithAudit(existing.UserID, "new-secret", 99, "auditor", "192.168.1.10")
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		var count int64
		if err := db.Model(&audit.AuditLog{}).Count(&count).Error; err != nil {
			return false
		}
		return count == 1
	}, time.Second, 20*time.Millisecond)

	var log audit.AuditLog
	require.NoError(t, db.First(&log).Error)
	assert.Equal(t, audit.StatusSuccess, log.Status)
	require.NotNil(t, log.ResourceName)
	assert.Equal(t, "audit-user", *log.ResourceName)

	var updated user.SysUser
	require.NoError(t, db.First(&updated, "user_id = ?", existing.UserID).Error)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(updated.Password), []byte("new-secret")))
}

func TestUserService_ResetPasswordWithAudit_NotFoundWritesFailedAudit(t *testing.T) {
	svc, db := setupUserServiceWithAuditTest(t)

	err := svc.ResetPasswordWithAudit(999, "new-secret", 88, "auditor", "192.168.1.11")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")

	require.Eventually(t, func() bool {
		var count int64
		if err := db.Model(&audit.AuditLog{}).Count(&count).Error; err != nil {
			return false
		}
		return count == 1
	}, time.Second, 20*time.Millisecond)

	var log audit.AuditLog
	require.NoError(t, db.First(&log).Error)
	assert.Equal(t, audit.StatusFailed, log.Status)
	require.NotNil(t, log.FailureReason)
	assert.Equal(t, "user not found", *log.FailureReason)
	require.NotNil(t, log.ResourceID)
	assert.Equal(t, int64(999), *log.ResourceID)
}

func TestUserService_CreateAndUpdateUsers(t *testing.T) {
	t.Run("create user success hashes password and uses defaults", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		t.Setenv(DefaultPasswordEnvName, "")

		createdID, err := svc.CreateUser(&user.UserCreateRequest{
			Username: "creator-user",
			Password: svc.ResolveDefaultPassword(),
			RealName: "Creator User",
			Email:    ptrUserStr("creator@example.com"),
			Phone:    ptrUserStr("13800138000"),
		})
		require.NoError(t, err)
		assert.NotZero(t, createdID)

		var stored user.SysUser
		require.NoError(t, db.Where("user_id = ?", createdID).First(&stored).Error)
		assert.Equal(t, "creator-user", stored.Username)
		assert.Equal(t, "Creator User", stored.NickName)
		assert.Equal(t, user.FromLocal, stored.From)
		assert.Equal(t, user.StatusEnabled, stored.Status)
		assert.Equal(t, user.DelFlagNormal, stored.DelFlag)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(FallbackDefaultPwd)))
	})

	t.Run("create user duplicate username returns error", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{Username: "duplicate-user", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		createdID, err := svc.CreateUser(&user.UserCreateRequest{Username: "duplicate-user", Password: "secret"})
		require.Error(t, err)
		assert.Zero(t, createdID)
		assert.Contains(t, err.Error(), "username already exists")
	})

	t.Run("create user wraps repository create error", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_user_insert BEFORE INSERT ON sys_user BEGIN SELECT RAISE(FAIL, 'deny user insert'); END;").Error)

		createdID, err := svc.CreateUser(&user.UserCreateRequest{Username: "broken-user", Password: "secret"})
		require.Error(t, err)
		assert.Zero(t, createdID)
		assert.Contains(t, err.Error(), "failed to create user")
	})

	t.Run("create user rolls back when org binding fails", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)
		require.NoError(t, db.Create(&domainorg.SysOrg{OrgID: 55, OrgName: "disabled-org", Status: domainorg.StatusEnabled}).Error)
		require.NoError(t, db.Model(&domainorg.SysOrg{}).Where("org_id = ?", 55).Update("status", domainorg.StatusDisabled).Error)
		orgID := int64(55)

		createdID, err := svc.CreateUser(&user.UserCreateRequest{Username: "rollback-user", Password: "secret", OrgID: &orgID})
		require.Error(t, err)
		assert.Zero(t, createdID)
		assert.Contains(t, err.Error(), "organization is disabled")

		deleted, getErr := svc.GetUserByUsername("rollback-user")
		require.Error(t, getErr)
		assert.Nil(t, deleted)

		var stored user.SysUser
		require.NoError(t, db.Unscoped().Where("username = ?", "rollback-user").First(&stored).Error)
		assert.Equal(t, user.DelFlagDeleted, stored.DelFlag)
	})

	t.Run("update user success updates selected fields and hashes password", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{UserID: 301, Username: "update-user", NickName: "Old Name", Password: "old-secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		email := ptrUserStr("updated@example.com")
		phone := ptrUserStr("13800138001")
		password := "NewSecret123!"
		status := user.StatusDisabled

		require.NoError(t, svc.UpdateUser(&user.UserUpdateRequest{ID: 301, Username: "updated-user", RealName: "Updated Name", Email: email, Phone: phone, Password: &password, Status: &status}))

		var stored user.SysUser
		require.NoError(t, db.Where("user_id = ?", 301).First(&stored).Error)
		assert.Equal(t, "updated-user", stored.Username)
		assert.Equal(t, "Updated Name", stored.NickName)
		require.NotNil(t, stored.Email)
		assert.Equal(t, "updated@example.com", *stored.Email)
		require.NotNil(t, stored.Phone)
		assert.Equal(t, "13800138001", *stored.Phone)
		assert.Equal(t, user.StatusDisabled, stored.Status)
		assert.NotNil(t, stored.UpdateTime)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte(password)))
	})

	t.Run("update user wraps not found and update error", func(t *testing.T) {
		t.Run("user not found", func(t *testing.T) {
			svc, _ := setupUserServiceRepoTest(t)

			err := svc.UpdateUser(&user.UserUpdateRequest{ID: 999, Username: "missing"})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "user not found")
		})

		t.Run("repository update error", func(t *testing.T) {
			svc, db := setupUserServiceRepoTest(t)
			require.NoError(t, db.Create(&user.SysUser{UserID: 302, Username: "update-error", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
			require.NoError(t, db.Exec("CREATE TRIGGER deny_user_update BEFORE UPDATE ON sys_user BEGIN SELECT RAISE(FAIL, 'deny user update'); END;").Error)

			err := svc.UpdateUser(&user.UserUpdateRequest{ID: 302, Username: "new-name"})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to update user")
		})
	})
}

func TestUserService_ReadAndDeletePaths(t *testing.T) {
	t.Run("get user by id found and not found", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{Username: "reader-id", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		var existing user.SysUser
		require.NoError(t, db.Where("username = ?", "reader-id").First(&existing).Error)

		loaded, err := svc.GetUserByID(existing.UserID)
		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "reader-id", loaded.Username)

		missing, err := svc.GetUserByID(999999)
		require.Error(t, err)
		assert.Nil(t, missing)
	})

	t.Run("get user by username found and not found", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{Username: "reader-name", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		loaded, err := svc.GetUserByUsername("reader-name")
		require.NoError(t, err)
		require.NotNil(t, loaded)
		assert.Equal(t, "reader-name", loaded.Username)

		missing, err := svc.GetUserByUsername("missing-user")
		require.Error(t, err)
		assert.Nil(t, missing)
	})

	t.Run("search users normalizes paging and returns total", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.AutoMigrate(&user.SysUserRole{}))
		keyword := "ali"
		orgID := int64(7)
		status := user.StatusEnabled
		require.NoError(t, db.Create(&user.SysUser{UserID: 101, Username: "alice", NickName: "Alice", Email: ptrUserStr("alice@example.com"), Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Create(&user.SysUser{UserID: 102, Username: "bob", NickName: "Bob", Email: ptrUserStr("bob@example.com"), Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Create(&user.SysUser{UserID: 103, Username: "alice-deleted", NickName: "Alice Deleted", Email: ptrUserStr("deleted@example.com"), Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagDeleted}).Error)
		require.NoError(t, db.Create(&user.SysUserRole{UserID: 101, OrgID: orgID, RoleID: 1}).Error)
		require.NoError(t, db.Create(&user.SysUserRole{UserID: 102, OrgID: 8, RoleID: 1}).Error)

		resp, err := svc.SearchUsers(&user.UserQueryRequest{Keyword: &keyword, OrgID: &orgID, Status: &status, Current: 0, Size: 0})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(1), resp.Total)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
		users, ok := resp.List.([]*user.SysUser)
		require.True(t, ok)
		require.Len(t, users, 1)
		assert.Equal(t, int64(101), users[0].UserID)
	})

	t.Run("search users wraps repository error", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		resp, err := svc.SearchUsers(&user.UserQueryRequest{Current: 1, Size: 10})
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to search users")
	})

	t.Run("delete user soft deletes and clears relations", func(t *testing.T) {
		svc, db := setupUserServiceFullRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{UserID: 201, Username: "deleter", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Create(&user.SysUserRole{UserID: 201, OrgID: 9, RoleID: 2}).Error)
		require.NoError(t, db.Create(&user.SysUserPerm{UserID: 201, PermID: 3}).Error)

		require.NoError(t, svc.DeleteUser(201))

		deleted, err := svc.GetUserByID(201)
		require.Error(t, err)
		assert.Nil(t, deleted)

		var stored user.SysUser
		require.NoError(t, db.Unscoped().Where("user_id = ?", 201).First(&stored).Error)
		assert.Equal(t, user.DelFlagDeleted, stored.DelFlag)

		var roleCount int64
		require.NoError(t, db.Model(&user.SysUserRole{}).Where("user_id = ?", 201).Count(&roleCount).Error)
		assert.Zero(t, roleCount)

		var permCount int64
		require.NoError(t, db.Model(&user.SysUserPerm{}).Where("user_id = ?", 201).Count(&permCount).Error)
		assert.Zero(t, permCount)
	})

	t.Run("delete user wraps repository error", func(t *testing.T) {
		svc, db := setupUserServiceFullRepoTest(t)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		err = svc.DeleteUser(202)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete user")
	})
}

func TestUserService_ResetPasswordPaths(t *testing.T) {
	t.Run("reset password success hashes and persists password", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{UserID: 401, Username: "reset-user", Password: "old-secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		require.NoError(t, svc.ResetPassword(401, "NewReset123!"))

		var stored user.SysUser
		require.NoError(t, db.Where("user_id = ?", 401).First(&stored).Error)
		assert.NotNil(t, stored.UpdateTime)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("NewReset123!")))
	})

	t.Run("reset password wraps user not found", func(t *testing.T) {
		svc, _ := setupUserServiceRepoTest(t)

		err := svc.ResetPassword(9999, "NewReset123!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("reset password wraps repository update error", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{UserID: 402, Username: "reset-error", Password: "old-secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_reset_update BEFORE UPDATE ON sys_user BEGIN SELECT RAISE(FAIL, 'deny reset update'); END;").Error)

		err := svc.ResetPassword(402, "NewReset123!")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to reset password")
	})

	t.Run("reset password with audit succeeds when audit service is nil", func(t *testing.T) {
		svc, db := setupUserServiceRepoTest(t)
		require.NoError(t, db.Create(&user.SysUser{UserID: 403, Username: "reset-audit-nil", Password: "old-secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

		require.NoError(t, svc.ResetPasswordWithAudit(403, "AnotherReset123!", 1, "tester", "127.0.0.1"))

		var stored user.SysUser
		require.NoError(t, db.Where("user_id = ?", 403).First(&stored).Error)
		require.NoError(t, bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("AnotherReset123!")))
	})

	t.Run("reset password with audit wraps user not found when audit service nil", func(t *testing.T) {
		svc, _ := setupUserServiceRepoTest(t)

		err := svc.ResetPasswordWithAudit(404, "AnotherReset123!", 1, "tester", "127.0.0.1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestUserService_BindUserToOrgBaseline(t *testing.T) {
	t.Run("fails when org repo missing", func(t *testing.T) {
		svc := NewUserService(nil, nil, nil)

		err := svc.bindUserToOrgBaseline(1, 2)
		assert.EqualError(t, err, "org repository is not configured")
	})

	t.Run("fails when user role repo missing", func(t *testing.T) {
		svc, _ := setupUserServiceOrgBindingTest(t)
		svc.userRoleRepo = nil

		err := svc.bindUserToOrgBaseline(1, 2)
		assert.EqualError(t, err, "user-role repository is not configured")
	})

	t.Run("fails when organization not found", func(t *testing.T) {
		svc, _ := setupUserServiceOrgBindingTest(t)

		err := svc.bindUserToOrgBaseline(1, 404)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "organization not found")
	})

	t.Run("fails when organization disabled", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)
		require.NoError(t, db.Create(&domainorg.SysOrg{OrgID: 9, OrgName: "disabled-org", Status: domainorg.StatusEnabled}).Error)
		require.NoError(t, db.Model(&domainorg.SysOrg{}).Where("org_id = ?", 9).Update("status", domainorg.StatusDisabled).Error)

		err := svc.bindUserToOrgBaseline(1, 9)
		assert.EqualError(t, err, "organization is disabled")
	})

	t.Run("propagates default role resolution error", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)
		require.NoError(t, db.Create(&domainorg.SysOrg{OrgID: 10, OrgName: "enabled-org", Status: domainorg.StatusEnabled}).Error)
		svc.roleRepo = nil

		err := svc.bindUserToOrgBaseline(1, 10)
		assert.EqualError(t, err, "role repository is not configured")
	})

	t.Run("creates default role and organization membership", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)
		require.NoError(t, db.Create(&domainorg.SysOrg{OrgID: 11, OrgName: "enabled-org", Status: domainorg.StatusEnabled}).Error)

		err := svc.bindUserToOrgBaseline(88, 11)
		require.NoError(t, err)

		var createdRole domainrole.SysRole
		require.NoError(t, db.Where("role_code = ?", domainrole.BuiltInOrgUserRoleCode).First(&createdRole).Error)
		assert.Equal(t, domainrole.BuiltInOrgUserRoleName, createdRole.RoleName)
		assert.Equal(t, domainrole.StatusEnabled, createdRole.Status)
		require.NotNil(t, createdRole.RoleType)
		require.NotNil(t, createdRole.DataScope)
		require.NotNil(t, createdRole.CreateBy)
		assert.Equal(t, domainrole.RoleTypeOrganization, *createdRole.RoleType)
		assert.Equal(t, domainrole.DataScopeSelf, *createdRole.DataScope)
		assert.Equal(t, systemActor, *createdRole.CreateBy)

		var membership user.SysUserRole
		require.NoError(t, db.Where("user_id = ? AND org_id = ?", 88, 11).First(&membership).Error)
		assert.Equal(t, createdRole.RoleID, membership.RoleID)
	})
}

func TestDeleteUser_BuiltInAdminProtected(t *testing.T) {
	svc, _ := setupUserServiceFullRepoTest(t)

	err := svc.DeleteUser(1)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBuiltInUserProtected)
}

func TestUpdateUserStatus_InvalidStatus(t *testing.T) {
	svc, _ := setupUserServiceRepoTest(t)

	err := svc.UpdateUserStatus(999, 5)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidStatus)
}

func TestUpdateUserStatus_BuiltInAdminDisableProtected(t *testing.T) {
	svc, db := setupUserServiceRepoTest(t)
	require.NoError(t, db.Create(&user.SysUser{UserID: 1, Username: "admin", Password: "secret", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}).Error)

	err := svc.UpdateUserStatus(1, user.StatusDisabled)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBuiltInUserProtected)
}

func TestResetPasswordWithAudit_BuiltInAdminProtected(t *testing.T) {
	svc, _ := setupUserServiceRepoTest(t)

	err := svc.ResetPasswordWithAudit(1, "new-secret", 99, "auditor", "192.168.1.10")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrBuiltInUserProtected)
}

func TestUpdateUserStatus_AllowEnableAdmin(t *testing.T) {
	svc, db := setupUserServiceRepoTest(t)
	require.NoError(t, db.Create(&user.SysUser{UserID: 1, Username: "admin", Password: "secret", Status: user.StatusDisabled, DelFlag: user.DelFlagNormal}).Error)

	err := svc.UpdateUserStatus(1, user.StatusEnabled)
	require.NoError(t, err)

	var updated user.SysUser
	require.NoError(t, db.First(&updated, "user_id = ?", 1).Error)
	assert.Equal(t, user.StatusEnabled, updated.Status)
}

func TestUserService_EnsureDefaultOrgUserRole(t *testing.T) {
	t.Run("fails when role repo missing", func(t *testing.T) {
		svc := NewUserService(nil, nil, nil)

		roleID, err := svc.ensureDefaultOrgUserRole()
		assert.Zero(t, roleID)
		assert.EqualError(t, err, "role repository is not configured")
	})

	t.Run("returns existing enabled role", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)
		roleType := domainrole.RoleTypeOrganization
		dataScope := domainrole.DataScopeSelf
		require.NoError(t, db.Create(&domainrole.SysRole{RoleID: 21, RoleName: "普通用户", RoleCode: domainrole.BuiltInOrgUserRoleCode, RoleType: &roleType, DataScope: &dataScope, Status: domainrole.StatusEnabled}).Error)

		roleID, err := svc.ensureDefaultOrgUserRole()
		require.NoError(t, err)
		assert.Equal(t, int64(21), roleID)
	})

	t.Run("enables existing disabled role", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)
		roleType := domainrole.RoleTypeOrganization
		dataScope := domainrole.DataScopeSelf
		require.NoError(t, db.Create(&domainrole.SysRole{RoleID: 22, RoleName: "普通用户", RoleCode: domainrole.BuiltInOrgUserRoleCode, RoleType: &roleType, DataScope: &dataScope, Status: domainrole.StatusEnabled}).Error)
		require.NoError(t, db.Model(&domainrole.SysRole{}).Where("role_id = ?", 22).Update("status", domainrole.StatusDisabled).Error)

		roleID, err := svc.ensureDefaultOrgUserRole()
		require.NoError(t, err)
		assert.Equal(t, int64(22), roleID)

		var updated domainrole.SysRole
		require.NoError(t, db.Where("role_id = ?", 22).First(&updated).Error)
		assert.Equal(t, domainrole.StatusEnabled, updated.Status)
	})

	t.Run("creates default role when missing", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)

		roleID, err := svc.ensureDefaultOrgUserRole()
		require.NoError(t, err)
		assert.NotZero(t, roleID)

		var created domainrole.SysRole
		require.NoError(t, db.Where("role_id = ?", roleID).First(&created).Error)
		assert.Equal(t, domainrole.BuiltInOrgUserRoleCode, created.RoleCode)
		assert.Equal(t, domainrole.BuiltInOrgUserRoleName, created.RoleName)
		assert.Equal(t, domainrole.StatusEnabled, created.Status)
	})

	t.Run("returns load error when repository unavailable", func(t *testing.T) {
		svc, db := setupUserServiceOrgBindingTest(t)
		sqlDB, err := db.DB()
		require.NoError(t, err)
		require.NoError(t, sqlDB.Close())

		roleID, err := svc.ensureDefaultOrgUserRole()
		assert.Zero(t, roleID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load default organization role")
	})
}
