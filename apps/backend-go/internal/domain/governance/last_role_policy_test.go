package governance

import (
	"testing"
	"time"
)

func TestLastRolePolicyConstants(t *testing.T) {
	tests := []struct {
		name     string
		policy   LastRolePolicy
		expected string
	}{
		{"BLOCK", LastRolePolicyBlock, "BLOCK"},
		{"WARN_ALLOW", LastRolePolicyWarnAllow, "WARN_ALLOW"},
		{"CASCADE", LastRolePolicyCascade, "CASCADE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.policy) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(tt.policy))
			}
		})
	}
}

func TestDefaultLastRolePolicy(t *testing.T) {
	if DefaultLastRolePolicy != LastRolePolicyBlock {
		t.Errorf("Expected DefaultLastRolePolicy = %v, got %v", LastRolePolicyBlock, DefaultLastRolePolicy)
	}
}

func TestPolicyKeyLastRole(t *testing.T) {
	if PolicyKeyLastRole != "last_role_policy" {
		t.Errorf("Expected PolicyKeyLastRole = 'last_role_policy', got '%s'", PolicyKeyLastRole)
	}
}

func TestLastRolePolicy_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		policy LastRolePolicy
		valid  bool
	}{
		{"BLOCK is valid", LastRolePolicyBlock, true},
		{"WARN_ALLOW is valid", LastRolePolicyWarnAllow, true},
		{"CASCADE is valid", LastRolePolicyCascade, true},
		{"empty string is invalid", LastRolePolicy(""), false},
		{"lowercase block is invalid", LastRolePolicy("block"), false},
		{"unknown value is invalid", LastRolePolicy("UNKNOWN"), false},
		{"partial match is invalid", LastRolePolicy("BLOCK_"), false},
		{"numeric is invalid", LastRolePolicy("1"), false},
		{"whitespace is invalid", LastRolePolicy(" "), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.policy.IsValid(); got != tt.valid {
				t.Errorf("LastRolePolicy(%q).IsValid() = %v, want %v", tt.policy, got, tt.valid)
			}
		})
	}
}

func TestSysGovernancePolicy_TableName(t *testing.T) {
	p := SysGovernancePolicy{}
	if p.TableName() != "sys_governance_policy" {
		t.Errorf("Expected table name 'sys_governance_policy', got '%s'", p.TableName())
	}
}

func TestSysGovernancePolicy_Fields(t *testing.T) {
	now := time.Now()
	p := SysGovernancePolicy{
		ID:          1,
		OrgID:       100,
		PolicyKey:   PolicyKeyLastRole,
		PolicyValue: string(LastRolePolicyBlock),
		UpdatedBy:   "admin",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if p.ID != 1 {
		t.Errorf("Expected ID 1, got %d", p.ID)
	}
	if p.OrgID != 100 {
		t.Errorf("Expected OrgID 100, got %d", p.OrgID)
	}
	if p.PolicyKey != "last_role_policy" {
		t.Errorf("Expected PolicyKey 'last_role_policy', got '%s'", p.PolicyKey)
	}
	if p.PolicyValue != "BLOCK" {
		t.Errorf("Expected PolicyValue 'BLOCK', got '%s'", p.PolicyValue)
	}
	if p.UpdatedBy != "admin" {
		t.Errorf("Expected UpdatedBy 'admin', got '%s'", p.UpdatedBy)
	}
	if !p.CreatedAt.Equal(now) {
		t.Errorf("Expected CreatedAt %v, got %v", now, p.CreatedAt)
	}
	if !p.UpdatedAt.Equal(now) {
		t.Errorf("Expected UpdatedAt %v, got %v", now, p.UpdatedAt)
	}
}

func TestSysGovernancePolicy_ZeroValues(t *testing.T) {
	p := SysGovernancePolicy{}

	if p.ID != 0 {
		t.Errorf("Expected zero ID, got %d", p.ID)
	}
	if p.OrgID != 0 {
		t.Errorf("Expected zero OrgID, got %d", p.OrgID)
	}
	if p.PolicyKey != "" {
		t.Errorf("Expected empty PolicyKey, got '%s'", p.PolicyKey)
	}
	if p.PolicyValue != "" {
		t.Errorf("Expected empty PolicyValue, got '%s'", p.PolicyValue)
	}
	if p.UpdatedBy != "" {
		t.Errorf("Expected empty UpdatedBy, got '%s'", p.UpdatedBy)
	}
}

func TestLastRolePolicy_ConversionRoundTrip(t *testing.T) {
	policies := []LastRolePolicy{LastRolePolicyBlock, LastRolePolicyWarnAllow, LastRolePolicyCascade}
	for _, policy := range policies {
		stored := string(policy)
		loaded := LastRolePolicy(stored)
		if loaded != policy {
			t.Errorf("Round-trip failed: stored %q, loaded %q, original %q", stored, loaded, policy)
		}
		if !loaded.IsValid() {
			t.Errorf("Loaded policy %q should be valid", loaded)
		}
	}
}

func TestDefaultPolicyIsValid(t *testing.T) {
	if !DefaultLastRolePolicy.IsValid() {
		t.Errorf("DefaultLastRolePolicy (%v) should be valid", DefaultLastRolePolicy)
	}
}
