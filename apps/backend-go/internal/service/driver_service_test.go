package service

import (
	"testing"

	"dataease/backend/internal/domain/driver"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDriverServiceRepoTest(t *testing.T) (*DriverService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&driver.Driver{}, &driver.DriverJar{}))

	repo := repository.NewDriverRepository(db)
	return NewDriverService(repo), db
}

func setupClosedDriverServiceRepoTest(t *testing.T) *DriverService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&driver.Driver{}, &driver.DriverJar{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewDriverService(repository.NewDriverRepository(db))
}

func TestDriverService_List_Unit(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		svc, _ := setupDriverServiceRepoTest(t)

		items, err := svc.List()
		require.NoError(t, err)
		assert.NotNil(t, items)
		assert.Empty(t, items)
	})

	t.Run("maps persisted driver fields", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		typeDesc := "MySQL"
		desc := "Driver for mysql"
		require.NoError(t, db.Create(&driver.Driver{Name: "MySQL Driver", Type: "mysql", TypeDesc: &typeDesc, Desc: &desc}).Error)

		items, err := svc.List()
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, "MySQL Driver", items[0].Name)
		assert.Equal(t, "mysql", items[0].Type)
		require.NotNil(t, items[0].TypeDesc)
		require.NotNil(t, items[0].Desc)
		assert.Equal(t, "MySQL", *items[0].TypeDesc)
		assert.Equal(t, "Driver for mysql", *items[0].Desc)
	})

	t.Run("maps nil optional fields", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		require.NoError(t, db.Create(&driver.Driver{Name: "Nil Optional Driver", Type: "sqlite"}).Error)

		items, err := svc.List()
		require.NoError(t, err)
		require.Len(t, items, 1)
		assert.Equal(t, "Nil Optional Driver", items[0].Name)
		assert.Nil(t, items[0].TypeDesc)
		assert.Nil(t, items[0].Desc)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedDriverServiceRepoTest(t)

		items, err := svc.List()
		require.Error(t, err)
		assert.Nil(t, items)
	})
}

func TestDriverService_ListByType_Unit(t *testing.T) {
	t.Run("returns empty list for missing type", func(t *testing.T) {
		svc, _ := setupDriverServiceRepoTest(t)

		items, err := svc.ListByType("missing")
		require.NoError(t, err)
		assert.NotNil(t, items)
		assert.Empty(t, items)
	})

	t.Run("maps drivers by type", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		require.NoError(t, db.Create(&driver.Driver{Name: "MySQL Driver 1", Type: "mysql"}).Error)
		require.NoError(t, db.Create(&driver.Driver{Name: "PostgreSQL Driver", Type: "postgresql"}).Error)
		require.NoError(t, db.Create(&driver.Driver{Name: "MySQL Driver 2", Type: "mysql"}).Error)

		items, err := svc.ListByType("mysql")
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, "mysql", items[0].Type)
		assert.Equal(t, "mysql", items[1].Type)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedDriverServiceRepoTest(t)

		items, err := svc.ListByType("mysql")
		require.Error(t, err)
		assert.Nil(t, items)
	})

	t.Run("preserves repository order", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		require.NoError(t, db.Create(&driver.Driver{ID: 21, Name: "MySQL Driver 21", Type: "mysql"}).Error)
		require.NoError(t, db.Create(&driver.Driver{ID: 22, Name: "MySQL Driver 22", Type: "mysql"}).Error)

		items, err := svc.ListByType("mysql")
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, int64(21), items[0].ID)
		assert.Equal(t, int64(22), items[1].ID)
	})
}

func TestDriverService_GetByID_Unit(t *testing.T) {
	t.Run("returns mapped driver", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		require.NoError(t, db.Create(&driver.Driver{Name: "Test Driver", Type: "test"}).Error)

		var row driver.Driver
		require.NoError(t, db.First(&row).Error)

		item, err := svc.GetByID(row.ID)
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, row.ID, item.ID)
		assert.Equal(t, "Test Driver", item.Name)
		assert.Equal(t, "test", item.Type)
	})

	t.Run("returns error when missing", func(t *testing.T) {
		svc, _ := setupDriverServiceRepoTest(t)

		item, err := svc.GetByID(999999)
		require.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedDriverServiceRepoTest(t)

		item, err := svc.GetByID(1)
		require.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("maps nil optional fields", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		require.NoError(t, db.Create(&driver.Driver{Name: "Nil Getter", Type: "sqlite"}).Error)

		var row driver.Driver
		require.NoError(t, db.First(&row).Error)

		item, err := svc.GetByID(row.ID)
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Nil(t, item.TypeDesc)
		assert.Nil(t, item.Desc)
	})
}

func TestDriverService_ListDriverJars_Unit(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		svc, _ := setupDriverServiceRepoTest(t)

		items, err := svc.ListDriverJars(999999)
		require.NoError(t, err)
		assert.NotNil(t, items)
		assert.Empty(t, items)
	})

	t.Run("maps persisted jar fields", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		version := "8.0.33"
		require.NoError(t, db.Create(&driver.Driver{Name: "MySQL Driver", Type: "mysql"}).Error)

		var drv driver.Driver
		require.NoError(t, db.First(&drv).Error)
		require.NoError(t, db.Create(&driver.DriverJar{DriverID: drv.ID, FileName: "mysql-connector.jar", FilePath: "/path/mysql-connector.jar", Version: &version}).Error)

		items, err := svc.ListDriverJars(drv.ID)
		require.NoError(t, err)
		assert.Len(t, items, 1)
		assert.Equal(t, drv.ID, items[0].DriverID)
		assert.Equal(t, "mysql-connector.jar", items[0].FileName)
		assert.Equal(t, "/path/mysql-connector.jar", items[0].FilePath)
		require.NotNil(t, items[0].Version)
		assert.Equal(t, "8.0.33", *items[0].Version)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedDriverServiceRepoTest(t)

		items, err := svc.ListDriverJars(1)
		require.Error(t, err)
		assert.Nil(t, items)
	})

	t.Run("maps nil version and preserves order", func(t *testing.T) {
		svc, db := setupDriverServiceRepoTest(t)
		require.NoError(t, db.Create(&driver.Driver{ID: 31, Name: "Ordered Driver", Type: "mysql"}).Error)
		require.NoError(t, db.Create(&driver.DriverJar{ID: 41, DriverID: 31, FileName: "jar-a", FilePath: "/a.jar", Version: nil}).Error)
		v := "2.0"
		require.NoError(t, db.Create(&driver.DriverJar{ID: 42, DriverID: 31, FileName: "jar-b", FilePath: "/b.jar", Version: &v}).Error)

		items, err := svc.ListDriverJars(31)
		require.NoError(t, err)
		require.Len(t, items, 2)
		assert.Equal(t, int64(41), items[0].ID)
		assert.Nil(t, items[0].Version)
		assert.Equal(t, int64(42), items[1].ID)
		require.NotNil(t, items[1].Version)
		assert.Equal(t, "2.0", *items[1].Version)
	})
}
