package repository

import (
	"fmt"
	"time"

	"dataease/backend/internal/domain/auto"

	"gorm.io/gorm"
)

const (
	tableChartView            = "core_chart_view"
	tableSnapshotChartView    = "snapshot_core_chart_view"
	tableLinkage              = "visualization_linkage"
	tableSnapshotLinkage      = "snapshot_visualization_linkage"
	tableLinkageField         = "visualization_linkage_field"
	tableSnapshotLinkageField = "snapshot_visualization_linkage_field"
)

type LinkageGatherRow struct {
	TargetViewID   int64  `json:"targetViewId"`
	TargetViewType string `json:"targetViewType"`
	TableID        int64  `json:"tableId"`
	TargetViewName string `json:"targetViewName"`
	SourceViewID   int64  `json:"sourceViewId"`
	LinkageActive  bool   `json:"linkageActive"`
	SourceField    *int64 `json:"sourceField"`
	TargetField    *int64 `json:"targetField"`
}

type DatasetFieldDTO struct {
	ID             int64  `json:"id"`
	DatasetTableID int64  `json:"datasetTableId"`
	OriginName     string `json:"originName"`
	Name           string `json:"name"`
	DeType         int    `json:"deType"`
}

type LinkageRepository struct {
	db *gorm.DB
}

func NewLinkageRepository(db *gorm.DB) *LinkageRepository {
	return &LinkageRepository{db: db}
}

func (r *LinkageRepository) GetViewLinkageGather(dvID, sourceViewID int64, targetViewIDs []int64, snapshot bool) ([]LinkageGatherRow, error) {
	if len(targetViewIDs) == 0 {
		return nil, nil
	}

	var chartTable, linkageTable, linkageFieldTable string
	if snapshot {
		chartTable = tableSnapshotChartView
		linkageTable = tableSnapshotLinkage
		linkageFieldTable = tableSnapshotLinkageField
	} else {
		chartTable = tableChartView
		linkageTable = tableLinkage
		linkageFieldTable = tableLinkageField
	}

	query := fmt.Sprintf(`
		SELECT
			ccv.title AS target_view_name,
			ccv.id AS target_view_id,
			ccv.type AS target_view_type,
			ccv.table_id,
			vl.source_view_id,
			CASE WHEN vl.target_view_id IS NULL THEN 0 ELSE vl.linkage_active END AS linkage_active,
			vlf.source_field,
			vlf.target_field
		FROM %s ccv
		LEFT JOIN %s vl ON ccv.id = vl.target_view_id
			AND vl.dv_id = ? AND vl.source_view_id = ?
		LEFT JOIN %s vlf ON vl.id = vlf.linkage_id
		WHERE ccv.type != 'VQuery' AND ccv.id IN (?)`,
		chartTable, linkageTable, linkageFieldTable)

	var rows []LinkageGatherRow
	err := r.db.Raw(query, dvID, sourceViewID, targetViewIDs).Scan(&rows).Error
	return rows, err
}

func (r *LinkageRepository) GetDatasetFieldsByGroupID(datasetGroupID int64) ([]DatasetFieldDTO, error) {
	var fields []DatasetFieldDTO
	err := r.db.Raw(`
		SELECT id, dataset_group_id AS dataset_table_id, origin_name, name, de_type
		FROM core_dataset_table_field
		WHERE dataset_group_id = ?`, datasetGroupID).Scan(&fields).Error
	return fields, err
}

func (r *LinkageRepository) DeleteLinkageAndFields(dvID, sourceViewID int64) error {
	err := r.db.Exec(fmt.Sprintf(`
		DELETE FROM %s
		WHERE linkage_id IN (
			SELECT id FROM %s
			WHERE dv_id = ? AND source_view_id = ?
		)`, tableSnapshotLinkageField, tableSnapshotLinkage), dvID, sourceViewID).Error
	if err != nil {
		return fmt.Errorf("delete linkage fields: %w", err)
	}

	err = r.db.Exec(fmt.Sprintf(`
		DELETE FROM %s
		WHERE dv_id = ? AND source_view_id = ?`, tableSnapshotLinkage), dvID, sourceViewID).Error
	if err != nil {
		return fmt.Errorf("delete linkage records: %w", err)
	}
	return nil
}

func (r *LinkageRepository) CreateLinkage(l *auto.SnapshotVisualizationLinkage) error {
	return r.db.Create(l).Error
}

func (r *LinkageRepository) CreateLinkageField(f *auto.SnapshotVisualizationLinkageField) error {
	return r.db.Create(f).Error
}

func (r *LinkageRepository) GetAllLinkageInfo(dvID int64, snapshot bool) (map[string][]string, error) {
	var linkageTable, chartTable, fieldTable string
	if snapshot {
		linkageTable = tableSnapshotLinkage
		chartTable = tableSnapshotChartView
		fieldTable = tableSnapshotLinkageField
	} else {
		linkageTable = tableLinkage
		chartTable = tableChartView
		fieldTable = tableLinkageField
	}

	query := fmt.Sprintf(`
		SELECT DISTINCT
			CONCAT(vl.source_view_id, '#', vlf.source_field) AS source_info,
			CONCAT(vl.target_view_id, '#', vlf.target_field) AS target_info
		FROM %s vl
		LEFT JOIN %s ccv ON vl.source_view_id = ccv.id
		LEFT JOIN %s vlf ON vl.id = vlf.linkage_id
		WHERE vl.dv_id = ?
			AND ccv.linkage_active = 1
			AND vl.linkage_active = 1
			AND vlf.id IS NOT NULL`, linkageTable, chartTable, fieldTable)

	type row struct {
		SourceInfo string
		TargetInfo string
	}
	var rows []row
	err := r.db.Raw(query, dvID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string][]string)
	for _, rr := range rows {
		result[rr.SourceInfo] = append(result[rr.SourceInfo], rr.TargetInfo)
	}
	return result, nil
}

func (r *LinkageRepository) UpdateChartLinkageActive(chartViewID int64, active bool) error {
	now := time.Now().UnixMilli()
	return r.db.Exec(fmt.Sprintf(`
		UPDATE %s SET linkage_active = ?, update_time = ? WHERE id = ?`,
		tableSnapshotChartView), active, now, chartViewID).Error
}
