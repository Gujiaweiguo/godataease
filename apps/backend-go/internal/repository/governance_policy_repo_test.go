package repository

import (
	"testing"

	"dataease/backend/internal/domain/governance"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGovernancePolicyRepoTest(t *testing.T) (*GovernancePolicyRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&governance.SysGovernancePolicy{}))
	return NewGovernancePolicyRepository(db), db
}

func TestGovernancePolicyRepository_GetLastRolePolicy_NotFound(t *testing.T) {
	repo, _ := setupGovernancePolicyRepoTest(t)

	policy, err := repo.GetLastRolePolicy(999)
	require.NoError(t, err)
	assert.Equal(t, governance.DefaultLastRolePolicy, policy)
}

func TestGovernancePolicyRepository_SetAndGetLastRolePolicy(t *testing.T) {
	repo, _ := setupGovernancePolicyRepoTest(t)

	require.NoError(t, repo.SetLastRolePolicy(1, governance.LastRolePolicyWarnAllow, "admin"))

	policy, err := repo.GetLastRolePolicy(1)
	require.NoError(t, err)
	assert.Equal(t, governance.LastRolePolicyWarnAllow, policy)
}

func TestGovernancePolicyRepository_SetLastRolePolicy_Upsert(t *testing.T) {
	repo, _ := setupGovernancePolicyRepoTest(t)

	require.NoError(t, repo.SetLastRolePolicy(2, governance.LastRolePolicyBlock, "admin"))
	require.NoError(t, repo.SetLastRolePolicy(2, governance.LastRolePolicyCascade, "editor"))

	policy, err := repo.GetLastRolePolicy(2)
	require.NoError(t, err)
	assert.Equal(t, governance.LastRolePolicyCascade, policy)
}

func TestGovernancePolicyRepository_GetLastRolePolicy_InvalidValue(t *testing.T) {
	repo, db := setupGovernancePolicyRepoTest(t)

	require.NoError(t, db.Create(&governance.SysGovernancePolicy{
		OrgID:       10,
		PolicyKey:   governance.PolicyKeyLastRole,
		PolicyValue: "INVALID_VALUE",
		UpdatedBy:   "admin",
	}).Error)

	policy, err := repo.GetLastRolePolicy(10)
	require.NoError(t, err)
	assert.Equal(t, governance.DefaultLastRolePolicy, policy)
}

func TestGovernancePolicyRepository_DifferentOrgs(t *testing.T) {
	repo, _ := setupGovernancePolicyRepoTest(t)

	require.NoError(t, repo.SetLastRolePolicy(1, governance.LastRolePolicyBlock, "admin"))
	require.NoError(t, repo.SetLastRolePolicy(2, governance.LastRolePolicyCascade, "admin"))

	policy1, err := repo.GetLastRolePolicy(1)
	require.NoError(t, err)
	assert.Equal(t, governance.LastRolePolicyBlock, policy1)

	policy2, err := repo.GetLastRolePolicy(2)
	require.NoError(t, err)
	assert.Equal(t, governance.LastRolePolicyCascade, policy2)
}
