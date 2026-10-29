package inbound

import (
	"context"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

type GetChatHistoryUseCase interface {
	GetMessageHistoryByChatID(ctx context.Context, chatID string) ([]model.Message, error)
}
