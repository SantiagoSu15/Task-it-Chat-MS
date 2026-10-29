package websocket

import (
	"context"
	"errors"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/outbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

var ErrHubNotConfigured = errors.New("WebSocket hub is not configured")

type ChatHub interface {
	Broadcast(ctx context.Context, message model.Message, chatID string) error
}

type ChatWebSocket struct {
	hub ChatHub
}

func NewChatWebSocket(hub ChatHub) *ChatWebSocket {
	return &ChatWebSocket{hub: hub}
}

func (w *ChatWebSocket) SendMessage(ctx context.Context, message model.Message, chatID string) error {
	if w.hub == nil {
		return ErrHubNotConfigured
	}
	return w.hub.Broadcast(ctx, message, chatID)
}

var _ outbound.ChatWs = (*ChatWebSocket)(nil)
