package service

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/repository"

	"github.com/google/uuid"
)

type LinkageRequest struct {
	DvID          int64            `json:"dvId"`
	SourceViewID  int64            `json:"sourceViewId"`
	TargetViewIds []int64          `json:"targetViewIds"`
	ResourceTable string           `json:"resourceTable"`
	LinkageInfo   []LinkageInfoDTO `json:"linkageInfo"`
	ActiveStatus  bool             `json:"activeStatus"`
}

type LinkageInfoDTO struct {
	TargetViewID     int64                        `json:"targetViewId"`
	TargetViewType   string                       `json:"targetViewType"`
	TableID          int64                        `json:"tableId"`
	TargetViewName   string                       `json:"targetViewName"`
	SourceViewID     int64                        `json:"sourceViewId"`
	LinkageActive    bool                         `json:"linkageActive"`
	TargetViewFields []repository.DatasetFieldDTO `json:"targetViewFields"`
	LinkageFields    []LinkageFieldVO             `json:"linkageFields"`
}

type LinkageFieldVO struct {
	SourceField int64 `json:"sourceField"`
	TargetField int64 `json:"targetField"`
}

type LinkageService struct {
	repo *repository.LinkageRepository
}

func NewLinkageService(repo *repository.LinkageRepository) *LinkageService {
	return &LinkageService{repo: repo}
}

func isSnapshot(resourceTable string) bool {
	return resourceTable == "snapshot"
}

func (s *LinkageService) GetViewLinkageGather(req *LinkageRequest) (map[string]LinkageInfoDTO, error) {
	if len(req.TargetViewIds) == 0 {
		return map[string]LinkageInfoDTO{}, nil
	}

	rows, err := s.repo.GetViewLinkageGather(req.DvID, req.SourceViewID, req.TargetViewIds, isSnapshot(req.ResourceTable))
	if err != nil {
		return nil, err
	}

	dtoMap := make(map[string]LinkageInfoDTO)
	for _, row := range rows {
		key := fmt.Sprintf("%d", row.TargetViewID)
		dto, exists := dtoMap[key]
		if !exists {
			dto = LinkageInfoDTO{
				TargetViewID:   row.TargetViewID,
				TargetViewType: row.TargetViewType,
				TableID:        row.TableID,
				TargetViewName: row.TargetViewName,
				SourceViewID:   row.SourceViewID,
				LinkageActive:  row.LinkageActive,
				LinkageFields:  []LinkageFieldVO{},
			}
			if row.TableID > 0 {
				fields, err := s.repo.GetDatasetFieldsByGroupID(row.TableID)
				if err == nil {
					dto.TargetViewFields = fields
				}
			}
		}
		if row.SourceField != nil && row.TargetField != nil {
			dto.LinkageFields = append(dto.LinkageFields, LinkageFieldVO{
				SourceField: *row.SourceField,
				TargetField: *row.TargetField,
			})
		}
		dtoMap[key] = dto
	}
	return dtoMap, nil
}

func (s *LinkageService) GetViewLinkageGatherArray(req *LinkageRequest) ([]LinkageInfoDTO, error) {
	m, err := s.GetViewLinkageGather(req)
	if err != nil {
		return nil, err
	}
	result := make([]LinkageInfoDTO, 0, len(m))
	for _, dto := range m {
		result = append(result, dto)
	}
	return result, nil
}

func (s *LinkageService) SaveLinkage(req *LinkageRequest) error {
	if req.SourceViewID <= 0 {
		return fmt.Errorf("sourceViewId is required")
	}
	if req.DvID <= 0 {
		return fmt.Errorf("dvId is required")
	}

	if err := s.repo.DeleteLinkageAndFields(req.DvID, req.SourceViewID); err != nil {
		return fmt.Errorf("delete existing linkage: %w", err)
	}

	now := time.Now().UnixMilli()
	for _, info := range req.LinkageInfo {
		if info.TargetViewID == req.SourceViewID {
			continue
		}

		linkageID := generateLinkageID()
		linkage := &auto.SnapshotVisualizationLinkage{
			ID:            linkageID,
			DvID:          req.DvID,
			SourceViewID:  req.SourceViewID,
			TargetViewID:  info.TargetViewID,
			UpdateTime:    now,
			UpdatePeople:  "",
			LinkageActive: info.LinkageActive,
		}
		if err := s.repo.CreateLinkage(linkage); err != nil {
			return fmt.Errorf("create linkage: %w", err)
		}

		if info.LinkageActive && len(info.LinkageFields) > 0 {
			for _, field := range info.LinkageFields {
				fieldID := generateLinkageID()
				linkageField := &auto.SnapshotVisualizationLinkageField{
					ID:          fieldID,
					LinkageID:   linkageID,
					SourceField: field.SourceField,
					TargetField: field.TargetField,
					UpdateTime:  now,
				}
				if err := s.repo.CreateLinkageField(linkageField); err != nil {
					return fmt.Errorf("create linkage field: %w", err)
				}
			}
		}
	}
	return nil
}

func (s *LinkageService) GetVisualizationAllLinkageInfo(dvID int64, resourceTable string) (map[string][]string, error) {
	return s.repo.GetAllLinkageInfo(dvID, isSnapshot(resourceTable))
}

func (s *LinkageService) UpdateLinkageActive(req *LinkageRequest) (map[string][]string, error) {
	if err := s.repo.UpdateChartLinkageActive(req.SourceViewID, req.ActiveStatus); err != nil {
		return nil, fmt.Errorf("update linkage active: %w", err)
	}
	return s.GetVisualizationAllLinkageInfo(req.DvID, "snapshot")
}

func (s *LinkageService) RemoveLinkage(req *LinkageRequest) error {
	return s.repo.DeleteLinkageAndFields(req.DvID, req.SourceViewID)
}

// generateLinkageID produces a unique int64 ID using UUID v4 hash.
func generateLinkageID() int64 {
	return int64(uuid.New().ID())
}
