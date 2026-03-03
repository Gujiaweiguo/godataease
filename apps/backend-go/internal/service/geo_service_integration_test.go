//go:build integration

package service

import (
	"testing"

	"dataease/backend/internal/domain/geo"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeoService_ListAreas(t *testing.T) {
	cleanupTables(&geo.GeometryArea{})

	repo := repository.NewGeoRepository(testDB)
	svc := NewGeoService(repo)

	t.Run("list empty areas", func(t *testing.T) {
		result, err := svc.ListAreas()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("list with data", func(t *testing.T) {
		// Create test areas
		areas := []*geo.GeometryArea{
			{ID: "110000", Name: "Beijing", Code: "110000", GeoJSON: `{"type":"Feature"}`},
			{ID: "310000", Name: "Shanghai", Code: "310000", GeoJSON: `{"type":"Feature"}`},
		}
		for _, a := range areas {
			require.NoError(t, testDB.Create(a).Error)
		}

		result, err := svc.ListAreas()
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
	})
}

func TestGeoService_GetArea(t *testing.T) {
	cleanupTables(&geo.GeometryArea{})

	repo := repository.NewGeoRepository(testDB)
	svc := NewGeoService(repo)

	t.Run("get by valid ID", func(t *testing.T) {
		testArea := &geo.GeometryArea{
			ID:      "110000",
			Name:    "Beijing",
			Code:    "110000",
			GeoJSON: `{"type":"Feature"}`,
		}
		require.NoError(t, testDB.Create(testArea).Error)

		result, err := svc.GetArea("110000")
		require.NoError(t, err)
		assert.Equal(t, "110000", result.ID)
		assert.Equal(t, "Beijing", result.Name)
	})

	t.Run("fail with non-existent ID", func(t *testing.T) {
		result, err := svc.GetArea("nonexistent")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}
