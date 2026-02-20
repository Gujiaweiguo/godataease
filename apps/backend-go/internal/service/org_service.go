package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/org"
	"dataease/backend/internal/pkg/logger"
	"dataease/backend/internal/repository"

	"go.uber.org/zap"
)

type OrgService struct {
	orgRepo *repository.OrgRepository
}

func NewOrgService(orgRepo *repository.OrgRepository) *OrgService {
	return &OrgService{
		orgRepo: orgRepo,
	}
}

func (s *OrgService) CreateOrg(req *org.OrgCreateRequest) error {
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

func (s *OrgService) UpdateOrg(req *org.OrgUpdateRequest) error {
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

	now := time.Now()
	existing.UpdateTime = &now

	if err := s.orgRepo.Update(existing); err != nil {
		logger.Error("Failed to update organization", zap.Error(err))
		return fmt.Errorf("failed to update organization: %w", err)
	}

	logger.Info("Organization updated", zap.Int64("orgId", req.OrgID))
	return nil
}

func (s *OrgService) DeleteOrg(orgID int64) error {
	childrenCount, err := s.orgRepo.CountChildren(orgID)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}
	if childrenCount > 0 {
		return fmt.Errorf("cannot delete organization with children")
	}

	if err := s.orgRepo.Delete(orgID); err != nil {
		logger.Error("Failed to delete organization", zap.Error(err))
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	logger.Info("Organization deleted", zap.Int64("orgId", orgID))
	return nil
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

func (s *OrgService) UpdateOrgStatus(orgID int64, status int) error {
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
