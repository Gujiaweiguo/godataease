//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/static"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStaticService_ListResources(t *testing.T) {
	cleanupTables(&static.StaticResource{})

	repo := repository.NewStaticRepository(testDB)
	storeRepo := repository.NewStoreRepository(testDB)
	typefaceRepo := repository.NewTypefaceRepository(testDB)
	svc := NewStaticService(repo, storeRepo, typefaceRepo)

	t.Run("list resources empty", func(t *testing.T) {
		result, err := svc.ListResources()
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("list resources with data", func(t *testing.T) {
		// Create test resources
		resources := []*static.StaticResource{
			{ID: "res1", Name: "Resource 1", Path: "/path/1", Type: "image"},
			{ID: "res2", Name: "Resource 2", Path: "/path/2", Type: "video"},
		}
		for _, r := range resources {
			require.NoError(t, testDB.Create(r).Error)
		}

		result, err := svc.ListResources()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
	})
}

func TestStaticService_GetResource(t *testing.T) {
	cleanupTables(&static.StaticResource{})

	repo := repository.NewStaticRepository(testDB)
	storeRepo := repository.NewStoreRepository(testDB)
	typefaceRepo := repository.NewTypefaceRepository(testDB)
	svc := NewStaticService(repo, storeRepo, typefaceRepo)

	t.Run("get resource by ID", func(t *testing.T) {
		// Create test resource
		resource := &static.StaticResource{
			ID:   "test-res",
			Name: "Test Resource",
			Path: "/test/path",
			Type: "image",
		}
		require.NoError(t, testDB.Create(resource).Error)

		result, err := svc.GetResource("test-res")
		require.NoError(t, err)
		assert.Equal(t, "test-res", result.ID)
		assert.Equal(t, "Test Resource", result.Name)
	})

	t.Run("get non-existent resource", func(t *testing.T) {
		result, err := svc.GetResource("non-existent")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestStaticService_ListStores(t *testing.T) {
	cleanupTables(&static.Store{})

	repo := repository.NewStaticRepository(testDB)
	storeRepo := repository.NewStoreRepository(testDB)
	typefaceRepo := repository.NewTypefaceRepository(testDB)
	svc := NewStaticService(repo, storeRepo, typefaceRepo)

	t.Run("list stores empty", func(t *testing.T) {
		result, err := svc.ListStores()
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("list stores with data", func(t *testing.T) {
		// Create test stores
		stores := []*static.Store{
			{ID: "store1", Name: "Store 1", URL: "http://store1.com"},
			{ID: "store2", Name: "Store 2", URL: "http://store2.com"},
		}
		for _, s := range stores {
			require.NoError(t, testDB.Create(s).Error)
		}

		result, err := svc.ListStores()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
	})
}

func TestStaticService_ListTypefaces(t *testing.T) {
	cleanupTables(&static.Typeface{})

	repo := repository.NewStaticRepository(testDB)
	storeRepo := repository.NewStoreRepository(testDB)
	typefaceRepo := repository.NewTypefaceRepository(testDB)
	svc := NewStaticService(repo, storeRepo, typefaceRepo)

	t.Run("list typefaces empty", func(t *testing.T) {
		result, err := svc.ListTypefaces()
		require.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("list typefaces with data", func(t *testing.T) {
		// Create test typefaces
		typefaces := []*static.Typeface{
			{ID: "font1", Name: "Font 1", File: "/fonts/font1.ttf"},
			{ID: "font2", Name: "Font 2", File: "/fonts/font2.ttf"},
		}
		for _, f := range typefaces {
			require.NoError(t, testDB.Create(f).Error)
		}

		result, err := svc.ListTypefaces()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
	})
}
