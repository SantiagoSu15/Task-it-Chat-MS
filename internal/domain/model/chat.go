package model

import "time"

type Chat struct {
	ChatID       string    `json:"chatId"`
	SellerID     string    `json:"sellerId"`
	BuyerID      *int64    `json:"buyerId"`
	Participants []User    `json:"participants"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"createdAt"`
	Messages     []Message `json:"messages"`
	TaskID       string    `json:"taskId"`
}

func (c *Chat) AddMessage(message Message) {
	c.Messages = append(c.Messages, message)
}

func (c Chat) GetMessages() []Message {
	if len(c.Messages) == 0 {
		return []Message{}
	}

	messages := make([]Message, len(c.Messages))
	copy(messages, c.Messages)
	return messages
}
