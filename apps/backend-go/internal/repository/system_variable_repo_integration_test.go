//go:build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemVariableRepository_CreateQueryAndDelete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&system.SysVariable{}, &system.SysVariableValue{}))
	cleanupTables("sys_variable_value", "sys_variable")

	repo := NewSystemVariableRepository(testDB)
	variable := &system.SysVariable{Type: "text", Name: "region", Root: false, Disabled: false}
	require.NoError(t, repo.Create(variable))

	value := &system.SysVariableValue{SysVariableID: variable.ID, Value: "华东", ValueDesc: "east"}
	require.NoError(t, repo.CreateValue(value))

	list, err := repo.Query(&system.SysVariableQueryRequest{Name: "reg"})
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, variable.ID, list[0].ID)

	values, total, err := repo.PageValues(1, 10, &system.SysVariableValueQueryRequest{SysVariableID: variable.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	require.Len(t, values, 1)

	require.NoError(t, repo.Delete(variable.ID))
	_, err = repo.GetByID(variable.ID)
	assert.Error(t, err)
	remaining, err := repo.ListValuesByVariableID(variable.ID)
	require.NoError(t, err)
	assert.Empty(t, remaining)
}
