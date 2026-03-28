package service

import (
	"testing"

	"dataease/backend/internal/domain/system"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupSystemVariableServiceRepoTest(t *testing.T) (*SystemVariableService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&system.SysVariable{}, &system.SysVariableValue{}))

	repo := repository.NewSystemVariableRepository(db)
	return NewSystemVariableService(repo), db
}

func setupClosedSystemVariableServiceRepoTest(t *testing.T) *SystemVariableService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&system.SysVariable{}, &system.SysVariableValue{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	repo := repository.NewSystemVariableRepository(db)
	return NewSystemVariableService(repo)
}

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

func TestSystemVariableService_CreateAndEdit(t *testing.T) {
	t.Run("create resets id and persists", func(t *testing.T) {
		svc, db := setupSystemVariableServiceRepoTest(t)

		created, err := svc.Create(&system.SysVariable{ID: 99, Type: "text", Name: "region", Min: 1, Max: 9, Root: true})
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.NotEqual(t, int64(99), created.ID)
		assert.Equal(t, "region", created.Name)

		var stored system.SysVariable
		require.NoError(t, db.First(&stored, created.ID).Error)
		assert.Equal(t, "region", stored.Name)
	})

	t.Run("create propagates repo error", func(t *testing.T) {
		svc := setupClosedSystemVariableServiceRepoTest(t)

		created, err := svc.Create(&system.SysVariable{Type: "text", Name: "broken"})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	t.Run("edit validates and updates fields", func(t *testing.T) {
		svc, db := setupSystemVariableServiceRepoTest(t)
		require.NoError(t, db.Create(&system.SysVariable{ID: 1, Type: "text", Name: "old", Min: 1, Max: 5, StartTime: "08:00", EndTime: "18:00"}).Error)

		updated, err := svc.Edit(&system.SysVariable{ID: 1, Type: "number", Name: "new", Min: 2, Max: 8, StartTime: "09:00", EndTime: "19:00", Root: true, Disabled: true})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "number", updated.Type)
		assert.Equal(t, "new", updated.Name)
		assert.Equal(t, int64(2), updated.Min)
		assert.True(t, updated.Root)
		assert.True(t, updated.Disabled)

		_, err = svc.Edit(&system.SysVariable{})
		require.ErrorIs(t, err, gorm.ErrInvalidData)
	})

	t.Run("edit propagates repo update error", func(t *testing.T) {
		svc := setupClosedSystemVariableServiceRepoTest(t)

		_, err := svc.Edit(&system.SysVariable{ID: 1, Type: "text", Name: "broken"})
		require.Error(t, err)
	})

	t.Run("edit propagates get by id error", func(t *testing.T) {
		svc := setupClosedSystemVariableServiceRepoTest(t)

		updated, err := svc.Edit(&system.SysVariable{ID: 2, Type: "text", Name: "broken-get"})
		require.Error(t, err)
		assert.Nil(t, updated)
	})
}

func TestSystemVariableService_CreateValueAndEditValue(t *testing.T) {
	t.Run("create value checks parent and resets id", func(t *testing.T) {
		svc, db := setupSystemVariableServiceRepoTest(t)
		require.NoError(t, db.Create(&system.SysVariable{ID: 10, Type: "text", Name: "region"}).Error)

		created, err := svc.CreateValue(&system.SysVariableValue{ID: 88, SysVariableID: 10, Value: "APAC", ValueDesc: "Asia Pacific"})
		require.NoError(t, err)
		require.NotNil(t, created)
		assert.NotEqual(t, int64(88), created.ID)
		assert.Equal(t, int64(10), created.SysVariableID)

		_, err = svc.CreateValue(&system.SysVariableValue{SysVariableID: 999, Value: "missing"})
		require.Error(t, err)
	})

	t.Run("edit value validates parent and updates fields", func(t *testing.T) {
		svc, db := setupSystemVariableServiceRepoTest(t)
		require.NoError(t, db.Create(&system.SysVariable{ID: 11, Type: "text", Name: "country"}).Error)
		require.NoError(t, db.Create(&system.SysVariableValue{ID: 21, SysVariableID: 11, Value: "CN", ValueDesc: "China", Begin: "2024-01-01", End: "2024-12-31"}).Error)

		updated, err := svc.EditValue(&system.SysVariableValue{ID: 21, SysVariableID: 11, Value: "US", ValueDesc: "United States", Begin: "2025-01-01", End: "2025-12-31"})
		require.NoError(t, err)
		require.NotNil(t, updated)
		assert.Equal(t, "US", updated.Value)
		assert.Equal(t, "United States", updated.ValueDesc)

		_, err = svc.EditValue(&system.SysVariableValue{})
		require.ErrorIs(t, err, gorm.ErrInvalidData)
	})

	t.Run("edit value propagates repo errors", func(t *testing.T) {
		svc := setupClosedSystemVariableServiceRepoTest(t)

		_, err := svc.EditValue(&system.SysVariableValue{ID: 1, SysVariableID: 1, Value: "broken"})
		require.Error(t, err)
	})

	t.Run("edit value propagates get value by id error", func(t *testing.T) {
		svc, db := setupSystemVariableServiceRepoTest(t)
		require.NoError(t, db.Create(&system.SysVariable{ID: 12, Type: "text", Name: "country"}).Error)
		require.NoError(t, db.Exec("DROP TABLE sys_variable_value").Error)

		updated, err := svc.EditValue(&system.SysVariableValue{ID: 22, SysVariableID: 12, Value: "JP"})
		require.Error(t, err)
		assert.Nil(t, updated)
	})
}

func TestSystemVariableService_SelectedValuePage(t *testing.T) {
	svc, db := setupSystemVariableServiceRepoTest(t)
	require.NoError(t, db.Create(&system.SysVariable{ID: 31, Type: "text", Name: "city"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 41, SysVariableID: 31, Value: "Shanghai", ValueDesc: "Shanghai"}).Error)
	require.NoError(t, db.Create(&system.SysVariableValue{ID: 42, SysVariableID: 31, Value: "Shenzhen", ValueDesc: "Shenzhen"}).Error)

	page, err := svc.SelectedValuePage(2, 1, &system.SysVariableValueQueryRequest{SysVariableID: 31})
	require.NoError(t, err)
	require.NotNil(t, page)
	assert.Equal(t, int64(2), page.Total)
	assert.Equal(t, 2, page.Current)
	assert.Equal(t, 1, page.Size)
	require.Len(t, page.Records, 1)
	assert.Equal(t, "Shenzhen", page.Records[0].Value)

	t.Run("repo error propagates", func(t *testing.T) {
		svc := setupClosedSystemVariableServiceRepoTest(t)

		page, err := svc.SelectedValuePage(1, 10, &system.SysVariableValueQueryRequest{SysVariableID: 31})
		require.Error(t, err)
		assert.Nil(t, page)
	})
}

func TestSystemVariableService_PassThroughWrappers(t *testing.T) {
	t.Run("detail query selected values and deletions work with sqlite repo", func(t *testing.T) {
		svc, db := setupSystemVariableServiceRepoTest(t)
		require.NoError(t, db.Create(&system.SysVariable{ID: 51, Type: "text", Name: "region", Disabled: false}).Error)
		require.NoError(t, db.Create(&system.SysVariable{ID: 52, Type: "number", Name: "age", Disabled: true}).Error)
		require.NoError(t, db.Create(&system.SysVariableValue{ID: 61, SysVariableID: 51, Value: "APAC", ValueDesc: "Asia Pacific"}).Error)
		require.NoError(t, db.Create(&system.SysVariableValue{ID: 62, SysVariableID: 51, Value: "EMEA", ValueDesc: "Europe"}).Error)
		require.NoError(t, db.Create(&system.SysVariableValue{ID: 63, SysVariableID: 52, Value: "18", ValueDesc: "Adult"}).Error)

		detail, err := svc.Detail(51)
		require.NoError(t, err)
		assert.Equal(t, "region", detail.Name)

		disabled := false
		rows, err := svc.Query(&system.SysVariableQueryRequest{Type: "text", Name: "reg", Disabled: &disabled})
		require.NoError(t, err)
		require.Len(t, rows, 1)
		assert.Equal(t, int64(51), rows[0].ID)

		values, err := svc.SelectedValues(51)
		require.NoError(t, err)
		require.Len(t, values, 2)
		assert.Equal(t, "APAC", values[0].Value)

		err = svc.DeleteValue(61)
		require.NoError(t, err)
		values, err = svc.SelectedValues(51)
		require.NoError(t, err)
		require.Len(t, values, 1)
		assert.Equal(t, int64(62), values[0].ID)

		err = svc.BatchDeleteValues([]int64{62, 63})
		require.NoError(t, err)
		values, err = svc.SelectedValues(51)
		require.NoError(t, err)
		assert.Empty(t, values)

		err = svc.Delete(51)
		require.NoError(t, err)
		_, err = svc.Detail(51)
		require.Error(t, err)
	})

	t.Run("pass-through methods propagate repo errors", func(t *testing.T) {
		svc := setupClosedSystemVariableServiceRepoTest(t)

		_, err := svc.Detail(1)
		require.Error(t, err)

		_, err = svc.Query(&system.SysVariableQueryRequest{})
		require.Error(t, err)

		err = svc.Delete(1)
		require.Error(t, err)

		err = svc.DeleteValue(1)
		require.Error(t, err)

		err = svc.BatchDeleteValues([]int64{1})
		require.Error(t, err)

		_, err = svc.SelectedValues(1)
		require.Error(t, err)
	})

	t.Run("query empty result selected values empty and batch delete empty slice", func(t *testing.T) {
		svc, _ := setupSystemVariableServiceRepoTest(t)

		rows, err := svc.Query(&system.SysVariableQueryRequest{Type: "missing"})
		require.NoError(t, err)
		assert.Empty(t, rows)

		values, err := svc.SelectedValues(999)
		require.NoError(t, err)
		assert.Empty(t, values)

		err = svc.BatchDeleteValues(nil)
		require.NoError(t, err)
	})
}
