package ws

import (
	"testing"
)

func TestNewHub(t *testing.T) {
	h := NewHub()
	if h == nil {
		t.Fatal("Expected Hub instance, got nil")
	}
	if h.clients == nil {
		t.Error("Expected clients map to be initialized")
	}
	if h.broadcast == nil {
		t.Error("Expected broadcast channel to be initialized")
	}
	if h.register == nil {
		t.Error("Expected register channel to be initialized")
	}
	if h.unregister == nil {
		t.Error("Expected unregister channel to be initialized")
	}
}

func TestHub_ClientCount(t *testing.T) {
	h := NewHub()
	if h.ClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", h.ClientCount())
	}
}

func TestClient_Fields(t *testing.T) {
	h := NewHub()
	c := &Client{
		ID:     "client-1",
		UserID: 1,
		Conn:   nil,
		Send:   make(chan []byte, 256),
		Hub:    h,
	}

	if c.ID != "client-1" {
		t.Errorf("Expected ID 'client-1', got '%s'", c.ID)
	}
	if c.UserID != 1 {
		t.Errorf("Expected UserID 1, got %d", c.UserID)
	}
	if c.Hub != h {
		t.Error("Expected Hub to be set")
	}
}

func TestHub_Broadcast(t *testing.T) {
	h := NewHub()
	h.Broadcast([]byte("test message"))
}

func TestHub_SendToUser_NoClients(t *testing.T) {
	h := NewHub()
	h.SendToUser(1, []byte("test message"))
}
