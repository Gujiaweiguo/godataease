//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/driver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestDriverRepository_ListAndListByType(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("de_driver", "de_driver_jar")

	seedDrivers(t)

	repo := NewDriverRepository(testDB)
	list, err := repo.List()
	require.NoError(t, err)
	require.Equal(t, 3, len(list))
	assert.Equal(t, int64(1), list[0].ID)
	assert.Equal(t, int64(2), list[1].ID)
	assert.Equal(t, int64(3), list[2].ID)

	filtered, err := repo.ListByType("mysql")
	require.NoError(t, err)
	require.Equal(t, 2, len(filtered))
	assert.Equal(t, "mysql-1", filtered[0].Name)
	assert.Equal(t, "mysql-2", filtered[1].Name)
}

func TestDriverRepository_GetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("de_driver", "de_driver_jar")

	seedDrivers(t)

	repo := NewDriverRepository(testDB)
	got, err := repo.GetByID(2)
	require.NoError(t, err)
	assert.Equal(t, int64(2), got.ID)
	assert.Equal(t, "mysql-2", got.Name)

	_, err = repo.GetByID(999)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestDriverRepository_ListDriverJars(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("de_driver", "de_driver_jar")

	require.NoError(t, testDB.Create(&driver.Driver{ID: 1, Name: "mysql-1", Type: "mysql"}).Error)
	jars := []driver.DriverJar{
		{ID: 20, DriverID: 1, FileName: "b.jar", FilePath: "/drivers/b.jar"},
		{ID: 10, DriverID: 1, FileName: "a.jar", FilePath: "/drivers/a.jar"},
		{ID: 30, DriverID: 2, FileName: "other.jar", FilePath: "/drivers/other.jar"},
	}
	for i := range jars {
		require.NoError(t, testDB.Create(&jars[i]).Error)
	}

	repo := NewDriverRepository(testDB)
	list, err := repo.ListDriverJars(1)
	require.NoError(t, err)
	require.Equal(t, 2, len(list))
	assert.Equal(t, int64(10), list[0].ID)
	assert.Equal(t, int64(20), list[1].ID)
	assert.Equal(t, int64(1), list[0].DriverID)
}

func seedDrivers(t *testing.T) {
	t.Helper()
	drivers := []driver.Driver{
		{ID: 2, Name: "mysql-2", Type: "mysql"},
		{ID: 1, Name: "mysql-1", Type: "mysql"},
		{ID: 3, Name: "postgres-1", Type: "postgres"},
	}
	for i := range drivers {
		require.NoError(t, testDB.Create(&drivers[i]).Error)
	}
}
