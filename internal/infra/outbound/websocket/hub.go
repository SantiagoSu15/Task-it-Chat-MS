package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

const clientBufferSize = 32

var (
	ErrInvalidRoom  = errors.New("chatId is required")
	ErrClientClosed = errors.New("client is closed")
)

type Client struct {
	send  chan []byte
	done  chan struct{}
	once  sync.Once
	mu    sync.RWMutex
	rooms map[string]struct{}
}

func (c *Client) Messages() <-chan []byte { return c.send }
func (c *Client) Done() <-chan struct{}   { return c.done }
func (c *Client) Enqueue(payload []byte) bool {
	select {
	case <-c.Done():
		return false
	case c.send <- payload:
		return true
	default:
		return false
	}
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	rooms   map[string]map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		rooms:   make(map[string]map[*Client]struct{}),
	}
}

func (h *Hub) Register() *Client {
	client := &Client{
		send:  make(chan []byte, clientBufferSize),
		done:  make(chan struct{}),
		rooms: make(map[string]struct{}),
	}
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()
	return client
}

func (h *Hub) Unregister(client *Client) {
	if client == nil {
		return
	}
	client.once.Do(func() {
		h.mu.Lock()
		delete(h.clients, client)
		client.mu.RLock()
		rooms := make([]string, 0, len(client.rooms))
		for room := range client.rooms {
			rooms = append(rooms, room)
		}
		client.mu.RUnlock()
		for _, room := range rooms {
			if members := h.rooms[room]; members != nil {
				delete(members, client)
				if len(members) == 0 {
					delete(h.rooms, room)
				}
			}
		}
		h.mu.Unlock()
		client.mu.Lock()
		client.rooms = make(map[string]struct{})
		client.mu.Unlock()
		close(client.done)
	})
}

func (h *Hub) Join(client *Client, chatID string) error {
	if client == nil {
		return ErrClientClosed
	}
	if chatID == "" {
		return ErrInvalidRoom
	}
	select {
	case <-client.Done():
		return ErrClientClosed
	default:
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[client]; !ok {
		return ErrClientClosed
	}
	if h.rooms[chatID] == nil {
		h.rooms[chatID] = make(map[*Client]struct{})
	}
	h.rooms[chatID][client] = struct{}{}
	client.mu.Lock()
	client.rooms[chatID] = struct{}{}
	client.mu.Unlock()
	return nil
}

func (h *Hub) InRoom(client *Client, chatID string) bool {
	if client == nil {
		return false
	}
	client.mu.RLock()
	_, ok := client.rooms[chatID]
	client.mu.RUnlock()
	return ok
}

func (h *Hub) Broadcast(ctx context.Context, message model.Message, chatID string) error {
	if chatID == "" {
		return ErrInvalidRoom
	}
	payload, err := json.Marshal(struct {
		Type string        `json:"type"`
		Data model.Message `json:"data"`
	}{Type: "chat:newMessage", Data: message})
	if err != nil {
		return err
	}

	h.mu.RLock()
	members := make([]*Client, 0, len(h.rooms[chatID]))
	for client := range h.rooms[chatID] {
		members = append(members, client)
	}
	h.mu.RUnlock()

	for _, client := range members {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-client.Done():
			continue
		default:
		}
		if !client.Enqueue(payload) {
			// A full queue is a slow client. Remove it without blocking the room.
			h.Unregister(client)
		}
	}
	return nil
}

func (h *Hub) Close() {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	for _, client := range clients {
		h.Unregister(client)
	}
}
