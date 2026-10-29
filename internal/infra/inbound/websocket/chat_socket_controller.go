package websocket

import (
	"context"
	"net/http"

	"github.com/coder/websocket"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
)

type ChatSocketController struct {
	sendMessage inbound.SendMessageUseCase
}

func NewChatSocketController(sendMessage inbound.SendMessageUseCase) *ChatSocketController {
	return &ChatSocketController{sendMessage: sendMessage}
}

// ServeHTTP accepts a WebSocket connection as the integration point for a
// future room protocol. Room joining and event decoding are intentionally not
// implemented until that protocol is defined.
func (c *ChatSocketController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "chat WebSocket scaffold")
}

type SendMessageEvent struct {
	Message  string `json:"message"`
	SenderID int64  `json:"senderId"`
}

func (c *ChatSocketController) HandleSendMessage(ctx context.Context, event SendMessageEvent) error {
	return c.sendMessage.SendMessage(ctx, event.Message, event.SenderID)
}
