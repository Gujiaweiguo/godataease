package governance

import "time"

// LastRolePolicy defines the configurable policy for removing a user's last role.
type LastRolePolicy string

const (
	LastRolePolicyBlock     LastRolePolicy = "BLOCK"
	LastRolePolicyWarnAllow LastRolePolicy = "WARN_ALLOW"
	LastRolePolicyCascade   LastRolePolicy = "CASCADE"
)

// DefaultLastRolePolicy is the default policy for new installations.
const DefaultLastRolePolicy = LastRolePolicyBlock

const PolicyKeyLastRole = "last_role_policy"

func (p LastRolePolicy) IsValid() bool {
	switch p {
	case LastRolePolicyBlock, LastRolePolicyWarnAllow, LastRolePolicyCascade:
		return true
	}
	return false
}

type SysGovernancePolicy struct {
	ID          uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	OrgID       int64     `gorm:"column:org_id;not null;uniqueIndex:uk_org_policy" json:"orgId"`
	PolicyKey   string    `gorm:"column:policy_key;size:100;not null;uniqueIndex:uk_org_policy" json:"policyKey"`
	PolicyValue string    `gorm:"column:policy_value;size:100;not null" json:"policyValue"`
	UpdatedBy   string    `gorm:"column:updated_by;size:100;default:''" json:"updatedBy"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`
}

func (SysGovernancePolicy) TableName() string {
	return "sys_governance_policy"
}
