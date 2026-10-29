package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	goredis "github.com/redis/go-redis/v9"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

const (
	chatKeyPrefix             = "chat:"
	participantIndexKeyPrefix = "chat:index:participant:"
	taskIndexKeyPrefix        = "chat:index:task:"
	messageDataKeyPrefix      = "chat:messages:"
	messageSetKeyPrefix       = "chat:message-order:"
)

var ErrMessageAlreadyExists = errors.New("message already exists")

type ChatRepository struct {
	client *goredis.Client
}

func NewChatRepository(client *goredis.Client) *ChatRepository {
	return &ChatRepository{client: client}
}

func (r *ChatRepository) Save(ctx context.Context, chat model.Chat) (model.Chat, error) {
	if strings.TrimSpace(chat.ChatID) == "" {
		return model.Chat{}, errors.New("chatId is required")
	}
	chat.Messages = nil
	payload, err := json.Marshal(chat)
	if err != nil {
		return model.Chat{}, err
	}
	indexKeys := chatIndexKeys(chat)
	args := make([]interface{}, 0, len(indexKeys)+1)
	args = append(args, payload)
	for _, key := range indexKeys {
		args = append(args, key)
	}
	result, err := createChatScript.Run(ctx, r.client, []string{chatKey(chat.ChatID)}, args...).Int()
	if err != nil {
		return model.Chat{}, err
	}
	if result == 0 {
		return r.FindByID(ctx, chat.ChatID)
	}
	return chat, nil
}

func (r *ChatRepository) FindByID(ctx context.Context, id string) (model.Chat, error) {
	payload, err := r.client.Get(ctx, chatKey(id)).Bytes()
	if errors.Is(err, goredis.Nil) {
		return model.Chat{}, &repository.ChatNotFoundError{ID: id}
	}
	if err != nil {
		return model.Chat{}, err
	}
	var chat model.Chat
	if err := json.Unmarshal(payload, &chat); err != nil {
		return model.Chat{}, err
	}
	return chat, nil
}

func (r *ChatRepository) FindByParticipantID(ctx context.Context, participantID int64) (model.Chat, error) {
	ids, err := r.client.SMembers(ctx, participantIndexKey(strconv.FormatInt(participantID, 10))).Result()
	if err != nil {
		return model.Chat{}, err
	}
	return r.findIndexedChat(ctx, ids, participantIndexKey(strconv.FormatInt(participantID, 10)))
}

func (r *ChatRepository) FindByTaskID(ctx context.Context, taskID string) (model.Chat, error) {
	ids, err := r.client.SMembers(ctx, taskIndexKey(taskID)).Result()
	if err != nil {
		return model.Chat{}, err
	}
	return r.findIndexedChat(ctx, ids, taskIndexKey(taskID))
}

func (r *ChatRepository) findIndexedChat(ctx context.Context, ids []string, indexKey string) (model.Chat, error) {
	for _, id := range ids {
		chat, err := r.FindByID(ctx, id)
		if err == nil {
			return chat, nil
		}
		if errors.Is(err, repository.ErrChatNotFound) {
			_ = r.client.SRem(ctx, indexKey, id).Err()
			continue
		}
		return model.Chat{}, err
	}
	return model.Chat{}, &repository.ChatNotFoundError{ID: indexKey}
}

func (r *ChatRepository) Update(ctx context.Context, chat model.Chat) (model.Chat, error) {
	current, err := r.FindByID(ctx, chat.ChatID)
	if err != nil {
		return model.Chat{}, err
	}
	chat.CreatedAt = current.CreatedAt
	chat.Messages = nil
	payload, err := json.Marshal(chat)
	if err != nil {
		return model.Chat{}, err
	}
	oldKeys := chatIndexKeys(current)
	newKeys := chatIndexKeys(chat)
	_, err = r.client.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		for _, key := range oldKeys {
			pipe.SRem(ctx, key, chat.ChatID)
		}
		pipe.Set(ctx, chatKey(chat.ChatID), payload, 0)
		for _, key := range newKeys {
			pipe.SAdd(ctx, key, chat.ChatID)
		}
		return nil
	})
	if err != nil {
		return model.Chat{}, err
	}
	return chat, nil
}

func (r *ChatRepository) DeleteByID(ctx context.Context, id string) error {
	chat, err := r.FindByID(ctx, id)
	if errors.Is(err, repository.ErrChatNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	keys := chatIndexKeys(chat)
	_, err = r.client.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
		pipe.Del(ctx, chatKey(id), messageDataKey(id), messageSetKey(id))
		for _, key := range keys {
			pipe.SRem(ctx, key, id)
		}
		return nil
	})
	return err
}

func (r *ChatRepository) DeleteByTaskID(ctx context.Context, taskID string) error {
	chat, err := r.FindByTaskID(ctx, taskID)
	if err != nil {
		if errors.Is(err, repository.ErrChatNotFound) {
			return nil
		}
		return err
	}
	return r.DeleteByID(ctx, chat.ChatID)
}

func (r *ChatRepository) Exists(ctx context.Context, id string) (bool, error) {
	exists, err := r.client.Exists(ctx, chatKey(id)).Result()
	return exists > 0, err
}

type MessageRepository struct {
	client *goredis.Client
}

func NewMessageRepository(client *goredis.Client) *MessageRepository {
	return &MessageRepository{client: client}
}

func (r *MessageRepository) Save(ctx context.Context, chatID string, message model.Message) (model.Message, bool, error) {
	if strings.TrimSpace(chatID) == "" {
		return model.Message{}, false, errors.New("chatId is required")
	}
	if strings.TrimSpace(message.MessageID) == "" {
		return model.Message{}, false, errors.New("messageId is required")
	}
	if message.ChatID != "" && message.ChatID != chatID {
		return model.Message{}, false, errors.New("message belongs to a different chat")
	}
	message.ChatID = chatID
	payload, err := json.Marshal(message)
	if err != nil {
		return model.Message{}, false, err
	}
	score := float64(message.Timestamp.UnixNano())
	result, err := saveMessageScript.Run(ctx, r.client, []string{
		chatKey(chatID), messageDataKey(chatID), messageSetKey(chatID),
	}, message.MessageID, payload, score).Int()
	if err != nil {
		return model.Message{}, false, err
	}
	if result == -1 {
		return model.Message{}, false, &repository.ChatNotFoundError{ID: chatID}
	}
	if result == 0 {
		existing, err := r.findOne(ctx, chatID, message.MessageID)
		return existing, false, err
	}
	return message, true, nil
}

func (r *MessageRepository) FindByChatID(ctx context.Context, chatID string) ([]model.Message, error) {
	exists, err := r.client.Exists(ctx, chatKey(chatID)).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, &repository.ChatNotFoundError{ID: chatID}
	}
	ids, err := r.client.ZRange(ctx, messageSetKey(chatID), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	values, err := r.client.HMGet(ctx, messageDataKey(chatID), ids...).Result()
	if err != nil {
		return nil, err
	}
	messages := make([]model.Message, 0, len(values))
	for _, value := range values {
		if value == nil {
			continue
		}
		var message model.Message
		if err := json.Unmarshal([]byte(value.(string)), &message); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (r *MessageRepository) findOne(ctx context.Context, chatID, messageID string) (model.Message, error) {
	value, err := r.client.HGet(ctx, messageDataKey(chatID), messageID).Result()
	if err != nil {
		return model.Message{}, err
	}
	var message model.Message
	err = json.Unmarshal([]byte(value), &message)
	return message, err
}

func (r *MessageRepository) DeleteByChatID(ctx context.Context, chatID string) error {
	return r.client.Del(ctx, messageDataKey(chatID), messageSetKey(chatID)).Err()
}

func chatKey(id string) string             { return chatKeyPrefix + id }
func participantIndexKey(id string) string { return participantIndexKeyPrefix + id }
func taskIndexKey(id string) string        { return taskIndexKeyPrefix + id }
func messageDataKey(id string) string      { return messageDataKeyPrefix + id }
func messageSetKey(id string) string       { return messageSetKeyPrefix + id }

func chatIndexKeys(chat model.Chat) []string {
	keys := make([]string, 0, 2)
	if participantID, err := strconv.ParseInt(chat.SellerID, 10, 64); err == nil {
		keys = append(keys, participantIndexKey(strconv.FormatInt(participantID, 10)))
	}
	if chat.BuyerID != nil {
		keys = append(keys, participantIndexKey(strconv.FormatInt(*chat.BuyerID, 10)))
	}
	if chat.TaskID != "" {
		keys = append(keys, taskIndexKey(chat.TaskID))
	}
	return unique(keys)
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

var createChatScript = goredis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
  return 0
end
redis.call("SET", KEYS[1], ARGV[1])
for i = 2, #ARGV do
  redis.call("SADD", ARGV[i], string.sub(KEYS[1], 6))
end
return 1
`)

var saveMessageScript = goredis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
  return -1
end
if redis.call("HEXISTS", KEYS[2], ARGV[1]) == 1 then
  return 0
end
redis.call("HSET", KEYS[2], ARGV[1], ARGV[2])
redis.call("ZADD", KEYS[3], ARGV[3], ARGV[1])
return 1
`)

var _ repository.ChatRepository = (*ChatRepository)(nil)
var _ repository.MessageRepository = (*MessageRepository)(nil)
