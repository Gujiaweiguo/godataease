package share

import (
	"testing"
	"time"
)

func TestShareTicket_Fields(t *testing.T) {
	now := time.Now()
	ticket := ShareTicket{
		ID:         1,
		UUID:       "test-uuid",
		Ticket:     "ticket-code",
		Exp:        3600,
		Args:       `{"key":"value"}`,
		AccessTime: &now,
	}

	if ticket.ID != 1 {
		t.Errorf("Expected ID 1, got %d", ticket.ID)
	}
	if ticket.UUID != "test-uuid" {
		t.Errorf("Expected UUID 'test-uuid', got '%s'", ticket.UUID)
	}
	if ticket.Ticket != "ticket-code" {
		t.Errorf("Expected Ticket 'ticket-code', got '%s'", ticket.Ticket)
	}
	if ticket.Exp != 3600 {
		t.Errorf("Expected Exp 3600, got %d", ticket.Exp)
	}
}

func TestShareTicket_NilAccessTime(t *testing.T) {
	ticket := ShareTicket{
		ID:     1,
		UUID:   "test-uuid",
		Ticket: "ticket-code",
	}

	if ticket.AccessTime != nil {
		t.Error("Expected AccessTime to be nil")
	}
}

func TestShare_Fields(t *testing.T) {
	now := time.Now()
	share := Share{
		ID:            1,
		Creator:       100,
		ResourceID:    200,
		ResourceType:  "dashboard",
		Time:          &now,
		Exp:           3600,
		UUID:          "share-uuid",
		Pwd:           "password123",
		AutoPwd:       true,
		TicketRequire: false,
	}

	if share.ID != 1 {
		t.Errorf("Expected ID 1, got %d", share.ID)
	}
	if share.Creator != 100 {
		t.Errorf("Expected Creator 100, got %d", share.Creator)
	}
	if share.ResourceID != 200 {
		t.Errorf("Expected ResourceID 200, got %d", share.ResourceID)
	}
	if share.ResourceType != "dashboard" {
		t.Errorf("Expected ResourceType 'dashboard', got '%s'", share.ResourceType)
	}
	if !share.AutoPwd {
		t.Error("Expected AutoPwd to be true")
	}
}

func TestShareCreateRequest_Fields(t *testing.T) {
	req := ShareCreateRequest{
		ResourceID:   1,
		ResourceType: "dashboard",
		Exp:          3600,
		AutoPwd:      true,
	}

	if req.ResourceID != 1 {
		t.Errorf("Expected ResourceID 1, got %d", req.ResourceID)
	}
	if req.ResourceType != "dashboard" {
		t.Errorf("Expected ResourceType 'dashboard', got '%s'", req.ResourceType)
	}
}

func TestShareCreateResponse_Fields(t *testing.T) {
	resp := ShareCreateResponse{
		ID:      1,
		UUID:    "new-uuid",
		Pwd:     "generated-pwd",
		AutoPwd: true,
	}

	if resp.ID != 1 {
		t.Errorf("Expected ID 1, got %d", resp.ID)
	}
	if resp.UUID != "new-uuid" {
		t.Errorf("Expected UUID 'new-uuid', got '%s'", resp.UUID)
	}
}

func TestShareValidateRequest_Fields(t *testing.T) {
	req := ShareValidateRequest{
		UUID: "test-uuid",
		Pwd:  "password",
	}

	if req.UUID != "test-uuid" {
		t.Errorf("Expected UUID 'test-uuid', got '%s'", req.UUID)
	}
	if req.Pwd != "password" {
		t.Errorf("Expected Pwd 'password', got '%s'", req.Pwd)
	}
}

func TestShareValidateResponse_Fields(t *testing.T) {
	resp := ShareValidateResponse{
		Valid:         true,
		ResourceID:    1,
		ResourceType:  "dashboard",
		TicketRequire: false,
	}

	if !resp.Valid {
		t.Error("Expected Valid to be true")
	}
	if resp.ResourceID != 1 {
		t.Errorf("Expected ResourceID 1, got %d", resp.ResourceID)
	}
}

func TestShareRevokeRequest_Fields(t *testing.T) {
	req := ShareRevokeRequest{ID: 1}

	if req.ID != 1 {
		t.Errorf("Expected ID 1, got %d", req.ID)
	}
}

func TestShareRevokeResponse_Fields(t *testing.T) {
	resp := ShareRevokeResponse{Success: true}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
}

func TestTicketCreateRequest_Fields(t *testing.T) {
	req := TicketCreateRequest{
		Ticket:      "ticket-code",
		Exp:         3600,
		Args:        `{"key":"value"}`,
		UUID:        "share-uuid",
		GenerateNew: true,
	}

	if req.Ticket != "ticket-code" {
		t.Errorf("Expected Ticket 'ticket-code', got '%s'", req.Ticket)
	}
	if !req.GenerateNew {
		t.Error("Expected GenerateNew to be true")
	}
}

func TestTicketValidateRequest_Fields(t *testing.T) {
	req := TicketValidateRequest{
		Ticket: "ticket-code",
		UUID:   "share-uuid",
	}

	if req.Ticket != "ticket-code" {
		t.Errorf("Expected Ticket 'ticket-code', got '%s'", req.Ticket)
	}
	if req.UUID != "share-uuid" {
		t.Errorf("Expected UUID 'share-uuid', got '%s'", req.UUID)
	}
}

func TestTicketValidateResponse_Fields(t *testing.T) {
	resp := TicketValidateResponse{
		TicketValid: true,
		TicketExp:   false,
		Args:        `{"key":"value"}`,
	}

	if !resp.TicketValid {
		t.Error("Expected TicketValid to be true")
	}
	if resp.TicketExp {
		t.Error("Expected TicketExp to be false")
	}
}

func TestShareDetailResponse_Fields(t *testing.T) {
	resp := ShareDetailResponse{
		ID:            1,
		Exp:           3600,
		UUID:          "detail-uuid",
		Pwd:           "password",
		AutoPwd:       true,
		TicketRequire: false,
	}

	if resp.ID != 1 {
		t.Errorf("Expected ID 1, got %d", resp.ID)
	}
	if resp.Exp != 3600 {
		t.Errorf("Expected Exp 3600, got %d", resp.Exp)
	}
}
