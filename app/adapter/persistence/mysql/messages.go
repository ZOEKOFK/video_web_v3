package mysql

import (
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
	"github.com/jinzhu/gorm"
)

type MessageRepositoryImpl struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) repository.MessageRepository {
	return &MessageRepositoryImpl{db: db}
}

func (r *MessageRepositoryImpl) CreateMessage(senderID, receiverID uint, content string, msgType model.MessageType, isAI bool) (*model.Message, error) {
	msg := &model.Message{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Type:       msgType,
		IsAI:       isAI,
	}
	if err := r.db.Create(msg).Error; err != nil {
		return nil, err
	}

	session := &model.ChatSession{
		UserID:      receiverID,
		PeerID:      senderID,
		LastMsgID:   msg.ID,
		UnreadCount: 1,
	}
	if err := r.db.Where(model.ChatSession{UserID: receiverID, PeerID: senderID}).
		Assign(session).
		FirstOrCreate(&model.ChatSession{}).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&model.ChatSession{}).
		Where("user_id = ? AND peer_id = ?", receiverID, senderID).
		UpdateColumn(map[string]interface{}{
			"last_msg_id":   msg.ID,
			"unread_count":  gorm.Expr("unread_count + 1"),
		}).Error; err != nil {
		return nil, err
	}

	return msg, nil
}

func (r *MessageRepositoryImpl) GetMessages(userID, peerID uint, page, pageSize int) ([]*model.Message, error) {
	var messages []*model.Message
	offset := (page - 1) * pageSize
	err := r.db.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		userID, peerID, peerID, userID).
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&messages).Error
	return messages, err
}

func (r *MessageRepositoryImpl) GetConversations(userID uint) ([]*model.ChatSession, error) {
	var sessions []*model.ChatSession
	err := r.db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&sessions).Error
	return sessions, err
}

func (r *MessageRepositoryImpl) UpdateSession(userID, peerID, lastMsgID uint) error {
	return r.db.Model(&model.ChatSession{}).
		Where("user_id = ? AND peer_id = ?", userID, peerID).
		UpdateColumns(map[string]interface{}{
			"last_msg_id": lastMsgID,
		}).Error
}

func (r *MessageRepositoryImpl) IncrementUnread(userID, peerID uint) error {
	return r.db.Model(&model.ChatSession{}).
		Where("user_id = ? AND peer_id = ?", userID, peerID).
		UpdateColumn("unread_count", gorm.Expr("unread_count + 1")).Error
}

func (r *MessageRepositoryImpl) ClearUnread(userID, peerID uint) error {
	return r.db.Model(&model.ChatSession{}).
		Where("user_id = ? AND peer_id = ?", userID, peerID).
		Update("unread_count", 0).Error
}

func (r *MessageRepositoryImpl) GetUnreadMessages(userID uint) ([]*model.Message, error) {
	var sessions []struct {
		PeerID      uint
		UnreadCount int
	}

	if err := r.db.Table("chat_sessions").
		Select("peer_id, unread_count").
		Where("user_id = ? AND unread_count > 0", userID).
		Find(&sessions).Error; err != nil {
		return nil, err
	}

	if len(sessions) == 0 {
		return []*model.Message{}, nil
	}

	var allMessages []*model.Message

	for _, session := range sessions {
		var sessionMessages []*model.Message
		limit := session.UnreadCount
		if limit <= 0 {
			continue
		}

		err := r.db.Where(
			"(sender_id = ? AND receiver_id = ?)",
			session.PeerID, userID,
		).
			Order("created_at DESC").
			Limit(limit).
			Find(&sessionMessages).Error
		if err != nil {
			return nil, err
		}

		allMessages = append(allMessages, sessionMessages...)
	}

	for i, j := 0, len(allMessages)-1; i < j; i, j = i+1, j-1 {
		allMessages[i], allMessages[j] = allMessages[j], allMessages[i]
	}

	return allMessages, nil
}