//go:build integration

package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/ticket"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTicketService_CreateTicket(t *testing.T) {
	// Clean up table
	testDB.Exec("DELETE FROM core_ticket")

	repo := repository.NewTicketRepository(testDB)
	svc := NewTicketService(repo)

	t.Run("create ticket with provided value", func(t *testing.T) {
		req := &ticket.TicketCreateRequest{
			Ticket: "test-ticket-1",
			UUID:   "user-1",
			Exp:    time.Now().Unix() + 3600,
			Args:   `{"key":"value"}`,
		}

		result, err := svc.CreateTicket(req)
		require.NoError(t, err)
		assert.Equal(t, "test-ticket-1", result.Ticket)
	})

	t.Run("create ticket with auto generation", func(t *testing.T) {
		req := &ticket.TicketCreateRequest{
			UUID:        "user-2",
			Exp:         time.Now().Unix() + 3600,
			GenerateNew: true,
		}

		result, err := svc.CreateTicket(req)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Ticket)
		assert.Len(t, result.Ticket, 32) // hex encoded 16 bytes = 32 chars
	})

	t.Run("create ticket with empty ticket generates new", func(t *testing.T) {
		req := &ticket.TicketCreateRequest{
			Ticket: "",
			UUID:   "user-3",
			Exp:    time.Now().Unix() + 3600,
		}

		result, err := svc.CreateTicket(req)
		require.NoError(t, err)
		assert.NotEmpty(t, result.Ticket)
	})
}

func TestTicketService_ValidateTicket(t *testing.T) {
	// Clean up table
	testDB.Exec("DELETE FROM core_ticket")

	repo := repository.NewTicketRepository(testDB)
	svc := NewTicketService(repo)

	t.Run("validate valid ticket", func(t *testing.T) {
		// Create ticket
		req := &ticket.TicketCreateRequest{
			Ticket: "valid-ticket",
			UUID:   "user-1",
			Exp:    time.Now().Unix() + 3600,
			Args:   `{"data":"test"}`,
		}
		_, err := svc.CreateTicket(req)
		require.NoError(t, err)

		// Validate
		result := svc.ValidateTicket("valid-ticket")
		assert.True(t, result.TicketValid)
		assert.False(t, result.TicketExp)
		assert.Equal(t, `{"data":"test"}`, result.Args)
	})

	t.Run("validate expired ticket", func(t *testing.T) {
		// Create expired ticket
		req := &ticket.TicketCreateRequest{
			Ticket: "expired-ticket",
			UUID:   "user-2",
			Exp:    time.Now().Unix() - 3600, // Expired 1 hour ago
		}
		_, err := svc.CreateTicket(req)
		require.NoError(t, err)

		// Validate
		result := svc.ValidateTicket("expired-ticket")
		assert.False(t, result.TicketValid)
		assert.True(t, result.TicketExp)
	})

	t.Run("validate non-existent ticket", func(t *testing.T) {
		result := svc.ValidateTicket("non-existent-ticket")
		assert.False(t, result.TicketValid)
		assert.False(t, result.TicketExp)
	})

	t.Run("validate ticket with no expiration", func(t *testing.T) {
		// Create ticket with no expiration (exp = 0)
		req := &ticket.TicketCreateRequest{
			Ticket: "no-exp-ticket",
			UUID:   "user-3",
			Exp:    0, // No expiration
		}
		_, err := svc.CreateTicket(req)
		require.NoError(t, err)

		// Validate
		result := svc.ValidateTicket("no-exp-ticket")
		assert.True(t, result.TicketValid)
		assert.False(t, result.TicketExp)
	})
}

func TestTicketService_DeleteTicket(t *testing.T) {
	// Clean up table
	testDB.Exec("DELETE FROM core_ticket")

	repo := repository.NewTicketRepository(testDB)
	svc := NewTicketService(repo)

	t.Run("delete existing ticket", func(t *testing.T) {
		// Create ticket
		req := &ticket.TicketCreateRequest{
			Ticket: "ticket-to-delete",
			UUID:   "user-1",
			Exp:    time.Now().Unix() + 3600,
		}
		_, err := svc.CreateTicket(req)
		require.NoError(t, err)

		// Delete
		err = svc.DeleteTicket(&ticket.TicketDeleteRequest{Ticket: "ticket-to-delete"})
		require.NoError(t, err)

		// Verify deleted
		result := svc.ValidateTicket("ticket-to-delete")
		assert.False(t, result.TicketValid)
	})
}

func TestTicketService_ListTickets(t *testing.T) {
	// Clean up table
	testDB.Exec("DELETE FROM core_ticket")

	repo := repository.NewTicketRepository(testDB)
	svc := NewTicketService(repo)

	uuid := "test-uuid-list"

	// Create multiple tickets
	for i := 1; i <= 15; i++ {
		req := &ticket.TicketCreateRequest{
			Ticket: "list-ticket-" + string(rune('0'+i)),
			UUID:   uuid,
			Exp:    time.Now().Unix() + 3600,
		}
		_, err := svc.CreateTicket(req)
		require.NoError(t, err)
	}

	t.Run("list first page", func(t *testing.T) {
		result := svc.ListTickets(uuid, 1, 10)
		assert.Len(t, result.List, 10)
		assert.Equal(t, int64(15), result.Total)
		assert.Equal(t, 1, result.Current)
		assert.Equal(t, 10, result.Size)
	})

	t.Run("list second page", func(t *testing.T) {
		result := svc.ListTickets(uuid, 2, 10)
		assert.Len(t, result.List, 5)
		assert.Equal(t, int64(15), result.Total)
		assert.Equal(t, 2, result.Current)
	})

	t.Run("list with invalid pagination", func(t *testing.T) {
		result := svc.ListTickets(uuid, 0, 0)
		assert.Equal(t, 1, result.Current)
		assert.Equal(t, 10, result.Size)
	})

	t.Run("list non-existent uuid", func(t *testing.T) {
		result := svc.ListTickets("non-existent-uuid", 1, 10)
		assert.Empty(t, result.List)
		assert.Equal(t, int64(0), result.Total)
	})
}

func TestTicketService_TempTicket(t *testing.T) {
	repo := repository.NewTicketRepository(testDB)
	svc := NewTicketService(repo)

	t.Run("generate temp ticket", func(t *testing.T) {
		result := svc.TempTicket()
		assert.NotEmpty(t, result)
		assert.Len(t, result, 32) // hex encoded 16 bytes = 32 chars
	})

	t.Run("generate unique temp tickets", func(t *testing.T) {
		ticket1 := svc.TempTicket()
		ticket2 := svc.TempTicket()
		assert.NotEqual(t, ticket1, ticket2)
	})
}
