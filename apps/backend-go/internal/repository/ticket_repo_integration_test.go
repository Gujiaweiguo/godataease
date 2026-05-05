//go:build integration
// +build integration

package repository

import (
	"testing"

	"dataease/backend/internal/domain/ticket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketRepository_CRUD(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_ticket")

	repo := NewTicketRepository(testDB)
	item := &ticket.Ticket{
		UUID:   "user-uuid-1",
		Ticket: "ticket-crud-1",
		Exp:    1710003001,
		Args:   `{"scope":"dashboard"}`,
	}

	require.NoError(t, repo.Create(item))
	assert.NotZero(t, item.ID)

	storedByID, err := repo.FindByID(item.ID)
	require.NoError(t, err)
	require.NotNil(t, storedByID)
	assert.Equal(t, item.Ticket, storedByID.Ticket)
	assert.Equal(t, item.AccessTime, storedByID.AccessTime)

	storedByTicket, err := repo.FindByTicket(item.Ticket)
	require.NoError(t, err)
	require.NotNil(t, storedByTicket)
	assert.Equal(t, item.ID, storedByTicket.ID)
	assert.Equal(t, item.UUID, storedByTicket.UUID)
	assert.Equal(t, item.Args, storedByTicket.Args)

	accessTime := int64(1710003999)
	require.NoError(t, repo.UpdateAccessTime(item.Ticket, accessTime))

	updated, err := repo.FindByTicket(item.Ticket)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, accessTime, updated.AccessTime)

	require.NoError(t, repo.Delete(item.Ticket))

	missingByTicket, err := repo.FindByTicket(item.Ticket)
	require.NoError(t, err)
	assert.Nil(t, missingByTicket)

	missingByID, err := repo.FindByID(item.ID)
	require.NoError(t, err)
	assert.Nil(t, missingByID)

	missingByTicket, err = repo.FindByTicket("missing-ticket")
	require.NoError(t, err)
	assert.Nil(t, missingByTicket)

	missingByID, err = repo.FindByID(999999)
	require.NoError(t, err)
	assert.Nil(t, missingByID)
}

func TestTicketRepository_ListByUUID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}
	cleanupTables("core_ticket")

	repo := NewTicketRepository(testDB)
	seed := []*ticket.Ticket{
		{UUID: "shared-uuid", Ticket: "ticket-list-1", Exp: 1710003101, Args: `{"idx":1}`},
		{UUID: "shared-uuid", Ticket: "ticket-list-2", Exp: 1710003102, Args: `{"idx":2}`},
		{UUID: "shared-uuid", Ticket: "ticket-list-3", Exp: 1710003103, Args: `{"idx":3}`},
		{UUID: "other-uuid", Ticket: "ticket-list-4", Exp: 1710003104, Args: `{"idx":4}`},
	}
	for _, item := range seed {
		require.NoError(t, repo.Create(item))
	}

	pageOne, total, err := repo.ListByUUID("shared-uuid", 1, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total)
	require.Len(t, pageOne, 2)
	assert.Equal(t, seed[0].Ticket, pageOne[0].Ticket)
	assert.Equal(t, seed[1].Ticket, pageOne[1].Ticket)

	pageTwo, secondTotal, err := repo.ListByUUID("shared-uuid", 2, 2)
	require.NoError(t, err)
	assert.Equal(t, int64(3), secondTotal)
	require.Len(t, pageTwo, 1)
	assert.Equal(t, seed[2].Ticket, pageTwo[0].Ticket)

	empty, emptyTotal, err := repo.ListByUUID("missing-uuid", 1, 10)
	require.NoError(t, err)
	assert.Equal(t, int64(0), emptyTotal)
	assert.Empty(t, empty)
}
