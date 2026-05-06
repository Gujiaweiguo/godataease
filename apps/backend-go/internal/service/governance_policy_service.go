package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/repository"
)

var ErrInvalidLastRolePolicy = errors.New("invalid last-role policy")

type GovernancePolicyService struct {
	policyRepo *repository.GovernancePolicyRepository
	auditSvc   *AuditService
}

func NewGovernancePolicyService(policyRepo *repository.GovernancePolicyRepository, auditSvc *AuditService) *GovernancePolicyService {
	return &GovernancePolicyService{policyRepo: policyRepo, auditSvc: auditSvc}
}

func (s *GovernancePolicyService) GetLastRolePolicy(orgID int64) (governance.LastRolePolicy, error) {
	if err := requireGovernedOrgContext(orgID); err != nil {
		return "", err
	}
	if s.policyRepo == nil {
		return governance.DefaultLastRolePolicy, nil
	}
	return s.policyRepo.GetLastRolePolicy(orgID)
}

func (s *GovernancePolicyService) SetLastRolePolicy(orgID int64, policy governance.LastRolePolicy, actorID int64) error {
	if err := requireGovernedOrgContext(orgID); err != nil {
		return err
	}
	if !policy.IsValid() {
		return fmt.Errorf("%w: %s", ErrInvalidLastRolePolicy, policy)
	}
	if s.policyRepo == nil {
		return errors.New("governance policy repository not initialized")
	}

	before, err := s.policyRepo.GetLastRolePolicy(orgID)
	if err != nil {
		return err
	}
	updatedBy := strconv.FormatInt(actorID, 10)
	if actorID <= 0 {
		updatedBy = systemActor
	}
	if err := s.policyRepo.SetLastRolePolicy(orgID, policy, updatedBy); err != nil {
		return err
	}

	if s.auditSvc != nil {
		beforeValue, _ := json.Marshal(map[string]string{"lastRolePolicy": string(before)})
		afterValue, _ := json.Marshal(map[string]string{"lastRolePolicy": string(policy)})
		resourceType := string(audit.ResourceTypeOrganization)
		_, _ = s.auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:         governanceInt64PtrIfPositive(actorID),
			Username:       governanceStringPtrIfNotEmpty(updatedBy),
			ActionType:     audit.ActionTypeSystemConfig,
			ActionName:     "更新最后角色策略",
			ResourceType:   &resourceType,
			ResourceID:     &orgID,
			OrganizationID: &orgID,
			Operation:      audit.OperationUpdate,
			BeforeValue:    governanceStringPtrIfNotEmpty(string(beforeValue)),
			AfterValue:     governanceStringPtrIfNotEmpty(string(afterValue)),
		})
	}

	return nil
}

func governanceInt64PtrIfPositive(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	return &v
}

func governanceStringPtrIfNotEmpty(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}
