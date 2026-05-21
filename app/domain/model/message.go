package model

import (
	"github.com/jinzhu/gorm"
)

type MessageType int32

const (
	MessageType_TEXT        MessageType = 0
	MessageType_IMAGE       MessageType = 1
	MessageType_AI_RESPONSE MessageType = 2
)

type Message struct {
	gorm.Model
	SenderID   uint        `gorm:"column:sender_id;index:idx_sender_receiver" json:"sender_id"`
	ReceiverID uint        `gorm:"column:receiver_id;index:idx_sender_receiver" json:"receiver_id"`
	Content    string      `gorm:"column:content;type:text" json:"content"`
	Type       MessageType `gorm:"column:type;default:0" json:"type"`
	IsAI       bool        `gorm:"column:is_ai;default:false" json:"is_ai"`
}

func (Message) TableName() string {
	return "messages"
}

type ChatSession struct {
	gorm.Model
	UserID      uint `gorm:"column:user_id;unique_index:idx_user_peer;not null" json:"user_id"`
	PeerID      uint `gorm:"column:peer_id;unique_index:idx_user_peer;not null" json:"peer_id"`
	LastMsgID   uint `gorm:"column:last_msg_id" json:"last_msg_id"`
	UnreadCount int  `gorm:"column:unread_count;default:0" json:"unread_count"`
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

type ChatMessageDTO struct {
	ID         uint64      `json:"id"`
	SenderID   uint64      `json:"sender_id"`
	ReceiverID uint64      `json:"receiver_id"`
	Content    string      `json:"content"`
	Timestamp  int64       `json:"timestamp"`
	Type       MessageType `json:"type"`
	IsAI       bool        `json:"is_ai"`
}

func (m *Message) ToDTO() *ChatMessageDTO {
	return &ChatMessageDTO{
		ID:         uint64(m.ID),
		SenderID:   uint64(m.SenderID),
		ReceiverID: uint64(m.ReceiverID),
		Content:    m.Content,
		Timestamp:  m.CreatedAt.Unix(),
		Type:       m.Type,
		IsAI:       m.IsAI,
	}
}

type ChatSessionDTO struct {
	PeerID      uint64 `json:"peer_id"`
	LastMessage string `json:"last_message"`
	LastTime    int64  `json:"last_time"`
	UnreadCount int    `json:"unread_count"`
}
