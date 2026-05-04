package service

import (
	"os"
	"path/filepath"
	"testing"

	"dataease/backend/internal/domain/areamap"
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

func TestGeoService_SaveMapGeoAndDeleteGeo(t *testing.T) {
	t.Run("saves custom geo metadata and file", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GEO_DIR", dir)

		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&geo.GeometryArea{}, &areamap.Area{}, &areamap.CoreAreaCustom{}))

		svc := NewGeoService(repository.NewGeoRepository(db))
		payload := []byte(`{"type":"FeatureCollection"}`)

		require.NoError(t, svc.SaveMapGeo("geo_123456", "Test Area", "geo_123000", payload, "123456.json"))

		stored, err := repository.NewGeoRepository(db).GetCustomAreaByID("geo_123456")
		require.NoError(t, err)
		assert.Equal(t, "Test Area", stored.Name)
		assert.Equal(t, "geo_123000", stored.Pid)

		content, err := os.ReadFile(filepath.Join(dir, "123", "123456.json"))
		require.NoError(t, err)
		assert.Equal(t, string(payload), string(content))
	})

	t.Run("deletes custom geo recursively", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GEO_DIR", dir)

		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&geo.GeometryArea{}, &areamap.Area{}, &areamap.CoreAreaCustom{}))
		repo := repository.NewGeoRepository(db)
		svc := NewGeoService(repo)

		require.NoError(t, repo.SaveCustomArea(&areamap.CoreAreaCustom{ID: "geo_123000", Name: "Parent"}))
		require.NoError(t, repo.SaveCustomArea(&areamap.CoreAreaCustom{ID: "geo_123100", Name: "Child", Pid: "geo_123000"}))
		require.NoError(t, repo.SaveCustomArea(&areamap.CoreAreaCustom{ID: "geo_123101", Name: "Grandchild", Pid: "geo_123100"}))
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "123"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "123", "123000.json"), []byte("root"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "123", "123100.json"), []byte("child"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "123", "123101.json"), []byte("grandchild"), 0o644))

		require.NoError(t, svc.DeleteGeo("geo_123000"))

		_, err = repo.GetCustomAreaByID("geo_123000")
		require.Error(t, err)
		_, err = repo.GetCustomAreaByID("geo_123100")
		require.Error(t, err)
		_, err = os.Stat(filepath.Join(dir, "123", "123000.json"))
		assert.True(t, os.IsNotExist(err))
		_, err = os.Stat(filepath.Join(dir, "123", "123101.json"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("collectChildIDs stops on missing descendants", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&areamap.CoreAreaCustom{}))
		repo := repository.NewGeoRepository(db)
		svc := NewGeoService(repo)
		require.NoError(t, repo.SaveCustomArea(&areamap.CoreAreaCustom{ID: "geo_1", Pid: "geo_root"}))
		require.NoError(t, repo.SaveCustomArea(&areamap.CoreAreaCustom{ID: "geo_2", Pid: "geo_1"}))

		ids := []string{"geo_root"}
		svc.collectChildIDs("geo_root", &ids)
		assert.Equal(t, []string{"geo_root", "geo_1", "geo_2"}, ids)
	})

	t.Run("rejects built in and duplicate geo codes", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("GEO_DIR", dir)

		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&areamap.Area{}, &areamap.CoreAreaCustom{}))
		repo := repository.NewGeoRepository(db)
		svc := NewGeoService(repo)

		require.NoError(t, db.Create(&areamap.Area{ID: "123456", Name: "Builtin"}).Error)
		err = svc.SaveMapGeo("123456", "Builtin", "", []byte("{}"), "123456.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists in built-in areas")

		require.NoError(t, repo.SaveCustomArea(&areamap.CoreAreaCustom{ID: "geo_654321", Name: "Existing Custom"}))
		err = svc.SaveMapGeo("654321", "Dup", "", []byte("{}"), "654321.json")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "already exists for [Existing Custom]")
	})

	t.Run("delete guards built in and missing geometries", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		require.NoError(t, db.AutoMigrate(&areamap.CoreAreaCustom{}))
		svc := NewGeoService(repository.NewGeoRepository(db))

		err = svc.DeleteGeo("123456")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot delete")

		err = svc.DeleteGeo("geo_missing")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})
}
