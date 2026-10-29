package redis

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"

	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/model"
	"github.com/SantiagoSu15/Task-it-Chat-MS/internal/domain/repository"
)

func TestRedisChatAndMessagePersistence(t *testing.T) {
	addr := os.Getenv("REDIS_TEST_ADDR")
	if addr == "" {
		addr = os.Getenv("REDIS_ADDR")
	}
	if addr == "" {
		addr = "localhost:6379"
	}
	client := goredis.NewClient(&goredis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis integration test skipped: %v", err)
	}
	defer client.Close()

	prefix := "chat-test-" + time.Now().Format("20060102150405.000000000")
	chat := model.Chat{
		ChatID:    prefix,
		SellerID:  "101",
		BuyerID:   int64Ptr(202),
		TaskID:    prefix + "-task",
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
		Active:    true,
	}
	chatRepository := NewChatRepository(client)
	messageRepository := NewMessageRepository(client)
	t.Cleanup(func() {
		_ = chatRepository.DeleteByID(context.Background(), chat.ChatID)
	})

	saved, err := chatRepository.Save(ctx, chat)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := chatRepository.Save(ctx, chat)
	if err != nil {
		t.Fatal(err)
	}
	if repeated.ChatID != saved.ChatID {
		t.Fatalf("idempotent save returned %q, want %q", repeated.ChatID, saved.ChatID)
	}
	if _, err := chatRepository.FindByParticipantID(ctx, 202); err != nil {
		t.Fatal(err)
	}
	if _, err := chatRepository.FindByTaskID(ctx, chat.TaskID); err != nil {
		t.Fatal(err)
	}

	first := model.Message{MessageID: "first", ChatID: chat.ChatID, SenderID: 101, Content: "one", Timestamp: chat.CreatedAt.Add(time.Second)}
	second := model.Message{MessageID: "second", ChatID: chat.ChatID, SenderID: 202, Content: "two", Timestamp: chat.CreatedAt.Add(2 * time.Second)}
	if _, created, err := messageRepository.Save(ctx, chat.ChatID, second); err != nil || !created {
		t.Fatalf("save second: created=%v err=%v", created, err)
	}
	if _, created, err := messageRepository.Save(ctx, chat.ChatID, first); err != nil || !created {
		t.Fatalf("save first: created=%v err=%v", created, err)
	}
	if _, created, err := messageRepository.Save(ctx, chat.ChatID, first); err != nil || created {
		t.Fatalf("duplicate message: created=%v err=%v", created, err)
	}
	if ttl, err := client.TTL(ctx, chatKey(chat.ChatID)).Result(); err != nil || ttl != -1 {
		t.Fatalf("chat TTL = %v, err=%v; expected no expiration", ttl, err)
	}
	if ttl, err := client.TTL(ctx, messageDataKey(chat.ChatID)).Result(); err != nil || ttl != -1 {
		t.Fatalf("message TTL = %v, err=%v; expected no expiration", ttl, err)
	}
	messages, err := messageRepository.FindByChatID(ctx, chat.ChatID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[0].MessageID != "first" || messages[1].MessageID != "second" {
		t.Fatalf("unexpected chronological messages: %+v", messages)
	}

	updated := chat
	updated.TaskID = chat.TaskID + "-updated"
	updated.CreatedAt = time.Time{}
	if _, err := chatRepository.Update(ctx, updated); err != nil {
		t.Fatal(err)
	}
	stored, err := chatRepository.FindByID(ctx, chat.ChatID)
	if err != nil {
		t.Fatal(err)
	}
	if !stored.CreatedAt.Equal(chat.CreatedAt) {
		t.Fatalf("createdAt changed from %v to %v", chat.CreatedAt, stored.CreatedAt)
	}
	if _, err := chatRepository.FindByTaskID(ctx, chat.TaskID); !isNotFound(err) {
		t.Fatalf("old task index still resolves: %v", err)
	}
	if err := chatRepository.DeleteByID(ctx, chat.ChatID); err != nil {
		t.Fatal(err)
	}
	if _, err := chatRepository.FindByID(ctx, chat.ChatID); !isNotFound(err) {
		t.Fatalf("chat still exists after deletion: %v", err)
	}
	if _, err := messageRepository.FindByChatID(ctx, chat.ChatID); !isNotFound(err) {
		t.Fatalf("message history still exists after deletion: %v", err)
	}
}

func isNotFound(err error) bool {
	return err != nil && errors.Is(err, repository.ErrChatNotFound)
}

func int64Ptr(value int64) *int64 {
	return &value
}
