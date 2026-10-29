package service

import (
	"context"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

type GetMessageHistoryService struct {
	repository        repository.ChatRepository
	messageRepository repository.MessageRepository
}

func NewGetMessageHistoryService(chatRepository repository.ChatRepository, messageRepositories ...repository.MessageRepository) *GetMessageHistoryService {
	var messageRepository repository.MessageRepository
	if len(messageRepositories) > 0 {
		messageRepository = messageRepositories[0]
	}
	return &GetMessageHistoryService{
		repository:        chatRepository,
		messageRepository: messageRepository,
	}
}

func (s *GetMessageHistoryService) GetMessageHistory(ctx context.Context, taskID string) ([]model.Message, error) {
	chat, err := s.repository.FindByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	if s.messageRepository == nil {
		return chat.GetMessages(), nil
	}
	return s.messageRepository.FindByChatID(ctx, chat.ChatID)
}

var _ inbound.GetMessageHistoryUseCase = (*GetMessageHistoryService)(nil)
var _ inbound.GetChatHistoryUseCase = (*GetMessageHistoryService)(nil)

func (s *GetMessageHistoryService) GetMessageHistoryByChatID(ctx context.Context, chatID string) ([]model.Message, error) {
	if _, err := s.repository.FindByID(ctx, chatID); err != nil {
		return nil, err
	}
	if s.messageRepository == nil {
		return []model.Message{}, nil
	}
	return s.messageRepository.FindByChatID(ctx, chatID)
}
