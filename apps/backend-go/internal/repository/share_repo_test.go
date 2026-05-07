package repository

import (
	"testing"
	"time"

	sharedomain "dataease/backend/internal/domain/share"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupShareRepositoryTest(t *testing.T) *ShareRepository {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&coreShare{}, &coreShareTicket{}))

	return NewShareRepository(db)
}

func TestShareRepository_ShareCRUDAndUUIDOperations(t *testing.T) {
	repo := setupShareRepositoryTest(t)

	item := &sharedomain.Share{
		Creator:       11,
		ResourceID:    101,
		ResourceType:  "dashboard",
		Exp:           3600,
		UUID:          "share-uuid-1",
		Pwd:           "pwd-1",
		AutoPwd:       true,
		TicketRequire: false,
	}
	require.NoError(t, repo.Create(item))
	require.Positive(t, item.ID)
	require.NotNil(t, item.Time)

	byID, err := repo.GetByID(item.ID)
	require.NoError(t, err)
	assert.Equal(t, item.ResourceID, byID.ResourceID)
	assert.Equal(t, item.UUID, byID.UUID)

	byUUID, err := repo.GetByUUID(item.UUID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, byUUID.ID)

	byResourceID, err := repo.GetByResourceID(item.ResourceID)
	require.NoError(t, err)
	assert.Equal(t, item.ID, byResourceID.ID)

	exists, err := repo.ExistsByUUID(item.UUID, 0)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = repo.ExistsByUUID(item.UUID, item.ID)
	require.NoError(t, err)
	assert.False(t, exists)

	item.UUID = "share-uuid-1-updated"
	item.Exp = 7200
	item.Pwd = "pwd-2"
	item.AutoPwd = false
	item.TicketRequire = true
	require.NoError(t, repo.Update(item))

	updated, err := repo.GetByID(item.ID)
	require.NoError(t, err)
	assert.Equal(t, "share-uuid-1-updated", updated.UUID)
	assert.Equal(t, int64(7200), updated.Exp)
	assert.Equal(t, "pwd-2", updated.Pwd)
	assert.False(t, updated.AutoPwd)
	assert.True(t, updated.TicketRequire)

	require.NoError(t, repo.Delete(item.ID))
	_, err = repo.GetByID(item.ID)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestShareRepository_TicketCRUDAndUUIDSync(t *testing.T) {
	repo := setupShareRepositoryTest(t)

	shareItem := &sharedomain.Share{
		Creator:       21,
		ResourceID:    202,
		ResourceType:  "screen",
		Exp:           3600,
		UUID:          "share-uuid-2",
		Pwd:           "pwd",
		AutoPwd:       false,
		TicketRequire: true,
	}
	require.NoError(t, repo.Create(shareItem))

	accessTime := time.Now().Add(-time.Hour).UTC().Truncate(time.Millisecond)
	ticket := &sharedomain.ShareTicket{
		UUID:       shareItem.UUID,
		Ticket:     "ticket-1",
		Exp:        1800,
		Args:       `{"page":1}`,
		AccessTime: &accessTime,
	}
	require.NoError(t, repo.CreateTicket(ticket))
	require.Positive(t, ticket.ID)

	byUUID, err := repo.GetTicketByUUID(shareItem.UUID)
	require.NoError(t, err)
	assert.Equal(t, ticket.Ticket, byUUID.Ticket)
	assert.Equal(t, ticket.Args, byUUID.Args)
	assert.NotNil(t, byUUID.AccessTime)

	byTicket, err := repo.GetTicketByTicket(ticket.Ticket)
	require.NoError(t, err)
	assert.Equal(t, shareItem.UUID, byTicket.UUID)

	require.NoError(t, repo.UpdateTicketUUID(shareItem.UUID, "share-uuid-2-mid"))
	renamedTicket, err := repo.GetTicketByTicket(ticket.Ticket)
	require.NoError(t, err)
	assert.Equal(t, "share-uuid-2-mid", renamedTicket.UUID)

	require.NoError(t, repo.UpdateUUIDWithTickets(shareItem.ID, "share-uuid-2-mid", "share-uuid-2-final"))
	updatedShare, err := repo.GetByID(shareItem.ID)
	require.NoError(t, err)
	assert.Equal(t, "share-uuid-2-final", updatedShare.UUID)

	updatedTicket, err := repo.GetTicketByTicket(ticket.Ticket)
	require.NoError(t, err)
	assert.Equal(t, "share-uuid-2-final", updatedTicket.UUID)

	beforeUpdate := updatedTicket.AccessTime
	require.NoError(t, repo.UpdateTicketAccessTime(ticket.Ticket))
	afterUpdate, err := repo.GetTicketByTicket(ticket.Ticket)
	require.NoError(t, err)
	require.NotNil(t, afterUpdate.AccessTime)
	if beforeUpdate != nil {
		assert.True(t, afterUpdate.AccessTime.After(*beforeUpdate) || afterUpdate.AccessTime.Equal(*beforeUpdate))
	}

	require.NoError(t, repo.DeleteTicket(ticket.ID))
	_, err = repo.GetTicketByTicket(ticket.Ticket)
	require.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
