package repository

import (
	"errors"

	"dataease/backend/internal/domain/governance"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GovernancePolicyRepository struct {
	db *gorm.DB
}

func NewGovernancePolicyRepository(db *gorm.DB) *GovernancePolicyRepository {
	return &GovernancePolicyRepository{db: db}
}

func (r *GovernancePolicyRepository) GetLastRolePolicy(orgID int64) (governance.LastRolePolicy, error) {
	var record governance.SysGovernancePolicy
	err := r.db.Where("org_id = ? AND policy_key = ?", orgID, governance.PolicyKeyLastRole).First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return governance.DefaultLastRolePolicy, nil
		}
		return "", err
	}

	policy := governance.LastRolePolicy(record.PolicyValue)
	if !policy.IsValid() {
		return governance.DefaultLastRolePolicy, nil
	}
	return policy, nil
}

func (r *GovernancePolicyRepository) SetLastRolePolicy(orgID int64, policy governance.LastRolePolicy, updatedBy string) error {
	record := &governance.SysGovernancePolicy{
		OrgID:       orgID,
		PolicyKey:   governance.PolicyKeyLastRole,
		PolicyValue: string(policy),
		UpdatedBy:   updatedBy,
	}

	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "policy_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"policy_value", "updated_by", "updated_at"}),
	}).Create(record).Error
}
