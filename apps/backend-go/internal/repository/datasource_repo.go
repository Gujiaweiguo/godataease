package repository

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"dataease/backend/internal/domain/datasource"

	"gorm.io/gorm"
)

var datasourceTableNamePattern = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

type datasourceTable struct {
	ID           int64   `gorm:"column:id"`
	Name         string  `gorm:"column:name"`
	PhysicalName string  `gorm:"column:table_name"`
	DatasourceID int64   `gorm:"column:datasource_id"`
	Type         *string `gorm:"column:type"`
}

func (datasourceTable) TableName() string {
	return "core_dataset_table"
}

type DatasourceRepository struct {
	db *gorm.DB
}

func NewDatasourceRepository(db *gorm.DB) *DatasourceRepository {
	return &DatasourceRepository{db: db}
}

func (r *DatasourceRepository) Query(req *datasource.ListRequest) ([]*datasource.CoreDatasource, int64, error) {
	var list []*datasource.CoreDatasource
	var total int64

	page := req.Current
	if page < 1 {
		page = 1
	}
	pageSize := req.Size
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	query := r.db.Model(&datasource.CoreDatasource{})
	query = query.Where("COALESCE(del_flag, 0) = 0")
	if req.Keyword != nil && *req.Keyword != "" {
		kw := "%" + *req.Keyword + "%"
		query = query.Where("name LIKE ?", kw)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *DatasourceRepository) GetByID(id int64) (*datasource.CoreDatasource, error) {
	var ds datasource.CoreDatasource
	if err := r.db.Where("id = ? AND COALESCE(del_flag, 0) = 0", id).First(&ds).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

func (r *DatasourceRepository) Create(ds *datasource.CoreDatasource) error {
	return r.db.Create(ds).Error
}

func (r *DatasourceRepository) Update(ds *datasource.CoreDatasource) error {
	return r.db.Save(ds).Error
}

func (r *DatasourceRepository) SoftDelete(id int64) error {
	return r.db.Model(&datasource.CoreDatasource{}).
		Where("id = ? AND COALESCE(del_flag, 0) = 0", id).
		Update("del_flag", 1).Error
}

func (r *DatasourceRepository) ListChildren(parentID int64) ([]*datasource.CoreDatasource, error) {
	var list []*datasource.CoreDatasource
	err := r.db.Model(&datasource.CoreDatasource{}).
		Where("pid = ? AND COALESCE(del_flag, 0) = 0", parentID).
		Order("id ASC").
		Find(&list).Error
	return list, err
}

func (r *DatasourceRepository) CountByNameAndPID(name string, pid int64, excludeID *int64) (int64, error) {
	var count int64
	query := r.db.Model(&datasource.CoreDatasource{}).
		Where("name = ? AND pid = ? AND COALESCE(del_flag, 0) = 0", name, pid)
	if excludeID != nil && *excludeID > 0 {
		query = query.Where("id <> ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *DatasourceRepository) ListAll(keyword *string) ([]*datasource.CoreDatasource, error) {
	var list []*datasource.CoreDatasource
	query := r.db.Model(&datasource.CoreDatasource{}).Where("COALESCE(del_flag, 0) = 0")
	if keyword != nil && *keyword != "" {
		kw := "%" + *keyword + "%"
		query = query.Where("name LIKE ?", kw)
	}
	if err := query.Order("id DESC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *DatasourceRepository) ListByType(dsType string, excludeID *int64) ([]*datasource.CoreDatasource, error) {
	var list []*datasource.CoreDatasource
	query := r.db.Model(&datasource.CoreDatasource{}).
		Where("type = ? AND COALESCE(del_flag, 0) = 0", dsType)
	if excludeID != nil && *excludeID > 0 {
		query = query.Where("id <> ?", *excludeID)
	}
	if err := query.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *DatasourceRepository) ListTables(datasourceID int64) ([]datasource.TableInfo, error) {
	var rows []datasourceTable
	if err := r.db.Model(&datasourceTable{}).
		Where("datasource_id = ?", datasourceID).
		Order("id DESC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]datasource.TableInfo, 0, len(rows))
	for _, row := range rows {
		typeVal := ""
		if row.Type != nil {
			typeVal = *row.Type
		}
		result = append(result, datasource.TableInfo{
			ID:           row.ID,
			DatasourceID: row.DatasourceID,
			Name:         row.Name,
			TableName:    row.PhysicalName,
			Type:         typeVal,
		})
	}

	return result, nil
}

func (r *DatasourceRepository) ListSchemas() ([]string, error) {
	rows, err := r.db.Raw("SELECT schema_name FROM information_schema.schemata ORDER BY schema_name ASC").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]string, 0)
	for rows.Next() {
		var schemaName sql.NullString
		if err := rows.Scan(&schemaName); err != nil {
			return nil, err
		}
		if !schemaName.Valid || strings.TrimSpace(schemaName.String) == "" {
			continue
		}
		result = append(result, schemaName.String)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *DatasourceRepository) ListTableFields(tableName string) ([]datasource.TableField, error) {
	if !datasourceTableNamePattern.MatchString(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}

	type columnRow struct {
		Field string `gorm:"column:Field"`
		Type  string `gorm:"column:Type"`
	}
	rows := make([]columnRow, 0)
	sql := fmt.Sprintf("SHOW COLUMNS FROM `%s`", tableName)
	if err := r.db.Raw(sql).Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make([]datasource.TableField, 0, len(rows))
	for _, row := range rows {
		result = append(result, datasource.TableField{
			OriginName: row.Field,
			Name:       row.Field,
			Type:       row.Type,
			DeType:     inferDeType(row.Type),
		})
	}
	return result, nil
}

func (r *DatasourceRepository) PreviewRows(tableName string, limit int) ([]map[string]interface{}, error) {
	if !datasourceTableNamePattern.MatchString(tableName) {
		return nil, fmt.Errorf("invalid table name")
	}
	if limit <= 0 {
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

func (r *DatasourceRepository) CountRows(tableName string) (int64, error) {
	if !datasourceTableNamePattern.MatchString(tableName) {
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

func (r *DatasourceRepository) CountDatasourceRelations(datasourceID int64) (int64, error) {
	var count int64
	err := r.db.Table("core_dataset_table").
		Where("datasource_id = ?", datasourceID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *DatasourceRepository) ListLatestTypesByCreator(createBy string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 5
	}
	if limit > 20 {
		limit = 20
	}

	var types []string
	err := r.db.Model(&datasource.CoreDatasource{}).
		Select("DISTINCT type").
		Where("create_by = ? AND COALESCE(del_flag, 0) = 0 AND type <> ?", createBy, "folder").
		Order("create_time DESC").
		Limit(limit).
		Pluck("type", &types).Error
	if err != nil {
		return nil, err
	}
	return types, nil
}

func (r *DatasourceRepository) ExistsFinishPageRecord(userID int64) (bool, error) {
	var count int64
	err := r.db.Table("core_ds_finish_page").
		Where("id = ?", userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *DatasourceRepository) CreateFinishPageRecord(userID int64) error {
	return r.db.Exec("INSERT IGNORE INTO core_ds_finish_page (id) VALUES (?)", userID).Error
}

func inferDeType(t string) int {
	typeName := strings.ToLower(t)
	switch {
	case strings.Contains(typeName, "int"), strings.Contains(typeName, "decimal"), strings.Contains(typeName, "double"), strings.Contains(typeName, "float"):
		return 2
	case strings.Contains(typeName, "date"), strings.Contains(typeName, "time"), strings.Contains(typeName, "year"):
		return 1
	default:
		return 0
	}
}
