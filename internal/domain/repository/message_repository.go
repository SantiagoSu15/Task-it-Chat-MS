package repository

import (
	"context"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

type MessageRepository interface {
	Save(ctx context.Context, chatID string, message model.Message) (model.Message, bool, error)
	FindByChatID(ctx context.Context, chatID string) ([]model.Message, error)
	DeleteByChatID(ctx context.Context, chatID string) error
}
