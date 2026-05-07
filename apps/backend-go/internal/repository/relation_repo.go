package repository

import (
	"fmt"

	"gorm.io/gorm"
)

type RelationRepository struct {
	db *gorm.DB
}

type RelationQueryRow struct {
	DatasourceID      *int64  `gorm:"column:datasource_id"`
	DatasourceName    *string `gorm:"column:datasource_name"`
	DatasourceCreator *string `gorm:"column:datasource_creator"`
	DatasourceUpdate  *int64  `gorm:"column:datasource_update"`
	DatasetID         *int64  `gorm:"column:dataset_id"`
	DatasetName       *string `gorm:"column:dataset_name"`
	DatasetCreator    *string `gorm:"column:dataset_creator"`
	DatasetUpdate     *int64  `gorm:"column:dataset_update"`
	ChartID           *int64  `gorm:"column:chart_id"`
	ChartName         *string `gorm:"column:chart_name"`
	ChartCreator      *string `gorm:"column:chart_creator"`
	ChartUpdate       *int64  `gorm:"column:chart_update"`
	DashboardID       *int64  `gorm:"column:dashboard_id"`
	DashboardName     *string `gorm:"column:dashboard_name"`
	DashboardCreator  *string `gorm:"column:dashboard_creator"`
	DashboardUpdate   *int64  `gorm:"column:dashboard_update"`
}

func NewRelationRepository(db *gorm.DB) *RelationRepository {
	return &RelationRepository{db: db}
}

func (r *RelationRepository) GetDatasourceRelations(id int64) ([]RelationQueryRow, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("relation repository is unavailable")
	}

	var rows []RelationQueryRow
	err := r.db.Raw(`
		SELECT
			dt.id AS dataset_id,
			dt.name AS dataset_name,
			dt.create_by AS dataset_creator,
			dt.update_time AS dataset_update,
			cv.id AS chart_id,
			cv.title AS chart_name,
			cv.create_by AS chart_creator,
			cv.update_time AS chart_update,
			cv.scene_id AS dashboard_id,
			dv.name AS dashboard_name,
			dv.create_by AS dashboard_creator,
			dv.update_time AS dashboard_update
		FROM core_dataset_table dt
		LEFT JOIN core_chart_view cv ON cv.table_id = dt.id
		LEFT JOIN data_visualization_info dv ON dv.id = cv.scene_id
		WHERE dt.datasource_id = ?
		ORDER BY dt.id ASC, cv.id ASC, cv.scene_id ASC
	`, id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *RelationRepository) GetDatasetRelations(id int64) ([]RelationQueryRow, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("relation repository is unavailable")
	}

	var rows []RelationQueryRow
	err := r.db.Raw(`
		SELECT
			dt.id AS dataset_id,
			dt.name AS dataset_name,
			dt.create_by AS dataset_creator,
			dt.update_time AS dataset_update,
			cv.id AS chart_id,
			cv.title AS chart_name,
			cv.create_by AS chart_creator,
			cv.update_time AS chart_update,
			cv.scene_id AS dashboard_id,
			dv.name AS dashboard_name,
			dv.create_by AS dashboard_creator,
			dv.update_time AS dashboard_update
		FROM core_dataset_table dt
		LEFT JOIN core_chart_view cv ON cv.table_id = dt.id
		LEFT JOIN data_visualization_info dv ON dv.id = cv.scene_id
		WHERE dt.dataset_group_id = ?
		ORDER BY dt.id ASC, cv.id ASC, cv.scene_id ASC
	`, id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *RelationRepository) GetPanelRelations(id int64) ([]RelationQueryRow, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("relation repository is unavailable")
	}

	var rows []RelationQueryRow
	err := r.db.Raw(`
		SELECT
			cv.id AS chart_id,
			cv.title AS chart_name,
			cv.create_by AS chart_creator,
			cv.update_time AS chart_update,
			dt.id AS dataset_id,
			dt.name AS dataset_name,
			dt.create_by AS dataset_creator,
			dt.update_time AS dataset_update,
			ds.id AS datasource_id,
			ds.name AS datasource_name,
			ds.create_by AS datasource_creator,
			ds.update_time AS datasource_update
		FROM core_chart_view cv
		LEFT JOIN core_dataset_table dt ON dt.id = cv.table_id
		LEFT JOIN core_datasource ds ON ds.id = dt.datasource_id
		WHERE cv.scene_id = ?
		ORDER BY cv.id ASC, dt.id ASC, ds.id ASC
	`, id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
