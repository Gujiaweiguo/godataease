package service

import (
	"fmt"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/repository"

	"github.com/google/uuid"
)

// OuterParamsDTO is the full outer params configuration (matches Java VisualizationOuterParamsDTO).
type OuterParamsDTO struct {
	ParamsID             string               `json:"paramsId"`
	VisualizationID      string               `json:"visualizationId"`
	Checked              bool                 `json:"checked"`
	Remark               string               `json:"remark"`
	CopyFrom             string               `json:"copyFrom"`
	CopyID               string               `json:"copyId"`
	OuterParamsInfoArray []OuterParamsInfoDTO `json:"outerParamsInfoArray"`
}

// OuterParamsInfoDTO represents a single param with target mappings (matches Java VisualizationOuterParamsInfoDTO).
type OuterParamsInfoDTO struct {
	ParamsInfoID       string                         `json:"paramsInfoId"`
	ParamsID           string                         `json:"paramsId"`
	ParamName          string                         `json:"paramName"`
	Checked            bool                           `json:"checked"`
	Required           bool                           `json:"required"`
	DefaultValue       string                         `json:"defaultValue"`
	EnabledDefault     bool                           `json:"enabledDefault"`
	CopyFrom           string                         `json:"copyFrom"`
	CopyID             string                         `json:"copyId"`
	TargetViewInfoList []OuterParamsTargetViewInfoDTO `json:"targetViewInfoList"`
	SourceInfo         string                         `json:"sourceInfo"`
	TargetInfoList     []string                       `json:"targetInfoList"`
}

// OuterParamsTargetViewInfoDTO represents a target view field mapping.
type OuterParamsTargetViewInfoDTO struct {
	TargetID      string `json:"targetId"`
	ParamsInfoID  string `json:"paramsInfoId"`
	TargetViewID  string `json:"targetViewId"`
	TargetDsID    string `json:"targetDsId"`
	TargetFieldID string `json:"targetFieldID"`
	CopyFrom      string `json:"copyFrom"`
	CopyID        string `json:"copyId"`
}

// OuterParamsBaseResponse is the runtime response (matches Java VisualizationOuterParamsBaseResponse).
type OuterParamsBaseResponse struct {
	OuterParamsInfoMap     map[string][]string            `json:"outerParamsInfoMap"`
	OuterParamsInfoBaseMap map[string]*OuterParamsInfoDTO `json:"outerParamsInfoBaseMap"`
}

// DatasetGroupVO represents a dataset group with fields and chart views.
type DatasetGroupVO struct {
	ID            int64            `json:"id"`
	Name          string           `json:"name"`
	Pid           int64            `json:"pid"`
	Level         int              `json:"level"`
	NodeType      string           `json:"nodeType"`
	Type          string           `json:"type"`
	Mode          int              `json:"mode"`
	DatasetFields []DatasetFieldVO `json:"datasetFields"`
	DatasetViews  []ChartBaseVO    `json:"datasetViews"`
}

// DatasetFieldVO represents a dataset field for outer params.
type DatasetFieldVO struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	DeType     int    `json:"deType"`
	OriginName string `json:"originName"`
}

// ChartBaseVO represents a chart view for outer params.
type ChartBaseVO struct {
	ChartID   int64  `json:"chartId"`
	ChartName string `json:"chartName"`
	ChartType string `json:"chartType"`
}

// OuterParamsService provides outer params business logic.
type OuterParamsService struct {
	repo *repository.OuterParamsRepository
}

func NewOuterParamsService(repo *repository.OuterParamsRepository) *OuterParamsService {
	return &OuterParamsService{repo: repo}
}

// QueryWithVisualizationId returns the full outer params config for a dashboard.
func (s *OuterParamsService) QueryWithVisualizationId(visualizationID string) (*OuterParamsDTO, error) {
	checkedRow, err := s.repo.QueryWithVisualizationId(visualizationID)
	if err != nil {
		return nil, err
	}

	dto := &OuterParamsDTO{
		ParamsID:             checkedRow.ParamsID,
		VisualizationID:      checkedRow.VisualizationID,
		Checked:              checkedRow.Checked,
		OuterParamsInfoArray: []OuterParamsInfoDTO{},
	}

	if checkedRow.Checked || checkedRow.ParamsID != "" {
		infoRows, err := s.repo.GetOuterParamsInfoSnapshot(visualizationID)
		if err != nil {
			return nil, err
		}
		dto.OuterParamsInfoArray = buildOuterParamsInfoArray(infoRows)
	}

	return dto, nil
}

// buildOuterParamsInfoArray groups flat rows into nested OuterParamsInfoDTO.
func buildOuterParamsInfoArray(rows []repository.OuterParamsInfoFlatRow) []OuterParamsInfoDTO {
	infoMap := make(map[string]*OuterParamsInfoDTO)
	var order []string

	for _, row := range rows {
		info, exists := infoMap[row.ParamsInfoID]
		if !exists {
			info = &OuterParamsInfoDTO{
				ParamsInfoID:       row.ParamsInfoID,
				ParamName:          row.ParamName,
				Checked:            row.Checked,
				Required:           row.Required,
				DefaultValue:       row.DefaultValue,
				EnabledDefault:     row.EnabledDefault,
				TargetViewInfoList: []OuterParamsTargetViewInfoDTO{},
			}
			infoMap[row.ParamsInfoID] = info
			order = append(order, row.ParamsInfoID)
		}

		if row.TargetViewID != "" {
			info.TargetViewInfoList = append(info.TargetViewInfoList, OuterParamsTargetViewInfoDTO{
				TargetViewID:  row.TargetViewID,
				ParamsInfoID:  row.ParamsInfoID,
				TargetDsID:    row.TargetDsID,
				TargetFieldID: row.TargetFieldID,
			})
		}
	}

	result := make([]OuterParamsInfoDTO, 0, len(order))
	for _, id := range order {
		result = append(result, *infoMap[id])
	}
	return result
}

// UpdateOuterParamsSet saves outer params configuration (delete+recreate).
func (s *OuterParamsService) UpdateOuterParamsSet(dto *OuterParamsDTO) error {
	visualizationID := dto.VisualizationID
	if visualizationID == "" {
		return fmt.Errorf("visualizationId is required")
	}

	// Preserve existing paramsInfoIds by param_name
	existingMap := make(map[string]string)
	existingRows, err := s.repo.GetOuterParamsInfoBase(visualizationID)
	if err == nil {
		for _, row := range existingRows {
			existingMap[row.ParamName] = row.ParamsInfoID
		}
	}

	// Delete existing data
	if err := s.repo.DeleteOuterParamsCascadeSnapshot(visualizationID); err != nil {
		return fmt.Errorf("delete existing outer params: %w", err)
	}

	if len(dto.OuterParamsInfoArray) == 0 {
		return nil
	}

	// Create new outer params record
	paramsID := uuid.New().String()
	dto.ParamsID = paramsID
	params := &auto.SnapshotVisualizationOuterParam{
		ParamsID:        paramsID,
		VisualizationID: visualizationID,
		Checked:         dto.Checked,
		Remark:          dto.Remark,
	}
	if err := s.repo.CreateSnapshotOuterParams(params); err != nil {
		return fmt.Errorf("create outer params: %w", err)
	}

	// Create params info records
	for _, infoDTO := range dto.OuterParamsInfoArray {
		// Reuse existing paramsInfoId if param_name matches
		paramsInfoID := infoDTO.ParamsInfoID
		if existingID, ok := existingMap[infoDTO.ParamName]; ok {
			paramsInfoID = existingID
		}
		if paramsInfoID == "" {
			paramsInfoID = uuid.New().String()
		}

		info := &auto.SnapshotVisualizationOuterParamsInfo{
			ParamsInfoID:   paramsInfoID,
			ParamsID:       paramsID,
			ParamName:      infoDTO.ParamName,
			Checked:        infoDTO.Checked,
			Required:       infoDTO.Required,
			DefaultValue:   infoDTO.DefaultValue,
			EnabledDefault: infoDTO.EnabledDefault,
		}
		if err := s.repo.CreateSnapshotOuterParamsInfo(info); err != nil {
			return fmt.Errorf("create outer params info: %w", err)
		}

		// Create target view info records
		for _, targetDTO := range infoDTO.TargetViewInfoList {
			targetID := uuid.New().String()
			target := &auto.SnapshotVisualizationOuterParamsTargetViewInfo{
				TargetID:      targetID,
				ParamsInfoID:  paramsInfoID,
				TargetViewID:  targetDTO.TargetViewID,
				TargetDsID:    targetDTO.TargetDsID,
				TargetFieldID: targetDTO.TargetFieldID,
			}
			if err := s.repo.CreateSnapshotOuterParamsTargetViewInfo(target); err != nil {
				return fmt.Errorf("create outer params target view info: %w", err)
			}
		}
	}
	return nil
}

// GetOuterParamsInfo returns runtime outer params info (for dashboard rendering).
func (s *OuterParamsService) GetOuterParamsInfo(visualizationID string) (*OuterParamsBaseResponse, error) {
	rows, err := s.repo.GetOuterParamsInfo(visualizationID)
	if err != nil {
		return nil, err
	}

	infoMap := make(map[string][]string)
	baseMap := make(map[string]*OuterParamsInfoDTO)

	for _, row := range rows {
		if row.SourceInfo == "" {
			continue
		}

		if row.TargetInfo != "" {
			infoMap[row.SourceInfo] = append(infoMap[row.SourceInfo], row.TargetInfo)
		}

		if _, exists := baseMap[row.SourceInfo]; !exists {
			baseMap[row.SourceInfo] = &OuterParamsInfoDTO{
				SourceInfo:     row.SourceInfo,
				Required:       row.Required,
				DefaultValue:   row.DefaultValue,
				EnabledDefault: row.EnabledDefault,
			}
		}
	}

	return &OuterParamsBaseResponse{
		OuterParamsInfoMap:     infoMap,
		OuterParamsInfoBaseMap: baseMap,
	}, nil
}

// QueryDsWithVisualizationId returns dataset groups with fields and chart views for a dashboard.
func (s *OuterParamsService) QueryDsWithVisualizationId(visualizationID string) ([]DatasetGroupVO, error) {
	rows, err := s.repo.GetDatasetGroupsWithFields(visualizationID)
	if err != nil {
		return nil, err
	}

	groupMap := make(map[int64]*DatasetGroupVO)
	var groupOrder []int64

	for _, row := range rows {
		group, exists := groupMap[row.ID]
		if !exists {
			group = &DatasetGroupVO{
				ID:       row.ID,
				Name:     row.Name,
				Pid:      row.Pid,
				Level:    row.Level,
				NodeType: row.NodeType,
				Type:     row.Type,
				Mode:     row.Mode,
			}
			groupMap[row.ID] = group
			groupOrder = append(groupOrder, row.ID)
		}

		if row.FieldID != nil {
			group.DatasetFields = append(group.DatasetFields, DatasetFieldVO{
				ID:         *row.FieldID,
				Name:       row.FieldName,
				DeType:     row.FieldDeType,
				OriginName: row.FieldOrigin,
			})
		}

		if row.ChartID != nil {
			// Avoid duplicate chart entries
			chartExists := false
			for _, v := range group.DatasetViews {
				if v.ChartID == *row.ChartID {
					chartExists = true
					break
				}
			}
			if !chartExists {
				group.DatasetViews = append(group.DatasetViews, ChartBaseVO{
					ChartID:   *row.ChartID,
					ChartName: row.ChartName,
					ChartType: row.ChartType,
				})
			}
		}
	}

	result := make([]DatasetGroupVO, 0, len(groupOrder))
	for _, id := range groupOrder {
		result = append(result, *groupMap[id])
	}
	return result, nil
}
