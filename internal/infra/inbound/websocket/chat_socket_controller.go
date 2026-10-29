package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/service"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
	wsout "github.com/SantiagoSu15/Task-it-Chat-MS/internal/infra/outbound/websocket"
)

const (
	maxWebSocketMessageSize = 64 * 1024
	writeTimeout            = 5 * time.Second
)

type ChatSocketController struct {
	sendMessage inbound.SendMessageUseCase
	history     inbound.GetChatHistoryUseCase
	hub         *wsout.Hub
}

func NewChatSocketController(sendMessage inbound.SendMessageUseCase, history inbound.GetChatHistoryUseCase, hub *wsout.Hub) *ChatSocketController {
	return &ChatSocketController{sendMessage: sendMessage, history: history, hub: hub}
}

func (c *ChatSocketController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if c.hub == nil || c.sendMessage == nil || c.history == nil {
		http.Error(w, "WebSocket controller is not configured", http.StatusInternalServerError)
		return
	}
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "chat WebSocket closed")

	client := c.hub.Register()
	defer c.hub.Unregister(client)

	ctx := r.Context()
	ctx = withAuthenticatedSender(ctx, r.Header.Get("X-User-ID"))
	go c.writeLoop(ctx, conn, client)
	conn.SetReadLimit(maxWebSocketMessageSize)

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		var event Event
		if err := json.Unmarshal(data, &event); err != nil {
			c.writeError(client, "invalid JSON event")
			continue
		}
		if err := c.handleEvent(ctx, client, event); err != nil {
			c.writeError(client, protocolError(err))
		}
	}
}

type Event struct {
	Type      string `json:"type"`
	ChatID    string `json:"chatId"`
	Content   string `json:"content"`
	MessageID string `json:"messageId"`
}

func (c *ChatSocketController) handleEvent(ctx context.Context, client *wsout.Client, event Event) error {
	switch event.Type {
	case "chat:EnterRoom":
		chatID := strings.TrimSpace(event.ChatID)
		if chatID == "" {
			return wsout.ErrInvalidRoom
		}
		history, err := c.history.GetMessageHistoryByChatID(ctx, chatID)
		if err != nil {
			return err
		}
		if err := c.hub.Join(client, chatID); err != nil {
			return err
		}
		payload, err := json.Marshal(struct {
			Type string          `json:"type"`
			Data []model.Message `json:"data"`
		}{Type: "chat:history", Data: history})
		if err != nil {
			return err
		}
		if !client.Enqueue(payload) {
			c.hub.Unregister(client)
		}
		return nil
	case "chat:sendMessage":
		if !c.hub.InRoom(client, strings.TrimSpace(event.ChatID)) {
			return errors.New("client must enter the room before sending")
		}
		senderID, ok := authenticatedSender(ctx)
		if !ok {
			return errors.New("authenticated sender is required; configure authentication middleware")
		}
		return c.sendMessage.SendMessage(ctx, event.ChatID, event.MessageID, event.Content, senderID)
	default:
		return errors.New("unsupported event type")
	}
}

func (c *ChatSocketController) writeLoop(ctx context.Context, conn *websocket.Conn, client *wsout.Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-client.Done():
			return
		case payload := <-client.Messages():
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := conn.Write(writeCtx, websocket.MessageText, payload)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

func (c *ChatSocketController) writeError(client *wsout.Client, message string) {
	payload, _ := json.Marshal(map[string]string{"type": "error", "error": message})
	if !client.Enqueue(payload) {
		c.hub.Unregister(client)
	}
}

type senderContextKey struct{}

func withAuthenticatedSender(ctx context.Context, value string) context.Context {
	id := strings.TrimSpace(value)
	if id == "" {
		return ctx
	}
	var senderID int64
	if _, err := fmt.Sscan(id, &senderID); err != nil || senderID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, senderContextKey{}, senderID)
}

func authenticatedSender(ctx context.Context) (int64, bool) {
	senderID, ok := ctx.Value(senderContextKey{}).(int64)
	return senderID, ok && senderID > 0
}

func protocolError(err error) string {
	switch {
	case errors.Is(err, service.ErrInvalidChatID), errors.Is(err, service.ErrInvalidMessage),
		errors.Is(err, service.ErrEmptyContent), errors.Is(err, service.ErrMessageTooLarge):
		return err.Error()
	default:
		return err.Error()
	}
}
