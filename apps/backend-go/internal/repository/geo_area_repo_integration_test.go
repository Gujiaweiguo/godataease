//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/areamap"
	"dataease/backend/internal/domain/geo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestGeoRepository_AreasAndCustomAreas(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("area_geo", "core_area_custom", "area")

	geoAreas := []geo.GeometryArea{
		{ID: "geo-1", Name: "China", Code: "CN", GeoJSON: `{"type":"Feature"}`},
		{ID: "geo-2", Name: "Japan", Code: "JP", GeoJSON: `{"type":"Feature"}`},
	}
	for i := range geoAreas {
		require.NoError(t, testDB.Create(&geoAreas[i]).Error)
	}
	require.NoError(t, testDB.Create(&areamap.Area{ID: "builtin-1", Level: "province", Name: "BuiltIn", Pid: "0"}).Error)

	repo := NewGeoRepository(testDB)
	areas, err := repo.ListAreas()
	require.NoError(t, err)
	require.Len(t, areas, 2)
	assert.ElementsMatch(t, []string{"geo-1", "geo-2"}, []string{areas[0].ID, areas[1].ID})

	area, err := repo.GetAreaByID("geo-2")
	require.NoError(t, err)
	assert.Equal(t, "Japan", area.Name)

	parent := &areamap.CoreAreaCustom{ID: "custom-parent", Level: "province", Name: "Parent", Pid: "0"}
	child1 := &areamap.CoreAreaCustom{ID: "custom-child-1", Level: "city", Name: "Child One", Pid: "custom-parent"}
	child2 := &areamap.CoreAreaCustom{ID: "custom-child-2", Level: "city", Name: "Child Two", Pid: "custom-parent"}
	orphan := &areamap.CoreAreaCustom{ID: "custom-orphan", Level: "district", Name: "Orphan", Pid: "custom-child-1"}
	for _, item := range []*areamap.CoreAreaCustom{parent, child1, child2, orphan} {
		require.NoError(t, repo.SaveCustomArea(item))
	}

	customArea, err := repo.GetCustomAreaByID("custom-child-1")
	require.NoError(t, err)
	assert.Equal(t, "Child One", customArea.Name)

	children, err := repo.GetCustomAreaChildren("custom-parent")
	require.NoError(t, err)
	require.Len(t, children, 2)
	assert.ElementsMatch(t, []string{"custom-child-1", "custom-child-2"}, []string{children[0].ID, children[1].ID})

	require.NoError(t, repo.DeleteCustomArea("custom-orphan"))
	_, err = repo.GetCustomAreaByID("custom-orphan")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	require.NoError(t, repo.DeleteCustomAreasBatch([]string{"custom-child-1", "custom-child-2"}))
	var count int64
	require.NoError(t, testDB.Model(&areamap.CoreAreaCustom{}).Where("id IN ?", []string{"custom-child-1", "custom-child-2"}).Count(&count).Error)
	assert.Zero(t, count)

	exists, err := repo.CheckAreaExists("builtin-1")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.CheckAreaExists("builtin-missing")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestAreaRepository_GetAllAreasAndCustomAreas(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("area", "core_area_custom")

	areas := []areamap.Area{
		{ID: "area-1", Level: "country", Name: "Country", Pid: "0"},
		{ID: "area-2", Level: "province", Name: "Province", Pid: "area-1"},
	}
	customAreas := []areamap.CoreAreaCustom{
		{ID: "custom-1", Level: "city", Name: "Custom City", Pid: "area-2"},
		{ID: "custom-2", Level: "district", Name: "Custom District", Pid: "custom-1"},
	}
	for i := range areas {
		require.NoError(t, testDB.Create(&areas[i]).Error)
	}
	for i := range customAreas {
		require.NoError(t, testDB.Create(&customAreas[i]).Error)
	}

	repo := NewAreaRepository(testDB)
	allAreas, err := repo.GetAllAreas()
	require.NoError(t, err)
	require.Len(t, allAreas, 2)
	assert.ElementsMatch(t, []string{"area-1", "area-2"}, []string{allAreas[0].ID, allAreas[1].ID})

	allCustomAreas, err := repo.GetAllCustomAreas()
	require.NoError(t, err)
	require.Len(t, allCustomAreas, 2)
	assert.ElementsMatch(t, []string{"custom-1", "custom-2"}, []string{allCustomAreas[0].ID, allCustomAreas[1].ID})
}
