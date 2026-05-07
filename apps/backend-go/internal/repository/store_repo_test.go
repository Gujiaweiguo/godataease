package repository

import (
	"testing"

	"dataease/backend/internal/domain/auto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFavoriteRepoTest(t *testing.T) (*FavoriteRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&auto.CoreStore{}))
	require.NoError(t, db.Exec(`CREATE TABLE data_visualization_info (
		id TEXT PRIMARY KEY,
		name TEXT,
		pid TEXT,
		org_id TEXT,
		level INTEGER DEFAULT 0,
		node_type TEXT,
		type INTEGER DEFAULT 0,
		canvas_style_data TEXT,
		component_data TEXT,
		mobile_layout INTEGER DEFAULT 0,
		status INTEGER DEFAULT 1,
		self_watermark_status INTEGER DEFAULT 0,
		sort INTEGER DEFAULT 0,
		create_time INTEGER DEFAULT 0,
		create_by TEXT,
		update_time INTEGER DEFAULT 0,
		update_by TEXT,
		remark TEXT,
		source TEXT,
		delete_flag INTEGER DEFAULT 0,
		delete_time INTEGER DEFAULT 0,
		delete_by TEXT,
		version INTEGER DEFAULT 3,
		content_id TEXT,
		check_version TEXT DEFAULT '1'
	)`).Error)
	return NewFavoriteRepository(db), db
}

func TestFavoriteRepository_IsFavorited(t *testing.T) {
	repo, db := setupFavoriteRepoTest(t)

	// Not favorited yet
	fav, err := repo.IsFavorited(100, 1)
	require.NoError(t, err)
	assert.False(t, fav)

	// Create a favorite
	require.NoError(t, db.Create(&auto.CoreStore{
		ResourceID:   100,
		UID:          1,
		ResourceType: 1,
		Time:         1000,
	}).Error)

	fav, err = repo.IsFavorited(100, 1)
	require.NoError(t, err)
	assert.True(t, fav)

	// Different user
	fav, err = repo.IsFavorited(100, 2)
	require.NoError(t, err)
	assert.False(t, fav)
}

func TestFavoriteRepository_CreateAndDeleteFavorite(t *testing.T) {
	repo, db := setupFavoriteRepoTest(t)

	store := &auto.CoreStore{
		ResourceID:   200,
		UID:          5,
		ResourceType: 2,
		Time:         2000,
	}
	require.NoError(t, repo.CreateFavorite(store))
	require.Positive(t, store.ID)

	// Verify it exists
	var count int64
	require.NoError(t, db.Model(&auto.CoreStore{}).Where("resource_id = ? AND uid = ?", 200, 5).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	// Delete it
	require.NoError(t, repo.DeleteFavorite(200, 5))

	require.NoError(t, db.Model(&auto.CoreStore{}).Where("resource_id = ? AND uid = ?", 200, 5).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}

func TestFavoriteRepository_DeleteFavorite_NotExist(t *testing.T) {
	repo, _ := setupFavoriteRepoTest(t)

	// Deleting non-existent favorite should not error
	err := repo.DeleteFavorite(999, 999)
	require.NoError(t, err)
}

func TestFavoriteRepository_QueryFavorites(t *testing.T) {
	repo, db := setupFavoriteRepoTest(t)

	// Seed visualizations
	require.NoError(t, db.Exec("INSERT INTO data_visualization_info (id, name, type, create_by, update_by, update_time) VALUES ('301', 'Dashboard A', 1, 'user1', 'user2', 3000)").Error)
	require.NoError(t, db.Exec("INSERT INTO data_visualization_info (id, name, type, create_by, update_by, update_time) VALUES ('302', 'Dashboard B', 1, 'user1', 'user1', 2000)").Error)
	require.NoError(t, db.Exec("INSERT INTO data_visualization_info (id, name, type, create_by, update_by, update_time) VALUES ('303', 'Report C', 2, 'user2', 'user2', 4000)").Error)

	// Seed favorites for user 10
	require.NoError(t, db.Create(&auto.CoreStore{ResourceID: 301, UID: 10, ResourceType: 1, Time: 1000}).Error)
	require.NoError(t, db.Create(&auto.CoreStore{ResourceID: 302, UID: 10, ResourceType: 1, Time: 2000}).Error)
	// User 20's favorite — should not appear
	require.NoError(t, db.Create(&auto.CoreStore{ResourceID: 303, UID: 20, ResourceType: 1, Time: 3000}).Error)

	// All favorites for user 10
	rows, err := repo.QueryFavorites(10, 0, "")
	require.NoError(t, err)
	require.Len(t, rows, 2)
	// Ordered by update_time DESC
	assert.Equal(t, int64(301), rows[0].ResourceID)
	assert.Equal(t, int64(302), rows[1].ResourceID)
}

func TestFavoriteRepository_QueryFavorites_EmptyResult(t *testing.T) {
	repo, _ := setupFavoriteRepoTest(t)

	rows, err := repo.QueryFavorites(999, 0, "")
	require.NoError(t, err)
	assert.Empty(t, rows)
}
