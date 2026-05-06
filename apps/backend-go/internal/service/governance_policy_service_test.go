package service

import (
	"testing"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/domain/role"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGovernancePolicyServiceTest(t *testing.T) (*GovernancePolicyService, *RoleService, *repository.RoleRepository, *repository.UserRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&audit.AuditLog{}, &audit.LoginFailure{}, &audit.AuditLogDetail{},
		&governance.SysGovernancePolicy{},
		&role.SysRole{}, &user.SysUser{}, &user.SysUserRole{},
	))

	auditSvc := NewAuditService(
		repository.NewAuditLogRepository(db),
		repository.NewLoginFailureRepository(db),
		repository.NewAuditLogDetailRepository(db),
	)
	policySvc := NewGovernancePolicyService(repository.NewGovernancePolicyRepository(db), auditSvc)
	roleRepo := repository.NewRoleRepository(db)
	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)
	roleSvc := NewRoleService(roleRepo, userRepo, userRoleRepo, policySvc)
	return policySvc, roleSvc, roleRepo, userRepo, db
}

func seedGovernanceRoleUser(t *testing.T, roleRepo *repository.RoleRepository, db *gorm.DB, userID int64) int64 {
	t.Helper()
	record := &user.SysUser{UserID: userID, Username: "user", NickName: "user", Status: user.StatusEnabled, DelFlag: user.DelFlagNormal}
	require.NoError(t, db.Create(record).Error)
	roleRecord := &role.SysRole{RoleName: "Role A", RoleCode: "role-a", Status: role.StatusEnabled}
	require.NoError(t, roleRepo.Create(roleRecord))
	require.NoError(t, db.Create(&user.SysUserRole{UserID: userID, RoleID: roleRecord.RoleID, OrgID: 7}).Error)
	return roleRecord.RoleID
}

func countLastRoleAudits(t *testing.T, db *gorm.DB) []audit.AuditLog {
	t.Helper()
	var logs []audit.AuditLog
	require.NoError(t, db.Where("action_name = ?", "移除用户最后角色").Order("id ASC").Find(&logs).Error)
	return logs
}

func TestGovernancePolicyService_DefaultToBlockWhenUnset(t *testing.T) {
	policySvc, _, _, _, _ := setupGovernancePolicyServiceTest(t)

	policy, err := policySvc.GetLastRolePolicy(7)
	require.NoError(t, err)
	assert.Equal(t, governance.DefaultLastRolePolicy, policy)
}

func TestGovernancePolicyService_SetLastRolePolicy_AuditsChange(t *testing.T) {
	policySvc, _, _, _, db := setupGovernancePolicyServiceTest(t)

	require.NoError(t, policySvc.SetLastRolePolicy(7, governance.LastRolePolicyWarnAllow, 9))

	policy, err := policySvc.GetLastRolePolicy(7)
	require.NoError(t, err)
	assert.Equal(t, governance.LastRolePolicyWarnAllow, policy)

	var logs []audit.AuditLog
	require.NoError(t, db.Where("action_name = ?", "更新最后角色策略").Find(&logs).Error)
	require.Len(t, logs, 1)
	assert.Equal(t, audit.ActionTypeSystemConfig, logs[0].ActionType)
	assert.Equal(t, audit.OperationUpdate, logs[0].Operation)
	assert.Equal(t, audit.StatusSuccess, logs[0].Status)
}

func TestRoleService_UnmountUser_BlockPolicyRejectsAndAudits(t *testing.T) {
	policySvc, roleSvc, roleRepo, _, db := setupGovernancePolicyServiceTest(t)
	roleID := seedGovernanceRoleUser(t, roleRepo, db, 21)
	require.NoError(t, policySvc.SetLastRolePolicy(7, governance.LastRolePolicyBlock, 1))

	err := roleSvc.UnmountUser(&role.UnmountUserRequest{Uid: 21, Rid: roleID, OrgId: 7})
	require.ErrorIs(t, err, ErrLastRoleRemovalBlocked)

	count, countErr := roleRepo.CountUserRolesByOrg(21, 7)
	require.NoError(t, countErr)
	assert.Equal(t, int64(1), count)

	logs := countLastRoleAudits(t, db)
	require.Len(t, logs, 1)
	assert.Equal(t, audit.StatusFailed, logs[0].Status)
	require.NotNil(t, logs[0].FailureReason)
	assert.Contains(t, *logs[0].FailureReason, ErrLastRoleRemovalBlocked.Error())
}

func TestRoleService_UnmountUser_WarnAllowPolicyUnbindsAndAudits(t *testing.T) {
	policySvc, roleSvc, roleRepo, userRepo, db := setupGovernancePolicyServiceTest(t)
	roleID := seedGovernanceRoleUser(t, roleRepo, db, 22)
	require.NoError(t, policySvc.SetLastRolePolicy(7, governance.LastRolePolicyWarnAllow, 1))

	err := roleSvc.UnmountUser(&role.UnmountUserRequest{Uid: 22, Rid: roleID, OrgId: 7})
	require.NoError(t, err)

	count, countErr := roleRepo.CountUserRolesByOrg(22, 7)
	require.NoError(t, countErr)
	assert.Equal(t, int64(0), count)

	reloaded, err := userRepo.GetByID(22)
	require.NoError(t, err)
	assert.Equal(t, user.StatusEnabled, reloaded.Status)

	logs := countLastRoleAudits(t, db)
	require.Len(t, logs, 1)
	assert.Equal(t, audit.StatusSuccess, logs[0].Status)
	require.NotNil(t, logs[0].AfterValue)
	assert.Contains(t, *logs[0].AfterValue, "warn_allow")
}

func TestRoleService_UnmountUser_CascadePolicyDisablesUserAndAudits(t *testing.T) {
	policySvc, roleSvc, roleRepo, userRepo, db := setupGovernancePolicyServiceTest(t)
	roleID := seedGovernanceRoleUser(t, roleRepo, db, 23)
	require.NoError(t, policySvc.SetLastRolePolicy(7, governance.LastRolePolicyCascade, 1))

	err := roleSvc.UnmountUser(&role.UnmountUserRequest{Uid: 23, Rid: roleID, OrgId: 7})
	require.NoError(t, err)

	count, countErr := roleRepo.CountUserRolesByOrg(23, 7)
	require.NoError(t, countErr)
	assert.Equal(t, int64(0), count)

	reloaded, err := userRepo.GetByID(23)
	require.NoError(t, err)
	assert.Equal(t, user.StatusDisabled, reloaded.Status)

	logs := countLastRoleAudits(t, db)
	require.Len(t, logs, 1)
	assert.Equal(t, audit.StatusSuccess, logs[0].Status)
	require.NotNil(t, logs[0].AfterValue)
	assert.Contains(t, *logs[0].AfterValue, "cascade")
}
