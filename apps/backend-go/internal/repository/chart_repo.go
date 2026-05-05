package repository

import (
	"encoding/json"
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

func (r *ChartRepository) QueryViewOption(resourceId int64) ([]chart.ViewSelectorVO, error) {
	var results []chart.ViewSelectorVO
	err := r.db.Raw(
		"SELECT id, scene_id AS pid, title, type FROM core_chart_view WHERE type != 'VQuery' AND scene_id = ?",
		resourceId,
	).Scan(&results).Error
	return results, err
}

func (r *ChartRepository) GetVisualizationComponentData(resourceId int64) (string, error) {
	var result struct{ ComponentData *string }
	err := r.db.Raw(
		"SELECT component_data FROM data_visualization_info WHERE id = ?",
		resourceId,
	).Scan(&result).Error
	if err != nil {
		return "", err
	}
	if result.ComponentData == nil {
		return "", nil
	}
	return *result.ComponentData, nil
}

func (r *ChartRepository) QueryChartBaseInfo(id int64, resourceTable string) (*chart.ChartBaseVO, error) {
	chartTable := "core_chart_view"
	dvTable := "data_visualization_info"
	if resourceTable == "snapshot" {
		chartTable = "snapshot_core_chart_view"
		dvTable = "snapshot_data_visualization_info"
	}

	type rawChartBase struct {
		ChartID          int64   `gorm:"column:chart_id"`
		ChartType        *string `gorm:"column:chart_type"`
		ChartName        *string `gorm:"column:chart_name"`
		TableID          *int64  `gorm:"column:table_id"`
		ResourceID       *int64  `gorm:"column:resource_id"`
		ResourceType     *string `gorm:"column:resource_type"`
		ResourceName     *string `gorm:"column:resource_name"`
		XAxis            *string `gorm:"column:x_axis"`
		XAxisExt         *string `gorm:"column:x_axis_ext"`
		YAxis            *string `gorm:"column:y_axis"`
		YAxisExt         *string `gorm:"column:y_axis_ext"`
		ExtStack         *string `gorm:"column:ext_stack"`
		ExtBubble        *string `gorm:"column:ext_bubble"`
		FlowMapStartName *string `gorm:"column:flow_map_start_name"`
		FlowMapEndName   *string `gorm:"column:flow_map_end_name"`
		ExtColor         *string `gorm:"column:ext_color"`
		ExtLabel         *string `gorm:"column:ext_label"`
		ExtTooltip       *string `gorm:"column:ext_tooltip"`
	}

	sql := fmt.Sprintf(`SELECT ccv.id AS chart_id, ccv.title AS chart_name, ccv.type AS chart_type, ccv.table_id,
		dvi.id AS resource_id, dvi.name AS resource_name, dvi.type AS resource_type,
		ccv.x_axis, ccv.x_axis_ext, ccv.y_axis, ccv.y_axis_ext,
		ccv.ext_stack, ccv.ext_bubble, ccv.flow_map_start_name, ccv.flow_map_end_name,
		ccv.ext_color, ccv.ext_label, ccv.ext_tooltip
		FROM %s ccv LEFT JOIN %s dvi ON dvi.id = ccv.scene_id WHERE ccv.id = ?`, chartTable, dvTable)

	var raw rawChartBase
	if err := r.db.Raw(sql, id).Scan(&raw).Error; err != nil {
		return nil, err
	}
	if raw.ChartID == 0 {
		return nil, nil
	}

	parseAxis := func(s *string) []map[string]interface{} {
		if s == nil || *s == "" || *s == "null" {
			return []map[string]interface{}{}
		}
		var result []map[string]interface{}
		if err := json.Unmarshal([]byte(*s), &result); err != nil {
			return []map[string]interface{}{}
		}
		return result
	}

	vo := &chart.ChartBaseVO{
		ChartID:          raw.ChartID,
		TableID:          raw.TableID,
		ResourceID:       ptrInt64(raw.ResourceID, 0),
		XAxis:            parseAxis(raw.XAxis),
		XAxisExt:         parseAxis(raw.XAxisExt),
		YAxis:            parseAxis(raw.YAxis),
		YAxisExt:         parseAxis(raw.YAxisExt),
		ExtStack:         parseAxis(raw.ExtStack),
		ExtBubble:        parseAxis(raw.ExtBubble),
		FlowMapStartName: parseAxis(raw.FlowMapStartName),
		FlowMapEndName:   parseAxis(raw.FlowMapEndName),
		ExtColor:         parseAxis(raw.ExtColor),
		ExtLabel:         parseAxis(raw.ExtLabel),
		ExtTooltip:       parseAxis(raw.ExtTooltip),
	}
	if raw.ChartType != nil {
		vo.ChartType = *raw.ChartType
	}
	if raw.ChartName != nil {
		vo.ChartName = *raw.ChartName
	}
	if raw.ResourceType != nil {
		vo.ResourceType = *raw.ResourceType
	}
	if raw.ResourceName != nil {
		vo.ResourceName = *raw.ResourceName
	}
	return vo, nil
}

func ptrInt64(p *int64, defaultVal int64) int64 {
	if p == nil {
		return defaultVal
	}
	return *p
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
