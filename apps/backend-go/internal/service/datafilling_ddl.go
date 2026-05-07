package service

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	datafillingdomain "dataease/backend/internal/domain/datafilling"
	"dataease/backend/internal/pkg/errno"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DDLProvider interface {
	CreateTable(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error
	DropTable(ctx context.Context, db *gorm.DB, tableName string) error
	InsertRow(ctx context.Context, db *gorm.DB, tableName string, rowData map[string]interface{}) error
	UpdateRow(ctx context.Context, db *gorm.DB, tableName string, id string, rowData map[string]interface{}) error
	DeleteRows(ctx context.Context, db *gorm.DB, tableName string, ids []string) error
	SearchRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}, limit, offset int64) ([]map[string]interface{}, error)
	CountRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}) (int64, error)
	TruncateTable(ctx context.Context, db *gorm.DB, tableName string) error
	ListColumnData(ctx context.Context, db *gorm.DB, tableName string, columnName string) ([]string, error)
	AddTableColumns(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error
	DropTableColumns(ctx context.Context, db *gorm.DB, tableName string, columnNames []string) error
}

type DatasourceConnectionProvider interface {
	GetDatasourceConnection(ctx context.Context, datasourceID int64) (*gorm.DB, error)
}

type MySQLDDLProvider struct{}

var ddlIdentifierPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func NewMySQLDDLProvider() *MySQLDDLProvider { return &MySQLDDLProvider{} }

func (p *MySQLDDLProvider) CreateTable(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error {
	sql, err := p.buildCreateTableSQL(tableName, fields)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Exec(sql).Error
}

func (p *MySQLDDLProvider) DropTable(ctx context.Context, db *gorm.DB, tableName string) error {
	if !isValidDDLIdentifier(tableName) {
		return fmt.Errorf(errno.ErrInvalidTableName)
	}
	return db.WithContext(ctx).Exec("DROP TABLE IF EXISTS `" + tableName + "`").Error
}

func (p *MySQLDDLProvider) InsertRow(ctx context.Context, db *gorm.DB, tableName string, rowData map[string]interface{}) error {
	sql, args, err := p.buildInsertRowSQL(tableName, rowData)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Exec(sql, args...).Error
}

func (p *MySQLDDLProvider) UpdateRow(ctx context.Context, db *gorm.DB, tableName string, id string, rowData map[string]interface{}) error {
	sql, args, err := p.buildUpdateRowSQL(tableName, id, rowData)
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Exec(sql, args...).Error
}

func (p *MySQLDDLProvider) DeleteRows(ctx context.Context, db *gorm.DB, tableName string, ids []string) error {
	sql, args, err := p.buildDeleteRowsSQL(tableName, ids)
	if err != nil || sql == "" {
		return err
	}
	return db.WithContext(ctx).Exec(sql, args...).Error
}

func (p *MySQLDDLProvider) SearchRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}, limit, offset int64) ([]map[string]interface{}, error) {
	sql, queryArgs, err := p.buildSearchRowsSQL(tableName, whereClause, args, limit, offset)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]interface{}, 0)
	if err := db.WithContext(ctx).Raw(sql, queryArgs...).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (p *MySQLDDLProvider) CountRows(ctx context.Context, db *gorm.DB, tableName string, whereClause string, args []interface{}) (int64, error) {
	sql, queryArgs, err := p.buildCountRowsSQL(tableName, whereClause, args)
	if err != nil {
		return 0, err
	}
	var total int64
	if err := db.WithContext(ctx).Raw(sql, queryArgs...).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (p *MySQLDDLProvider) TruncateTable(ctx context.Context, db *gorm.DB, tableName string) error {
	if !isValidDDLIdentifier(tableName) {
		return fmt.Errorf(errno.ErrInvalidTableName)
	}
	return db.WithContext(ctx).Exec("TRUNCATE TABLE `" + tableName + "`").Error
}

func (p *MySQLDDLProvider) ListColumnData(ctx context.Context, db *gorm.DB, tableName string, columnName string) ([]string, error) {
	sql, err := p.buildListColumnDataSQL(tableName, columnName)
	if err != nil {
		return nil, err
	}
	rows := make([]string, 0)
	if err := db.WithContext(ctx).Raw(sql).Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (p *MySQLDDLProvider) AddTableColumns(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error {
	sql, err := p.buildAddTableColumnsSQL(tableName, fields)
	if err != nil || sql == "" {
		return err
	}
	return db.WithContext(ctx).Exec(sql).Error
}

func (p *MySQLDDLProvider) DropTableColumns(ctx context.Context, db *gorm.DB, tableName string, columnNames []string) error {
	sql, err := p.buildDropTableColumnsSQL(tableName, columnNames)
	if err != nil || sql == "" {
		return err
	}
	return db.WithContext(ctx).Exec(sql).Error
}

func (p *MySQLDDLProvider) buildCreateTableSQL(tableName string, fields []datafillingdomain.ExtTableField) (string, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", fmt.Errorf(errno.ErrInvalidTableName)
	}
	columns := []string{"`id` VARCHAR(64) PRIMARY KEY"}
	for _, field := range fields {
		if field.Removed {
			continue
		}
		mapping := field.Settings.Mapping
		if !isValidDDLIdentifier(mapping.ColumnName) {
			return "", fmt.Errorf(errno.ErrInvalidColumnName)
		}
		columnType, err := mysqlColumnType(mapping)
		if err != nil {
			return "", err
		}
		definition := fmt.Sprintf("`%s` %s", mapping.ColumnName, columnType)
		if field.Settings.Required {
			definition += " NOT NULL"
		}
		if field.Settings.Unique {
			definition += " UNIQUE"
		}
		columns = append(columns, definition)
	}
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s)", tableName, strings.Join(columns, ", ")), nil
}

func (p *MySQLDDLProvider) buildInsertRowSQL(tableName string, rowData map[string]interface{}) (string, []interface{}, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", nil, fmt.Errorf(errno.ErrInvalidTableName)
	}
	prepared := copyRowData(rowData)
	if id, ok := prepared["id"]; !ok || strings.TrimSpace(fmt.Sprint(id)) == "" {
		prepared["id"] = uuid.NewString()
		if rowData != nil {
			rowData["id"] = prepared["id"]
		}
	}
	columns, args, err := sortedRowColumns(prepared)
	if err != nil {
		return "", nil, err
	}
	placeholders := make([]string, 0, len(columns))
	quotedColumns := make([]string, 0, len(columns))
	for _, column := range columns {
		placeholders = append(placeholders, "?")
		quotedColumns = append(quotedColumns, "`"+column+"`")
	}
	return fmt.Sprintf("INSERT INTO `%s` (%s) VALUES (%s)", tableName, strings.Join(quotedColumns, ", "), strings.Join(placeholders, ", ")), args, nil
}

func (p *MySQLDDLProvider) buildUpdateRowSQL(tableName string, id string, rowData map[string]interface{}) (string, []interface{}, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", nil, fmt.Errorf(errno.ErrInvalidTableName)
	}
	if strings.TrimSpace(id) == "" {
		return "", nil, fmt.Errorf("invalid row id")
	}
	prepared := copyRowData(rowData)
	delete(prepared, "id")
	columns, args, err := sortedRowColumns(prepared)
	if err != nil {
		return "", nil, err
	}
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no columns to update")
	}
	sets := make([]string, 0, len(columns))
	for _, column := range columns {
		sets = append(sets, fmt.Sprintf("`%s` = ?", column))
	}
	args = append(args, id)
	return fmt.Sprintf("UPDATE `%s` SET %s WHERE `id` = ?", tableName, strings.Join(sets, ", ")), args, nil
}

func (p *MySQLDDLProvider) buildDeleteRowsSQL(tableName string, ids []string) (string, []interface{}, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", nil, fmt.Errorf(errno.ErrInvalidTableName)
	}
	cleaned := make([]string, 0, len(ids))
	for _, id := range ids {
		if trimmed := strings.TrimSpace(id); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	if len(cleaned) == 0 {
		return "", nil, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(cleaned)), ",")
	args := make([]interface{}, 0, len(cleaned))
	for _, id := range cleaned {
		args = append(args, id)
	}
	return fmt.Sprintf("DELETE FROM `%s` WHERE `id` IN (%s)", tableName, placeholders), args, nil
}

func (p *MySQLDDLProvider) buildSearchRowsSQL(tableName string, whereClause string, args []interface{}, limit, offset int64) (string, []interface{}, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", nil, fmt.Errorf(errno.ErrInvalidTableName)
	}
	queryArgs := append([]interface{}{}, args...)
	sql := fmt.Sprintf("SELECT * FROM `%s`", tableName)
	if strings.TrimSpace(whereClause) != "" {
		sql += " WHERE " + whereClause
	}
	sql += " LIMIT ? OFFSET ?"
	queryArgs = append(queryArgs, limit, offset)
	return sql, queryArgs, nil
}

func (p *MySQLDDLProvider) buildCountRowsSQL(tableName string, whereClause string, args []interface{}) (string, []interface{}, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", nil, fmt.Errorf(errno.ErrInvalidTableName)
	}
	sql := fmt.Sprintf("SELECT COUNT(*) FROM `%s`", tableName)
	if strings.TrimSpace(whereClause) != "" {
		sql += " WHERE " + whereClause
	}
	return sql, append([]interface{}{}, args...), nil
}

func (p *MySQLDDLProvider) buildListColumnDataSQL(tableName string, columnName string) (string, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", fmt.Errorf(errno.ErrInvalidTableName)
	}
	if !isValidDDLIdentifier(columnName) {
		return "", fmt.Errorf(errno.ErrInvalidColumnName)
	}
	return fmt.Sprintf("SELECT DISTINCT `%s` FROM `%s` ORDER BY `%s` ASC", columnName, tableName, columnName), nil
}

func (p *MySQLDDLProvider) buildAddTableColumnsSQL(tableName string, fields []datafillingdomain.ExtTableField) (string, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", fmt.Errorf(errno.ErrInvalidTableName)
	}
	clauses := make([]string, 0)
	for _, field := range fields {
		if field.Removed {
			continue
		}
		columnDefinition, err := buildColumnDefinition(field)
		if err != nil {
			return "", err
		}
		clauses = append(clauses, "ADD COLUMN "+columnDefinition)
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return fmt.Sprintf("ALTER TABLE `%s` %s", tableName, strings.Join(clauses, ", ")), nil
}

func (p *MySQLDDLProvider) buildDropTableColumnsSQL(tableName string, columnNames []string) (string, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", fmt.Errorf(errno.ErrInvalidTableName)
	}
	clauses := make([]string, 0, len(columnNames))
	for _, name := range columnNames {
		if !isValidDDLIdentifier(name) {
			return "", fmt.Errorf(errno.ErrInvalidColumnName)
		}
		if name == "id" {
			return "", fmt.Errorf("cannot drop primary key column")
		}
		clauses = append(clauses, "DROP COLUMN `"+name+"`")
	}
	if len(clauses) == 0 {
		return "", nil
	}
	return fmt.Sprintf("ALTER TABLE `%s` %s", tableName, strings.Join(clauses, ", ")), nil
}

func mysqlColumnType(mapping datafillingdomain.ExtTableFieldMapping) (string, error) {
	switch mapping.Type {
	case datafillingdomain.BaseTypeNvarchar:
		size := mapping.Size
		if size <= 0 {
			size = 255
		}
		return fmt.Sprintf("VARCHAR(%d)", size), nil
	case datafillingdomain.BaseTypeText:
		return "TEXT", nil
	case datafillingdomain.BaseTypeNumber:
		return "BIGINT", nil
	case datafillingdomain.BaseTypeDecimal:
		size := mapping.Size
		if size <= 0 {
			size = 19
		}
		accuracy := mapping.Accuracy
		if accuracy < 0 {
			accuracy = 0
		}
		return fmt.Sprintf("DECIMAL(%d,%d)", size, accuracy), nil
	case datafillingdomain.BaseTypeDatetime:
		return "DATETIME", nil
	default:
		return "", fmt.Errorf("unsupported base type: %s", mapping.Type)
	}
}

func buildColumnDefinition(field datafillingdomain.ExtTableField) (string, error) {
	mapping := field.Settings.Mapping
	if !isValidDDLIdentifier(mapping.ColumnName) {
		return "", fmt.Errorf(errno.ErrInvalidColumnName)
	}
	columnType, err := mysqlColumnType(mapping)
	if err != nil {
		return "", err
	}
	definition := fmt.Sprintf("`%s` %s", mapping.ColumnName, columnType)
	if field.Settings.Required {
		definition += " NOT NULL"
	}
	if field.Settings.Unique {
		definition += " UNIQUE"
	}
	return definition, nil
}

func buildWhereClause(params []datafillingdomain.SearchParam) (string, []interface{}, error) {
	if len(params) == 0 {
		return "", nil, nil
	}
	clauses := make([]string, 0, len(params))
	args := make([]interface{}, 0)
	for _, param := range params {
		field := strings.TrimSpace(param.Field)
		if !isValidDDLIdentifier(field) {
			return "", nil, fmt.Errorf("invalid field name")
		}
		column := "`" + field + "`"
		term := strings.ToLower(strings.TrimSpace(param.Term))
		switch {
		case param.Multiple:
			if len(param.Values) == 0 {
				return "", nil, fmt.Errorf("missing values for IN")
			}
			placeholders := make([]string, 0, len(param.Values))
			for _, value := range param.Values {
				placeholders = append(placeholders, "?")
				args = append(args, value)
			}
			clauses = append(clauses, fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", ")))
		case term == "eq":
			clauses = append(clauses, column+" = ?")
			args = append(args, param.Value)
		case term == "not_eq":
			clauses = append(clauses, column+" <> ?")
			args = append(args, param.Value)
		case term == "lt":
			clauses = append(clauses, column+" < ?")
			args = append(args, param.Value)
		case term == "gt":
			clauses = append(clauses, column+" > ?")
			args = append(args, param.Value)
		case term == "le":
			clauses = append(clauses, column+" <= ?")
			args = append(args, param.Value)
		case term == "ge":
			clauses = append(clauses, column+" >= ?")
			args = append(args, param.Value)
		case term == "null":
			clauses = append(clauses, column+" IS NULL")
		case term == "not_null":
			clauses = append(clauses, column+" IS NOT NULL")
		default:
			return "", nil, fmt.Errorf("unsupported search term: %s", param.Term)
		}
	}
	return strings.Join(clauses, " AND "), args, nil
}

func copyRowData(rowData map[string]interface{}) map[string]interface{} {
	cloned := make(map[string]interface{}, len(rowData))
	for key, value := range rowData {
		cloned[key] = value
	}
	return cloned
}

func sortedRowColumns(rowData map[string]interface{}) ([]string, []interface{}, error) {
	columns := make([]string, 0, len(rowData))
	for column := range rowData {
		if !isValidDDLIdentifier(column) {
			return nil, nil, fmt.Errorf(errno.ErrInvalidColumnName)
		}
		columns = append(columns, column)
	}
	sort.Strings(columns)
	args := make([]interface{}, 0, len(columns))
	for _, column := range columns {
		args = append(args, rowData[column])
	}
	return columns, args, nil
}

func isValidDDLIdentifier(name string) bool {
	return ddlIdentifierPattern.MatchString(strings.TrimSpace(name))
}
