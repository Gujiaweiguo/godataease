package ticket

import (
	"testing"
)

func TestTicket_Fields(t *testing.T) {
	ticket := Ticket{
		ID:         1,
		UUID:       "test-uuid",
		Ticket:     "ticket-code",
		Exp:        3600,
		Args:       `{"key":"value"}`,
		AccessTime: 1700000000,
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
	if ticket.Args != `{"key":"value"}` {
		t.Errorf("Expected Args '{\"key\":\"value\"}', got '%s'", ticket.Args)
	}
	if ticket.AccessTime != 1700000000 {
		t.Errorf("Expected AccessTime 1700000000, got %d", ticket.AccessTime)
	}
}

func TestTicketCreateRequest_Fields(t *testing.T) {
	req := TicketCreateRequest{
		Ticket:      "new-ticket",
		Exp:         7200,
		Args:        `{"filter":"value"}`,
		UUID:        "share-uuid",
		GenerateNew: true,
	}

	if req.Ticket != "new-ticket" {
		t.Errorf("Expected Ticket 'new-ticket', got '%s'", req.Ticket)
	}
	if req.Exp != 7200 {
		t.Errorf("Expected Exp 7200, got %d", req.Exp)
	}
	if req.UUID != "share-uuid" {
		t.Errorf("Expected UUID 'share-uuid', got '%s'", req.UUID)
	}
	if !req.GenerateNew {
		t.Error("Expected GenerateNew to be true")
	}
}

func TestTicketCreateResponse_Fields(t *testing.T) {
	resp := TicketCreateResponse{
		Ticket: "generated-ticket-code",
	}

	if resp.Ticket != "generated-ticket-code" {
		t.Errorf("Expected Ticket 'generated-ticket-code', got '%s'", resp.Ticket)
	}
}

func TestTicketDeleteRequest_Fields(t *testing.T) {
	req := TicketDeleteRequest{
		Ticket: "ticket-to-delete",
	}

	if req.Ticket != "ticket-to-delete" {
		t.Errorf("Expected Ticket 'ticket-to-delete', got '%s'", req.Ticket)
	}
}

func TestTicketValidateResponse_Fields(t *testing.T) {
	resp := TicketValidateResponse{
		TicketValid: true,
		TicketExp:   false,
		Args:        `{"user":"admin"}`,
	}

	if !resp.TicketValid {
		t.Error("Expected TicketValid to be true")
	}
	if resp.TicketExp {
		t.Error("Expected TicketExp to be false")
	}
	if resp.Args != `{"user":"admin"}` {
		t.Errorf("Expected Args '{\"user\":\"admin\"}', got '%s'", resp.Args)
	}
}

func TestTicketValidateResponse_Expired(t *testing.T) {
	resp := TicketValidateResponse{
		TicketValid: false,
		TicketExp:   true,
		Args:        "",
	}

	if resp.TicketValid {
		t.Error("Expected TicketValid to be false")
	}
	if !resp.TicketExp {
		t.Error("Expected TicketExp to be true")
	}
}

func TestTicketListResponse_Fields(t *testing.T) {
	tickets := []Ticket{
		{ID: 1, UUID: "uuid-1", Ticket: "ticket-1"},
		{ID: 2, UUID: "uuid-2", Ticket: "ticket-2"},
	}

	resp := TicketListResponse{
		List:    tickets,
		Total:   2,
		Current: 1,
		Size:    10,
	}

	if len(resp.List) != 2 {
		t.Errorf("Expected 2 tickets, got %d", len(resp.List))
	}
	if resp.Total != 2 {
		t.Errorf("Expected Total 2, got %d", resp.Total)
	}
	if resp.Current != 1 {
		t.Errorf("Expected Current 1, got %d", resp.Current)
	}
	if resp.Size != 10 {
		t.Errorf("Expected Size 10, got %d", resp.Size)
	}
}

func TestTicketListResponse_Empty(t *testing.T) {
	resp := TicketListResponse{
		List:    []Ticket{},
		Total:   0,
		Current: 1,
		Size:    10,
	}

	if len(resp.List) != 0 {
		t.Errorf("Expected 0 tickets, got %d", len(resp.List))
	}
	if resp.Total != 0 {
		t.Errorf("Expected Total 0, got %d", resp.Total)
	}
}
