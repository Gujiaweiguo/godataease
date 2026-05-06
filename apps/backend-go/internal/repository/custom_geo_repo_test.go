package repository

import (
	"testing"

	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/auto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupCustomGeoRepoTest(t *testing.T) (*CustomGeoRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.CoreCustomGeoArea{}, &auto.CoreCustomGeoSubArea{}, &areamap.Area{}))
	return NewCustomGeoRepository(db), db
}

func TestCustomGeoRepository_ListGeoAreas_Empty(t *testing.T) {
	repo, _ := setupCustomGeoRepoTest(t)

	areas, err := repo.ListGeoAreas()
	require.NoError(t, err)
	assert.Empty(t, areas)
}

func TestCustomGeoRepository_SaveGeoArea_Create(t *testing.T) {
	repo, _ := setupCustomGeoRepoTest(t)

	area := &auto.CoreCustomGeoArea{ID: "ga1", Name: "Region A"}
	require.NoError(t, repo.SaveGeoArea(area))

	areas, err := repo.ListGeoAreas()
	require.NoError(t, err)
	require.Len(t, areas, 1)
	assert.Equal(t, "Region A", areas[0].Name)
}

func TestCustomGeoRepository_SaveGeoArea_Update(t *testing.T) {
	repo, _ := setupCustomGeoRepoTest(t)

	area := &auto.CoreCustomGeoArea{ID: "ga1", Name: "Original"}
	require.NoError(t, repo.SaveGeoArea(area))

	area.Name = "Updated"
	require.NoError(t, repo.SaveGeoArea(area))

	areas, err := repo.ListGeoAreas()
	require.NoError(t, err)
	require.Len(t, areas, 1)
	assert.Equal(t, "Updated", areas[0].Name)
}

func TestCustomGeoRepository_GetGeoArea(t *testing.T) {
	repo, db := setupCustomGeoRepoTest(t)

	require.NoError(t, db.Create(&auto.CoreCustomGeoArea{ID: "ga1", Name: "Area1"}).Error)
	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 1, Name: "Sub1", GeoAreaID: "ga1", Scope: "scope1"}).Error)
	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 2, Name: "Sub2", GeoAreaID: "ga1", Scope: "scope2"}).Error)
	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 3, Name: "Sub3", GeoAreaID: "ga2", Scope: "scope3"}).Error)

	subAreas, err := repo.GetGeoArea("ga1")
	require.NoError(t, err)
	require.Len(t, subAreas, 2)
	assert.Equal(t, "Sub1", subAreas[0].Name)
	assert.Equal(t, "Sub2", subAreas[1].Name)
}

func TestCustomGeoRepository_DeleteGeoArea(t *testing.T) {
	repo, db := setupCustomGeoRepoTest(t)

	require.NoError(t, db.Create(&auto.CoreCustomGeoArea{ID: "ga1", Name: "Area1"}).Error)
	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 1, Name: "Sub1", GeoAreaID: "ga1"}).Error)
	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 2, Name: "Sub2", GeoAreaID: "ga1"}).Error)

	require.NoError(t, repo.DeleteGeoArea("ga1"))

	// Area and its sub-areas should be deleted
	areas, err := repo.ListGeoAreas()
	require.NoError(t, err)
	assert.Empty(t, areas)

	subAreas, err := repo.GetGeoArea("ga1")
	require.NoError(t, err)
	assert.Empty(t, subAreas)
}

func TestCustomGeoRepository_CheckGeoAreaName(t *testing.T) {
	repo, db := setupCustomGeoRepoTest(t)

	require.NoError(t, db.Create(&auto.CoreCustomGeoArea{ID: "ga1", Name: "Beijing Area"}).Error)

	exists, err := repo.CheckGeoAreaName("Beijing Area", "")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.CheckGeoAreaName("Beijing Area", "ga1")
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = repo.CheckGeoAreaName("Shanghai Area", "")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCustomGeoRepository_SaveGeoSubArea_Create(t *testing.T) {
	repo, _ := setupCustomGeoRepoTest(t)

	subArea := &auto.CoreCustomGeoSubArea{ID: 10, Name: "Sub Area 1", GeoAreaID: "ga1", Scope: "scope_data"}
	require.NoError(t, repo.SaveGeoSubArea(subArea))

	subAreas, err := repo.GetGeoArea("ga1")
	require.NoError(t, err)
	require.Len(t, subAreas, 1)
	assert.Equal(t, "Sub Area 1", subAreas[0].Name)
}

func TestCustomGeoRepository_SaveGeoSubArea_Update(t *testing.T) {
	repo, db := setupCustomGeoRepoTest(t)

	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 10, Name: "Original", GeoAreaID: "ga1", Scope: "s1"}).Error)

	subArea := &auto.CoreCustomGeoSubArea{ID: 10, Name: "Updated", GeoAreaID: "ga1", Scope: "s2"}
	require.NoError(t, repo.SaveGeoSubArea(subArea))

	found, err := repo.GetGeoArea("ga1")
	require.NoError(t, err)
	require.Len(t, found, 1)
	assert.Equal(t, "Updated", found[0].Name)
	assert.Equal(t, "s2", found[0].Scope)
}

func TestCustomGeoRepository_DeleteGeoSubArea(t *testing.T) {
	repo, db := setupCustomGeoRepoTest(t)

	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 10, Name: "Sub1", GeoAreaID: "ga1"}).Error)
	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 11, Name: "Sub2", GeoAreaID: "ga1"}).Error)

	require.NoError(t, repo.DeleteGeoSubArea(10))

	subAreas, err := repo.GetGeoArea("ga1")
	require.NoError(t, err)
	require.Len(t, subAreas, 1)
	assert.Equal(t, int64(11), subAreas[0].ID)
}

func TestCustomGeoRepository_CheckGeoSubAreaName(t *testing.T) {
	repo, db := setupCustomGeoRepoTest(t)

	require.NoError(t, db.Create(&auto.CoreCustomGeoSubArea{ID: 1, Name: "Dongcheng", GeoAreaID: "ga1"}).Error)

	exists, err := repo.CheckGeoSubAreaName("Dongcheng", "ga1", 0)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.CheckGeoSubAreaName("Dongcheng", "ga1", 1)
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = repo.CheckGeoSubAreaName("Dongcheng", "ga2", 0)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestCustomGeoRepository_ListAreaOptions(t *testing.T) {
	repo, db := setupCustomGeoRepoTest(t)

	require.NoError(t, db.Create(&areamap.Area{ID: "110000", Level: "1", Name: "Beijing", Pid: "156"}).Error)
	require.NoError(t, db.Create(&areamap.Area{ID: "310000", Level: "1", Name: "Shanghai", Pid: "156"}).Error)
	require.NoError(t, db.Create(&areamap.Area{ID: "999999", Level: "1", Name: "Other", Pid: "000"}).Error)

	areas, err := repo.ListAreaOptions()
	require.NoError(t, err)
	require.Len(t, areas, 2)
	assert.Equal(t, "Beijing", areas[0].Name)
	assert.Equal(t, "Shanghai", areas[1].Name)
}
