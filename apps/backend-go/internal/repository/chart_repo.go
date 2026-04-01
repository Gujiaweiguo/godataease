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
	tableName, _, err := r.resolveChartTable(chartID)
	if err != nil {
		return nil, 0, err
	}
	return r.queryRowsFromTable(tableName, "*", "", nil, limit)
}

func (r *ChartRepository) QueryRowsWithFilter(chartID int64, selectColumns string, whereClause string, whereArgs []interface{}, limit int) ([]map[string]interface{}, int64, error) {
	tableName, _, err := r.resolveChartTable(chartID)
	if err != nil {
		return nil, 0, err
	}
	return r.queryRowsFromTable(tableName, selectColumns, whereClause, whereArgs, limit)
}

func (r *ChartRepository) GetDatasetGroupIDByChartID(chartID int64) (int64, error) {
	_, datasetGroupID, err := r.resolveChartTable(chartID)
	if err != nil {
		return 0, err
	}
	return datasetGroupID, nil
}

func (r *ChartRepository) resolveChartTable(chartID int64) (string, int64, error) {
	view, err := r.GetByID(chartID)
	if err != nil {
		return "", 0, err
	}
	if view.TableID == nil {
		return "", 0, fmt.Errorf("chart does not bind dataset table")
	}

	var dsTable struct {
		TableName string `gorm:"column:table_name"`
	}
	err = r.db.Table("core_dataset_table").
		Select("table_name").
		Where("id = ?", *view.TableID).
		First(&dsTable).Error
	if err != nil {
		return "", 0, err
	}
	if dsTable.TableName == "" || !chartTableNamePattern.MatchString(dsTable.TableName) {
		return "", 0, fmt.Errorf("invalid dataset table name")
	}

	var datasetGroup struct {
		DatasetGroupID int64 `gorm:"column:dataset_group_id"`
	}
	err = r.db.Table("core_dataset_table").
		Select("dataset_group_id").
		Where("id = ?", *view.TableID).
		First(&datasetGroup).Error
	if err != nil {
		return "", 0, err
	}

	return dsTable.TableName, datasetGroup.DatasetGroupID, nil
}

func (r *ChartRepository) queryRowsFromTable(tableName string, selectColumns string, whereClause string, whereArgs []interface{}, limit int) ([]map[string]interface{}, int64, error) {
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	if selectColumns == "" {
		selectColumns = "*"
	}

	rows := make([]map[string]interface{}, 0)
	querySQL := fmt.Sprintf("SELECT %s FROM `%s`", selectColumns, tableName)
	args := make([]interface{}, 0, len(whereArgs)+1)
	if whereClause != "" {
		querySQL += " WHERE " + whereClause
		args = append(args, whereArgs...)
	}
	querySQL += " LIMIT ?"
	args = append(args, limit)
	if err := r.db.Raw(querySQL, args...).Scan(&rows).Error; err != nil {
		return nil, 0, err
	}

	countSQL := fmt.Sprintf("SELECT COUNT(1) AS c FROM `%s`", tableName)
	countArgs := make([]interface{}, 0, len(whereArgs))
	if whereClause != "" {
		countSQL += " WHERE " + whereClause
		countArgs = append(countArgs, whereArgs...)
	}
	var countResult struct {
		C int64 `gorm:"column:c"`
	}
	if err := r.db.Raw(countSQL, countArgs...).Scan(&countResult).Error; err != nil {
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
