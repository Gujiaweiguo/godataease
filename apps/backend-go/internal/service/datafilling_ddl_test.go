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

func TestMySQLDDLProviderBuildDMLSQL(t *testing.T) {
	provider := NewMySQLDDLProvider()

	insertSQL, insertArgs, err := provider.buildInsertRowSQL("df_demo", map[string]interface{}{"name": "alice", "age": 18, "id": "row-1"})
	require.NoError(t, err)
	assert.Equal(t, "INSERT INTO `df_demo` (`age`, `id`, `name`) VALUES (?, ?, ?)", insertSQL)
	assert.Equal(t, []interface{}{18, "row-1", "alice"}, insertArgs)

	updateSQL, updateArgs, err := provider.buildUpdateRowSQL("df_demo", "row-1", map[string]interface{}{"id": "row-1", "name": "alice"})
	require.NoError(t, err)
	assert.Equal(t, "UPDATE `df_demo` SET `name` = ? WHERE `id` = ?", updateSQL)
	assert.Equal(t, []interface{}{"alice", "row-1"}, updateArgs)

	deleteSQL, deleteArgs, err := provider.buildDeleteRowsSQL("df_demo", []string{"a", "b"})
	require.NoError(t, err)
	assert.Equal(t, "DELETE FROM `df_demo` WHERE `id` IN (?,?)", deleteSQL)
	assert.Equal(t, []interface{}{"a", "b"}, deleteArgs)

	searchSQL, searchArgs, err := provider.buildSearchRowsSQL("df_demo", "`name` = ?", []interface{}{"alice"}, 10, 20)
	require.NoError(t, err)
	assert.Equal(t, "SELECT * FROM `df_demo` WHERE `name` = ? LIMIT ? OFFSET ?", searchSQL)
	assert.Equal(t, []interface{}{"alice", int64(10), int64(20)}, searchArgs)

	countSQL, countArgs, err := provider.buildCountRowsSQL("df_demo", "`name` = ?", []interface{}{"alice"})
	require.NoError(t, err)
	assert.Equal(t, "SELECT COUNT(*) FROM `df_demo` WHERE `name` = ?", countSQL)
	assert.Equal(t, []interface{}{"alice"}, countArgs)

	columnSQL, err := provider.buildListColumnDataSQL("df_demo", "name")
	require.NoError(t, err)
	assert.Equal(t, "SELECT DISTINCT `name` FROM `df_demo` ORDER BY `name` ASC", columnSQL)

	addSQL, err := provider.buildAddTableColumnsSQL("df_demo", []datafillingdomain.ExtTableField{{Settings: datafillingdomain.ExtTableFieldSetting{Required: true, Mapping: datafillingdomain.ExtTableFieldMapping{ColumnName: "nickname", Type: datafillingdomain.BaseTypeNvarchar, Size: 32}}}})
	require.NoError(t, err)
	assert.Equal(t, "ALTER TABLE `df_demo` ADD COLUMN `nickname` VARCHAR(32) NOT NULL", addSQL)

	dropSQL, err := provider.buildDropTableColumnsSQL("df_demo", []string{"nickname", "age"})
	require.NoError(t, err)
	assert.Equal(t, "ALTER TABLE `df_demo` DROP COLUMN `nickname`, DROP COLUMN `age`", dropSQL)
}

func TestBuildWhereClause(t *testing.T) {
	tests := []struct {
		name     string
		params   []datafillingdomain.SearchParam
		wantSQL  string
		wantArgs []interface{}
		wantErr  bool
	}{
		{name: "empty", params: nil, wantSQL: "", wantArgs: nil},
		{name: "eq", params: []datafillingdomain.SearchParam{{Field: "name", Term: "eq", Value: "alice"}}, wantSQL: "`name` = ?", wantArgs: []interface{}{"alice"}},
		{name: "not_eq", params: []datafillingdomain.SearchParam{{Field: "name", Term: "not_eq", Value: "alice"}}, wantSQL: "`name` <> ?", wantArgs: []interface{}{"alice"}},
		{name: "lt", params: []datafillingdomain.SearchParam{{Field: "age", Term: "lt", Value: 18}}, wantSQL: "`age` < ?", wantArgs: []interface{}{18}},
		{name: "gt", params: []datafillingdomain.SearchParam{{Field: "age", Term: "gt", Value: 18}}, wantSQL: "`age` > ?", wantArgs: []interface{}{18}},
		{name: "le", params: []datafillingdomain.SearchParam{{Field: "age", Term: "le", Value: 18}}, wantSQL: "`age` <= ?", wantArgs: []interface{}{18}},
		{name: "ge", params: []datafillingdomain.SearchParam{{Field: "age", Term: "ge", Value: 18}}, wantSQL: "`age` >= ?", wantArgs: []interface{}{18}},
		{name: "null", params: []datafillingdomain.SearchParam{{Field: "deleted_at", Term: "null"}}, wantSQL: "`deleted_at` IS NULL", wantArgs: []interface{}{}},
		{name: "not_null", params: []datafillingdomain.SearchParam{{Field: "deleted_at", Term: "not_null"}}, wantSQL: "`deleted_at` IS NOT NULL", wantArgs: []interface{}{}},
		{name: "in", params: []datafillingdomain.SearchParam{{Field: "id", Multiple: true, Values: []interface{}{"a", "b"}}}, wantSQL: "`id` IN (?, ?)", wantArgs: []interface{}{"a", "b"}},
		{name: "invalid field", params: []datafillingdomain.SearchParam{{Field: "bad-name", Term: "eq", Value: 1}}, wantErr: true},
		{name: "invalid term", params: []datafillingdomain.SearchParam{{Field: "name", Term: "contains", Value: "a"}}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := buildWhereClause(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantSQL, sql)
			assert.Equal(t, tt.wantArgs, args)
		})
	}
}
