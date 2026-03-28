package service

import (
	"testing"

	"dataease/backend/internal/domain/geo"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGeoServiceRepoTest(t *testing.T) (*GeoService, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&geo.GeometryArea{}))

	repo := repository.NewGeoRepository(db)
	return NewGeoService(repo), db
}

func setupClosedGeoServiceRepoTest(t *testing.T) *GeoService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&geo.GeometryArea{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewGeoService(repository.NewGeoRepository(db))
}

func TestGeoService_ListAreas_Unit(t *testing.T) {
	t.Run("returns empty slice when table empty", func(t *testing.T) {
		svc, _ := setupGeoServiceRepoTest(t)

		areas, err := svc.ListAreas()
		require.NoError(t, err)
		assert.NotNil(t, areas)
		assert.Empty(t, areas)
	})

	t.Run("returns persisted areas", func(t *testing.T) {
		svc, db := setupGeoServiceRepoTest(t)
		require.NoError(t, db.Create(&geo.GeometryArea{ID: "110000", Name: "Beijing", Code: "110000", GeoJSON: `{"type":"Feature"}`}).Error)
		require.NoError(t, db.Create(&geo.GeometryArea{ID: "310000", Name: "Shanghai", Code: "310000", GeoJSON: `{"type":"Feature"}`}).Error)

		areas, err := svc.ListAreas()
		require.NoError(t, err)
		assert.Len(t, areas, 2)
		assert.Equal(t, "110000", areas[0].ID)
		assert.Equal(t, "Beijing", areas[0].Name)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedGeoServiceRepoTest(t)

		areas, err := svc.ListAreas()
		require.Error(t, err)
		assert.Nil(t, areas)
	})
}

func TestGeoService_GetArea_Unit(t *testing.T) {
	t.Run("returns persisted area", func(t *testing.T) {
		svc, db := setupGeoServiceRepoTest(t)
		require.NoError(t, db.Create(&geo.GeometryArea{ID: "440100", Name: "Guangzhou", Code: "440100", GeoJSON: `{"type":"Feature"}`}).Error)

		area, err := svc.GetArea("440100")
		require.NoError(t, err)
		require.NotNil(t, area)
		assert.Equal(t, "440100", area.ID)
		assert.Equal(t, "Guangzhou", area.Name)
		assert.Equal(t, "440100", area.Code)
	})

	t.Run("returns error when area missing", func(t *testing.T) {
		svc, _ := setupGeoServiceRepoTest(t)

		area, err := svc.GetArea("missing")
		require.Error(t, err)
		assert.Nil(t, area)
	})

	t.Run("propagates repository error", func(t *testing.T) {
		svc := setupClosedGeoServiceRepoTest(t)

		area, err := svc.GetArea("440100")
		require.Error(t, err)
		assert.Nil(t, area)
	})
}
