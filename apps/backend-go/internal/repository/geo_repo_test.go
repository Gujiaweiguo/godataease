package repository

import (
	"testing"

	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/geo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupGeoRepoTest(t *testing.T) (*GeoRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&geo.GeometryArea{}, &areamap.CoreAreaCustom{}))
	// GeoRepository.CheckAreaExists uses raw table name "area"
	require.NoError(t, db.Exec("CREATE TABLE area (id TEXT PRIMARY KEY, level TEXT, name TEXT, pid TEXT)").Error)
	return NewGeoRepository(db), db
}

func TestGeoRepository_ListAreas_Empty(t *testing.T) {
	repo, _ := setupGeoRepoTest(t)

	areas, err := repo.ListAreas()
	require.NoError(t, err)
	assert.Empty(t, areas)
}

func TestGeoRepository_ListAreas_WithData(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	require.NoError(t, db.Create(&geo.GeometryArea{ID: "110000", Name: "Beijing", Code: "110000", GeoJSON: `{"type":"Polygon"}`}).Error)
	require.NoError(t, db.Create(&geo.GeometryArea{ID: "310000", Name: "Shanghai", Code: "310000", GeoJSON: `{"type":"Polygon"}`}).Error)

	areas, err := repo.ListAreas()
	require.NoError(t, err)
	require.Len(t, areas, 2)
}

func TestGeoRepository_GetAreaByID_Found(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	require.NoError(t, db.Create(&geo.GeometryArea{ID: "110000", Name: "Beijing", Code: "110000", GeoJSON: `{"type":"Polygon"}`}).Error)

	area, err := repo.GetAreaByID("110000")
	require.NoError(t, err)
	assert.Equal(t, "Beijing", area.Name)
	assert.Equal(t, "110000", area.Code)
}

func TestGeoRepository_GetAreaByID_NotFound(t *testing.T) {
	repo, _ := setupGeoRepoTest(t)

	_, err := repo.GetAreaByID("nonexistent")
	require.Error(t, err)
}

func TestGeoRepository_SaveCustomArea(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	area := &areamap.CoreAreaCustom{ID: "custom1", Level: "2", Name: "Custom Region", Pid: "156"}
	require.NoError(t, repo.SaveCustomArea(area))

	var found areamap.CoreAreaCustom
	require.NoError(t, db.First(&found, "id = ?", "custom1").Error)
	assert.Equal(t, "Custom Region", found.Name)
}

func TestGeoRepository_GetCustomAreaByID(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c1", Level: "2", Name: "Test Area", Pid: "156"}).Error)

	area, err := repo.GetCustomAreaByID("c1")
	require.NoError(t, err)
	assert.Equal(t, "Test Area", area.Name)
}

func TestGeoRepository_GetCustomAreaByID_NotFound(t *testing.T) {
	repo, _ := setupGeoRepoTest(t)

	_, err := repo.GetCustomAreaByID("nonexistent")
	require.Error(t, err)
}

func TestGeoRepository_DeleteCustomArea(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c1", Level: "2", Name: "ToDelete", Pid: "156"}).Error)
	require.NoError(t, repo.DeleteCustomArea("c1"))

	var count int64
	require.NoError(t, db.Model(&areamap.CoreAreaCustom{}).Where("id = ?", "c1").Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestGeoRepository_GetCustomAreaChildren(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c1", Level: "1", Name: "Parent", Pid: "0"}).Error)
	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c2", Level: "2", Name: "Child1", Pid: "c1"}).Error)
	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c3", Level: "2", Name: "Child2", Pid: "c1"}).Error)
	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c4", Level: "2", Name: "OtherChild", Pid: "c99"}).Error)

	children, err := repo.GetCustomAreaChildren("c1")
	require.NoError(t, err)
	require.Len(t, children, 2)
}

func TestGeoRepository_DeleteCustomAreasBatch(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c1", Level: "1", Name: "A", Pid: "0"}).Error)
	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c2", Level: "1", Name: "B", Pid: "0"}).Error)
	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c3", Level: "1", Name: "C", Pid: "0"}).Error)

	require.NoError(t, repo.DeleteCustomAreasBatch([]string{"c1", "c3"}))

	var count int64
	require.NoError(t, db.Model(&areamap.CoreAreaCustom{}).Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

func TestGeoRepository_CheckAreaExists(t *testing.T) {
	repo, db := setupGeoRepoTest(t)

	require.NoError(t, db.Exec("INSERT INTO area (id, level, name, pid) VALUES ('110000', '1', 'Beijing', '156')").Error)

	exists, err := repo.CheckAreaExists("110000")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.CheckAreaExists("999999")
	require.NoError(t, err)
	assert.False(t, exists)
}
