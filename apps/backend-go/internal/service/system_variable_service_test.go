package service

import (
	"testing"

	"dataease/backend/internal/domain/system"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestSystemVariableService_NilRepoBranches(t *testing.T) {
	svc := NewSystemVariableService(nil)

	_, err := svc.Create(&system.SysVariable{Type: "text", Name: "a"})
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.Edit(&system.SysVariable{ID: 1, Type: "text", Name: "a"})
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.Detail(1)
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	err = svc.Delete(1)
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.Query(&system.SysVariableQueryRequest{})
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.CreateValue(&system.SysVariableValue{SysVariableID: 1, Value: "x"})
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.EditValue(&system.SysVariableValue{ID: 1, SysVariableID: 1, Value: "x"})
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	err = svc.DeleteValue(1)
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	err = svc.BatchDeleteValues([]int64{1})
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.SelectedValues(1)
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.SelectedValuePage(1, 10, &system.SysVariableValueQueryRequest{})
	assert.Equal(t, errSystemVariableRepoNotReady, err)
}

func TestSystemVariableService_ValidationHelpers(t *testing.T) {
	assert.Equal(t, gorm.ErrInvalidData, validateVariable(nil))
	assert.Equal(t, gorm.ErrInvalidData, validateVariable(&system.SysVariable{}))
	assert.Equal(t, gorm.ErrInvalidData, validateVariable(&system.SysVariable{Type: "text", Name: "bad", Min: 10, Max: 1}))
	assert.NoError(t, validateVariable(&system.SysVariable{Type: "text", Name: "ok", Min: 1, Max: 10}))

	assert.Equal(t, gorm.ErrInvalidData, validateVariableValue(nil))
	assert.Equal(t, gorm.ErrInvalidData, validateVariableValue(&system.SysVariableValue{}))
	assert.NoError(t, validateVariableValue(&system.SysVariableValue{SysVariableID: 1, Value: "v"}))
}

func TestSystemVariableService_InvalidEditRequests(t *testing.T) {
	svc := NewSystemVariableService(nil)

	_, err := svc.Edit(nil)
	assert.Equal(t, errSystemVariableRepoNotReady, err)

	_, err = svc.EditValue(nil)
	assert.Equal(t, errSystemVariableRepoNotReady, err)
}
