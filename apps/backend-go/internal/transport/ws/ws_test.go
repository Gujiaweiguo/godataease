package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startHub(t *testing.T) *Hub {
	t.Helper()
	h := NewHub()
	go h.Run()
	return h
}

// --- Existing tests (updated to testify) ---

func TestNewHub(t *testing.T) {
	h := NewHub()
	require.NotNil(t, h, "Expected Hub instance, got nil")
	assert.NotNil(t, h.clients)
	assert.NotNil(t, h.broadcast)
	assert.NotNil(t, h.register)
	assert.NotNil(t, h.unregister)
}

func TestHub_ClientCount(t *testing.T) {
	h := NewHub()
	assert.Equal(t, 0, h.ClientCount())
}

func TestClient_Fields(t *testing.T) {
	h := NewHub()
	c := &Client{ID: "client-1", UserID: 1, Conn: nil, Send: make(chan []byte, 256), Hub: h}
	assert.Equal(t, "client-1", c.ID)
	assert.Equal(t, uint64(1), c.UserID)
	assert.Same(t, h, c.Hub)
}

func TestHub_Broadcast_NoRun(t *testing.T) {
	h := NewHub()
	h.Broadcast([]byte("test message"))
}

func TestHub_SendToUser_NoClients(t *testing.T) {
	h := NewHub()
	h.SendToUser(1, []byte("test message"))
}

// --- Hub.Run + Register + Unregister ---

func TestHub_RegisterIncreasesClientCount(t *testing.T) {
	h := startHub(t)
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	h.Register(c)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, h.ClientCount())
}

func TestHub_UnregisterDecreasesClientCount(t *testing.T) {
	h := startHub(t)
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	h.Register(c)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, h.ClientCount())

	h.Unregister(c)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 0, h.ClientCount())
}

func TestHub_UnregisterClosesSendChannel(t *testing.T) {
	h := startHub(t)
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	h.Unregister(c)
	time.Sleep(10 * time.Millisecond)

	_, ok := <-c.Send
	assert.False(t, ok, "Send channel should be closed after unregister")
}

func TestHub_RegisterDuplicateClient_Overwrites(t *testing.T) {
	h := startHub(t)
	c1 := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	c2 := &Client{ID: "c1", UserID: 2, Send: make(chan []byte, 256), Hub: h}

	h.Register(c1)
	time.Sleep(10 * time.Millisecond)
	h.Register(c2)
	time.Sleep(10 * time.Millisecond)

	assert.Equal(t, 1, h.ClientCount())
}

func TestHub_UnregisterNonExistent_NoPanic(t *testing.T) {
	h := startHub(t)
	c := &Client{ID: "ghost", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	require.NotPanics(t, func() {
		h.Unregister(c)
		time.Sleep(10 * time.Millisecond)
	})
	assert.Equal(t, 0, h.ClientCount())
}

// --- Broadcast via Run() ---

func TestHub_BroadcastSendsToAllClients(t *testing.T) {
	h := startHub(t)
	c1 := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	c2 := &Client{ID: "c2", UserID: 2, Send: make(chan []byte, 256), Hub: h}
	h.Register(c1)
	h.Register(c2)
	time.Sleep(10 * time.Millisecond)

	h.Broadcast([]byte("hello"))

	select {
	case msg := <-c1.Send:
		assert.Equal(t, []byte("hello"), msg)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for c1 to receive broadcast")
	}
	select {
	case msg := <-c2.Send:
		assert.Equal(t, []byte("hello"), msg)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for c2 to receive broadcast")
	}
}

func TestHub_BroadcastRemovesFullChannelClient(t *testing.T) {
	h := startHub(t)
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 1), Hub: h}
	c.Send <- []byte("filler")
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	h.Broadcast([]byte("overflow"))
	time.Sleep(50 * time.Millisecond)

	_, ok := <-c.Send
	if ok {
		_, ok = <-c.Send
	}
	assert.False(t, ok, "Send channel should be closed after broadcast removes full client")
}

// --- SendToUser (matching path) ---

func TestHub_SendToUser_MatchesUserID(t *testing.T) {
	h := startHub(t)
	c1 := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	c2 := &Client{ID: "c2", UserID: 2, Send: make(chan []byte, 256), Hub: h}
	h.Register(c1)
	h.Register(c2)
	time.Sleep(10 * time.Millisecond)

	h.SendToUser(1, []byte("for-user-1"))

	select {
	case msg := <-c1.Send:
		assert.Equal(t, []byte("for-user-1"), msg)
	case <-time.After(time.Second):
		t.Fatal("timeout: c1 should receive message for user 1")
	}

	select {
	case <-c2.Send:
		t.Fatal("c2 should NOT receive message for user 1")
	default:
	}
}

func TestHub_SendToUser_MultipleClientsSameUser(t *testing.T) {
	h := startHub(t)
	c1 := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	c2 := &Client{ID: "c2", UserID: 1, Send: make(chan []byte, 256), Hub: h}
	h.Register(c1)
	h.Register(c2)
	time.Sleep(10 * time.Millisecond)

	h.SendToUser(1, []byte("multi"))

	select {
	case msg := <-c1.Send:
		assert.Equal(t, []byte("multi"), msg)
	case <-time.After(time.Second):
		t.Fatal("timeout: c1 should receive message")
	}
	select {
	case msg := <-c2.Send:
		assert.Equal(t, []byte("multi"), msg)
	case <-time.After(time.Second):
		t.Fatal("timeout: c2 should receive message")
	}
}

func TestHub_SendToUser_DropsOnFullChannel(t *testing.T) {
	h := startHub(t)
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 1), Hub: h}
	c.Send <- []byte("filler")
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		h.SendToUser(1, []byte("overflow"))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("SendToUser blocked on full channel")
	}
}

// --- ReadPump / WritePump ---

func TestClient_ReadPump_UnregistersOnError(t *testing.T) {
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	h := startHub(t)
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h, Conn: conn}
	h.Register(c)
	time.Sleep(10 * time.Millisecond)
	assert.Equal(t, 1, h.ClientCount())

	done := make(chan struct{})
	go func() {
		c.ReadPump()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ReadPump should have returned after connection close")
	}

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, h.ClientCount(), "client should be unregistered after ReadPump error")
}

func TestClient_WritePump_SendsMessages(t *testing.T) {
	upgrader := websocket.Upgrader{}
	received := make(chan []byte, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_, msg, err := conn.ReadMessage()
		if err == nil {
			received <- msg
		}
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	h := NewHub()
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h, Conn: conn}

	go c.WritePump()
	c.Send <- []byte("hello")

	select {
	case msg := <-received:
		assert.Equal(t, []byte("hello"), msg)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for message via WritePump")
	}

	close(c.Send)
}

func TestClient_WritePump_StopsOnConnError(t *testing.T) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	conn.Close()

	h := NewHub()
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h, Conn: conn}

	done := make(chan struct{})
	go func() {
		c.WritePump()
		close(done)
	}()

	c.Send <- []byte("fail")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("WritePump should have returned after write error")
	}
}

func TestClient_ReadPump_ReadsMultipleMessages(t *testing.T) {
	upgrader := websocket.Upgrader{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.WriteMessage(websocket.TextMessage, []byte("msg1"))
		_ = conn.WriteMessage(websocket.TextMessage, []byte("msg2"))
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	h := startHub(t)
	c := &Client{ID: "c1", UserID: 1, Send: make(chan []byte, 256), Hub: h, Conn: conn}
	h.Register(c)
	time.Sleep(10 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		c.ReadPump()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("ReadPump should have returned")
	}

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 0, h.ClientCount())
}
