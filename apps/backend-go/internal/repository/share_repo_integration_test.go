//go:build integration
// +build integration

package repository

import (
	"testing"
	"time"

	"dataease/backend/internal/domain/share"

	"github.com/google/uuid"
)

func TestShareRepository_CreateAndGetByID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share")

	s := &share.Share{
		Creator:       1,
		ResourceID:    100,
		ResourceType:  "dashboard",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          uuid.New().String(),
		Pwd:           "test123",
		AutoPwd:       true,
		TicketRequire: false,
	}

	err := repo.Create(s)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if s.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}

	found, err := repo.GetByID(s.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if found.ResourceID != 100 {
		t.Errorf("Expected ResourceID 100, got %d", found.ResourceID)
	}
}

func TestShareRepository_GetByUUID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share")

	testUUID := uuid.New().String()
	s := &share.Share{
		Creator:       1,
		ResourceID:    200,
		ResourceType:  "screen",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          testUUID,
		AutoPwd:       false,
		TicketRequire: true,
	}
	_ = repo.Create(s)

	found, err := repo.GetByUUID(testUUID)
	if err != nil {
		t.Fatalf("GetByUUID failed: %v", err)
	}

	if found.UUID != testUUID {
		t.Errorf("Expected UUID '%s', got '%s'", testUUID, found.UUID)
	}
}

func TestShareRepository_GetByResourceID(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share")

	s := &share.Share{
		Creator:       1,
		ResourceID:    300,
		ResourceType:  "dashboard",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          uuid.New().String(),
		AutoPwd:       true,
		TicketRequire: false,
	}
	_ = repo.Create(s)

	found, err := repo.GetByResourceID(300)
	if err != nil {
		t.Fatalf("GetByResourceID failed: %v", err)
	}

	if found.ResourceID != 300 {
		t.Errorf("Expected ResourceID 300, got %d", found.ResourceID)
	}
}

func TestShareRepository_Update(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share")

	s := &share.Share{
		Creator:       1,
		ResourceID:    400,
		ResourceType:  "dashboard",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          uuid.New().String(),
		Pwd:           "oldpwd",
		AutoPwd:       true,
		TicketRequire: false,
	}
	_ = repo.Create(s)

	s.Pwd = "newpwd"
	s.AutoPwd = false
	s.TicketRequire = true

	err := repo.Update(s)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.GetByID(s.ID)
	if found.Pwd != "newpwd" {
		t.Errorf("Expected Pwd 'newpwd', got '%s'", found.Pwd)
	}
}

func TestShareRepository_Delete(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share")

	s := &share.Share{
		Creator:       1,
		ResourceID:    500,
		ResourceType:  "dashboard",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          uuid.New().String(),
		AutoPwd:       true,
		TicketRequire: false,
	}
	_ = repo.Create(s)

	err := repo.Delete(s.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.GetByID(s.ID)
	if err == nil {
		t.Error("Expected error when getting deleted share")
	}
}

func TestShareRepository_Ticket(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share", "core_share_ticket")

	testUUID := uuid.New().String()
	s := &share.Share{
		Creator:       1,
		ResourceID:    600,
		ResourceType:  "dashboard",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          testUUID,
		AutoPwd:       true,
		TicketRequire: true,
	}
	_ = repo.Create(s)

	ticket := &share.ShareTicket{
		UUID:   testUUID,
		Ticket: "ticket123",
		Exp:    time.Now().Add(1 * time.Hour).Unix(),
		Args:   `{"param": "value"}`,
	}

	err := repo.CreateTicket(ticket)
	if err != nil {
		t.Fatalf("CreateTicket failed: %v", err)
	}

	if ticket.ID == 0 {
		t.Error("Expected ID to be set after ticket creation")
	}

	found, err := repo.GetTicketByUUID(testUUID)
	if err != nil {
		t.Fatalf("GetTicketByUUID failed: %v", err)
	}

	if found.Ticket != "ticket123" {
		t.Errorf("Expected Ticket 'ticket123', got '%s'", found.Ticket)
	}

	foundByTicket, err := repo.GetTicketByTicket("ticket123")
	if err != nil {
		t.Fatalf("GetTicketByTicket failed: %v", err)
	}

	if foundByTicket.UUID != testUUID {
		t.Errorf("Expected UUID '%s', got '%s'", testUUID, foundByTicket.UUID)
	}
}

func TestShareRepository_UpdateTicketAccessTime(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share", "core_share_ticket")

	testUUID := uuid.New().String()
	s := &share.Share{
		Creator:       1,
		ResourceID:    700,
		ResourceType:  "dashboard",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          testUUID,
		AutoPwd:       true,
		TicketRequire: true,
	}
	_ = repo.Create(s)

	ticket := &share.ShareTicket{
		UUID:   testUUID,
		Ticket: "access-test-ticket",
		Exp:    time.Now().Add(1 * time.Hour).Unix(),
	}
	_ = repo.CreateTicket(ticket)

	err := repo.UpdateTicketAccessTime("access-test-ticket")
	if err != nil {
		t.Fatalf("UpdateTicketAccessTime failed: %v", err)
	}

	found, _ := repo.GetTicketByTicket("access-test-ticket")
	if found.AccessTime == nil {
		t.Error("Expected AccessTime to be set")
	}
}

func TestShareRepository_DeleteTicket(t *testing.T) {
	if testDB == nil {
		t.Skip("Test database not available")
	}

	repo := NewShareRepository(testDB)
	cleanupTables("core_share", "core_share_ticket")

	testUUID := uuid.New().String()
	s := &share.Share{
		Creator:       1,
		ResourceID:    800,
		ResourceType:  "dashboard",
		Exp:           time.Now().Add(24 * time.Hour).Unix(),
		UUID:          testUUID,
		AutoPwd:       true,
		TicketRequire: true,
	}
	_ = repo.Create(s)

	ticket := &share.ShareTicket{
		UUID:   testUUID,
		Ticket: "delete-test-ticket",
		Exp:    time.Now().Add(1 * time.Hour).Unix(),
	}
	_ = repo.CreateTicket(ticket)

	err := repo.DeleteTicket(ticket.ID)
	if err != nil {
		t.Fatalf("DeleteTicket failed: %v", err)
	}

	_, err = repo.GetTicketByTicket("delete-test-ticket")
	if err == nil {
		t.Error("Expected error when getting deleted ticket")
	}
}
