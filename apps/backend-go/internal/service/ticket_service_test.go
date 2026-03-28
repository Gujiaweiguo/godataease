package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/ticket"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTicketServiceRepoTest(t *testing.T) (*TicketService, *repository.TicketRepository, *gorm.DB) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE core_ticket (id INTEGER PRIMARY KEY AUTOINCREMENT, uuid TEXT, ticket TEXT UNIQUE, exp INTEGER, args TEXT, access_time INTEGER, create_time DATETIME)`).Error)

	repo := repository.NewTicketRepository(db)
	return NewTicketService(repo), repo, db
}

func setupClosedTicketServiceRepoTest(t *testing.T) *TicketService {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE core_ticket (id INTEGER PRIMARY KEY AUTOINCREMENT, uuid TEXT, ticket TEXT UNIQUE, exp INTEGER, args TEXT, access_time INTEGER, create_time DATETIME)`).Error)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	return NewTicketService(repository.NewTicketRepository(db))
}

func TestTicketService_CreateTicket_Unit(t *testing.T) {
	t.Run("uses provided ticket", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)

		resp, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "test-ticket-1", UUID: "user-1", Exp: time.Now().Unix() + 3600, Args: `{"key":"value"}`})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "test-ticket-1", resp.Ticket)
	})

	t.Run("generates when requested", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)

		resp, err := svc.CreateTicket(&ticket.TicketCreateRequest{UUID: "user-2", Exp: time.Now().Unix() + 3600, GenerateNew: true})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Ticket, 32)
	})

	t.Run("generates when ticket empty", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)

		resp, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "", UUID: "user-3", Exp: time.Now().Unix() + 3600})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Ticket, 32)
	})

	t.Run("generate new overrides provided ticket", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)

		resp, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "keep-me", UUID: "user-4", Exp: time.Now().Unix() + 3600, GenerateNew: true})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Len(t, resp.Ticket, 32)
		assert.NotEqual(t, "keep-me", resp.Ticket)
	})

	t.Run("repo error", func(t *testing.T) {
		svc := setupClosedTicketServiceRepoTest(t)

		resp, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "ticket-error", UUID: "user-5", Exp: time.Now().Unix() + 3600})
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestTicketService_ValidateTicket_Unit(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)

		resp := svc.ValidateTicket("missing")
		assert.False(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
	})

	t.Run("expired", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)
		_, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "expired-ticket", UUID: "user-1", Exp: time.Now().Unix() - 3600})
		require.NoError(t, err)

		resp := svc.ValidateTicket("expired-ticket")
		assert.False(t, resp.TicketValid)
		assert.True(t, resp.TicketExp)
	})

	t.Run("no expiration", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)
		_, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "no-exp-ticket", UUID: "user-2", Exp: 0, Args: `{"data":"test"}`})
		require.NoError(t, err)

		resp := svc.ValidateTicket("no-exp-ticket")
		assert.True(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
		assert.Equal(t, `{"data":"test"}`, resp.Args)
	})

	t.Run("valid updates access time", func(t *testing.T) {
		svc, repo, _ := setupTicketServiceRepoTest(t)
		_, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "valid-ticket", UUID: "user-3", Exp: time.Now().Unix() + 3600, Args: `{"k":"v"}`})
		require.NoError(t, err)

		resp := svc.ValidateTicket("valid-ticket")
		assert.True(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
		assert.Equal(t, `{"k":"v"}`, resp.Args)

		stored, findErr := repo.FindByTicket("valid-ticket")
		require.NoError(t, findErr)
		require.NotNil(t, stored)
		assert.NotZero(t, stored.AccessTime)
	})

	t.Run("repo error returns invalid", func(t *testing.T) {
		svc := setupClosedTicketServiceRepoTest(t)

		resp := svc.ValidateTicket("repo-error")
		assert.False(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
	})

	t.Run("update access time failure still returns valid", func(t *testing.T) {
		svc, repo, db := setupTicketServiceRepoTest(t)
		_, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "access-time-fail", UUID: "user-6", Exp: time.Now().Unix() + 3600, Args: `{"keep":true}`})
		require.NoError(t, err)
		require.NoError(t, db.Exec("CREATE TRIGGER deny_ticket_access_update BEFORE UPDATE ON core_ticket BEGIN SELECT RAISE(FAIL, 'deny access update'); END;").Error)

		resp := svc.ValidateTicket("access-time-fail")
		assert.True(t, resp.TicketValid)
		assert.False(t, resp.TicketExp)
		assert.Equal(t, `{"keep":true}`, resp.Args)

		stored, findErr := repo.FindByTicket("access-time-fail")
		require.NoError(t, findErr)
		require.NotNil(t, stored)
		assert.Zero(t, stored.AccessTime)
	})
}

func TestTicketService_ListTickets_Unit(t *testing.T) {
	t.Run("default pagination", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)

		resp := svc.ListTickets("uuid-1", 0, 0)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
		assert.Empty(t, resp.List)
	})

	t.Run("repo data", func(t *testing.T) {
		svc, _, _ := setupTicketServiceRepoTest(t)
		for i := 0; i < 3; i++ {
			_, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "ticket-list-" + string(rune('a'+i)), UUID: "uuid-list", Exp: time.Now().Unix() + 3600})
			require.NoError(t, err)
		}

		resp := svc.ListTickets("uuid-list", 1, 10)
		assert.Len(t, resp.List, 3)
		assert.Equal(t, int64(3), resp.Total)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
	})

	t.Run("repo error returns empty page", func(t *testing.T) {
		svc := setupClosedTicketServiceRepoTest(t)

		resp := svc.ListTickets("uuid-list", 1, 10)
		assert.Empty(t, resp.List)
		assert.Zero(t, resp.Total)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
	})

	t.Run("repo error preserves normalized paging", func(t *testing.T) {
		svc := setupClosedTicketServiceRepoTest(t)

		resp := svc.ListTickets("uuid-list", 0, 0)
		assert.Empty(t, resp.List)
		assert.Zero(t, resp.Total)
		assert.Equal(t, 1, resp.Current)
		assert.Equal(t, 10, resp.Size)
	})
}

func TestTicketService_DeleteTicket_Unit(t *testing.T) {
	svc, repo, _ := setupTicketServiceRepoTest(t)
	_, err := svc.CreateTicket(&ticket.TicketCreateRequest{Ticket: "ticket-to-delete", UUID: "user-del", Exp: time.Now().Unix() + 3600})
	require.NoError(t, err)

	require.NoError(t, svc.DeleteTicket(&ticket.TicketDeleteRequest{Ticket: "ticket-to-delete"}))
	stored, findErr := repo.FindByTicket("ticket-to-delete")
	require.NoError(t, findErr)
	assert.Nil(t, stored)

	svc = setupClosedTicketServiceRepoTest(t)
	err = svc.DeleteTicket(&ticket.TicketDeleteRequest{Ticket: "ticket-to-delete"})
	require.Error(t, err)
}

func TestTicketService_TempTicket_Unit(t *testing.T) {
	svc, _, _ := setupTicketServiceRepoTest(t)

	ticket1 := svc.TempTicket()
	ticket2 := svc.TempTicket()
	assert.Len(t, ticket1, 32)
	assert.Len(t, ticket2, 32)
	assert.NotEqual(t, ticket1, ticket2)
}
