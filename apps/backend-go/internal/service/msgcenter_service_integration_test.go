//go:build integration

package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/msgcenter"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
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
