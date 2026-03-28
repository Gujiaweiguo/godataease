package service

import (
	"testing"

	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupStaticServiceRepoTest(t *testing.T) (*StaticService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&static.StaticResource{}, &static.Store{}, &static.Typeface{}))

	repo := repository.NewStaticRepository(db)
	storeRepo := repository.NewStoreRepository(db)
	typefaceRepo := repository.NewTypefaceRepository(db)

	return NewStaticService(repo, storeRepo, typefaceRepo), db
}

func setupClosedStaticServiceRepoTest(t *testing.T) *StaticService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&static.StaticResource{}, &static.Store{}, &static.Typeface{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewStaticService(
		repository.NewStaticRepository(db),
		repository.NewStoreRepository(db),
		repository.NewTypefaceRepository(db),
	)
}

func TestStaticService_ListResources_Unit(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		svc, _ := setupStaticServiceRepoTest(t)

		items, err := svc.ListResources()
		require.NoError(t, err)
		assert.NotNil(t, items)
		assert.Empty(t, items)
	})

	t.Run("returns persisted resources", func(t *testing.T) {
		svc, db := setupStaticServiceRepoTest(t)
		require.NoError(t, db.Create(&static.StaticResource{ID: "res1", Name: "Resource 1", Path: "/path/1", Type: "image"}).Error)
		require.NoError(t, db.Create(&static.StaticResource{ID: "res2", Name: "Resource 2", Path: "/path/2", Type: "video"}).Error)

		items, err := svc.ListResources()
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, "res1", items[0].ID)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedStaticServiceRepoTest(t)

		items, err := svc.ListResources()
		require.Error(t, err)
		assert.Nil(t, items)
	})
}

func TestStaticService_GetResource_Unit(t *testing.T) {
	t.Run("returns persisted resource", func(t *testing.T) {
		svc, db := setupStaticServiceRepoTest(t)
		require.NoError(t, db.Create(&static.StaticResource{ID: "test-res", Name: "Test Resource", Path: "/test/path", Type: "image"}).Error)

		item, err := svc.GetResource("test-res")
		require.NoError(t, err)
		require.NotNil(t, item)
		assert.Equal(t, "test-res", item.ID)
		assert.Equal(t, "Test Resource", item.Name)
	})

	t.Run("returns error when resource missing", func(t *testing.T) {
		svc, _ := setupStaticServiceRepoTest(t)

		item, err := svc.GetResource("missing")
		require.Error(t, err)
		assert.Nil(t, item)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedStaticServiceRepoTest(t)

		item, err := svc.GetResource("test-res")
		require.Error(t, err)
		assert.Nil(t, item)
	})
}

func TestStaticService_ListStores_Unit(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		svc, _ := setupStaticServiceRepoTest(t)

		items, err := svc.ListStores()
		require.NoError(t, err)
		assert.NotNil(t, items)
		assert.Empty(t, items)
	})

	t.Run("returns persisted stores", func(t *testing.T) {
		svc, db := setupStaticServiceRepoTest(t)
		require.NoError(t, db.Create(&static.Store{ID: "store1", Name: "Store 1", URL: "http://store1.example"}).Error)
		require.NoError(t, db.Create(&static.Store{ID: "store2", Name: "Store 2", URL: "http://store2.example"}).Error)

		items, err := svc.ListStores()
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, "store1", items[0].ID)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedStaticServiceRepoTest(t)

		items, err := svc.ListStores()
		require.Error(t, err)
		assert.Nil(t, items)
	})
}

func TestStaticService_ListTypefaces_Unit(t *testing.T) {
	t.Run("returns empty list", func(t *testing.T) {
		svc, _ := setupStaticServiceRepoTest(t)

		items, err := svc.ListTypefaces()
		require.NoError(t, err)
		assert.NotNil(t, items)
		assert.Empty(t, items)
	})

	t.Run("returns persisted typefaces", func(t *testing.T) {
		svc, db := setupStaticServiceRepoTest(t)
		require.NoError(t, db.Create(&static.Typeface{ID: "font1", Name: "Font 1", File: "/fonts/font1.ttf"}).Error)
		require.NoError(t, db.Create(&static.Typeface{ID: "font2", Name: "Font 2", File: "/fonts/font2.ttf"}).Error)

		items, err := svc.ListTypefaces()
		require.NoError(t, err)
		assert.Len(t, items, 2)
		assert.Equal(t, "font1", items[0].ID)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := setupClosedStaticServiceRepoTest(t)

		items, err := svc.ListTypefaces()
		require.Error(t, err)
		assert.Nil(t, items)
	})
}
