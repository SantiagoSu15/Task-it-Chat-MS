package inbound

import "context"

type SendMessageUseCase interface {
	SendMessage(ctx context.Context, chatID, messageID, content string, senderID int64) error
}
