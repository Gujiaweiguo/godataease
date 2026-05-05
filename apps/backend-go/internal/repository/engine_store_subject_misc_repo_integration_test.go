//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"
	"dataease/backend/internal/domain/engine"
	"dataease/backend/internal/domain/visualization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestEngineRepository_Get(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_engine")

	repo := NewEngineRepository(testDB)
	_, err := repo.Get()
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	config := `{"endpoint":"http://localhost:9000"}`
	creator := "tester"
	createTime := int64(1710000100)
	seed := &engine.Engine{
		ID:            1,
		Name:          "Spark Engine",
		Type:          "spark",
		Configuration: &config,
		CreateBy:      &creator,
		CreateTime:    &createTime,
	}
	require.NoError(t, testDB.Create(seed).Error)

	got, err := repo.Get()
	require.NoError(t, err)
	assert.Equal(t, seed.ID, got.ID)
	assert.Equal(t, seed.Name, got.Name)
	assert.Equal(t, seed.Type, got.Type)
	assert.Equal(t, seed.Configuration, got.Configuration)
	assert.Equal(t, seed.CreateBy, got.CreateBy)
	assert.Equal(t, seed.CreateTime, got.CreateTime)
}

func TestFavoriteRepository_CRUDAndQueryFavorites(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_store", "data_visualization_info")

	repo := NewFavoriteRepository(testDB)
	isFavorited, err := repo.IsFavorited(1001, 2001)
	require.NoError(t, err)
	assert.False(t, isFavorited)

	creator := "creator"
	editor := "editor"
	typeDashboard := "2"
	typeScreen := "5"
	updateTime1 := int64(1710000200)
	updateTime2 := int64(1710000300)
	visualizations := []*visualization.DataVisualizationInfo{
		{ID: 1001, Name: "Alpha Dashboard", Type: &typeDashboard, CreateBy: &creator, UpdateBy: &editor, UpdateTime: &updateTime1},
		{ID: 1002, Name: "Beta Screen", Type: &typeScreen, CreateBy: &creator, UpdateBy: &editor, UpdateTime: &updateTime2},
	}
	for _, item := range visualizations {
		require.NoError(t, testDB.Create(item).Error)
	}

	stores := []*auto.CoreStore{
		{ID: 1, ResourceID: 1001, UID: 2001, ResourceType: 10, Time: 1710000201},
		{ID: 2, ResourceID: 1002, UID: 2001, ResourceType: 11, Time: 1710000301},
		{ID: 3, ResourceID: 1001, UID: 2002, ResourceType: 10, Time: 1710000401},
	}
	for _, store := range stores {
		require.NoError(t, repo.CreateFavorite(store))
	}

	isFavorited, err = repo.IsFavorited(1001, 2001)
	require.NoError(t, err)
	assert.True(t, isFavorited)

	allRows, err := repo.QueryFavorites(2001, 0, "")
	require.NoError(t, err)
	require.Len(t, allRows, 2)
	assert.Equal(t, int64(2), allRows[0].StoreID)
	assert.Equal(t, int64(1002), allRows[0].ResourceID)
	assert.Equal(t, "Beta Screen", allRows[0].Name)
	assert.Equal(t, int32(5), allRows[0].Type)
	assert.Equal(t, "editor", allRows[0].Editor)
	assert.Equal(t, updateTime2, allRows[0].EditTime)
	assert.Equal(t, int64(1), allRows[1].StoreID)

	filteredRows, err := repo.QueryFavorites(2001, 11, "beta")
	require.NoError(t, err)
	require.Len(t, filteredRows, 1)
	assert.Equal(t, int64(1002), filteredRows[0].ResourceID)

	require.NoError(t, repo.DeleteFavorite(1001, 2001))
	isFavorited, err = repo.IsFavorited(1001, 2001)
	require.NoError(t, err)
	assert.False(t, isFavorited)

	remainingRows, err := repo.QueryFavorites(2001, 0, "")
	require.NoError(t, err)
	require.Len(t, remainingRows, 1)
	assert.Equal(t, int64(1002), remainingRows[0].ResourceID)
}

func TestSubjectRepository_CRUDAndQueries(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("visualization_subject")

	repo := NewSubjectRepository(testDB)
	subjects := []*auto.VisualizationSubject{
		{ID: "subject-1", Name: "Subject One", Type: "system", Details: `{"color":"blue"}`, DeleteFlag: false, CoverURL: "cover-1", CreateNum: 1, CreateTime: 10, CreateBy: "alice", UpdateTime: 10, UpdateBy: "alice"},
		{ID: "subject-2", Name: "Deleted Subject", Type: "self", Details: `{"color":"gray"}`, DeleteFlag: true, CoverURL: "cover-2", CreateNum: 2, CreateTime: 20, CreateBy: "bob", UpdateTime: 20, UpdateBy: "bob", DeleteTime: 21, DeleteBy: 2},
		{ID: "subject-3", Name: "Conflict Subject", Type: "self", Details: `{"color":"green"}`, DeleteFlag: false, CoverURL: "cover-3", CreateNum: 3, CreateTime: 30, CreateBy: "carol", UpdateTime: 30, UpdateBy: "carol"},
		{ID: "subject-4", Name: "Conflict Subject", Type: "self", Details: `{"color":"yellow"}`, DeleteFlag: false, CoverURL: "cover-4", CreateNum: 4, CreateTime: 40, CreateBy: "dave", UpdateTime: 40, UpdateBy: "dave"},
	}
	for _, subject := range subjects {
		require.NoError(t, repo.Create(subject))
	}

	got, err := repo.GetByID("subject-1")
	require.NoError(t, err)
	assert.Equal(t, "Subject One", got.Name)

	_, err = repo.GetByID("subject-2")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	list, err := repo.List()
	require.NoError(t, err)
	require.Len(t, list, 3)
	assert.Equal(t, "subject-1", list[0].ID)
	assert.Equal(t, "subject-3", list[1].ID)
	assert.Equal(t, "subject-4", list[2].ID)

	listAll, err := repo.ListAll()
	require.NoError(t, err)
	require.Len(t, listAll, 4)
	assert.Equal(t, "subject-2", listAll[1].ID)

	byName, err := repo.FindByName("Subject One")
	require.NoError(t, err)
	assert.Equal(t, "subject-1", byName.ID)

	byNameExcludeID, err := repo.FindByNameExcludeID("Conflict Subject", "subject-3")
	require.NoError(t, err)
	assert.Equal(t, "subject-4", byNameExcludeID.ID)

	got.Name = "Subject One Updated"
	got.Details = `{"color":"purple"}`
	got.UpdateTime = 99
	got.UpdateBy = "eve"
	require.NoError(t, repo.Update(got))

	updated, err := repo.GetByID("subject-1")
	require.NoError(t, err)
	assert.Equal(t, "Subject One Updated", updated.Name)
	assert.Equal(t, `{"color":"purple"}`, updated.Details)
	assert.Equal(t, int64(99), updated.UpdateTime)
	assert.Equal(t, "eve", updated.UpdateBy)

	require.NoError(t, repo.Delete("subject-4"))
	var count int64
	require.NoError(t, testDB.Model(&auto.VisualizationSubject{}).Where("id = ?", "subject-4").Count(&count).Error)
	assert.Zero(t, count)
}

func TestTemplateExtendDataRepository_BatchCreate(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("visualization_template_extend_data")

	repo := NewTemplateExtendDataRepository(testDB)
	require.NoError(t, repo.BatchCreate(nil))
	require.NoError(t, repo.BatchCreate([]auto.VisualizationTemplateExtendDatum{}))

	records := []auto.VisualizationTemplateExtendDatum{
		{ID: 1, DvID: 101, ViewID: 10001, ViewDetails: `{"name":"view-a"}`, CopyFrom: "template", CopyID: "copy-a"},
		{ID: 2, DvID: 101, ViewID: 10002, ViewDetails: `{"name":"view-b"}`, CopyFrom: "template", CopyID: "copy-b"},
	}
	require.NoError(t, repo.BatchCreate(records))

	var saved []auto.VisualizationTemplateExtendDatum
	require.NoError(t, testDB.Order("id ASC").Find(&saved).Error)
	require.Len(t, saved, 2)
	assert.Equal(t, int64(10001), saved[0].ViewID)
	assert.Equal(t, "copy-b", saved[1].CopyID)
}

func TestVisualizationBackgroundRepository_FindAll(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("visualization_background")

	seed := []auto.VisualizationBackground{
		{ID: "bg-3", Name: "Third", Classification: "dashboard", Content: "c3", Sort: 30, UploadTime: 3, BaseURL: "/base", URL: "/3"},
		{ID: "bg-1", Name: "First", Classification: "dashboard", Content: "c1", Sort: 10, UploadTime: 1, BaseURL: "/base", URL: "/1"},
		{ID: "bg-2", Name: "Second", Classification: "dashboard", Content: "c2", Sort: 20, UploadTime: 2, BaseURL: "/base", URL: "/2"},
	}
	for i := range seed {
		require.NoError(t, testDB.Create(&seed[i]).Error)
	}

	repo := NewVisualizationBackgroundRepository(testDB)
	backgrounds, err := repo.FindAll()
	require.NoError(t, err)
	require.Len(t, backgrounds, 3)
	assert.Equal(t, "bg-1", backgrounds[0].ID)
	assert.Equal(t, int32(10), backgrounds[0].Sort)
	assert.Equal(t, "bg-2", backgrounds[1].ID)
	assert.Equal(t, "bg-3", backgrounds[2].ID)
}
