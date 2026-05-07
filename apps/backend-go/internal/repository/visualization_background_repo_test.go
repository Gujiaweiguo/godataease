package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupVisualizationBackgroundRepoTest(t *testing.T) (*VisualizationBackgroundRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.VisualizationBackground{}))
	return NewVisualizationBackgroundRepository(db), db
}

func TestVisualizationBackgroundRepository_FindAll_Empty(t *testing.T) {
	repo, _ := setupVisualizationBackgroundRepoTest(t)

	backgrounds, err := repo.FindAll()
	require.NoError(t, err)
	assert.Empty(t, backgrounds)
}

func TestVisualizationBackgroundRepository_FindAll_OrderedBySort(t *testing.T) {
	repo, db := setupVisualizationBackgroundRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationBackground{ID: "bg3", Name: "Background 3", Sort: 30}).Error)
	require.NoError(t, db.Create(&auto.VisualizationBackground{ID: "bg1", Name: "Background 1", Sort: 10}).Error)
	require.NoError(t, db.Create(&auto.VisualizationBackground{ID: "bg2", Name: "Background 2", Sort: 20}).Error)

	backgrounds, err := repo.FindAll()
	require.NoError(t, err)
	require.Len(t, backgrounds, 3)
	assert.Equal(t, "bg1", backgrounds[0].ID)
	assert.Equal(t, "bg2", backgrounds[1].ID)
	assert.Equal(t, "bg3", backgrounds[2].ID)
}

func TestVisualizationBackgroundRepository_FindAll_WithClassification(t *testing.T) {
	repo, db := setupVisualizationBackgroundRepoTest(t)

	require.NoError(t, db.Create(&auto.VisualizationBackground{ID: "bg1", Name: "Gradient", Classification: "gradient", Sort: 1, Content: "linear-gradient(...)"}).Error)
	require.NoError(t, db.Create(&auto.VisualizationBackground{ID: "bg2", Name: "Solid", Classification: "solid", Sort: 2, URL: "http://example.com/bg.png"}).Error)

	backgrounds, err := repo.FindAll()
	require.NoError(t, err)
	require.Len(t, backgrounds, 2)
	assert.Equal(t, "gradient", backgrounds[0].Classification)
	assert.Equal(t, "solid", backgrounds[1].Classification)
}
