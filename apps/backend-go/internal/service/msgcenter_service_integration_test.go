//go:build integration

package service

import (
	"errors"
	"testing"
	"time"

	"dataease/backend/internal/domain/msgcenter"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestMsgCenterService_Count(t *testing.T) {
	repo := repository.NewMsgCenterRepository(testDB)
	svc := NewMsgCenterService(repo)

	t.Run("count always returns 0", func(t *testing.T) {
		result := svc.Count(&msgcenter.CountRequest{})
		assert.Equal(t, int64(0), result)
	})
}

func TestMsgCenterService_List(t *testing.T) {
	repo := repository.NewMsgCenterRepository(testDB)
	svc := NewMsgCenterService(repo)

	t.Run("list with default pagination", func(t *testing.T) {
		req := &msgcenter.ListRequest{}
		result := svc.List(req)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Current)
		assert.Equal(t, 10, result.Size)
		assert.Equal(t, int64(0), result.Total)
		assert.Empty(t, result.List)
	})

	t.Run("list with custom pagination", func(t *testing.T) {
		req := &msgcenter.ListRequest{
			Current: 2,
			Size:    20,
		}
		result := svc.List(req)
		assert.Equal(t, 2, result.Current)
		assert.Equal(t, 20, result.Size)
	})

	t.Run("list with invalid pagination", func(t *testing.T) {
		req := &msgcenter.ListRequest{
			Current: 0,
			Size:    0,
		}
		result := svc.List(req)
		assert.Equal(t, 1, result.Current)
		assert.Equal(t, 10, result.Size)
	})
}

func TestMsgCenterService_Read(t *testing.T) {
	// Clean up table
	testDB.Exec("DELETE FROM core_msg_setting")

	repo := repository.NewMsgCenterRepository(testDB)
	svc := NewMsgCenterService(repo)

	userID := time.Now().UnixNano()

	t.Run("read new message", func(t *testing.T) {
		req := &msgcenter.ReadRequest{ID: "msg-1"}
		result := svc.Read(req, userID)
		assert.True(t, result.Success)
		assert.False(t, result.AlreadyRead)
	})

	t.Run("read already read message", func(t *testing.T) {
		req := &msgcenter.ReadRequest{ID: "msg-1"}
		result := svc.Read(req, userID)
		assert.True(t, result.Success)
		assert.True(t, result.AlreadyRead)
	})

	t.Run("read different message", func(t *testing.T) {
		req := &msgcenter.ReadRequest{ID: "msg-2"}
		result := svc.Read(req, userID)
		assert.True(t, result.Success)
		assert.False(t, result.AlreadyRead)
	})
}

func TestMsgCenterService_ReadBatch(t *testing.T) {
	// Clean up table
	testDB.Exec("DELETE FROM core_msg_setting")

	repo := repository.NewMsgCenterRepository(testDB)
	svc := NewMsgCenterService(repo)

	userID := time.Now().UnixNano()

	t.Run("read batch messages", func(t *testing.T) {
		req := &msgcenter.ReadBatchRequest{
			IDs: []string{"batch-1", "batch-2", "batch-3"},
		}
		result := svc.ReadBatch(req, userID)
		assert.True(t, result.Success)
		assert.Equal(t, 3, result.Updated)
	})

	t.Run("read already read batch", func(t *testing.T) {
		req := &msgcenter.ReadBatchRequest{
			IDs: []string{"batch-1", "batch-2", "batch-3"},
		}
		result := svc.ReadBatch(req, userID)
		assert.True(t, result.Success)
		assert.Equal(t, 0, result.Updated)
	})

	t.Run("read empty batch", func(t *testing.T) {
		req := &msgcenter.ReadBatchRequest{
			IDs: []string{},
		}
		result := svc.ReadBatch(req, userID)
		assert.True(t, result.Success)
		assert.Equal(t, 0, result.Updated)
	})
}

func TestMsgCenterService_Read_IsReadError(t *testing.T) {
	err := testDB.Migrator().DropTable("core_msg_setting")
	assert.NoError(t, err)
	defer func() {
		_ = testDB.Exec(`CREATE TABLE IF NOT EXISTS core_msg_setting (
	id BIGINT AUTO_INCREMENT PRIMARY KEY,
	msg_id VARCHAR(100),
	user_id BIGINT,
	status VARCHAR(20),
	read_at DATETIME,
	UNIQUE INDEX idx_msg_user (msg_id, user_id),
	INDEX idx_user_id (user_id)
)`).Error
	}()

	repo := repository.NewMsgCenterRepository(testDB)
	svc := NewMsgCenterService(repo)

	resp := svc.Read(&msgcenter.ReadRequest{ID: "err-msg"}, time.Now().UnixNano())
	assert.False(t, resp.Success)
	assert.False(t, resp.AlreadyRead)
}

func TestMsgCenterService_Read_MarkAsReadError(t *testing.T) {
	repoDB := testDB.Session(&gorm.Session{NewDB: true})
	callbackName := "test:force_msgcenter_create_error"

	err := repoDB.Callback().Create().Before("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		tx.AddError(errors.New("forced create error"))
	})
	assert.NoError(t, err)
	defer func() {
		_ = repoDB.Callback().Create().Remove(callbackName)
	}()

	repo := repository.NewMsgCenterRepository(repoDB)
	svc := NewMsgCenterService(repo)

	resp := svc.Read(&msgcenter.ReadRequest{ID: "mark-error-msg"}, time.Now().UnixNano())
	assert.False(t, resp.Success)
	assert.False(t, resp.AlreadyRead)
}
