package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
)

var ErrChatNotFound = errors.New("chat not found")

type ChatNotFoundError struct {
	ID string
}

func (e *ChatNotFoundError) Error() string {
	return fmt.Sprintf("%v: %s", ErrChatNotFound, e.ID)
}

func (e *ChatNotFoundError) Unwrap() error {
	return ErrChatNotFound
}

type ChatRepository interface {
	Save(ctx context.Context, chat model.Chat) (model.Chat, error)
	FindByID(ctx context.Context, id string) (model.Chat, error)
	// FindByParticipantID returns one chat for the participant, preserving the
	// current public contract even though the Redis index can contain several.
	FindByParticipantID(ctx context.Context, participantID int64) (model.Chat, error)
	FindByTaskID(ctx context.Context, taskID string) (model.Chat, error)
	Update(ctx context.Context, chat model.Chat) (model.Chat, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteByTaskID(ctx context.Context, taskID string) error
	Exists(ctx context.Context, id string) (bool, error)
}
