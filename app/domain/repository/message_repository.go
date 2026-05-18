package repository

import (
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
)

type MessageRepository interface {
	CreateMessage(senderID, receiverID uint, content string, msgType model.MessageType, isAI bool) (*model.Message, error)
	GetMessages(userID, peerID uint, page, pageSize int) ([]*model.Message, error)
	GetConversations(userID uint) ([]*model.ChatSession, error)
	UpdateSession(userID, peerID, lastMsgID uint) error
	IncrementUnread(userID, peerID uint) error
	ClearUnread(userID, peerID uint) error
	GetUnreadMessages(userID uint) ([]*model.Message, error)
}
