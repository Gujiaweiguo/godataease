//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemVariableService_FullLifecycle(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	require.NoError(t, testDB.AutoMigrate(&system.SysVariable{}, &system.SysVariableValue{}))
	cleanupTables(&system.SysVariable{})

	repo := repository.NewSystemVariableRepository(testDB)
	svc := NewSystemVariableService(repo)

	created, err := svc.Create(&system.SysVariable{Type: "num", Name: "amount", Min: 1, Max: 10})
	require.NoError(t, err)
	require.NotNil(t, created)

	created.Name = "amount2"
	updated, err := svc.Edit(created)
	require.NoError(t, err)
	assert.Equal(t, "amount2", updated.Name)

	detail, err := svc.Detail(created.ID)
	require.NoError(t, err)
	assert.Equal(t, "amount2", detail.Name)

	query, err := svc.Query(&system.SysVariableQueryRequest{Type: "num"})
	require.NoError(t, err)
	require.Len(t, query, 1)

	val, err := svc.CreateValue(&system.SysVariableValue{SysVariableID: created.ID, Value: "1", ValueDesc: "one"})
	require.NoError(t, err)
	val.ValueDesc = "ONE"
	val, err = svc.EditValue(val)
	require.NoError(t, err)
	assert.Equal(t, "ONE", val.ValueDesc)

	page, err := svc.SelectedValuePage(1, 10, &system.SysVariableValueQueryRequest{SysVariableID: created.ID})
	require.NoError(t, err)
	assert.Equal(t, int64(1), page.Total)

	selected, err := svc.SelectedValues(created.ID)
	require.NoError(t, err)
	require.Len(t, selected, 1)

	require.NoError(t, svc.BatchDeleteValues([]int64{val.ID}))
	selected, err = svc.SelectedValues(created.ID)
	require.NoError(t, err)
	assert.Empty(t, selected)

	require.NoError(t, svc.Delete(created.ID))
	_, err = svc.Detail(created.ID)
	assert.Error(t, err)
}
