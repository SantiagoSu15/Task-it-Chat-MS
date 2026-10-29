package service

import (
	"context"
	"errors"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/outbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

var ErrChatResolutionNotDefined = errors.New("resolving a chat from message and sender is not defined")

type SendMessageService struct {
	repository repository.ChatRepository
	websocket  outbound.ChatWs
}

func NewSendMessageService(chatRepository repository.ChatRepository, chatWebsocket outbound.ChatWs) *SendMessageService {
	return &SendMessageService{
		repository: chatRepository,
		websocket:  chatWebsocket,
	}
}

func (s *SendMessageService) SendMessage(_ context.Context, _ string, _ int64) error {
	// The inbound contract does not include a chat or task identifier, so no
	// repository lookup or WebSocket delivery can be performed safely yet.
	return ErrChatResolutionNotDefined
}

var _ inbound.SendMessageUseCase = (*SendMessageService)(nil)
