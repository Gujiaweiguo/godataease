package repository

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

// Table name constants for link jump queries.
const (
	ljChartViewSnap  = "snapshot_core_chart_view"
	ljChartViewCore  = "core_chart_view"
	ljJumpSnap       = "snapshot_visualization_link_jump"
	ljJumpCore       = "visualization_link_jump"
	ljJumpInfoSnap   = "snapshot_visualization_link_jump_info"
	ljJumpInfoCore   = "visualization_link_jump_info"
	ljJumpTargetSnap = "snapshot_visualization_link_jump_target_view_info"
	ljJumpTargetCore = "visualization_link_jump_target_view_info"
)

// LinkJumpRepository handles database operations for chart-to-dashboard jump navigation.
type LinkJumpRepository struct {
	db *gorm.DB
}

func NewLinkJumpRepository(db *gorm.DB) *LinkJumpRepository {
	return &LinkJumpRepository{db: db}
}

// ViewTableDetailRow is a flat row from the viewTableDetailList query.
type ViewTableDetailRow struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	DvID       int64  `json:"dvId"`
	TableID    *int64 `json:"tableId"`
	FieldID    *int64 `json:"fieldId"`
	OriginName string `json:"originName"`
	FieldName  string `json:"fieldName"`
	FieldType  string `json:"fieldType"`
	DeType     string `json:"deType"`
}

// LinkJumpInfoFlatRow is a flat row from the getLinkJumpInfo query.
type LinkJumpInfoFlatRow struct {
	SourceFieldID       int64  `json:"sourceFieldId"`
	SourceDeType        int    `json:"sourceDeType"`
	SourceFieldName     string `json:"sourceFieldName"`
	InfoID              int64  `json:"id"`
	LinkJumpID          int64  `json:"linkJumpId"`
	LinkType            string `json:"linkType"`
	JumpType            string `json:"jumpType"`
	WindowSize          string `json:"windowSize"`
	TargetDvID          int64  `json:"targetDvId"`
	TargetDvType        string `json:"targetDvType"`
	Content             string `json:"content"`
	Checked             bool   `json:"checked"`
	AttachParams        bool   `json:"attachParams"`
	TargetID            int64  `json:"targetId"`
	TargetViewID        string `json:"targetViewId"`
	TargetFieldID       string `json:"targetFieldId"`
	TargetType          string `json:"targetType"`
	SourceFieldActiveID int64  `json:"sourceFieldActiveId"`
	OuterParamsName     string `json:"outerParamsName"`
}

// LinkJumpWithDvIDRow represents the result of queryWithDvId / queryWithViewId.
type LinkJumpWithDvIDRow struct {
	SourceViewID int64  `json:"sourceViewId"`
	ID           int64  `json:"id"`
	SourceDvID   int64  `json:"sourceDvId"`
	LinkJumpInfo string `json:"linkJumpInfo"`
	Checked      bool   `json:"checked"`
}

// TargetJumpInfoRow represents a sourceInfo→targetInfo mapping.
type TargetJumpInfoRow struct {
	SourceInfo string `gorm:"column:sourceInfo" json:"sourceInfo"`
	TargetInfo string `gorm:"column:targetInfo" json:"targetInfo"`
}

// OutParamsJumpRow represents an outer params target for jump.
type OutParamsJumpRow struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	Type  string `json:"type"`
}

// GetTableFieldWithViewID returns dataset fields for a chart view.
func (r *LinkJumpRepository) GetTableFieldWithViewID(viewID int64) ([]DatasetFieldDTO, error) {
	var fields []DatasetFieldDTO
	err := r.db.Raw(`
		SELECT cdtf.id, cdtf.dataset_group_id AS dataset_table_id, cdtf.origin_name, cdtf.name, cdtf.de_type
		FROM core_dataset_table_field cdtf
		LEFT JOIN core_chart_view ccv ON ccv.table_id = cdtf.dataset_group_id
		WHERE ccv.id = ?`, viewID).Scan(&fields).Error
	return fields, err
}

// QueryWithViewId returns the jump config for a specific view in a dashboard.
func (r *LinkJumpRepository) QueryWithViewId(dvID, viewID int64) (*LinkJumpWithDvIDRow, error) {
	var row LinkJumpWithDvIDRow
	err := r.db.Raw(`
		SELECT ccv.id AS source_view_id,
			vlj.id,
			? AS source_dv_id,
			vlj.link_jump_info,
			IFNULL(vlj.checked, 0) AS checked
		FROM snapshot_core_chart_view ccv
		LEFT JOIN snapshot_visualization_link_jump vlj ON ccv.id = vlj.source_view_id
			AND vlj.source_dv_id = ?
		WHERE ccv.id = ?`, dvID, dvID, viewID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetLinkJumpInfo returns detailed jump info for a given link_jump ID and source_view_id.
func (r *LinkJumpRepository) GetLinkJumpInfo(linkJumpID, sourceViewID int64, snapshot bool) ([]LinkJumpInfoFlatRow, error) {
	var chartTable, jumpTable, jumpInfoTable, jumpTargetTable string
	if snapshot {
		chartTable = ljChartViewSnap
		jumpTable = ljJumpSnap
		jumpInfoTable = ljJumpInfoSnap
		jumpTargetTable = ljJumpTargetSnap
	} else {
		chartTable = ljChartViewCore
		jumpTable = ljJumpCore
		jumpInfoTable = ljJumpInfoCore
		jumpTargetTable = ljJumpTargetCore
	}

	query := fmt.Sprintf(`
		SELECT
			cdtf.id AS source_field_id,
			cdtf.de_type AS source_de_type,
			cdtf.name AS source_field_name,
			vlji.id,
			vlji.link_jump_id,
			vlji.link_type,
			vlji.jump_type,
			vlji.window_size,
			vlji.target_dv_id,
			dvi.type AS target_dv_type,
			vlji.content,
			IFNULL(vlji.checked, 0) AS checked,
			IFNULL(vlji.attach_params, 0) AS attach_params,
			vljtvi.target_id,
			vljtvi.target_view_id,
			vljtvi.target_field_id,
			vljtvi.target_type,
			vljtvi.source_field_active_id
		FROM %s ccv
		LEFT JOIN core_dataset_table_field cdtf ON ccv.table_id = cdtf.dataset_group_id
		LEFT JOIN %s vlj ON ccv.id = vlj.source_view_id AND vlj.id = ?
		LEFT JOIN %s vlji ON vlj.id = vlji.link_jump_id AND cdtf.id = vlji.source_field_id
		LEFT JOIN data_visualization_info dvi ON vlji.target_dv_id = dvi.id
		LEFT JOIN %s vljtvi ON vlji.id = vljtvi.link_jump_info_id
		WHERE ccv.id = ?
		  AND ccv.type != 'VQuery'
		ORDER BY cdtf.name`,
		chartTable, jumpTable, jumpInfoTable, jumpTargetTable)

	var rows []LinkJumpInfoFlatRow
	err := r.db.Raw(query, linkJumpID, sourceViewID).Scan(&rows).Error
	return rows, err
}

// QueryWithDvId returns all active jump configs for a dashboard.
func (r *LinkJumpRepository) QueryWithDvId(dvID int64, snapshot bool) ([]LinkJumpWithDvIDRow, error) {
	var chartTable, jumpTable string
	if snapshot {
		chartTable = ljChartViewSnap
		jumpTable = ljJumpSnap
	} else {
		chartTable = ljChartViewCore
		jumpTable = ljJumpCore
	}

	query := fmt.Sprintf(`
		SELECT ccv.id AS source_view_id,
			vlj.id,
			? AS source_dv_id,
			vlj.link_jump_info,
			IFNULL(ccv.jump_active, 0) AS checked
		FROM %s ccv
		LEFT JOIN %s vlj ON ccv.id = vlj.source_view_id
		WHERE vlj.source_dv_id = ?
		  AND ccv.jump_active = 1`, chartTable, jumpTable)

	var rows []LinkJumpWithDvIDRow
	err := r.db.Raw(query, dvID, dvID).Scan(&rows).Error
	return rows, err
}

// DeleteJumpCascadeSnapshot deletes all jump data for a view in the snapshot tables.
func (r *LinkJumpRepository) DeleteJumpCascadeSnapshot(dvID, viewID int64) error {
	err := r.db.Exec(`
		DELETE FROM snapshot_visualization_link_jump_target_view_info
		WHERE link_jump_info_id IN (
			SELECT lji.id FROM snapshot_visualization_link_jump_info lji
			JOIN snapshot_visualization_link_jump lj ON lji.link_jump_id = lj.id
			WHERE lj.source_dv_id = ? AND lj.source_view_id = ?
		)`, dvID, viewID).Error
	if err != nil {
		return fmt.Errorf("delete jump target view info: %w", err)
	}

	err = r.db.Exec(`
		DELETE FROM snapshot_visualization_link_jump_info
		WHERE link_jump_id IN (
			SELECT lj.id FROM snapshot_visualization_link_jump lj
			WHERE lj.source_dv_id = ? AND lj.source_view_id = ?
		)`, dvID, viewID).Error
	if err != nil {
		return fmt.Errorf("delete jump info: %w", err)
	}

	err = r.db.Exec(`
		DELETE FROM snapshot_visualization_link_jump
		WHERE source_dv_id = ? AND source_view_id = ?`, dvID, viewID).Error
	if err != nil {
		return fmt.Errorf("delete jump: %w", err)
	}
	return nil
}

// CreateSnapshotJump creates a new snapshot link jump record.
func (r *LinkJumpRepository) CreateSnapshotJump(j *auto.SnapshotVisualizationLinkJump) error {
	return r.db.Create(j).Error
}

// CreateSnapshotJumpInfo creates a new snapshot link jump info record.
func (r *LinkJumpRepository) CreateSnapshotJumpInfo(info *auto.SnapshotVisualizationLinkJumpInfo) error {
	return r.db.Create(info).Error
}

// CreateSnapshotJumpTargetViewInfo creates a new snapshot target view info record.
func (r *LinkJumpRepository) CreateSnapshotJumpTargetViewInfo(t *auto.SnapshotVisualizationLinkJumpTargetViewInfo) error {
	return r.db.Create(t).Error
}

// GetTargetVisualizationJumpInfo returns sourceInfo→targetInfo mappings for target dashboard navigation.
func (r *LinkJumpRepository) GetTargetVisualizationJumpInfo(sourceDvID, sourceViewID, targetDvID int64, sourceFieldID *int64, snapshot bool) ([]TargetJumpInfoRow, error) {
	var jtviTable, ljiTable, ljTable string
	if snapshot {
		jtviTable = ljJumpTargetSnap
		ljiTable = ljJumpInfoSnap
		ljTable = ljJumpSnap
	} else {
		jtviTable = ljJumpTargetCore
		ljiTable = ljJumpInfoCore
		ljTable = ljJumpCore
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT
			CONCAT(lj.source_view_id, '#', jtvi.source_field_active_id, '#', lji.source_field_id) AS sourceInfo,
			CONCAT(jtvi.target_view_id, '#', jtvi.target_field_id) AS targetInfo
		FROM %s jtvi
		LEFT JOIN %s lji ON jtvi.link_jump_info_id = lji.id
		LEFT JOIN %s lj ON lji.link_jump_id = lj.id
		WHERE lji.checked = 1
		  AND lj.source_dv_id = ?
		  AND lj.source_view_id = ?
		  AND lji.target_dv_id = ?`, jtviTable, ljiTable, ljTable)

	args := []interface{}{sourceDvID, sourceViewID, targetDvID}
	if sourceFieldID != nil {
		query += " AND lji.source_field_id = ?"
		args = append(args, *sourceFieldID)
	}

	var rows []TargetJumpInfoRow
	err := r.db.Raw(query, args...).Scan(&rows).Error
	return rows, err
}

// GetViewTableDetails returns chart views with their field details for a dashboard.
func (r *LinkJumpRepository) GetViewTableDetails(dvID int64) ([]ViewTableDetailRow, error) {
	var rows []ViewTableDetailRow
		err := r.db.Raw(`
		SELECT
			core_chart_view.id,
			core_chart_view.title,
			core_chart_view.type,
			core_chart_view.scene_id AS dv_id,
			core_dataset_table_field.id AS field_id,
			core_dataset_table_field.origin_name,
			core_dataset_table_field.name AS field_name,
			core_dataset_table_field.type AS field_type,
			core_dataset_table_field.de_type
		FROM core_chart_view
		LEFT JOIN core_dataset_table_field ON core_chart_view.table_id = core_dataset_table_field.dataset_group_id
		INNER JOIN data_visualization_info dvi ON core_chart_view.scene_id = dvi.id
		WHERE core_chart_view.scene_id = ?
		  AND core_chart_view.type != 'VQuery'
		  AND core_chart_view.table_id IS NOT NULL
		  AND dvi.id = ?
		  AND LOCATE(CONCAT('"', core_chart_view.id, '"'), dvi.component_data)`, dvID, dvID).Scan(&rows).Error
	return rows, err
}

// GetOutParamsTargetWithDvID returns outer params targets for a dashboard.
func (r *LinkJumpRepository) GetOutParamsTargetWithDvID(dvID int64) ([]OutParamsJumpRow, error) {
	var rows []OutParamsJumpRow
	err := r.db.Raw(`
		SELECT
			vopi.params_info_id AS id,
			vopi.param_name AS name,
			vopi.param_name AS title,
			'outerParams' AS type
		FROM visualization_outer_params_info vopi
		LEFT JOIN visualization_outer_params vop ON vopi.params_id = vop.params_id
		WHERE vop.visualization_id = ?`, dvID).Scan(&rows).Error
	return rows, err
}

// GetComponentData returns the component_data JSON string for a dashboard.
func (r *LinkJumpRepository) GetComponentData(dvID int64) (string, error) {
	var componentData string
	err := r.db.Raw(`SELECT component_data FROM data_visualization_info WHERE id = ?`, dvID).Scan(&componentData).Error
	return componentData, err
}

// UpdateChartJumpActive toggles jump_active on a snapshot chart view.
func (r *LinkJumpRepository) UpdateChartJumpActive(chartViewID int64, active bool) error {
	now := time.Now().UnixMilli()
	return r.db.Exec(`
		UPDATE snapshot_core_chart_view SET jump_active = ?, update_time = ? WHERE id = ?`,
		active, now, chartViewID).Error
}
