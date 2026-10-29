package model

import "time"

type Message struct {
	MessageID string    `json:"messageId"`
	SenderID  int64     `json:"senderId"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
