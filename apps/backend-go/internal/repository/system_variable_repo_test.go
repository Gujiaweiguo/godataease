package repository

import (
	"testing"

	"dataease/backend/internal/domain/system"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSystemVariableRepositoryTest(t *testing.T) (*SystemVariableRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&system.SysVariable{}, &system.SysVariableValue{}))
	return NewSystemVariableRepository(db), db
}

func TestSystemVariableRepository_Query(t *testing.T) {
	repo, db := setupSystemVariableRepositoryTest(t)
	require.NoError(t, db.Create(&system.SysVariable{ID: 1, Type: "text", Name: "Region", Disabled: false}).Error)
	require.NoError(t, db.Create(&system.SysVariable{ID: 2, Type: "number", Name: "Age", Disabled: true}).Error)
	require.NoError(t, db.Create(&system.SysVariable{ID: 3, Type: "text", Name: "Country", Disabled: false}).Error)

	rows, err := repo.Query(nil)
	require.NoError(t, err)
	require.Len(t, rows, 3)
	assert.Equal(t, int64(1), rows[0].ID)
	assert.Equal(t, int64(3), rows[2].ID)

	disabled := false
	rows, err = repo.Query(&system.SysVariableQueryRequest{ID: 1, Type: " text ", Name: "gio", Disabled: &disabled})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(1), rows[0].ID)

	disabled = true
	rows, err = repo.Query(&system.SysVariableQueryRequest{Disabled: &disabled})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, int64(2), rows[0].ID)
}

func TestSystemVariableRepository_CreateUpdateAndGetters(t *testing.T) {
	repo, db := setupSystemVariableRepositoryTest(t)
	variable := &system.SysVariable{Type: "text", Name: "Region", Root: true}
	require.NoError(t, repo.Create(variable))
	require.Positive(t, variable.ID)

	loaded, err := repo.GetByID(variable.ID)
	require.NoError(t, err)
	assert.Equal(t, "Region", loaded.Name)

	variable.Name = "Region Updated"
	variable.Disabled = true
	require.NoError(t, repo.Update(variable))
	loaded, err = repo.GetByID(variable.ID)
	require.NoError(t, err)
	assert.Equal(t, "Region Updated", loaded.Name)
	assert.True(t, loaded.Disabled)

	value := &system.SysVariableValue{SysVariableID: variable.ID, Value: "APAC", ValueDesc: "Asia Pacific"}
	require.NoError(t, repo.CreateValue(value))
	require.Positive(t, value.ID)

	value.ValueDesc = "Asia"
	require.NoError(t, repo.UpdateValue(value))
	loadedValue, err := repo.GetValueByID(value.ID)
	require.NoError(t, err)
	assert.Equal(t, "Asia", loadedValue.ValueDesc)

	var raw system.SysVariableValue
	require.NoError(t, db.Where("id = ?", value.ID).First(&raw).Error)
	assert.Equal(t, value.Value, raw.Value)
}

func TestSystemVariableRepository_PageValuesAndValueHelpers(t *testing.T) {
	repo, db := setupSystemVariableRepositoryTest(t)
	require.NoError(t, db.Create(&system.SysVariable{ID: 10, Type: "text", Name: "Region"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 101, SysVariableID: 10, Value: "APAC", ValueDesc: "Asia Pacific"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 102, SysVariableID: 10, Value: "EMEA", ValueDesc: "Europe"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 103, SysVariableID: 10, Value: "AMER", ValueDesc: "America"}).Error)

	values, err := repo.ListValuesByVariableID(10)
	require.NoError(t, err)
	require.Len(t, values, 3)
	assert.Equal(t, int64(101), values[0].ID)
	assert.Equal(t, int64(103), values[2].ID)

	pageRows, total, err := repo.PageValues(0, 0, &system.SysVariableValueQueryRequest{SysVariableID: 10, Value: "A", ValueDesc: "a"})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, pageRows, 2)

	pageRows, total, err = repo.PageValues(2, 1, &system.SysVariableValueQueryRequest{SysVariableID: 10})
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, pageRows, 1)
	assert.Equal(t, int64(102), pageRows[0].ID)

	_, err = repo.GetValueByID(999)
	require.Error(t, err)
}

func TestSystemVariableRepository_DeleteAndBatchDeleteValues(t *testing.T) {
	repo, db := setupSystemVariableRepositoryTest(t)
	require.NoError(t, db.Create(&system.SysVariable{ID: 20, Type: "text", Name: "Region"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 201, SysVariableID: 20, Value: "APAC"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 202, SysVariableID: 20, Value: "EMEA"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 203, SysVariableID: 20, Value: "AMER"}).Error)

	require.NoError(t, repo.BatchDeleteValues(nil))

	require.NoError(t, repo.BatchDeleteValues([]int64{201, 203}))
	values, err := repo.ListValuesByVariableID(20)
	require.NoError(t, err)
	require.Len(t, values, 1)
	assert.Equal(t, int64(202), values[0].ID)

	require.NoError(t, repo.DeleteValue(202))
	values, err = repo.ListValuesByVariableID(20)
	require.NoError(t, err)
	assert.Empty(t, values)

	require.NoError(t, db.Create(&system.SysVariableValue{ID: 204, SysVariableID: 20, Value: "LATAM"}).Error)
	require.NoError(t, repo.Delete(20))

	_, err = repo.GetByID(20)
	require.Error(t, err)
	values, err = repo.ListValuesByVariableID(20)
	require.NoError(t, err)
	assert.Empty(t, values)
}
