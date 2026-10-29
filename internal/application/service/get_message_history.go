package service

import (
	"context"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

type GetMessageHistoryService struct {
	repository repository.ChatRepository
}

func NewGetMessageHistoryService(chatRepository repository.ChatRepository) *GetMessageHistoryService {
	return &GetMessageHistoryService{repository: chatRepository}
}

func (s *GetMessageHistoryService) GetMessageHistory(ctx context.Context, taskID string) ([]model.Message, error) {
	chat, err := s.repository.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return chat.GetMessages(), nil
}

var _ inbound.GetMessageHistoryUseCase = (*GetMessageHistoryService)(nil)
