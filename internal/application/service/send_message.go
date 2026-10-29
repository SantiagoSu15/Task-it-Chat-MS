package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/inbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/application/port/outbound"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

var (
	ErrInvalidChatID   = errors.New("chatId is required")
	ErrInvalidMessage  = errors.New("messageId is required")
	ErrEmptyContent    = errors.New("message content is required")
	ErrMessageTooLarge = errors.New("message content exceeds the maximum size")
)

const MaxMessageContentLength = 4096

type SendMessageService struct {
	repository        repository.ChatRepository
	messageRepository repository.MessageRepository
	websocket         outbound.ChatWs
}

func NewSendMessageService(chatRepository repository.ChatRepository, messageRepository repository.MessageRepository, chatWebsocket outbound.ChatWs) *SendMessageService {
	return &SendMessageService{
		repository:        chatRepository,
		messageRepository: messageRepository,
		websocket:         chatWebsocket,
	}
}

func (s *SendMessageService) SendMessage(ctx context.Context, chatID, messageID, content string, senderID int64) error {
	chatID = strings.TrimSpace(chatID)
	messageID = strings.TrimSpace(messageID)
	content = strings.TrimSpace(content)
	if chatID == "" {
		return ErrInvalidChatID
	}
	if messageID == "" {
		return ErrInvalidMessage
	}
	if content == "" {
		return ErrEmptyContent
	}
	if len([]rune(content)) > MaxMessageContentLength {
		return ErrMessageTooLarge
	}
	if senderID <= 0 {
		return errors.New("authenticated sender is required")
	}

	if s.repository == nil {
		return errors.New("chat repository is not configured")
	}
	if s.messageRepository == nil {
		return errors.New("message repository is not configured")
	}
	if s.websocket == nil {
		return errors.New("chat websocket is not configured")
	}
	if _, err := s.repository.FindByID(ctx, chatID); err != nil {
		return err
	}

	message := model.Message{
		MessageID: messageID,
		ChatID:    chatID,
		SenderID:  senderID,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}
	persisted, created, err := s.messageRepository.Save(ctx, chatID, message)
	if err != nil {
		return err
	}
	if !created {
		return nil
	}
	return s.websocket.SendMessage(ctx, persisted, chatID)
}

var _ inbound.SendMessageUseCase = (*SendMessageService)(nil)
