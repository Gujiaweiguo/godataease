package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/msgcenter"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testCoreMsgSetting struct {
	ID     int64      `gorm:"column:id;primaryKey;autoIncrement"`
	MsgID  string     `gorm:"column:msg_id;size:100;uniqueIndex:idx_msg_user"`
	UserID int64      `gorm:"column:user_id;index;uniqueIndex:idx_msg_user"`
	Status string     `gorm:"column:status;size:20"`
	ReadAt *time.Time `gorm:"column:read_at"`
}

func (testCoreMsgSetting) TableName() string {
	return "core_msg_setting"
}

func setupMsgCenterServiceRepoTest(t *testing.T) (*MsgCenterService, *repository.MsgCenterRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&testCoreMsgSetting{}))

	repo := repository.NewMsgCenterRepository(db)
	return NewMsgCenterService(repo), repo, db
}

func setupClosedMsgCenterServiceRepoTest(t *testing.T) *MsgCenterService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&testCoreMsgSetting{}))
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewMsgCenterService(repository.NewMsgCenterRepository(db))
}

func TestMsgCenterService_Read(t *testing.T) {
	t.Run("returns failure when is read check errors", func(t *testing.T) {
		svc := setupClosedMsgCenterServiceRepoTest(t)

		resp := svc.Read(&msgcenter.ReadRequest{ID: "msg-1"}, 1)
		require.NotNil(t, resp)
		assert.False(t, resp.Success)
		assert.False(t, resp.AlreadyRead)
	})

	t.Run("returns already read when message was already marked", func(t *testing.T) {
		svc, repo, _ := setupMsgCenterServiceRepoTest(t)
		require.NoError(t, repo.MarkAsRead("msg-1", 1))

		resp := svc.Read(&msgcenter.ReadRequest{ID: "msg-1"}, 1)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.True(t, resp.AlreadyRead)
	})

	t.Run("marks unread message as read", func(t *testing.T) {
		svc, repo, _ := setupMsgCenterServiceRepoTest(t)

		resp := svc.Read(&msgcenter.ReadRequest{ID: "msg-2"}, 1)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.False(t, resp.AlreadyRead)

		read, err := repo.IsRead("msg-2", 1)
		require.NoError(t, err)
		assert.True(t, read)
	})

	t.Run("returns failure when mark as read fails", func(t *testing.T) {
		svc, repo, db := setupMsgCenterServiceRepoTest(t)
		require.NoError(t, db.Exec(`
			CREATE TRIGGER msgcenter_insert_fail
			BEFORE INSERT ON core_msg_setting
			BEGIN
				SELECT RAISE(FAIL, 'forced insert failure');
			END;
		`).Error)

		resp := svc.Read(&msgcenter.ReadRequest{ID: "msg-4"}, 1)
		require.NotNil(t, resp)
		assert.False(t, resp.Success)
		assert.False(t, resp.AlreadyRead)

		read, err := repo.IsRead("msg-4", 1)
		assert.NoError(t, err)
		assert.False(t, read)
	})

	t.Run("isolates read state per user", func(t *testing.T) {
		svc, _, _ := setupMsgCenterServiceRepoTest(t)

		firstUser := svc.Read(&msgcenter.ReadRequest{ID: "msg-3"}, 1)
		secondUser := svc.Read(&msgcenter.ReadRequest{ID: "msg-3"}, 2)
		firstUserAgain := svc.Read(&msgcenter.ReadRequest{ID: "msg-3"}, 1)

		assert.False(t, firstUser.AlreadyRead)
		assert.False(t, secondUser.AlreadyRead)
		assert.True(t, firstUserAgain.AlreadyRead)
	})
}

func TestMsgCenterService_ReadBatch(t *testing.T) {
	t.Run("returns zero updates when repository swallows errors", func(t *testing.T) {
		svc := setupClosedMsgCenterServiceRepoTest(t)

		resp := svc.ReadBatch(&msgcenter.ReadBatchRequest{IDs: []string{"msg-1"}}, 1)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Zero(t, resp.Updated)
	})

	t.Run("empty input succeeds with zero updates", func(t *testing.T) {
		svc, _, _ := setupMsgCenterServiceRepoTest(t)

		resp := svc.ReadBatch(&msgcenter.ReadBatchRequest{IDs: nil}, 1)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Zero(t, resp.Updated)
	})

	t.Run("marks only unread messages", func(t *testing.T) {
		svc, repo, _ := setupMsgCenterServiceRepoTest(t)
		require.NoError(t, repo.MarkAsRead("msg-a", 1))

		resp := svc.ReadBatch(&msgcenter.ReadBatchRequest{IDs: []string{"msg-a", "msg-b", "msg-c"}}, 1)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, 2, resp.Updated)
	})

	t.Run("returns zero updates when all messages were already read", func(t *testing.T) {
		svc, repo, _ := setupMsgCenterServiceRepoTest(t)
		require.NoError(t, repo.MarkAsRead("msg-a", 1))
		require.NoError(t, repo.MarkAsRead("msg-b", 1))

		resp := svc.ReadBatch(&msgcenter.ReadBatchRequest{IDs: []string{"msg-a", "msg-b"}}, 1)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Zero(t, resp.Updated)
	})

	t.Run("isolates read batch state per user", func(t *testing.T) {
		svc, repo, _ := setupMsgCenterServiceRepoTest(t)

		resp := svc.ReadBatch(&msgcenter.ReadBatchRequest{IDs: []string{"msg-user-scope"}}, 1)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
		assert.Equal(t, 1, resp.Updated)

		firstUserRead, err := repo.IsRead("msg-user-scope", 1)
		require.NoError(t, err)
		secondUserRead, err := repo.IsRead("msg-user-scope", 2)
		require.NoError(t, err)
		assert.True(t, firstUserRead)
		assert.False(t, secondUserRead)
	})
}

func TestMsgCenterService_List(t *testing.T) {
	t.Run("defaults invalid pagination", func(t *testing.T) {
		svc, _, _ := setupMsgCenterServiceRepoTest(t)

		resp := svc.List(&msgcenter.ListRequest{Current: 0, Size: 0})
		require.NotNil(t, resp)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
		assert.Empty(t, resp.List)
		assert.Zero(t, resp.Total)
	})

	t.Run("preserves valid pagination", func(t *testing.T) {
		svc, _, _ := setupMsgCenterServiceRepoTest(t)

		resp := svc.List(&msgcenter.ListRequest{Current: 2, Size: 20})
		require.NotNil(t, resp)
		assert.Equal(t, 2, resp.Current)
		assert.Equal(t, 20, resp.Size)
	})
}

func TestMsgCenterService_Count(t *testing.T) {
	svc, _, _ := setupMsgCenterServiceRepoTest(t)

	count := svc.Count(&msgcenter.CountRequest{})
	assert.Zero(t, count)
}
