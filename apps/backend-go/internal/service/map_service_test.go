package service

import (
	"testing"

	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMapServiceRepoTest(t *testing.T) (*MapService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&areamap.Area{}, &areamap.CoreAreaCustom{}))

	repo := repository.NewAreaRepository(db)
	return NewMapService(repo), db
}

func setupClosedMapServiceRepoTest(t *testing.T) *MapService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&areamap.Area{}, &areamap.CoreAreaCustom{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewMapService(repository.NewAreaRepository(db))
}

func TestMapService_GetWorldTree(t *testing.T) {
	t.Run("empty data returns world root", func(t *testing.T) {
		svc, _ := setupMapServiceRepoTest(t)

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "000", result.ID)
		assert.Equal(t, "world", result.Level)
		assert.Equal(t, "世界村", result.Name)
		assert.False(t, result.Custom)
		assert.Nil(t, result.Children)
	})

	t.Run("base area attaches to world", func(t *testing.T) {
		svc, db := setupMapServiceRepoTest(t)
		require.NoError(t, db.Create(&areamap.Area{ID: "156", Level: "country", Name: "China", Pid: "000"}).Error)

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.Len(t, result.Children, 1)
		assert.Equal(t, "156", result.Children[0].ID)
		assert.Equal(t, "China", result.Children[0].Name)
		assert.False(t, result.Children[0].Custom)
	})

	t.Run("custom area attaches under parent and marks custom", func(t *testing.T) {
		svc, db := setupMapServiceRepoTest(t)
		require.NoError(t, db.Create(&areamap.Area{ID: "156", Level: "country", Name: "China", Pid: "000"}).Error)
		require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "custom1", Level: "custom", Name: "Custom Area", Pid: "156"}).Error)

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.Len(t, result.Children, 1)
		china := result.Children[0]
		require.Len(t, china.Children, 1)
		assert.Equal(t, "custom1", china.Children[0].ID)
		assert.Equal(t, "Custom Area", china.Children[0].Name)
		assert.True(t, china.Children[0].Custom)
	})

	t.Run("missing parent skips attachment", func(t *testing.T) {
		svc, db := setupMapServiceRepoTest(t)
		require.NoError(t, db.Create(&areamap.Area{ID: "orphan", Level: "country", Name: "Orphan", Pid: "missing-parent"}).Error)
		require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "custom-orphan", Level: "custom", Name: "Custom Orphan", Pid: "missing-parent"}).Error)

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		assert.Nil(t, result.Children)
	})

	t.Run("propagates get all areas error", func(t *testing.T) {
		svc := setupClosedMapServiceRepoTest(t)

		result, err := svc.GetWorldTree()
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("propagates get all custom areas error", func(t *testing.T) {
		svc, db := setupMapServiceRepoTest(t)
		require.NoError(t, db.Create(&areamap.Area{ID: "156", Level: "country", Name: "China", Pid: "000"}).Error)
		require.NoError(t, db.Migrator().DropTable(&areamap.CoreAreaCustom{}))

		result, err := svc.GetWorldTree()
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("base and custom siblings both attach under same parent", func(t *testing.T) {
		svc, db := setupMapServiceRepoTest(t)
		require.NoError(t, db.Create(&areamap.Area{ID: "156", Level: "country", Name: "China", Pid: "000"}).Error)
		require.NoError(t, db.Create(&areamap.Area{ID: "310000", Level: "city", Name: "Shanghai", Pid: "156"}).Error)
		require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "custom-sibling", Level: "custom", Name: "Custom Sibling", Pid: "156"}).Error)

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.Len(t, result.Children, 1)
		china := result.Children[0]
		require.Len(t, china.Children, 2)
		assert.Equal(t, "310000", china.Children[0].ID)
		assert.False(t, china.Children[0].Custom)
		assert.Equal(t, "custom-sibling", china.Children[1].ID)
		assert.True(t, china.Children[1].Custom)
	})

	t.Run("custom area with world parent attaches to root", func(t *testing.T) {
		svc, db := setupMapServiceRepoTest(t)
		require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "world-custom", Level: "custom", Name: "World Custom", Pid: "000"}).Error)

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.Len(t, result.Children, 1)
		assert.Equal(t, "world-custom", result.Children[0].ID)
		assert.True(t, result.Children[0].Custom)
	})

	t.Run("base area can attach under custom parent", func(t *testing.T) {
		svc, db := setupMapServiceRepoTest(t)
		require.NoError(t, db.Create(&areamap.Area{ID: "156", Level: "country", Name: "China", Pid: "000"}).Error)
		require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "custom-parent", Level: "custom", Name: "Custom Parent", Pid: "156"}).Error)
		require.NoError(t, db.Create(&areamap.Area{ID: "custom-child-base", Level: "city", Name: "Base Under Custom", Pid: "custom-parent"}).Error)

		result, err := svc.GetWorldTree()
		require.NoError(t, err)
		require.Len(t, result.Children, 1)
		china := result.Children[0]
		require.Len(t, china.Children, 1)
		customParent := china.Children[0]
		assert.Equal(t, "custom-parent", customParent.ID)
		require.Len(t, customParent.Children, 1)
		assert.Equal(t, "custom-child-base", customParent.Children[0].ID)
		assert.False(t, customParent.Children[0].Custom)
	})
}
