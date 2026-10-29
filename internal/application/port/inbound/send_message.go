package inbound

import "context"

type SendMessageUseCase interface {
	SendMessage(ctx context.Context, message string, senderID int64) error
}
