package outbound

import (
	"context"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

type ChatWs interface {
	SendMessage(ctx context.Context, message model.Message, chatID string) error
}
