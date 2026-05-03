package service

import (
	"testing"

	datafillingdomain "dataease/backend/internal/domain/datafilling"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMySQLDDLProviderBuildCreateTableSQL(t *testing.T) {
	provider := NewMySQLDDLProvider()
	sql, err := provider.buildCreateTableSQL("df_demo", []datafillingdomain.ExtTableField{
		{Settings: datafillingdomain.ExtTableFieldSetting{Required: true, Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "name", Type: datafillingdomain.BaseTypeNvarchar, Size: 64}}},
		{Settings: datafillingdomain.ExtTableFieldSetting{Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "content", Type: datafillingdomain.BaseTypeText}}},
		{Settings: datafillingdomain.ExtTableFieldSetting{Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "count_num", Type: datafillingdomain.BaseTypeNumber}}},
		{Settings: datafillingdomain.ExtTableFieldSetting{Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "amount", Type: datafillingdomain.BaseTypeDecimal, Size: 12, Accuracy: 2}}},
		{Settings: datafillingdomain.ExtTableFieldSetting{Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "created_at", Type: datafillingdomain.BaseTypeDatetime}}},
	})
	require.NoError(t, err)
	assert.Contains(t, sql, "`id` VARCHAR(64) PRIMARY KEY")
	assert.Contains(t, sql, "`name` VARCHAR(64) NOT NULL")
	assert.Contains(t, sql, "`content` TEXT")
	assert.Contains(t, sql, "`count_num` BIGINT")
	assert.Contains(t, sql, "`amount` DECIMAL(12,2)")
	assert.Contains(t, sql, "`created_at` DATETIME")
}

func TestMySQLDDLProviderRejectsInvalidIdentifier(t *testing.T) {
	provider := NewMySQLDDLProvider()
	_, err := provider.buildCreateTableSQL("bad-name", nil)
	assert.Error(t, err)
}
