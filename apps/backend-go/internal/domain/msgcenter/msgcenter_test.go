package msgcenter

import (
	"testing"
)

func TestMessage_Fields(t *testing.T) {
	msg := Message{
		ID:         "msg-1",
		Title:      "System Update",
		Content:    "New version available",
		Type:       "system",
		Level:      "info",
		Read:       false,
		CreateTime: 1700000000,
		ReadTime:   0,
	}

	if msg.ID != "msg-1" {
		t.Errorf("Expected ID 'msg-1', got '%s'", msg.ID)
	}
	if msg.Title != "System Update" {
		t.Errorf("Expected Title 'System Update', got '%s'", msg.Title)
	}
	if msg.Read {
		t.Error("Expected Read to be false")
	}
}

func TestListRequest_Fields(t *testing.T) {
	req := ListRequest{
		Current:    1,
		Size:       10,
		ReadStatus: "unread",
		Type:       "system",
	}

	if req.Current != 1 {
		t.Errorf("Expected Current 1, got %d", req.Current)
	}
	if req.ReadStatus != "unread" {
		t.Errorf("Expected ReadStatus 'unread', got '%s'", req.ReadStatus)
	}
}

func TestListResponse_Fields(t *testing.T) {
	msgs := []Message{
		{ID: "msg-1", Title: "Message 1"},
		{ID: "msg-2", Title: "Message 2"},
	}

	resp := ListResponse{
		List:    msgs,
		Total:   2,
		Current: 1,
		Size:    10,
	}

	if len(resp.List) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(resp.List))
	}
	if resp.Total != 2 {
		t.Errorf("Expected Total 2, got %d", resp.Total)
	}
}

func TestReadRequest_Fields(t *testing.T) {
	req := ReadRequest{ID: "msg-1"}
	if req.ID != "msg-1" {
		t.Errorf("Expected ID 'msg-1', got '%s'", req.ID)
	}
}

func TestReadBatchRequest_Fields(t *testing.T) {
	req := ReadBatchRequest{
		IDs: []string{"msg-1", "msg-2", "msg-3"},
	}

	if len(req.IDs) != 3 {
		t.Errorf("Expected 3 IDs, got %d", len(req.IDs))
	}
}

func TestReadResponse_Fields(t *testing.T) {
	resp := ReadResponse{
		Success:     true,
		AlreadyRead: false,
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.AlreadyRead {
		t.Error("Expected AlreadyRead to be false")
	}
}

func TestReadBatchResponse_Fields(t *testing.T) {
	resp := ReadBatchResponse{
		Success: true,
		Updated: 5,
	}

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Updated != 5 {
		t.Errorf("Expected Updated 5, got %d", resp.Updated)
	}
}
