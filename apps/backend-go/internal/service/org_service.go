package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"dataease/backend/internal/domain/audit"
	"dataease/backend/internal/domain/governance"
	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/domain/user"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrgService struct {
	orgRepo             *repository.OrgRepository
	auditSvc            *AuditService
	userRepo            *repository.UserRepository
	roleRepo            *repository.RoleRepository
	userRoleRepo        *repository.UserRoleRepository
	governancePolicySvc *GovernancePolicyService
}

func NewOrgService(orgRepo *repository.OrgRepository, auditSvc *AuditService, userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, userRoleRepo *repository.UserRoleRepository, governancePolicySvc *GovernancePolicyService) *OrgService {
	return &OrgService{
		orgRepo:             orgRepo,
		auditSvc:            auditSvc,
		userRepo:            userRepo,
		roleRepo:            roleRepo,
		userRoleRepo:        userRoleRepo,
		governancePolicySvc: governancePolicySvc,
	}
}

// TransferUserOrg transfers a user from source org to target org atomically.
func (s *OrgService) TransferUserOrg(sourceOrgID, targetOrgID, userID, actorID int64) error {
	if sourceOrgID <= 0 || targetOrgID <= 0 || userID <= 0 {
		return fmt.Errorf("source org, target org, and user id are required")
	}
	if sourceOrgID == targetOrgID {
		return fmt.Errorf("source and target organizations must be different")
	}
	if s.orgRepo == nil || s.orgRepo.DB() == nil {
		return fmt.Errorf("organization repository is not configured")
	}
	if s.roleRepo == nil {
		return fmt.Errorf("role repository is not configured")
	}
	if s.userRoleRepo == nil {
		return fmt.Errorf("user role repository is not configured")
	}

	sourceOrgInfo, err := s.orgRepo.GetByID(sourceOrgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("source organization not found")
		}
		return fmt.Errorf("failed to load source organization: %w", err)
	}
	targetOrgInfo, err := s.orgRepo.GetByID(targetOrgID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("target organization not found")
		}
		return fmt.Errorf("failed to load target organization: %w", err)
	}
	if targetOrgInfo.Status != org.StatusEnabled {
		return fmt.Errorf("target organization is disabled")
	}

	isMember, err := s.userRoleRepo.IsUserInOrg(userID, sourceOrgID)
	if err != nil {
		return fmt.Errorf("failed to validate source organization membership: %w", err)
	}
	if !isMember {
		return fmt.Errorf("user is not a member of source organization")
	}

	sourceRoleCount, err := s.roleRepo.CountUserRolesByOrg(userID, sourceOrgID)
	if err != nil {
		return fmt.Errorf("failed to count source organization roles: %w", err)
	}
	policy := governance.DefaultLastRolePolicy
	if sourceRoleCount <= 1 {
		policy, err = s.getEffectiveLastRolePolicy(sourceOrgID)
		if err != nil {
			return fmt.Errorf("failed to resolve last-role policy: %w", err)
		}
		if policy == governance.LastRolePolicyBlock {
			return fmt.Errorf("%w", ErrLastRoleRemovalBlocked)
		}
	}

	var assignedRoleID int64
	err = s.orgRepo.DB().Transaction(func(txDB *gorm.DB) error {
		txOrgRepo := repository.NewOrgRepository(txDB)
		txRoleRepo := repository.NewRoleRepository(txDB)
		txUserRoleRepo := repository.NewUserRoleRepository(txDB)

		if err := txDB.Where("user_id = ? AND org_id = ?", userID, sourceOrgID).Delete(&user.SysUserRole{}).Error; err != nil {
			return fmt.Errorf("failed to remove source organization roles: %w", err)
		}

		if err := ensureUserOrgBaseline(txDB, txOrgRepo, txRoleRepo, txUserRoleRepo, userID, targetOrgID); err != nil {
			return err
		}

		assignedRoleID, err = ensureDefaultOrgUserRole(txRoleRepo)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	s.recordTransferAudit(sourceOrgInfo, targetOrgInfo, userID, actorID, assignedRoleID, policy)
	logger.Info("User transferred between organizations",
		zap.Int64("userId", userID),
		zap.Int64("sourceOrgId", sourceOrgID),
		zap.Int64("targetOrgId", targetOrgID),
		zap.Int64("assignedRoleId", assignedRoleID),
	)
	return nil
}

func (s *OrgService) getEffectiveLastRolePolicy(orgID int64) (governance.LastRolePolicy, error) {
	if s.governancePolicySvc == nil {
		return governance.DefaultLastRolePolicy, nil
	}
	return s.governancePolicySvc.GetLastRolePolicy(orgID)
}

func (s *OrgService) recordTransferAudit(sourceOrgInfo, targetOrgInfo *org.SysOrg, userID, actorID, assignedRoleID int64, policy governance.LastRolePolicy) {
	if s.auditSvc == nil {
		return
	}

	beforeValue, _ := json.Marshal(map[string]interface{}{
		"sourceOrgId": sourceOrgInfo.OrgID,
		"sourceOrg":   sourceOrgInfo.OrgName,
		"userId":      userID,
		"policy":      policy,
	})
	afterValue, _ := json.Marshal(map[string]interface{}{
		"targetOrgId":      targetOrgInfo.OrgID,
		"targetOrg":        targetOrgInfo.OrgName,
		"userId":           userID,
		"actorId":          actorID,
		"assignedRoleId":   assignedRoleID,
		"lastRolePolicy":   policy,
		"transferOccurred": true,
	})
	resourceType := string(audit.ResourceTypeUser)
	username := strconv.FormatInt(actorID, 10)
	if actorID <= 0 {
		username = systemActor
	}

	for _, visibilityOrgID := range []int64{sourceOrgInfo.OrgID, targetOrgInfo.OrgID} {
		_, _ = s.auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:         governanceInt64PtrIfPositive(actorID),
			Username:       governanceStringPtrIfNotEmpty(username),
			ActionType:     audit.ActionTypeSystemConfig,
			ActionName:     "转移用户组织归属",
			ResourceType:   &resourceType,
			ResourceID:     &userID,
			Operation:      audit.OperationUpdate,
			Status:         ptrStatus(audit.StatusSuccess),
			BeforeValue:    ptrStr(string(beforeValue)),
			AfterValue:     ptrStr(string(afterValue)),
			OrganizationID: &visibilityOrgID,
		})
	}
}

func (s *OrgService) CreateOrg(req *org.OrgCreateRequest, callerOrgID int64) error {
	if err := requireGovernedOrgContext(callerOrgID); err != nil {
		return err
	}

	count, err := s.orgRepo.CheckNameExists(req.OrgName, 0)
	if err != nil {
		return fmt.Errorf("failed to check org name: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("organization name already exists")
	}

	parentID := int64(org.RootParentID)
	level := 1

	if req.ParentID != nil && *req.ParentID > 0 {
		parent, err := s.orgRepo.GetByID(*req.ParentID)
		if err != nil {
			return fmt.Errorf("parent organization not found: %w", err)
		}
		parentID = parent.OrgID
		level = parent.Level + 1
	}

	o := &org.SysOrg{
		OrgName:  req.OrgName,
		OrgDesc:  req.OrgDesc,
		ParentID: parentID,
		Level:    level,
		Status:   org.StatusEnabled,
		DelFlag:  org.DelFlagNormal,
	}

	if err := s.orgRepo.Create(o); err != nil {
		logger.Error("Failed to create organization", zap.Error(err))
		return fmt.Errorf("failed to create organization: %w", err)
	}

	logger.Info("Organization created", zap.Int64("orgId", o.OrgID), zap.String("orgName", o.OrgName))
	return nil
}

func (s *OrgService) UpdateOrg(req *org.OrgUpdateRequest, callerOrgID int64) error {
	if err := requireGovernedOrgContext(callerOrgID); err != nil {
		return err
	}

	existing, err := s.orgRepo.GetByID(req.OrgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	if req.OrgName != "" && req.OrgName != existing.OrgName {
		count, err := s.orgRepo.CheckNameExists(req.OrgName, req.OrgID)
		if err != nil {
			return fmt.Errorf("failed to check org name: %w", err)
		}
		if count > 0 {
			return fmt.Errorf("organization name already exists")
		}
		existing.OrgName = req.OrgName
	}

	if req.OrgDesc != nil {
		existing.OrgDesc = req.OrgDesc
	}

	if req.ParentID != nil {
		if err := s.validateAndUpdateParent(existing, *req.ParentID); err != nil {
			return err
		}
	}

	now := time.Now()
	existing.UpdateTime = &now

	if err := s.orgRepo.Update(existing); err != nil {
		logger.Error("Failed to update organization", zap.Error(err))
		return fmt.Errorf("failed to update organization: %w", err)
	}

	logger.Info("Organization updated", zap.Int64("orgId", req.OrgID))
	return nil
}

func (s *OrgService) validateAndUpdateParent(existing *org.SysOrg, newParentID int64) error {
	if newParentID == existing.OrgID {
		return fmt.Errorf("organization cannot be its own parent")
	}
	if newParentID == existing.ParentID {
		return nil
	}
	if newParentID > 0 {
		parent, err := s.orgRepo.GetByID(newParentID)
		if err != nil {
			return fmt.Errorf("parent organization not found: %w", err)
		}
		if parent.Status != org.StatusEnabled {
			return fmt.Errorf("parent organization is disabled")
		}
		isDesc, err := s.orgRepo.IsDescendant(existing.OrgID, newParentID)
		if err != nil {
			return fmt.Errorf("failed to check descendant: %w", err)
		}
		if isDesc {
			return fmt.Errorf("cannot move organization under its own descendant")
		}
		existing.ParentID = newParentID
		existing.Level = parent.Level + 1
	} else {
		existing.ParentID = org.RootParentID
		existing.Level = 1
	}
	return nil
}

// DeleteOrg deletes an organization with deterministic C1 policy:
//  1. Child rejection — reject if child organizations exist
//  2. Soft delete — mark org as deleted without cascade
//  3. Auditable deferred disposition — record affected users in audit log;
//     actual resource cleanup is explicitly deferred to downstream lifecycle work (C2+),
//     not treated as implicitly complete.
func (s *OrgService) DeleteOrg(orgID int64, callerOrgID int64, operatorID int64, operatorName string, ipAddress string) error {
	if err := requireGovernedOrgContext(callerOrgID); err != nil {
		return err
	}

	// 1. 获取组织信息
	orgInfo, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	// 2. 检查子组织
	childrenCount, err := s.orgRepo.CountChildren(orgID)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}
	if childrenCount > 0 {
		// 记录删除失败的审计日志
		if s.auditSvc != nil {
			resourceType := string(audit.ResourceTypeOrganization)
			_, _ = s.auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
				UserID:         &operatorID,
				Username:       &operatorName,
				ActionType:     audit.ActionTypeSystemConfig,
				ActionName:     "删除组织",
				ResourceType:   &resourceType,
				ResourceID:     &orgID,
				ResourceName:   &orgInfo.OrgName,
				OrganizationID: &orgID,
				Operation:      audit.OperationDelete,
				IPAddress:      &ipAddress,
				Status:         ptrStatus(audit.StatusFailed),
				FailureReason:  ptrStr(fmt.Sprintf("组织下存在 %d 个子组织，无法删除", childrenCount)),
			})
		}
		return fmt.Errorf("cannot delete organization with %d child organizations - please delete or move child organizations first", childrenCount)
	}

	// 3. 检查关联资源（可选，用于审计记录）
	affectedUsers := "0"
	if s.userRepo != nil {
		// 检查组织下用户
		userCount, countErr := s.userRepo.CountByOrgID(orgID)
		if countErr != nil {
			affectedUsers = "unknown"
		} else {
			affectedUsers = fmt.Sprintf("%d", userCount)
		}
	}

	// 4. 执行删除（软删除）
	if err := s.orgRepo.Delete(orgID); err != nil {
		logger.Error("Failed to delete organization", zap.Error(err))
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	// 5. 记录成功的审计日志
	if s.auditSvc != nil {
		afterValue := fmt.Sprintf("disposition=soft-delete; org_name=%s; affected_users=%s", orgInfo.OrgName, affectedUsers)
		resourceType := string(audit.ResourceTypeOrganization)
		_, _ = s.auditSvc.CreateAuditLog(&audit.AuditLogCreateRequest{
			UserID:         &operatorID,
			Username:       &operatorName,
			ActionType:     audit.ActionTypeSystemConfig,
			ActionName:     "删除组织",
			ResourceType:   &resourceType,
			ResourceID:     &orgID,
			ResourceName:   &orgInfo.OrgName,
			OrganizationID: &orgID,
			Operation:      audit.OperationDelete,
			IPAddress:      &ipAddress,
			BeforeValue:    ptrStr(fmt.Sprintf("OrgName: %s, Level: %d", orgInfo.OrgName, orgInfo.Level)),
			AfterValue:     ptrStr(afterValue),
			Status:         ptrStatus(audit.StatusSuccess),
		})
	}
	logger.Info("Organization deleted", zap.Int64("orgId", orgID), zap.String("orgName", orgInfo.OrgName))
	return nil
}

// ptrStr 辅助函数
func ptrStr(v string) *string {
	return &v
}

// ptrStatus 辅助函数
func ptrStatus(v audit.Status) *audit.Status {
	return &v
}
func (s *OrgService) GetOrgByID(orgID int64) (*org.SysOrg, error) {
	return s.orgRepo.GetByID(orgID)
}

func (s *OrgService) ListOrgs() ([]*org.SysOrg, error) {
	return s.orgRepo.List()
}

func (s *OrgService) ListByParentID(parentID int64) ([]*org.SysOrg, error) {
	return s.orgRepo.ListByParentID(parentID)
}

func (s *OrgService) GetOrgTree() ([]*org.OrgTreeNode, error) {
	orgs, err := s.orgRepo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	orgMap := make(map[int64]*org.OrgTreeNode)
	var rootNodes []*org.OrgTreeNode

	for _, o := range orgs {
		node := o.ToTreeNode()
		orgMap[o.OrgID] = node
	}

	for _, o := range orgs {
		node := orgMap[o.OrgID]
		if o.ParentID == org.RootParentID {
			rootNodes = append(rootNodes, node)
		} else {
			if parent, ok := orgMap[o.ParentID]; ok {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return rootNodes, nil
}

func (s *OrgService) UpdateOrgStatus(orgID int64, status int, callerOrgID int64) error {
	if err := requireGovernedOrgContext(callerOrgID); err != nil {
		return err
	}

	existing, err := s.orgRepo.GetByID(orgID)
	if err != nil {
		return fmt.Errorf("organization not found: %w", err)
	}

	existing.Status = status
	now := time.Now()
	existing.UpdateTime = &now

	if err := s.orgRepo.Update(existing); err != nil {
		logger.Error("Failed to update organization status", zap.Error(err))
		return fmt.Errorf("failed to update organization status: %w", err)
	}

	logger.Info("Organization status updated", zap.Int64("orgId", orgID), zap.Int("status", status))
	return nil
}

func (s *OrgService) CheckOrgNameExists(orgName string) (bool, error) {
	count, err := s.orgRepo.CheckNameExists(orgName, 0)
	if err != nil {
		return false, fmt.Errorf("failed to check org name: %w", err)
	}
	return count > 0, nil
}
