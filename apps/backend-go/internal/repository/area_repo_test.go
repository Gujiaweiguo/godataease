package repository

import (
	"testing"

	"dataease/backend/internal/domain/areamap"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupAreaRepoTest(t *testing.T) (*AreaRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&areamap.Area{}, &areamap.CoreAreaCustom{}))
	return NewAreaRepository(db), db
}

func TestAreaRepository_GetAllAreas_Empty(t *testing.T) {
	repo, _ := setupAreaRepoTest(t)

	areas, err := repo.GetAllAreas()
	require.NoError(t, err)
	assert.Empty(t, areas)
}

func TestAreaRepository_GetAllAreas_WithData(t *testing.T) {
	repo, db := setupAreaRepoTest(t)

	require.NoError(t, db.Create(&areamap.Area{ID: "156", Level: "0", Name: "China", Pid: "0"}).Error)
	require.NoError(t, db.Create(&areamap.Area{ID: "110000", Level: "1", Name: "Beijing", Pid: "156"}).Error)
	require.NoError(t, db.Create(&areamap.Area{ID: "310000", Level: "1", Name: "Shanghai", Pid: "156"}).Error)

	areas, err := repo.GetAllAreas()
	require.NoError(t, err)
	require.Len(t, areas, 3)
	assert.Equal(t, "China", areas[0].Name)
	assert.Equal(t, "Beijing", areas[1].Name)
	assert.Equal(t, "Shanghai", areas[2].Name)
}

func TestAreaRepository_GetAllCustomAreas_Empty(t *testing.T) {
	repo, _ := setupAreaRepoTest(t)

	areas, err := repo.GetAllCustomAreas()
	require.NoError(t, err)
	assert.Empty(t, areas)
}

func TestAreaRepository_GetAllCustomAreas_WithData(t *testing.T) {
	repo, db := setupAreaRepoTest(t)

	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c1", Level: "1", Name: "Custom Area 1", Pid: "156"}).Error)
	require.NoError(t, db.Create(&areamap.CoreAreaCustom{ID: "c2", Level: "2", Name: "Custom Area 2", Pid: "c1"}).Error)

	areas, err := repo.GetAllCustomAreas()
	require.NoError(t, err)
	require.Len(t, areas, 2)
	assert.Equal(t, "Custom Area 1", areas[0].Name)
	assert.Equal(t, "Custom Area 2", areas[1].Name)
}
