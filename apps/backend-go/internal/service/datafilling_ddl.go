package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"gorm.io/gorm"
)

type DDLProvider interface {
	CreateTable(ctx context.Context, db *gorm.DB, tableName string, fields []datafillingdomain.ExtTableField) error
	DropTable(ctx context.Context, db *gorm.DB, tableName string) error
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
		return fmt.Errorf("invalid table name")
	}
	return db.WithContext(ctx).Exec("DROP TABLE IF EXISTS `" + tableName + "`").Error
}

func (p *MySQLDDLProvider) buildCreateTableSQL(tableName string, fields []datafillingdomain.ExtTableField) (string, error) {
	if !isValidDDLIdentifier(tableName) {
		return "", fmt.Errorf("invalid table name")
	}
	columns := []string{"`id` VARCHAR(64) PRIMARY KEY"}
	for _, field := range fields {
		if field.Removed {
			continue
		}
		mapping := field.Settings.Mapping
		if !isValidDDLIdentifier(mapping.ColumnName) {
			return "", fmt.Errorf("invalid column name")
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

func isValidDDLIdentifier(name string) bool {
	return ddlIdentifierPattern.MatchString(strings.TrimSpace(name))
}
