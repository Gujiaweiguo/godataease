package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupMsgCenterRepositoryTest(t *testing.T) (*MsgCenterRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&coreMsgSetting{}))
	return NewMsgCenterRepository(db), db
}

func setupClosedMsgCenterRepositoryTest(t *testing.T) *MsgCenterRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&coreMsgSetting{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())
	return NewMsgCenterRepository(db)
}

func TestMsgCenterRepository_MarkAsRead(t *testing.T) {
	t.Run("creates new read record", func(t *testing.T) {
		repo, db := setupMsgCenterRepositoryTest(t)

		require.NoError(t, repo.MarkAsRead("msg-1", 7))

		var row coreMsgSetting
		require.NoError(t, db.Where("msg_id = ? AND user_id = ?", "msg-1", 7).First(&row).Error)
		assert.Equal(t, msgStatusRead, row.Status)
		require.NotNil(t, row.ReadAt)
	})

	t.Run("updates existing unread record", func(t *testing.T) {
		repo, db := setupMsgCenterRepositoryTest(t)
		require.NoError(t, db.Create(&coreMsgSetting{MsgID: "msg-2", UserID: 8, Status: "unread"}).Error)

		require.NoError(t, repo.MarkAsRead("msg-2", 8))

		var row coreMsgSetting
		require.NoError(t, db.Where("msg_id = ? AND user_id = ?", "msg-2", 8).First(&row).Error)
		assert.Equal(t, msgStatusRead, row.Status)
		require.NotNil(t, row.ReadAt)
	})

	t.Run("propagates unexpected db error", func(t *testing.T) {
		repo := setupClosedMsgCenterRepositoryTest(t)

		err := repo.MarkAsRead("msg-3", 9)
		require.Error(t, err)
	})
}

func TestMsgCenterRepository_MarkBatchAsRead(t *testing.T) {
	t.Run("empty input short circuits", func(t *testing.T) {
		repo, _ := setupMsgCenterRepositoryTest(t)

		updated, err := repo.MarkBatchAsRead(nil, 10)
		require.NoError(t, err)
		assert.Zero(t, updated)
	})

	t.Run("creates updates and skips blank or already read ids", func(t *testing.T) {
		repo, db := setupMsgCenterRepositoryTest(t)
		require.NoError(t, db.Create(&coreMsgSetting{MsgID: "msg-b", UserID: 10, Status: "unread"}).Error)
		require.NoError(t, db.Create(&coreMsgSetting{MsgID: "msg-c", UserID: 10, Status: msgStatusRead}).Error)

		updated, err := repo.MarkBatchAsRead([]string{"", "msg-a", "msg-b", "msg-c"}, 10)
		require.NoError(t, err)
		assert.Equal(t, 2, updated)

		statusMap, err := repo.GetReadStatusMap([]string{"msg-a", "msg-b", "msg-c"}, 10)
		require.NoError(t, err)
		assert.True(t, statusMap["msg-a"])
		assert.True(t, statusMap["msg-b"])
		assert.True(t, statusMap["msg-c"])
	})
}

func TestMsgCenterRepository_IsRead(t *testing.T) {
	t.Run("returns false on record not found", func(t *testing.T) {
		repo, _ := setupMsgCenterRepositoryTest(t)

		read, err := repo.IsRead("msg-x", 1)
		require.NoError(t, err)
		assert.False(t, read)
	})

	t.Run("returns true for read record", func(t *testing.T) {
		repo, db := setupMsgCenterRepositoryTest(t)
		require.NoError(t, db.Create(&coreMsgSetting{MsgID: "msg-y", UserID: 1, Status: msgStatusRead}).Error)

		read, err := repo.IsRead("msg-y", 1)
		require.NoError(t, err)
		assert.True(t, read)
	})

	t.Run("propagates unexpected db error", func(t *testing.T) {
		repo := setupClosedMsgCenterRepositoryTest(t)

		read, err := repo.IsRead("msg-z", 1)
		require.Error(t, err)
		assert.False(t, read)
	})
}

func TestMsgCenterRepository_GetReadStatusMap(t *testing.T) {
	t.Run("empty input returns empty map", func(t *testing.T) {
		repo, _ := setupMsgCenterRepositoryTest(t)

		statusMap, err := repo.GetReadStatusMap(nil, 1)
		require.NoError(t, err)
		assert.Empty(t, statusMap)
	})

	t.Run("returns only read records", func(t *testing.T) {
		repo, db := setupMsgCenterRepositoryTest(t)
		require.NoError(t, db.Create(&coreMsgSetting{MsgID: "msg-1", UserID: 3, Status: msgStatusRead}).Error)
		require.NoError(t, db.Create(&coreMsgSetting{MsgID: "msg-2", UserID: 3, Status: "unread"}).Error)
		require.NoError(t, db.Create(&coreMsgSetting{MsgID: "msg-3", UserID: 3, Status: msgStatusRead}).Error)

		statusMap, err := repo.GetReadStatusMap([]string{"msg-1", "msg-2", "msg-3"}, 3)
		require.NoError(t, err)
		assert.Equal(t, map[string]bool{"msg-1": true, "msg-3": true}, statusMap)
	})

	t.Run("propagates db error", func(t *testing.T) {
		repo := setupClosedMsgCenterRepositoryTest(t)

		statusMap, err := repo.GetReadStatusMap([]string{"msg-err"}, 5)
		require.Error(t, err)
		assert.Nil(t, statusMap)
	})
}
