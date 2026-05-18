package usecase

import (
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/redis"
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
)

type ChatUseCase interface {
	SendMessage(senderID, receiverID uint, content string, msgType model.MessageType) (*model.ChatMessageDTO, error)
	GetChatHistory(userID, peerID uint, page, pageSize int) ([]*model.ChatMessageDTO, error)
	GetChatSessions(userID uint) ([]*model.ChatSessionDTO, error)
	GetUnreadMessages(userID uint) ([]*model.ChatMessageDTO, error)
	ClearUnread(userID, peerID uint) error
}

type chatUsecase struct {
	msgRepo repository.MessageRepository
	cache   *redis.ChatCache
}

func NewChatUsecase(msgRepo repository.MessageRepository, cache *redis.ChatCache) ChatUseCase {
	return &chatUsecase{
		msgRepo: msgRepo,
		cache:   cache,
	}
}

func (u *chatUsecase) SendMessage(senderID, receiverID uint, content string, msgType model.MessageType) (*model.ChatMessageDTO, error) {
	msg, err := u.msgRepo.CreateMessage(senderID, receiverID, content, msgType, false)
	if err != nil {
		log.Printf("创建消息失败: %v", err)
		return nil, err
	}
	msgDTO := msg.ToDTO()
	if u.cache != nil {
		u.cache.CacheMessage(senderID, receiverID, msgDTO)
		u.cache.PublishMessage(receiverID, msgDTO)
	}
	u.msgRepo.UpdateSession(receiverID, senderID, msg.ID)
	return msgDTO, nil
}

func (u *chatUsecase) GetChatHistory(userID, peerID uint, page, pageSize int) ([]*model.ChatMessageDTO, error) {
	if u.cache != nil {
		cached, err := u.cache.GetRecentMessages(userID, peerID, int64(page*pageSize))
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
	}
	messages, err := u.msgRepo.GetMessages(userID, peerID, page, pageSize)
	if err != nil {
		return nil, err
	}
	dtos := make([]*model.ChatMessageDTO, len(messages))
	for i, msg := range messages {
		dtos[i] = msg.ToDTO()
	}
	return dtos, nil
}

func (u *chatUsecase) GetChatSessions(userID uint) ([]*model.ChatSessionDTO, error) {
	sessions, err := u.msgRepo.GetConversations(userID)
	if err != nil {
		return nil, err
	}
	dtos := make([]*model.ChatSessionDTO, len(sessions))
	for i, session := range sessions {
		dtos[i] = &model.ChatSessionDTO{
			PeerID:      uint64(session.PeerID),
			UnreadCount: session.UnreadCount,
		}
	}
	return dtos, nil
}

func (u *chatUsecase) GetUnreadMessages(userID uint) ([]*model.ChatMessageDTO, error) {
	messages, err := u.msgRepo.GetUnreadMessages(userID)
	if err != nil {
		return nil, err
	}
	dtos := make([]*model.ChatMessageDTO, len(messages))
	for i, msg := range messages {
		dtos[i] = msg.ToDTO()
	}
	return dtos, nil
}

func (u *chatUsecase) ClearUnread(userID, peerID uint) error {
	return u.msgRepo.ClearUnread(userID, peerID)
}
