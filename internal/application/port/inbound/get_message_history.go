package inbound

import (
	"context"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

type GetMessageHistoryUseCase interface {
	GetMessageHistory(ctx context.Context, taskID string) ([]model.Message, error)
}
