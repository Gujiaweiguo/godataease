package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTemplateExtendDataRepoTest(t *testing.T) (*TemplateExtendDataRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.VisualizationTemplateExtendDatum{}))
	return NewTemplateExtendDataRepository(db), db
}

func TestTemplateExtendDataRepository_BatchCreate_EmptySlice(t *testing.T) {
	repo, _ := setupTemplateExtendDataRepoTest(t)

	err := repo.BatchCreate(nil)
	require.NoError(t, err)

	err = repo.BatchCreate([]auto.VisualizationTemplateExtendDatum{})
	require.NoError(t, err)
}

func TestTemplateExtendDataRepository_BatchCreate_Success(t *testing.T) {
	repo, db := setupTemplateExtendDataRepoTest(t)

	records := []auto.VisualizationTemplateExtendDatum{
		{ID: 1, DvID: 100, ViewID: 200, ViewDetails: `{"key":"val1"}`, CopyFrom: "t1", CopyID: "c1"},
		{ID: 2, DvID: 100, ViewID: 201, ViewDetails: `{"key":"val2"}`, CopyFrom: "t2", CopyID: "c2"},
		{ID: 3, DvID: 101, ViewID: 300, ViewDetails: `{"key":"val3"}`},
	}
	require.NoError(t, repo.BatchCreate(records))

	var count int64
	require.NoError(t, db.Model(&auto.VisualizationTemplateExtendDatum{}).Count(&count).Error)
	assert.Equal(t, int64(3), count)
}

func TestTemplateExtendDataRepository_BatchCreate_VerifyData(t *testing.T) {
	repo, db := setupTemplateExtendDataRepoTest(t)

	records := []auto.VisualizationTemplateExtendDatum{
		{ID: 10, DvID: 500, ViewID: 600, ViewDetails: `{"chart":"bar"}`},
	}
	require.NoError(t, repo.BatchCreate(records))

	var found auto.VisualizationTemplateExtendDatum
	require.NoError(t, db.First(&found, 10).Error)
	assert.Equal(t, int64(500), found.DvID)
	assert.Equal(t, int64(600), found.ViewID)
	assert.Equal(t, `{"chart":"bar"}`, found.ViewDetails)
}

func TestTemplateExtendDataRepository_BatchCreate_DuplicateID(t *testing.T) {
	repo, _ := setupTemplateExtendDataRepoTest(t)

	records := []auto.VisualizationTemplateExtendDatum{
		{ID: 1, DvID: 100, ViewID: 200},
	}
	require.NoError(t, repo.BatchCreate(records))

	// Duplicate primary key should error
	dupRecords := []auto.VisualizationTemplateExtendDatum{
		{ID: 1, DvID: 999, ViewID: 888},
	}
	err := repo.BatchCreate(dupRecords)
	require.Error(t, err)
}
