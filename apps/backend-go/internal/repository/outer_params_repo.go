package repository

import (
	"fmt"

	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

// OuterParamsRepository handles database operations for dashboard external parameters.
type OuterParamsRepository struct {
	db *gorm.DB
}

func NewOuterParamsRepository(db *gorm.DB) *OuterParamsRepository {
	return &OuterParamsRepository{db: db}
}

// OuterParamsInfoFlatRow is a flat row from getOuterParamsInfoSnapshot query.
type OuterParamsInfoFlatRow struct {
	VisualizationID string `json:"visualizationId"`
	ParamsInfoID    string `json:"paramsInfoId"`
	ParamName       string `json:"paramName"`
	EnabledDefault  bool   `json:"enabledDefault"`
	Required        bool   `json:"required"`
	DefaultValue    string `json:"defaultValue"`
	Checked         bool   `json:"checked"`
	TargetViewID    string `json:"targetViewId"`
	TargetDsID      string `json:"targetDsId"`
	TargetFieldID   string `json:"targetFieldId"`
}

// OuterParamsCheckedRow is the result of queryWithVisualizationId.
type OuterParamsCheckedRow struct {
	VisualizationID string `json:"visualizationId"`
	Checked         bool   `json:"checked"`
	ParamsID        string `json:"paramsId"`
}

// OuterParamsRuntimeRow is a flat row from getOuterParamsInfo (runtime, regular tables).
type OuterParamsRuntimeRow struct {
	SourceInfo     string `json:"sourceInfo"`
	Required       bool   `json:"required"`
	DefaultValue   string `json:"defaultValue"`
	EnabledDefault bool   `json:"enabledDefault"`
	TargetInfo     string `json:"targetInfo"`
}

// OuterParamsInfoBaseRow maps param_name → params_info_id for ID preservation.
type OuterParamsInfoBaseRow struct {
	ParamName    string `json:"paramName"`
	ParamsInfoID string `json:"paramsInfoId"`
}

// DatasetGroupRow is a flat row from dataset group + fields query.
type DatasetGroupRow struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Pid         int64  `json:"pid"`
	Level       int    `json:"level"`
	NodeType    string `json:"nodeType"`
	Type        string `json:"type"`
	Mode        int    `json:"mode"`
	FieldID     *int64 `json:"fieldId"`
	FieldName   string `json:"fieldName"`
	FieldDeType int    `json:"fieldDeType"`
	FieldOrigin string `json:"fieldOrigin"`
	ChartID     *int64 `json:"chartId"`
	ChartName   string `json:"chartName"`
	ChartType   string `json:"chartType"`
}

// QueryWithVisualizationId returns the outer params config for a dashboard (snapshot tables).
func (r *OuterParamsRepository) QueryWithVisualizationId(visualizationID string) (*OuterParamsCheckedRow, error) {
	var row OuterParamsCheckedRow
	err := r.db.Raw(`
		SELECT ? AS visualization_id,
			IFNULL(vop.checked, 0) AS checked,
			IFNULL(vop.params_id, '') AS params_id
		FROM snapshot_data_visualization_info dvi
		LEFT JOIN snapshot_visualization_outer_params vop ON dvi.id = vop.visualization_id
		WHERE dvi.id = ?`, visualizationID, visualizationID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetOuterParamsInfoSnapshot returns the full outer params info hierarchy (snapshot tables).
func (r *OuterParamsRepository) GetOuterParamsInfoSnapshot(visualizationID string) ([]OuterParamsInfoFlatRow, error) {
	var rows []OuterParamsInfoFlatRow
	err := r.db.Raw(`
		SELECT
			pop.visualization_id,
			popi.params_info_id,
			popi.param_name,
			popi.enabled_default,
			popi.required,
			popi.default_value,
			IFNULL(popi.checked, 0) AS checked,
			poptvi.target_view_id,
			poptvi.target_ds_id,
			poptvi.target_field_id
		FROM snapshot_visualization_outer_params pop
		LEFT JOIN snapshot_visualization_outer_params_info popi ON pop.params_id = popi.params_id
		LEFT JOIN snapshot_visualization_outer_params_target_view_info poptvi ON popi.params_info_id = poptvi.params_info_id
		WHERE pop.visualization_id = ?
		ORDER BY IFNULL(popi.checked, 0) DESC`, visualizationID).Scan(&rows).Error
	return rows, err
}

// GetOuterParamsInfoBase returns param_name→params_info_id mapping for ID preservation.
func (r *OuterParamsRepository) GetOuterParamsInfoBase(visualizationID string) ([]OuterParamsInfoBaseRow, error) {
	var rows []OuterParamsInfoBaseRow
	err := r.db.Raw(`
		SELECT vopi.param_name, vopi.params_info_id
		FROM snapshot_visualization_outer_params_info vopi
		INNER JOIN snapshot_visualization_outer_params vop ON vop.params_id = vopi.params_id
		WHERE vop.visualization_id = ?`, visualizationID).Scan(&rows).Error
	return rows, err
}

// DeleteOuterParamsCascadeSnapshot deletes all outer params data for a dashboard in snapshot tables.
func (r *OuterParamsRepository) DeleteOuterParamsCascadeSnapshot(visualizationID string) error {
	err := r.db.Exec(`
		DELETE FROM snapshot_visualization_outer_params_target_view_info
		WHERE params_info_id IN (
			SELECT poptvi.params_info_id FROM
				snapshot_visualization_outer_params_target_view_info poptvi
				INNER JOIN snapshot_visualization_outer_params_info popi ON poptvi.params_info_id = popi.params_info_id
				INNER JOIN snapshot_visualization_outer_params pop ON popi.params_id = pop.params_id
			WHERE pop.visualization_id = ?
		)`, visualizationID).Error
	if err != nil {
		return fmt.Errorf("delete outer params target view info: %w", err)
	}

	err = r.db.Exec(`
		DELETE FROM snapshot_visualization_outer_params_info
		WHERE params_id IN (
			SELECT popi.params_id FROM
				snapshot_visualization_outer_params_info popi
				INNER JOIN snapshot_visualization_outer_params pop ON popi.params_id = pop.params_id
			WHERE pop.visualization_id = ?
		)`, visualizationID).Error
	if err != nil {
		return fmt.Errorf("delete outer params info: %w", err)
	}

	err = r.db.Exec(`
		DELETE FROM snapshot_visualization_outer_params
		WHERE visualization_id = ?`, visualizationID).Error
	if err != nil {
		return fmt.Errorf("delete outer params: %w", err)
	}
	return nil
}

// CreateSnapshotOuterParams creates a new snapshot outer params record.
func (r *OuterParamsRepository) CreateSnapshotOuterParams(p *auto.SnapshotVisualizationOuterParam) error {
	return r.db.Create(p).Error
}

// CreateSnapshotOuterParamsInfo creates a new snapshot outer params info record.
func (r *OuterParamsRepository) CreateSnapshotOuterParamsInfo(info *auto.SnapshotVisualizationOuterParamsInfo) error {
	return r.db.Create(info).Error
}

// CreateSnapshotOuterParamsTargetViewInfo creates a new snapshot target view info record.
func (r *OuterParamsRepository) CreateSnapshotOuterParamsTargetViewInfo(t *auto.SnapshotVisualizationOuterParamsTargetViewInfo) error {
	return r.db.Create(t).Error
}

// GetOuterParamsInfo returns runtime outer params info (regular tables).
func (r *OuterParamsRepository) GetOuterParamsInfo(visualizationID string) ([]OuterParamsRuntimeRow, error) {
	var rows []OuterParamsRuntimeRow
	err := r.db.Raw(`
		SELECT DISTINCT
			popi.param_name AS sourceInfo,
			popi.required AS required,
			popi.default_value AS default_value,
			popi.enabled_default AS enabled_default,
			CONCAT(poptvi.target_view_id, '#', poptvi.target_field_id) AS targetInfo
		FROM visualization_outer_params pop
		LEFT JOIN visualization_outer_params_info popi ON pop.params_id = popi.params_id
		LEFT JOIN visualization_outer_params_target_view_info poptvi ON popi.params_info_id = poptvi.params_info_id
		WHERE pop.visualization_id = ? AND pop.checked = 1 AND popi.checked = 1`, visualizationID).Scan(&rows).Error
	return rows, err
}

// GetDatasetGroupsWithFields returns dataset groups with fields and chart views for a dashboard.
func (r *OuterParamsRepository) GetDatasetGroupsWithFields(visualizationID string) ([]DatasetGroupRow, error) {
	query := `
		SELECT DISTINCT
			cdg.id,
			cdg.name,
			cdg.pid,
			cdg.level,
			cdg.node_type,
			cdg.type,
			cdg.mode,
			cdtf.id AS field_id,
			cdtf.name AS field_name,
			cdtf.de_type AS field_de_type,
			cdtf.origin_name AS field_origin,
			ccv2.chart_id,
			ccv2.chart_name,
			ccv2.chart_type
		FROM core_dataset_group cdg
		INNER JOIN snapshot_core_chart_view ccv ON cdg.id = ccv.table_id AND ccv.type != 'VQuery'
		INNER JOIN snapshot_data_visualization_info dvi ON ccv.scene_id = dvi.id
		LEFT JOIN core_dataset_table_field cdtf ON cdtf.dataset_group_id = cdg.id
		LEFT JOIN (
			SELECT DISTINCT ccv.id AS chart_id, ccv.title AS chart_name, ccv.type AS chart_type, ccv.table_id
			FROM snapshot_core_chart_view ccv
			INNER JOIN snapshot_data_visualization_info dvi ON ccv.scene_id = dvi.id
			WHERE ccv.type != 'VQuery' AND dvi.id = ? AND LOCATE(CONCAT('"', ccv.id, '"'), dvi.component_data)
		) ccv2 ON ccv2.table_id = cdg.id
		WHERE ccv.scene_id = ? AND dvi.id = ?
		  AND LOCATE(CONCAT('"', ccv.id, '"'), dvi.component_data)
		ORDER BY cdtf.de_type, cdtf.origin_name`

	var rows []DatasetGroupRow
	err := r.db.Raw(query, visualizationID, visualizationID, visualizationID).Scan(&rows).Error
	return rows, err
}
