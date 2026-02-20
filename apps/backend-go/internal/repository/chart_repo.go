package repository

import (
	"fmt"
	"regexp"

	"dataease/backend/internal/domain/chart"
	"dataease/backend/internal/domain/dataset"

	"gorm.io/gorm"
)

var chartTableNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type ChartRepository struct {
	db *gorm.DB
}

func NewChartRepository(db *gorm.DB) *ChartRepository {
	return &ChartRepository{db: db}
}

func (r *ChartRepository) GetByID(id int64) (*chart.CoreChartView, error) {
	var c chart.CoreChartView
	err := r.db.Model(&chart.CoreChartView{}).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *ChartRepository) Update(view *chart.CoreChartView) error {
	return r.db.Save(view).Error
}

func (r *ChartRepository) QueryRows(chartID int64, limit int) ([]map[string]interface{}, int64, error) {
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	view, err := r.GetByID(chartID)
	if err != nil {
		return nil, 0, err
	}
	if view.TableID == nil {
		return nil, 0, fmt.Errorf("chart does not bind dataset table")
	}

	var dsTable struct {
		TableName string `gorm:"column:table_name"`
	}
	err = r.db.Table("core_dataset_table").
		Select("table_name").
		Where("id = ?", *view.TableID).
		First(&dsTable).Error
	if err != nil {
		return nil, 0, err
	}
	if dsTable.TableName == "" || !chartTableNamePattern.MatchString(dsTable.TableName) {
		return nil, 0, fmt.Errorf("invalid dataset table name")
	}

	rows := make([]map[string]interface{}, 0)
	querySQL := fmt.Sprintf("SELECT * FROM `%s` LIMIT ?", dsTable.TableName)
	if err = r.db.Raw(querySQL, limit).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	var countResult struct {
		C int64 `gorm:"column:c"`
	}
	countSQL := fmt.Sprintf("SELECT COUNT(1) AS c FROM `%s`", dsTable.TableName)
	if err = r.db.Raw(countSQL).Scan(&countResult).Error; err != nil {
		return nil, 0, err
	}

	return rows, countResult.C, nil
}

func (r *ChartRepository) ListDatasetFieldsByGroup(datasetGroupID int64) ([]*dataset.CoreDatasetTableField, error) {
	list := make([]*dataset.CoreDatasetTableField, 0)
	err := r.db.Model(&dataset.CoreDatasetTableField{}).
		Where("dataset_group_id = ?", datasetGroupID).
		Where("chart_id IS NULL").
		Where("COALESCE(checked, 1) = 1").
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ChartRepository) ListDatasetFieldsByChart(chartID int64) ([]*dataset.CoreDatasetTableField, error) {
	list := make([]*dataset.CoreDatasetTableField, 0)
	err := r.db.Model(&dataset.CoreDatasetTableField{}).
		Where("chart_id = ?", chartID).
		Order("id ASC").
		Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (r *ChartRepository) GetDatasetFieldByID(id int64) (*dataset.CoreDatasetTableField, error) {
	var field dataset.CoreDatasetTableField
	err := r.db.Model(&dataset.CoreDatasetTableField{}).Where("id = ?", id).First(&field).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *ChartRepository) CountDatasetFieldName(datasetGroupID int64, name string) (int64, error) {
	var count int64
	err := r.db.Model(&dataset.CoreDatasetTableField{}).
		Where("dataset_group_id = ? AND name = ?", datasetGroupID, name).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ChartRepository) CreateDatasetField(field *dataset.CoreDatasetTableField) error {
	return r.db.Create(field).Error
}

func (r *ChartRepository) UpdateDatasetFieldNames(id int64, dataeaseName string, fieldShortName string) error {
	updates := map[string]interface{}{
		"dataease_name":    dataeaseName,
		"field_short_name": fieldShortName,
	}
	return r.db.Model(&dataset.CoreDatasetTableField{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ChartRepository) DeleteDatasetField(id int64) error {
	return r.db.Where("id = ?", id).Delete(&dataset.CoreDatasetTableField{}).Error
}

func (r *ChartRepository) DeleteDatasetFieldsByChart(chartID int64) error {
	return r.db.Where("chart_id = ?", chartID).Delete(&dataset.CoreDatasetTableField{}).Error
}
