//go:build integration

package service

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/share"
	"dataease/backend/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShareServiceIntegration_CreateShare(t *testing.T) {
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share")

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	detail, err := svc.GetDetail(999)
	assert.NoError(t, err)
	assert.Nil(t, detail)
}

func TestShareServiceIntegration_EditUUID(t *testing.T) {
	cleanupTables("core_share", "core_share_ticket")

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	first, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 5001, ResourceType: "dashboard", AutoPwd: true}, 1)
	require.NoError(t, err)
	_, err = svc.CreateShare(&share.ShareCreateRequest{ResourceID: 5002, ResourceType: "dashboard", AutoPwd: true}, 1)
	require.NoError(t, err)

	ticket, err := svc.CreateTicket(&share.TicketCreateRequest{UUID: first.UUID, Ticket: "ticket-edit-uuid", Exp: time.Now().UnixMilli() + 60000})
	require.NoError(t, err)

	msg, err := svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 5001, UUID: "newuuid88"}, 1)
	require.NoError(t, err)
	assert.Equal(t, "", msg)

	detail, err := svc.GetDetail(5001)
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, "newuuid88", detail.UUID)

	updatedTicket, err := repo.GetTicketByTicket(ticket.Ticket)
	require.NoError(t, err)
	assert.Equal(t, "newuuid88", updatedTicket.UUID)

	msg, err = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 5001, UUID: "bad-uuid"}, 1)
	require.NoError(t, err)
	assert.Equal(t, "invalid uuid format", msg)

	msg, err = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 5002, UUID: "otheruuid9"}, 1)
	require.NoError(t, err)
	assert.Equal(t, "", msg)

	msg, err = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 5001, UUID: "otheruuid9"}, 1)
	require.NoError(t, err)
	assert.Equal(t, "uuid already exists", msg)

	msg, err = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 5001, UUID: "newuuid88"}, 1)
	require.NoError(t, err)
	assert.Equal(t, "", msg)
}

func TestShareServiceIntegration_EditExpAndPwdReflectInDetail(t *testing.T) {
	cleanupTables("core_share")

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 6001, ResourceType: "dashboard", AutoPwd: true}, 1)
	require.NoError(t, err)

	newExp := time.Now().Add(time.Hour).UnixMilli()
	require.NoError(t, svc.EditExp(&share.ShareEditExpRequest{ResourceID: 6001, Exp: newExp}, 1))
	require.NoError(t, svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 6001, Pwd: "Ab1!", AutoPwd: false}, 1))

	detail, err := svc.GetDetail(6001)
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Equal(t, newExp, detail.Exp)
	assert.Equal(t, "Ab1!", detail.Pwd)
	assert.False(t, detail.AutoPwd)

	require.NoError(t, svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 6001, Pwd: "", AutoPwd: true}, 1))
	detail, err = svc.GetDetail(6001)
	require.NoError(t, err)
	require.NotNil(t, detail)
	assert.Empty(t, detail.Pwd)
	assert.True(t, detail.AutoPwd)
}

func TestShareServiceIntegration_EditExpAndPwdRejectInvalidInput(t *testing.T) {
	cleanupTables("core_share")

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 7001, ResourceType: "dashboard", AutoPwd: true}, 1)
	require.NoError(t, err)

	err = svc.EditExp(&share.ShareEditExpRequest{ResourceID: 7001, Exp: time.Now().Add(-time.Hour).UnixMilli()}, 1)
	assert.EqualError(t, err, "invalid expiration")

	err = svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 7001, Pwd: "bad", AutoPwd: false}, 1)
	assert.EqualError(t, err, "invalid password format")
}

func TestShareServiceIntegration_EditRejectedForNonCreator(t *testing.T) {
	cleanupTables("core_share")

	repo := repository.NewShareRepository(testDB)
	svc := NewShareService(repo)

	_, err := svc.CreateShare(&share.ShareCreateRequest{ResourceID: 7101, ResourceType: "dashboard", AutoPwd: true}, 1)
	require.NoError(t, err)

	_, err = svc.EditUUID(&share.ShareEditUUIDRequest{ResourceID: 7101, UUID: "secure888"}, 2)
	assert.EqualError(t, err, "forbidden")

	err = svc.EditExp(&share.ShareEditExpRequest{ResourceID: 7101, Exp: time.Now().Add(time.Hour).UnixMilli()}, 2)
	assert.EqualError(t, err, "forbidden")

	err = svc.EditPwd(&share.ShareEditPwdRequest{ResourceID: 7101, Pwd: "Ab1!", AutoPwd: false}, 2)
	assert.EqualError(t, err, "forbidden")
}

func TestShareServiceIntegration_SwitchStatus_Create(t *testing.T) {
	cleanupTables("core_share")

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
	cleanupTables("core_share")

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
	cleanupTables("core_share", "core_share_ticket")

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
	cleanupTables("core_share", "core_share_ticket")

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
	cleanupTables("core_share", "core_share_ticket")

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
		UUID:        created.UUID,
		GenerateNew: true,
		Exp:         0,
		Args:        "",
	}
	ticket, err := svc.CreateTicket(ticketReq)
	assert.NoError(t, err)
	assert.NotNil(t, ticket)
	assert.NotEmpty(t, ticket.Ticket)
	assert.Len(t, ticket.Ticket, 32) // hex encoded 16 bytes
}

func TestShareServiceIntegration_Ticket_Expired(t *testing.T) {
	cleanupTables("core_share", "core_share_ticket")

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
