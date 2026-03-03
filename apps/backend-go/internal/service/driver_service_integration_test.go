//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/driver"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriverService_List(t *testing.T) {
	cleanupTables(&driver.Driver{})

	repo := repository.NewDriverRepository(testDB)
	svc := NewDriverService(repo)

	t.Run("list empty drivers", func(t *testing.T) {
		result, err := svc.List()
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("list with data", func(t *testing.T) {
		// Create test driver
		testDriver := &driver.Driver{
			Name: "MySQL Driver",
			Type: "mysql",
		}
		require.NoError(t, testDB.Create(testDriver).Error)

		result, err := svc.List()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)

		// Find our driver
		found := false
		for _, d := range result {
			if d.Name == "MySQL Driver" {
				found = true
				assert.Equal(t, "mysql", d.Type)
				break
			}
		}
		assert.True(t, found)
	})
}

func TestDriverService_ListByType(t *testing.T) {
	cleanupTables(&driver.Driver{})

	repo := repository.NewDriverRepository(testDB)
	svc := NewDriverService(repo)

	// Create test drivers
	drivers := []driver.Driver{
		{Name: "MySQL Driver 1", Type: "mysql"},
		{Name: "MySQL Driver 2", Type: "mysql"},
		{Name: "PostgreSQL Driver", Type: "postgresql"},
	}
	for _, d := range drivers {
		require.NoError(t, testDB.Create(&d).Error)
	}

	t.Run("list mysql drivers", func(t *testing.T) {
		result, err := svc.ListByType("mysql")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)

		for _, d := range result {
			assert.Equal(t, "mysql", d.Type)
		}
	})

	t.Run("list postgresql drivers", func(t *testing.T) {
		result, err := svc.ListByType("postgresql")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
	})

	t.Run("list non-existent type", func(t *testing.T) {
		result, err := svc.ListByType("nonexistent")
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestDriverService_GetByID(t *testing.T) {
	cleanupTables(&driver.Driver{})

	repo := repository.NewDriverRepository(testDB)
	svc := NewDriverService(repo)

	t.Run("get by valid ID", func(t *testing.T) {
		testDriver := &driver.Driver{
			Name: "Test Driver",
			Type: "test",
		}
		require.NoError(t, testDB.Create(testDriver).Error)

		result, err := svc.GetByID(testDriver.ID)
		require.NoError(t, err)
		assert.Equal(t, testDriver.ID, result.ID)
		assert.Equal(t, "Test Driver", result.Name)
	})

	t.Run("fail with non-existent ID", func(t *testing.T) {
		result, err := svc.GetByID(999999)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestDriverService_ListDriverJars(t *testing.T) {
	cleanupTables(&driver.Driver{}, &driver.DriverJar{})

	repo := repository.NewDriverRepository(testDB)
	svc := NewDriverService(repo)

	t.Run("list jars for driver", func(t *testing.T) {
		// Create driver
		testDriver := &driver.Driver{
			Name: "MySQL Driver",
			Type: "mysql",
		}
		require.NoError(t, testDB.Create(testDriver).Error)

		// Create driver jars
		jars := []driver.DriverJar{
			{DriverID: testDriver.ID, FileName: "mysql-connector.jar", FilePath: "/path/to/mysql.jar"},
			{DriverID: testDriver.ID, FileName: "mysql-addon.jar", FilePath: "/path/to/addon.jar"},
		}
		for _, j := range jars {
			require.NoError(t, testDB.Create(&j).Error)
		}

		result, err := svc.ListDriverJars(testDriver.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
	})

	t.Run("list jars for non-existent driver", func(t *testing.T) {
		result, err := svc.ListDriverJars(999999)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
