package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/permission"
	"dataease/backend/internal/domain/visualization"
	"dataease/backend/internal/repository"
)

type VisualizationService struct {
	repo                *repository.VisualizationRepository
	resourcePermService *ResourcePermissionService
}

func NewVisualizationService(repo *repository.VisualizationRepository) *VisualizationService {
	return &VisualizationService{repo: repo}
}

func (s *VisualizationService) SetResourcePermissionService(resourcePermSvc *ResourcePermissionService) {
	s.resourcePermService = resourcePermSvc
}

func (s *VisualizationService) Save(req *visualization.SaveRequest, updateBy string) (int64, error) {
	now := time.Now().UnixMilli()
	nodeType := "panel"
	if req.NodeType != nil && *req.NodeType != "" {
		nodeType = *req.NodeType
	}
	status := 0
	if nodeType == "folder" {
		status = 1
	}

	v := &visualization.DataVisualizationInfo{
		Name:            req.Name,
		PID:             req.PID,
		Type:            req.Type,
		NodeType:        &nodeType,
		CanvasStyleData: req.CanvasStyleData,
		ComponentData:   req.ComponentData,
		MobileLayout:    req.MobileLayout,
		ContentID:       req.ContentID,
		CheckVersion:    req.CheckVersion,
		Status:          &status,
		CreateTime:      &now,
		UpdateTime:      &now,
		CreateBy:        &updateBy,
		UpdateBy:        &updateBy,
	}

	if err := s.repo.Create(v); err != nil {
		return 0, err
	}
	if err := s.applyInheritedPermissionsOnCreate(v.ID, req.Name, req.PID, req.Type); err != nil {
		_ = s.repo.DeleteLogic(v.ID, updateBy)
		return 0, err
	}
	return v.ID, nil
}

func (s *VisualizationService) applyInheritedPermissionsOnCreate(resourceID int64, resourceName string, pid *int64, visualizationType *string) error {
	if s.resourcePermService == nil || pid == nil || *pid <= 0 {
		return nil
	}
	return s.resourcePermService.InheritParentResourcePermissions(*pid, resourceID, resourceName, normalizeVisualizationResourceType(visualizationType))
}

func (s *VisualizationService) BackfillGovernedResources() (*VisualizationGovernanceBackfillReport, error) {
	return s.BackfillGovernedVisualizationResourcesWithOptions(nil)
}

func (s *VisualizationService) BackfillGovernedVisualizationResourcesWithOptions(options *GovernanceBackfillOptions) (*VisualizationGovernanceBackfillReport, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("visualization repository not initialized")
	}
	if s.resourcePermService == nil {
		return nil, fmt.Errorf("resource permission service not initialized")
	}

	normalized := normalizeGovernanceBackfillOptions(options)
	items, err := s.repo.ListAllByTypesBatch(nil, normalized.AfterID, normalized.Limit, normalized.OrgID)
	if err != nil {
		return nil, err
	}

	report := newGovernanceBackfillReport("visualization", normalized)
	for _, item := range items {
		if item == nil || item.ID <= 0 {
			continue
		}
		resourceType := normalizeVisualizationResourceType(item.Type)
		report.observe(item.ID)
		if item.PID == nil || *item.PID <= 0 {
			report.addSkipped(item.ID, resourceType, 0, GovernanceBackfillSkipReasonMissingParent)
			continue
		}
		inherited, err := s.resourcePermService.TryInheritParentResourcePermissions(*item.PID, item.ID, item.Name, resourceType)
		if err != nil {
			return nil, err
		}
		if !inherited {
			report.addSkipped(item.ID, resourceType, *item.PID, GovernanceBackfillSkipReasonParentNotGoverned)
			continue
		}
		report.addGoverned(item.ID)
	}

	return report, nil
}

func normalizeVisualizationResourceType(visualizationType *string) string {
	if visualizationType != nil {
		switch *visualizationType {
		case "dataV", permission.ResourceTypeScreen:
			return permission.ResourceTypeScreen
		}
	}
	return permission.ResourceTypeDashboard
}

func (s *VisualizationService) Copy(req *visualization.CopyRequest, updateBy string) (int64, error) {
	if req == nil {
		return 0, fmt.Errorf("copy request is required")
	}
	if req.ID <= 0 {
		return 0, fmt.Errorf("source id is required")
	}
	if req.Name == "" {
		return 0, fmt.Errorf("name is required")
	}

	source, err := s.repo.GetByID(req.ID)
	if err != nil {
		return 0, err
	}

	nodeType := source.NodeType
	if req.NodeType != nil && *req.NodeType != "" {
		nodeType = req.NodeType
	}
	typ := source.Type
	if req.Type != nil && *req.Type != "" {
		typ = req.Type
	}
	mobileLayout := source.MobileLayout
	if req.MobileLayout != nil {
		mobileLayout = req.MobileLayout
	}

	return s.Save(&visualization.SaveRequest{
		Name:            req.Name,
		PID:             req.PID,
		Type:            typ,
		NodeType:        nodeType,
		CanvasStyleData: source.CanvasStyleData,
		ComponentData:   source.ComponentData,
		MobileLayout:    mobileLayout,
		ContentID:       source.ContentID,
		CheckVersion:    source.CheckVersion,
	}, updateBy)
}

func (s *VisualizationService) Update(req *visualization.UpdateRequest, updateBy string) error {
	v, err := s.repo.GetByID(req.ID)
	if err != nil {
		return fmt.Errorf("visualization not found: %w", err)
	}

	if req.Name != nil {
		v.Name = *req.Name
	}
	if req.PID != nil {
		v.PID = req.PID
	}
	if req.Type != nil {
		v.Type = req.Type
	}
	if req.CanvasStyleData != nil {
		v.CanvasStyleData = req.CanvasStyleData
	}
	if req.ComponentData != nil {
		v.ComponentData = req.ComponentData
	}
	if req.MobileLayout != nil {
		v.MobileLayout = req.MobileLayout
	}
	if req.ContentID != nil {
		v.ContentID = req.ContentID
	}
	if req.CheckVersion != nil {
		v.CheckVersion = req.CheckVersion
	}
	if req.Status != nil {
		v.Status = req.Status
	}
	now := time.Now().UnixMilli()
	v.UpdateTime = &now
	v.UpdateBy = &updateBy

	return s.repo.Update(v)
}

func (s *VisualizationService) Detail(req *visualization.DetailRequest) (*visualization.DataVisualizationInfo, error) {
	return s.repo.GetByID(req.ID)
}

func (s *VisualizationService) List(req *visualization.ListRequest) (*visualization.ListResponse, error) {
	list, total, err := s.repo.Query(req)
	if err != nil {
		return nil, err
	}

	current := req.Current
	if current < 1 {
		current = 1
	}
	size := req.Size
	if size < 1 {
		size = 10
	}

	return &visualization.ListResponse{
		List:    list,
		Total:   total,
		Current: current,
		Size:    size,
	}, nil
}

func (s *VisualizationService) InteractiveTree(busiFlag string) ([]*visualization.DataVisualizationInfo, error) {
	types, err := resolveInteractiveVisualizationTypes(busiFlag)
	if err != nil {
		return nil, err
	}
	return s.repo.ListAllByTypes(types)
}

func (s *VisualizationService) DeleteLogic(id int64, updateBy string) error {
	return s.repo.DeleteLogic(id, updateBy)
}

func (s *VisualizationService) FindDvType(id int64) (string, error) {
	item, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	if item.Type == nil {
		return "", nil
	}
	return *item.Type, nil
}

func resolveInteractiveVisualizationTypes(busiFlag string) ([]string, error) {
	flag := busiFlag
	switch flag {
	case "", "dashboard-dataV":
		return []string{"dashboard", "dataV"}, nil
	case "panel", "dashboard":
		return []string{"dashboard"}, nil
	case "screen", "dataV":
		return []string{"dataV"}, nil
	default:
		return nil, fmt.Errorf("unsupported busiFlag: %s", flag)
	}
}

func (s *VisualizationService) NameCheck(req *visualization.NameCheckRequest) (string, error) {
	if req == nil {
		return "success", nil
	}
	var excludeID *int64
	if req.ID > 0 {
		excludeID = &req.ID
	}
	count, err := s.repo.CountByNameAndPID(req.Name, req.PID, excludeID)
	if err != nil {
		return "", err
	}
	if count > 0 {
		return "repeat", nil
	}
	return "success", nil
}

func (s *VisualizationService) CheckCanvasChange(req *visualization.CanvasChangeRequest) (string, error) {
	if req == nil || req.ID <= 0 {
		return "", nil
	}
	item, err := s.repo.GetByID(req.ID)
	if err != nil {
		return "", err
	}
	if req.ContentID != nil && item.ContentID != nil && *req.ContentID != "" && *item.ContentID != "" && *req.ContentID != *item.ContentID {
		return "Repeat", nil
	}
	if req.CheckVersion != nil && item.CheckVersion != nil && *req.CheckVersion != "" && *item.CheckVersion != "" && *req.CheckVersion != *item.CheckVersion {
		return "Repeat", nil
	}
	return "", nil
}

func (s *VisualizationService) UpdateBase(req *visualization.UpdateRequest, updateBy string) error {
	return s.Update(req, updateBy)
}

func (s *VisualizationService) Move(req *visualization.MoveRequest, updateBy string) error {
	if req == nil {
		return nil
	}
	return s.Update(&visualization.UpdateRequest{ID: req.ID, PID: req.PID}, updateBy)
}

func (s *VisualizationService) UpdatePublishStatus(req *visualization.UpdateRequest, updateBy string) (*visualization.DataVisualizationInfo, error) {
	if err := s.Update(req, updateBy); err != nil {
		return nil, err
	}
	return s.repo.GetByID(req.ID)
}

func (s *VisualizationService) RecoverToPublished(id int64, updateBy string) (*visualization.DataVisualizationInfo, error) {
	status := 1
	if err := s.Update(&visualization.UpdateRequest{ID: id, Status: &status}, updateBy); err != nil {
		return nil, err
	}
	return s.repo.GetByID(id)
}
