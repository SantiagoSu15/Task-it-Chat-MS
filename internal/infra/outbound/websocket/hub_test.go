package websocket

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

func TestHubBroadcastsOnlyToRoomMembers(t *testing.T) {
	hub := NewHub()
	first, second, other := hub.Register(), hub.Register(), hub.Register()
	if err := hub.Join(first, "chat-1"); err != nil {
		t.Fatal(err)
	}
	if err := hub.Join(second, "chat-1"); err != nil {
		t.Fatal(err)
	}
	if err := hub.Join(other, "chat-2"); err != nil {
		t.Fatal(err)
	}

	if err := hub.Broadcast(context.Background(), model.Message{ChatID: "chat-1", MessageID: "m1"}, "chat-1"); err != nil {
		t.Fatal(err)
	}
	assertMessage(t, first)
	assertMessage(t, second)
	select {
	case <-other.Messages():
		t.Fatal("client in another room received the message")
	default:
	}
}

func TestHubJoinIsIdempotentAndUnregistersAllRooms(t *testing.T) {
	hub := NewHub()
	client := hub.Register()
	if err := hub.Join(client, "chat-1"); err != nil {
		t.Fatal(err)
	}
	if err := hub.Join(client, "chat-1"); err != nil {
		t.Fatal(err)
	}
	if !hub.InRoom(client, "chat-1") {
		t.Fatal("client was not joined")
	}
	hub.Unregister(client)
	if hub.InRoom(client, "chat-1") {
		t.Fatal("client remained in room after unregister")
	}
	if err := hub.Join(client, "chat-1"); err != ErrClientClosed {
		t.Fatalf("expected ErrClientClosed, got %v", err)
	}
}

func TestHubConcurrentAccessAndSlowClient(t *testing.T) {
	hub := NewHub()
	fast, slow := hub.Register(), hub.Register()
	if err := hub.Join(fast, "chat-1"); err != nil {
		t.Fatal(err)
	}
	if err := hub.Join(slow, "chat-1"); err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = hub.Broadcast(context.Background(), model.Message{ChatID: "chat-1"}, "chat-1")
		}()
	}
	wg.Wait()
	select {
	case <-slow.Done():
	case <-time.After(time.Second):
		t.Fatal("slow client was not disconnected")
	}
}

func assertMessage(t *testing.T, client *Client) {
	t.Helper()
	select {
	case payload := <-client.Messages():
		var event struct {
			Type string        `json:"type"`
			Data model.Message `json:"data"`
		}
		if err := json.Unmarshal(payload, &event); err != nil {
			t.Fatal(err)
		}
		if event.Type != "chat:newMessage" {
			t.Fatalf("unexpected event type %q", event.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("message was not delivered")
	}
}
