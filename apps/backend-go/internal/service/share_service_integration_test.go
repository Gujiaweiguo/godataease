//go:build integration

package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/share"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
)

func TestShareServiceIntegration_CreateShare(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	req := &share.ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	resp, err := svc.CreateShare(req, 1)
	assert.NoError(t, err)
	assert.Greater(t, resp.ID, int64(0))
	assert.NotEmpty(t, resp.UUID)
	assert.NotEmpty(t, resp.Pwd)
	assert.True(t, resp.AutoPwd)
}

func TestShareServiceIntegration_CreateShare_NoPassword(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	req := &share.ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		AutoPwd:      false,
	}
	resp, err := svc.CreateShare(req, 1)
	assert.NoError(t, err)
	assert.Greater(t, resp.ID, int64(0))
	assert.NotEmpty(t, resp.UUID)
	assert.Empty(t, resp.Pwd)
}

func TestShareServiceIntegration_ValidateShare_Success(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share first
	createReq := &share.ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		AutoPwd:      false,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Validate without password
	validateReq := &share.ShareValidateRequest{
		UUID: created.UUID,
		Pwd:  "",
	}
	resp, err := svc.ValidateShare(validateReq)
	assert.NoError(t, err)
	assert.True(t, resp.Valid)
	assert.Equal(t, int64(1), resp.ResourceID)
	assert.Equal(t, "dashboard", resp.ResourceType)
}

func TestShareServiceIntegration_ValidateShare_NotFound(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	req := &share.ShareValidateRequest{
		UUID: "nonexistent",
		Pwd:  "",
	}
	resp, err := svc.ValidateShare(req)
	assert.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestShareServiceIntegration_ValidateShare_Expired(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create expired share
	createReq := &share.ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		Exp:          time.Now().Unix() - 100, // already expired
		AutoPwd:      false,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Validate expired share
	validateReq := &share.ShareValidateRequest{
		UUID: created.UUID,
		Pwd:  "",
	}
	resp, err := svc.ValidateShare(validateReq)
	assert.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestShareServiceIntegration_ValidateShare_WrongPassword(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share with password
	createReq := &share.ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Validate with wrong password
	validateReq := &share.ShareValidateRequest{
		UUID: created.UUID,
		Pwd:  "wrongpassword",
	}
	resp, err := svc.ValidateShare(validateReq)
	assert.NoError(t, err)
	assert.False(t, resp.Valid)
}

func TestShareServiceIntegration_RevokeShare_Success(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share
	createReq := &share.ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Revoke with correct creator
	resp, err := svc.RevokeShare(created.ID, 1)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestShareServiceIntegration_RevokeShare_WrongCreator(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share
	createReq := &share.ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Revoke with wrong creator
	resp, err := svc.RevokeShare(created.ID, 999)
	assert.NoError(t, err)
	assert.False(t, resp.Success)
}

func TestShareServiceIntegration_GetDetail(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share with unique resource ID
	createReq := &share.ShareCreateRequest{
		ResourceID:   1001,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Get detail by resource ID
	detail, err := svc.GetDetail(1001)
	assert.NoError(t, err)
	assert.NotNil(t, detail)
	assert.Equal(t, created.ID, detail.ID)
}

func TestShareServiceIntegration_GetDetail_NotFound(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	detail, err := svc.GetDetail(999)
	assert.NoError(t, err)
	assert.Nil(t, detail)
}

func TestShareServiceIntegration_SwitchStatus_Create(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Switch status on non-existent share (should create new)
	err := svc.SwitchStatus(1, 1)
	assert.NoError(t, err)

	// Verify share was created
	detail, err := svc.GetDetail(1)
	assert.NoError(t, err)
	assert.NotNil(t, detail)
}

func TestShareServiceIntegration_SwitchStatus_Delete(t *testing.T) {
	cleanupTables(&share.Share{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share first with unique resource ID
	createReq := &share.ShareCreateRequest{
		ResourceID:   2001,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	_, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Switch status (should delete)
	err = svc.SwitchStatus(2001, 1)
	assert.NoError(t, err)

	// Verify share was deleted
	detail, err := svc.GetDetail(2001)
	assert.NoError(t, err)
	assert.Nil(t, detail)
}

func TestShareServiceIntegration_Ticket(t *testing.T) {
	cleanupTables(&share.Share{}, &share.ShareTicket{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share first with unique resource ID
	createReq := &share.ShareCreateRequest{
		ResourceID:   3001,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Create ticket with unique ticket name
	ticketReq := &share.TicketCreateRequest{
		UUID:   created.UUID,
		Ticket: "test-ticket-3001",
		Exp:    time.Now().Unix() + 3600,
		Args:   "test-args",
	}
	ticket, err := svc.CreateTicket(ticketReq)
	assert.NoError(t, err)
	assert.NotNil(t, ticket)
	assert.Equal(t, created.UUID, ticket.UUID)
	assert.Equal(t, "test-ticket-3001", ticket.Ticket)

	// Validate ticket
	validateReq := &share.TicketValidateRequest{
		Ticket: "test-ticket-3001",
		UUID:   created.UUID,
	}
	resp, err := svc.ValidateTicket(validateReq)
	assert.NoError(t, err)
	assert.True(t, resp.TicketValid)
	assert.False(t, resp.TicketExp)
	assert.Equal(t, "test-args", resp.Args)
}

func TestShareServiceIntegration_Ticket_WrongUUID(t *testing.T) {
	cleanupTables(&share.Share{}, &share.ShareTicket{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share first with unique resource ID
	createReq := &share.ShareCreateRequest{
		ResourceID:   4001,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Create ticket with unique ticket name
	ticketReq := &share.TicketCreateRequest{
		UUID:   created.UUID,
		Ticket: "test-ticket-4001",
		Exp:    time.Now().Unix() + 3600,
	}
	_, err = svc.CreateTicket(ticketReq)
	assert.NoError(t, err)

	// Validate with wrong UUID
	validateReq := &share.TicketValidateRequest{
		Ticket: "test-ticket-4001",
		UUID:   "wrong-uuid",
	}
	resp, err := svc.ValidateTicket(validateReq)
	assert.NoError(t, err)
	assert.False(t, resp.TicketValid)
}

func TestShareServiceIntegration_Ticket_GenerateNew(t *testing.T) {
	cleanupTables(&share.Share{}, &share.ShareTicket{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share first
	createReq := &share.ShareCreateRequest{
		ResourceID:   5001,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Create ticket with GenerateNew=true
	ticketReq := &share.TicketCreateRequest{
		UUID:       created.UUID,
		GenerateNew: true,
		Exp:        0,
		Args:       "",
	}
	ticket, err := svc.CreateTicket(ticketReq)
	assert.NoError(t, err)
	assert.NotNil(t, ticket)
	assert.NotEmpty(t, ticket.Ticket)
	assert.Len(t, ticket.Ticket, 32) // hex encoded 16 bytes
}

func TestShareServiceIntegration_Ticket_Expired(t *testing.T) {
	cleanupTables(&share.Share{}, &share.ShareTicket{})

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Create share first
	createReq := &share.ShareCreateRequest{
		ResourceID:   5002,
		ResourceType: "dashboard",
		AutoPwd:      true,
	}
	created, err := svc.CreateShare(createReq, 1)
	assert.NoError(t, err)

	// Create expired ticket
	ticketReq := &share.TicketCreateRequest{
		UUID:   created.UUID,
		Ticket: "expired-ticket-5002",
		Exp:    time.Now().Unix() - 3600, // expired 1 hour ago
		Args:   "",
	}
	_, err = svc.CreateTicket(ticketReq)
	assert.NoError(t, err)

	// Validate expired ticket
	validateReq := &share.TicketValidateRequest{
		Ticket: "expired-ticket-5002",
		UUID:   created.UUID,
	}
	resp, err := svc.ValidateTicket(validateReq)
	assert.NoError(t, err)
	assert.True(t, resp.TicketValid)
	assert.True(t, resp.TicketExp) // ticket is expired
}

func TestShareServiceIntegration_Ticket_NotFound(t *testing.T) {
	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Validate non-existent ticket
	validateReq := &share.TicketValidateRequest{
		Ticket: "nonexistent-ticket",
		UUID:   "any-uuid",
	}
	resp, err := svc.ValidateTicket(validateReq)
	assert.NoError(t, err)
	assert.False(t, resp.TicketValid)
	assert.False(t, resp.TicketExp)
}

func TestShareServiceIntegration_RevokeShare_NotFound(t *testing.T) {
	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	// Revoke non-existent share
	resp, err := svc.RevokeShare(99999, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.False(t, resp.Success)
}
