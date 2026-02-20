package repository

import (
	"fmt"
	"regexp"
	"strings"

	"dataease/backend/internal/domain/dataset"

	"gorm.io/gorm"
)

var tableNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type DatasetRepository struct {
	db *gorm.DB
}

func NewDatasetRepository(db *gorm.DB) *DatasetRepository {
	return &DatasetRepository{db: db}
}

func (r *DatasetRepository) ListGroups(keyword *string) ([]*dataset.CoreDatasetGroup, error) {
	var groups []*dataset.CoreDatasetGroup
	q := r.db.Model(&dataset.CoreDatasetGroup{}).Where("COALESCE(del_flag, 0) = 0")
	if keyword != nil && *keyword != "" {
		kw := "%" + *keyword + "%"
		q = q.Where("name LIKE ?", kw)
	}
	err := q.Order("level ASC, id ASC").Find(&groups).Error
	return groups, err
}

func (r *DatasetRepository) GetGroupByID(id int64) (*dataset.CoreDatasetGroup, error) {
	var group dataset.CoreDatasetGroup
	err := r.db.Model(&dataset.CoreDatasetGroup{}).
		Where("id = ? AND COALESCE(del_flag, 0) = 0", id).
		First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *DatasetRepository) CreateGroup(group *dataset.CoreDatasetGroup) error {
	return r.db.Create(group).Error
}

func (r *DatasetRepository) UpdateGroup(group *dataset.CoreDatasetGroup) error {
	return r.db.Save(group).Error
}

func (r *DatasetRepository) CountGroupByNameAndPID(name string, pid int64, excludeID *int64) (int64, error) {
	var count int64
	q := r.db.Model(&dataset.CoreDatasetGroup{}).
		Where("name = ? AND COALESCE(pid, 0) = ? AND COALESCE(del_flag, 0) = 0", name, pid)
	if excludeID != nil && *excludeID > 0 {
		q = q.Where("id <> ?", *excludeID)
	}
	err := q.Count(&count).Error
	return count, err
}

func (r *DatasetRepository) ListGroupChildren(parentID int64) ([]*dataset.CoreDatasetGroup, error) {
	var list []*dataset.CoreDatasetGroup
	err := r.db.Model(&dataset.CoreDatasetGroup{}).
		Where("COALESCE(pid, 0) = ? AND COALESCE(del_flag, 0) = 0", parentID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *DatasetRepository) SoftDeleteGroup(id int64) error {
	return r.db.Model(&dataset.CoreDatasetGroup{}).
		Where("id = ? AND COALESCE(del_flag, 0) = 0", id).
		Update("del_flag", 1).Error
}

func (r *DatasetRepository) ListTablesByDatasetGroupID(datasetGroupID int64) ([]*dataset.CoreDatasetTable, error) {
	var tables []*dataset.CoreDatasetTable
	err := r.db.Model(&dataset.CoreDatasetTable{}).
		Where("dataset_group_id = ?", datasetGroupID).
		Order("id ASC").
		Find(&tables).Error
	return tables, err
}

func (r *DatasetRepository) PreviewSQL(rawSQL string, limit int) ([]map[string]interface{}, error) {
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	rows := make([]map[string]interface{}, 0)
	query := fmt.Sprintf("SELECT * FROM (%s) AS de_preview LIMIT ?", rawSQL)
	if err := r.db.Raw(query, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *DatasetRepository) CountChartRelations(datasetGroupID int64) (int64, error) {
	var count int64
	err := r.db.Table("core_chart_view AS cv").
		Joins("JOIN core_dataset_table AS dt ON dt.id = cv.table_id").
		Where("dt.dataset_group_id = ?", datasetGroupID).
		Count(&count).Error
	return count, err
}

func (r *DatasetRepository) ListFields(datasetGroupID int64) ([]*dataset.CoreDatasetTableField, error) {
	var fields []*dataset.CoreDatasetTableField
	err := r.db.Model(&dataset.CoreDatasetTableField{}).
		Where("dataset_group_id = ?", datasetGroupID).
		Order("id ASC").
		Find(&fields).Error
	return fields, err
}

func (r *DatasetRepository) GetFieldByID(id int64) (*dataset.CoreDatasetTableField, error) {
	var field dataset.CoreDatasetTableField
	err := r.db.Model(&dataset.CoreDatasetTableField{}).Where("id = ?", id).First(&field).Error
	if err != nil {
		return nil, err
	}
	return &field, nil
}

func (r *DatasetRepository) GetTableByID(id int64) (*dataset.CoreDatasetTable, error) {
	var table dataset.CoreDatasetTable
	err := r.db.Model(&dataset.CoreDatasetTable{}).Where("id = ?", id).First(&table).Error
	if err != nil {
		return nil, err
	}
	return &table, nil
}

func (r *DatasetRepository) FindPrimaryTableName(datasetGroupID int64) (string, error) {
	var table dataset.CoreDatasetTable
	err := r.db.Model(&dataset.CoreDatasetTable{}).
		Where("dataset_group_id = ?", datasetGroupID).
		Order("id ASC").
		First(&table).Error
	if err != nil {
		return "", err
	}
	if table.PhysicalTable == nil || *table.PhysicalTable == "" {
		return "", fmt.Errorf("dataset table_name is empty")
	}
	if !tableNamePattern.MatchString(*table.PhysicalTable) {
		return "", fmt.Errorf("invalid dataset table name")
	}
	return *table.PhysicalTable, nil
}

func (r *DatasetRepository) PreviewRows(tableName string, limit int) ([]map[string]interface{}, error) {
	if !tableNamePattern.MatchString(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}
	if limit < 1 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	rows := make([]map[string]interface{}, 0)
	sql := fmt.Sprintf("SELECT * FROM `%s` LIMIT ?", tableName)
	if err := r.db.Raw(sql, limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *DatasetRepository) QueryDistinctValues(tableName string, columnName string, filters []dataset.EnumFilterClause, limit int) ([]string, error) {
	if !tableNamePattern.MatchString(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}

	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return nil, err
	}
	quotedColumn, err := quoteIdentifier(columnName)
	if err != nil {
		return nil, err
	}

	query := strings.Builder{}
	query.WriteString("SELECT DISTINCT ")
	query.WriteString(quotedColumn)
	query.WriteString(" AS `de_value` FROM ")
	query.WriteString(quotedTable)

	args := make([]interface{}, 0)
	whereParts := make([]string, 0)
	for _, filter := range filters {
		if strings.TrimSpace(filter.Column) == "" || len(filter.Values) == 0 {
			continue
		}
		quotedFilterColumn, quoteErr := quoteIdentifier(filter.Column)
		if quoteErr != nil {
			continue
		}
		placeholders := make([]string, 0, len(filter.Values))
		for _, value := range filter.Values {
			placeholders = append(placeholders, "?")
			args = append(args, value)
		}
		whereParts = append(whereParts, fmt.Sprintf("%s IN (%s)", quotedFilterColumn, strings.Join(placeholders, ", ")))
	}

	if len(whereParts) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(whereParts, " AND "))
	}
	query.WriteString(" ORDER BY ")
	query.WriteString(quotedColumn)
	query.WriteString(" ASC")
	if limit > 0 {
		query.WriteString(" LIMIT ?")
		args = append(args, limit)
	}

	type row struct {
		Value *string `gorm:"column:de_value"`
	}
	rows := make([]row, 0)
	if err = r.db.Raw(query.String(), args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]string, 0, len(rows))
	for _, item := range rows {
		if item.Value == nil {
			continue
		}
		result = append(result, *item.Value)
	}
	return result, nil
}

func (r *DatasetRepository) QueryDistinctObjectValues(tableName string, columns []dataset.EnumObjectColumn, filters []dataset.EnumFilterClause, searchColumn string, searchText string, sortColumn string, sortDirection string, limit int) ([]map[string]interface{}, error) {
	if !tableNamePattern.MatchString(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}
	if len(columns) == 0 {
		return []map[string]interface{}{}, nil
	}

	quotedTable, err := quoteIdentifier(tableName)
	if err != nil {
		return nil, err
	}

	selectParts := make([]string, 0, len(columns))
	for _, column := range columns {
		quotedColumn, quoteErr := quoteIdentifier(column.Column)
		if quoteErr != nil {
			return nil, quoteErr
		}
		quotedAlias, quoteErr := quoteIdentifier(column.Alias)
		if quoteErr != nil {
			return nil, quoteErr
		}
		selectParts = append(selectParts, fmt.Sprintf("%s AS %s", quotedColumn, quotedAlias))
	}

	query := strings.Builder{}
	query.WriteString("SELECT DISTINCT ")
	query.WriteString(strings.Join(selectParts, ", "))
	query.WriteString(" FROM ")
	query.WriteString(quotedTable)

	args := make([]interface{}, 0)
	whereParts := make([]string, 0)
	for _, filter := range filters {
		if strings.TrimSpace(filter.Column) == "" || len(filter.Values) == 0 {
			continue
		}
		quotedFilterColumn, quoteErr := quoteIdentifier(filter.Column)
		if quoteErr != nil {
			continue
		}
		placeholders := make([]string, 0, len(filter.Values))
		for _, value := range filter.Values {
			placeholders = append(placeholders, "?")
			args = append(args, value)
		}
		whereParts = append(whereParts, fmt.Sprintf("%s IN (%s)", quotedFilterColumn, strings.Join(placeholders, ", ")))
	}

	if strings.TrimSpace(searchText) != "" && strings.TrimSpace(searchColumn) != "" {
		quotedSearchColumn, quoteErr := quoteIdentifier(searchColumn)
		if quoteErr == nil {
			whereParts = append(whereParts, fmt.Sprintf("%s LIKE ?", quotedSearchColumn))
			args = append(args, "%"+strings.TrimSpace(searchText)+"%")
		}
	}

	if len(whereParts) > 0 {
		query.WriteString(" WHERE ")
		query.WriteString(strings.Join(whereParts, " AND "))
	}

	if strings.TrimSpace(sortColumn) != "" {
		quotedSortColumn, quoteErr := quoteIdentifier(sortColumn)
		if quoteErr == nil {
			direction := strings.ToUpper(strings.TrimSpace(sortDirection))
			if direction != "DESC" {
				direction = "ASC"
			}
			query.WriteString(" ORDER BY ")
			query.WriteString(quotedSortColumn)
			query.WriteString(" ")
			query.WriteString(direction)
		}
	}
	if limit > 0 {
		query.WriteString(" LIMIT ?")
		args = append(args, limit)
	}

	rows := make([]map[string]interface{}, 0)
	if err = r.db.Raw(query.String(), args...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *DatasetRepository) CountRows(tableName string) (int64, error) {
	if !tableNamePattern.MatchString(tableName) {
		return 0, fmt.Errorf("invalid table name")
	}
	var result struct {
		C int64 `gorm:"column:c"`
	}
	sql := fmt.Sprintf("SELECT COUNT(1) AS c FROM `%s`", tableName)
	if err := r.db.Raw(sql).Scan(&result).Error; err != nil {
		return 0, err
	}
	return result.C, nil
}

func quoteIdentifier(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("invalid identifier")
	}
	if strings.ContainsRune(trimmed, 0) {
		return "", fmt.Errorf("invalid identifier")
	}
	return "`" + strings.ReplaceAll(trimmed, "`", "``") + "`", nil
}
