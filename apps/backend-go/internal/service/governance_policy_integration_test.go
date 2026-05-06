//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGovernancePolicyServiceIntegration_DefaultBlock(t *testing.T) {
	cleanupTables(&governance.SysGovernancePolicy{})
	svc := NewGovernancePolicyService(repository.NewGovernancePolicyRepository(testDB), nil)

	policy, err := svc.GetLastRolePolicy(11)
	require.NoError(t, err)
	assert.Equal(t, governance.DefaultLastRolePolicy, policy)
}

func TestGovernancePolicyServiceIntegration_SetAndGetLastRolePolicy(t *testing.T) {
	cleanupTables(&governance.SysGovernancePolicy{})
	svc := NewGovernancePolicyService(repository.NewGovernancePolicyRepository(testDB), nil)

	tests := []struct {
		name   string
		orgID  int64
		policy governance.LastRolePolicy
	}{
		{name: "block", orgID: 21, policy: governance.LastRolePolicyBlock},
		{name: "warn allow", orgID: 22, policy: governance.LastRolePolicyWarnAllow},
		{name: "cascade", orgID: 23, policy: governance.LastRolePolicyCascade},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, svc.SetLastRolePolicy(tt.orgID, tt.policy, 9))

			got, err := svc.GetLastRolePolicy(tt.orgID)
			require.NoError(t, err)
			assert.Equal(t, tt.policy, got)
		})
	}
}
