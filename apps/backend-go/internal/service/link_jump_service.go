package service

import (
	"fmt"
	"strings"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/repository"

	"github.com/google/uuid"
)

// LinkJumpRequest represents a request for link jump operations.
type LinkJumpRequest struct {
	SourceDvID    int64  `json:"sourceDvId"`
	SourceViewID  int64  `json:"sourceViewId"`
	SourceFieldID *int64 `json:"sourceFieldId"`
	TargetDvID    int64  `json:"targetDvId"`
	TargetViewID  int64  `json:"targetViewId"`
	LinkJumpID    int64  `json:"linkJumpId"`
	ActiveStatus  bool   `json:"activeStatus"`
	ResourceTable string `json:"resourceTable"`
}

// LinkJumpDTO is the full jump configuration DTO (matches Java VisualizationLinkJumpDTO).
type LinkJumpDTO struct {
	ID                int64             `json:"id"`
	SourceDvID        int64             `json:"sourceDvId"`
	SourceViewID      int64             `json:"sourceViewId"`
	LinkJumpInfo      string            `json:"linkJumpInfo"`
	Checked           bool              `json:"checked"`
	LinkJumpInfoArray []LinkJumpInfoDTO `json:"linkJumpInfoArray"`
}

// LinkJumpInfoDTO represents a single jump info with target mappings (matches Java VisualizationLinkJumpInfoDTO).
type LinkJumpInfoDTO struct {
	ID                 int64                       `json:"id"`
	LinkJumpID         int64                       `json:"linkJumpId"`
	LinkType           string                      `json:"linkType"`
	JumpType           string                      `json:"jumpType"`
	WindowSize         string                      `json:"windowSize"`
	TargetDvID         int64                       `json:"targetDvId"`
	SourceFieldID      int64                       `json:"sourceFieldId"`
	Content            string                      `json:"content"`
	Checked            bool                        `json:"checked"`
	AttachParams       bool                        `json:"attachParams"`
	SourceFieldName    string                      `json:"sourceFieldName"`
	SourceDeType       int                         `json:"sourceDeType"`
	PublicJumpID       string                      `json:"publicJumpId"`
	TargetDvType       string                      `json:"targetDvType"`
	TargetViewInfoList []LinkJumpTargetViewInfoDTO `json:"targetViewInfoList"`
}

// LinkJumpTargetViewInfoDTO represents a target view field mapping (matches Java VisualizationLinkJumpTargetViewInfoVO).
type LinkJumpTargetViewInfoDTO struct {
	TargetID            int64  `json:"targetId"`
	LinkJumpInfoID      int64  `json:"linkJumpInfoId"`
	SourceFieldActiveID int64  `json:"sourceFieldActiveId"`
	TargetViewID        string `json:"targetViewId"`
	TargetFieldID       string `json:"targetFieldId"`
	CopyFrom            int64  `json:"copyFrom"`
	CopyID              int64  `json:"copyId"`
	TargetType          string `json:"targetType"`
	OuterParamsName     string `json:"outerParamsName"`
}

// ViewTableVO represents a chart view with fields (matches Java VisualizationViewTableVO).
type ViewTableVO struct {
	ID          int64                        `json:"id"`
	Title       string                       `json:"title"`
	Type        string                       `json:"type"`
	DvID        int64                        `json:"dvId"`
	TableFields []repository.DatasetFieldDTO `json:"tableFields"`
}

// VisualizationComponentDTO is the response for viewTableDetailList (matches Java VisualizationComponentDTO).
type VisualizationComponentDTO struct {
	ComponentData     string            `json:"componentData"`
	ComponentView     []ViewTableVO     `json:"componentView"`
	OutParamsJumpInfo []OutParamsJumpVO `json:"outParamsJumpInfo"`
}

// OutParamsJumpVO represents an outer params target (matches Java VisualizationOutParamsJumpVO).
type OutParamsJumpVO struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Title string `json:"title"`
}

// LinkJumpBaseResponse is the response for jump info queries (matches Java VisualizationLinkJumpBaseResponse).
type LinkJumpBaseResponse struct {
	BaseJumpInfoMap              map[string]*LinkJumpInfoDTO `json:"baseJumpInfoMap"`
	BaseJumpInfoVisualizationMap map[string][]string         `json:"baseJumpInfoVisualizationMap"`
}

// LinkJumpService provides link jump business logic.
type LinkJumpService struct {
	repo *repository.LinkJumpRepository
}

func NewLinkJumpService(repo *repository.LinkJumpRepository) *LinkJumpService {
	return &LinkJumpService{repo: repo}
}

func hasQuotedComponentID(componentData string, viewID int64) bool {
	return strings.Contains(componentData, fmt.Sprintf(`"%d"`, viewID))
}

func isSnapshotTable(resourceTable string) bool {
	return resourceTable == "snapshot"
}

// GetTableFieldWithViewID returns dataset fields for a chart view.
func (s *LinkJumpService) GetTableFieldWithViewID(viewID int64) ([]repository.DatasetFieldDTO, error) {
	return s.repo.GetTableFieldWithViewID(viewID)
}

// QueryWithViewId returns the full jump configuration for a specific view in a dashboard.
func (s *LinkJumpService) QueryWithViewId(dvID, viewID int64) (*LinkJumpDTO, error) {
	row, err := s.repo.QueryWithViewId(dvID, viewID)
	if err != nil {
		return nil, err
	}

	dto := &LinkJumpDTO{
		ID:                row.ID,
		SourceDvID:        row.SourceDvID,
		SourceViewID:      row.SourceViewID,
		LinkJumpInfo:      row.LinkJumpInfo,
		Checked:           row.Checked,
		LinkJumpInfoArray: []LinkJumpInfoDTO{},
	}

	// If a jump config exists, fetch the detailed info
	if row.ID > 0 {
		infoRows, err := s.repo.GetLinkJumpInfo(row.ID, viewID, true)
		if err != nil {
			return nil, err
		}
		dto.LinkJumpInfoArray = buildLinkJumpInfoArray(infoRows)
	} else {
		// Even without a saved config, fetch dataset fields so the UI can show them
		infoRows, err := s.repo.GetLinkJumpInfo(0, viewID, true)
		if err != nil {
			return nil, err
		}
		dto.LinkJumpInfoArray = buildLinkJumpInfoArray(infoRows)
	}

	return dto, nil
}

// buildLinkJumpInfoArray groups flat rows into a structured LinkJumpInfoDTO array.
func buildLinkJumpInfoArray(rows []repository.LinkJumpInfoFlatRow) []LinkJumpInfoDTO {
	infoMap := make(map[int64]*LinkJumpInfoDTO)
	var order []int64

	for _, row := range rows {
		info, exists := infoMap[row.InfoID]
		if !exists {
			info = &LinkJumpInfoDTO{
				ID:                 row.InfoID,
				LinkJumpID:         row.LinkJumpID,
				LinkType:           row.LinkType,
				JumpType:           row.JumpType,
				WindowSize:         row.WindowSize,
				TargetDvID:         row.TargetDvID,
				SourceFieldID:      row.SourceFieldID,
				Content:            row.Content,
				Checked:            row.Checked,
				AttachParams:       row.AttachParams,
				SourceFieldName:    row.SourceFieldName,
				SourceDeType:       row.SourceDeType,
				TargetDvType:       row.TargetDvType,
				TargetViewInfoList: []LinkJumpTargetViewInfoDTO{},
			}
			infoMap[row.InfoID] = info
			order = append(order, row.InfoID)
		}

		if row.TargetID > 0 {
			info.TargetViewInfoList = append(info.TargetViewInfoList, LinkJumpTargetViewInfoDTO{
				TargetID:            row.TargetID,
				LinkJumpInfoID:      row.InfoID,
				SourceFieldActiveID: row.SourceFieldActiveID,
				TargetViewID:        row.TargetViewID,
				TargetFieldID:       row.TargetFieldID,
				TargetType:          row.TargetType,
			})
		}
	}

	result := make([]LinkJumpInfoDTO, 0, len(order))
	for _, id := range order {
		result = append(result, *infoMap[id])
	}
	return result
}

// UpdateJumpSet saves a complete jump configuration (delete + recreate pattern).
func (s *LinkJumpService) UpdateJumpSet(dto *LinkJumpDTO) error {
	if dto.SourceDvID <= 0 {
		return fmt.Errorf("sourceDvId is required")
	}
	if dto.SourceViewID <= 0 {
		return fmt.Errorf("sourceViewId is required")
	}

	// Delete existing data
	if err := s.repo.DeleteJumpCascadeSnapshot(dto.SourceDvID, dto.SourceViewID); err != nil {
		return fmt.Errorf("delete existing jump data: %w", err)
	}

	// Create new jump record
	linkJumpID := generateLinkJumpID()
	dto.ID = linkJumpID
	jump := &auto.SnapshotVisualizationLinkJump{
		ID:           linkJumpID,
		SourceDvID:   dto.SourceDvID,
		SourceViewID: dto.SourceViewID,
		LinkJumpInfo: dto.LinkJumpInfo,
		Checked:      dto.Checked,
	}
	if err := s.repo.CreateSnapshotJump(jump); err != nil {
		return fmt.Errorf("create jump: %w", err)
	}

	// Create jump info records
	for _, infoDTO := range dto.LinkJumpInfoArray {
		infoID := generateLinkJumpID()
		info := &auto.SnapshotVisualizationLinkJumpInfo{
			ID:            infoID,
			LinkJumpID:    linkJumpID,
			LinkType:      infoDTO.LinkType,
			JumpType:      infoDTO.JumpType,
			WindowSize:    infoDTO.WindowSize,
			TargetDvID:    infoDTO.TargetDvID,
			SourceFieldID: infoDTO.SourceFieldID,
			Content:       infoDTO.Content,
			Checked:       infoDTO.Checked,
			AttachParams:  infoDTO.AttachParams,
		}
		if err := s.repo.CreateSnapshotJumpInfo(info); err != nil {
			return fmt.Errorf("create jump info: %w", err)
		}

		// Create target view info records
		for _, targetDTO := range infoDTO.TargetViewInfoList {
			targetID := generateLinkJumpID()
			target := &auto.SnapshotVisualizationLinkJumpTargetViewInfo{
				TargetID:            targetID,
				LinkJumpInfoID:      infoID,
				SourceFieldActiveID: targetDTO.SourceFieldActiveID,
				TargetViewID:        targetDTO.TargetViewID,
				TargetFieldID:       targetDTO.TargetFieldID,
				TargetType:          targetDTO.TargetType,
			}
			if err := s.repo.CreateSnapshotJumpTargetViewInfo(target); err != nil {
				return fmt.Errorf("create jump target view info: %w", err)
			}
		}
	}
	return nil
}

// QueryVisualizationJumpInfo returns all active jump info for a dashboard.
func (s *LinkJumpService) QueryVisualizationJumpInfo(dvID int64, resourceTable string) (*LinkJumpBaseResponse, error) {
	snapshot := isSnapshotTable(resourceTable)
	rows, err := s.repo.QueryWithDvId(dvID, snapshot)
	if err != nil {
		return nil, err
	}

	resultBase := make(map[string]*LinkJumpInfoDTO)

	for _, jumpRow := range rows {
		if !jumpRow.Checked {
			continue
		}
		infoRows, err := s.repo.GetLinkJumpInfo(jumpRow.ID, jumpRow.SourceViewID, snapshot)
		if err != nil {
			return nil, err
		}

		infoList := buildLinkJumpInfoArray(infoRows)
		for _, info := range infoList {
			if !info.Checked {
				continue
			}
			sourceJumpInfo := fmt.Sprintf("%d#%d", jumpRow.SourceViewID, info.SourceFieldID)
			// Inner dashboard jump requires a target dashboard
			if info.LinkType == "inner" && info.TargetDvID == 0 {
				continue
			}
			infoCopy := info
			resultBase[sourceJumpInfo] = &infoCopy
		}
	}

	return &LinkJumpBaseResponse{BaseJumpInfoMap: resultBase}, nil
}

// QueryTargetVisualizationJumpInfo returns sourceInfo→targetInfo mappings for target dashboard navigation.
func (s *LinkJumpService) QueryTargetVisualizationJumpInfo(req *LinkJumpRequest) (map[string][]string, error) {
	snapshot := isSnapshotTable(req.ResourceTable)
	rows, err := s.repo.GetTargetVisualizationJumpInfo(req.SourceDvID, req.SourceViewID, req.TargetDvID, req.SourceFieldID, snapshot)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, row := range rows {
		if row.SourceInfo == "" {
			continue
		}
		result[row.SourceInfo] = append(result[row.SourceInfo], row.TargetInfo)
	}
	return result, nil
}

// ViewTableDetailList returns chart views with field details for a dashboard.
func (s *LinkJumpService) ViewTableDetailList(dvID int64) (*VisualizationComponentDTO, error) {
	componentData, err := s.repo.GetComponentData(dvID)
	if err != nil {
		return nil, err
	}
	if componentData == "" {
		componentData = "[]"
	}

	detailRows, err := s.repo.GetViewTableDetails(dvID)
	if err != nil {
		return nil, err
	}

	// Filter: only include views whose ID appears in component_data
	var views []ViewTableVO
	viewMap := make(map[int64]*ViewTableVO)
	var viewOrder []int64

	for _, row := range detailRows {
		if !hasQuotedComponentID(componentData, row.ID) {
			continue
		}

		vt, exists := viewMap[row.ID]
		if !exists {
			vt = &ViewTableVO{
				ID:    row.ID,
				Title: row.Title,
				Type:  row.Type,
				DvID:  row.DvID,
			}
			viewMap[row.ID] = vt
			viewOrder = append(viewOrder, row.ID)
		}

		if row.FieldID != nil {
			vt.TableFields = append(vt.TableFields, repository.DatasetFieldDTO{
				ID:             *row.FieldID,
				DatasetTableID: 0,
				OriginName:     row.OriginName,
				Name:           row.FieldName,
				DeType:         0,
			})
		}
	}

	for _, id := range viewOrder {
		views = append(views, *viewMap[id])
	}

	// Fetch outer params targets
	outParamsRows, err := s.repo.GetOutParamsTargetWithDvID(dvID)
	if err != nil {
		return nil, err
	}
	outParams := make([]OutParamsJumpVO, 0, len(outParamsRows))
	for _, r := range outParamsRows {
		outParams = append(outParams, OutParamsJumpVO{
			ID:    r.ID,
			Name:  r.Name,
			Title: r.Title,
			Type:  r.Type,
		})
	}

	return &VisualizationComponentDTO{
		ComponentData:     componentData,
		ComponentView:     views,
		OutParamsJumpInfo: outParams,
	}, nil
}

// UpdateJumpActive toggles jump_active on a chart and returns the updated jump info.
func (s *LinkJumpService) UpdateJumpActive(req *LinkJumpRequest) (*LinkJumpBaseResponse, error) {
	if err := s.repo.UpdateChartJumpActive(req.SourceViewID, req.ActiveStatus); err != nil {
		return nil, fmt.Errorf("update jump active: %w", err)
	}
	return s.QueryVisualizationJumpInfo(req.SourceDvID, "snapshot")
}

// RemoveJumpSet removes all jump data for a view.
func (s *LinkJumpService) RemoveJumpSet(dto *LinkJumpDTO) error {
	return s.repo.DeleteJumpCascadeSnapshot(dto.SourceDvID, dto.SourceViewID)
}

// generateLinkJumpID produces a unique int64 ID using UUID v4 hash.
func generateLinkJumpID() int64 {
	return int64(uuid.New().ID())
}
