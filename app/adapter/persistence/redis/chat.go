package redis

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/redis/go-redis/v9"
)

const (
	ChatMsgKeyPrefix   = "chat:messages:"
	OnlineUsersKey     = "chat:online:"
	UserSessionPrefix  = "chat:session:"
	MessageListMaxLen  = 100
	MessageCacheTTL    = 24 * time.Hour
	OnlineStatusTTL    = 5 * time.Minute
)

type ChatCache struct{}

func NewChatCache() *ChatCache {
	return &ChatCache{}
}

func (c *ChatCache) getMessageKey(userID1, userID2 uint) string {
	if userID1 < userID2 {
		return fmt.Sprintf("%s%d:%d", ChatMsgKeyPrefix, userID1, userID2)
	}
	return fmt.Sprintf("%s%d:%d", ChatMsgKeyPrefix, userID2, userID1)
}

func (c *ChatCache) CacheMessage(senderID, receiverID uint, msg *model.ChatMessageDTO) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	key := c.getMessageKey(senderID, receiverID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	pipe := Client.Pipeline()
	pipe.LPush(Ctx, key, data)
	pipe.LTrim(Ctx, key, 0, MessageListMaxLen-1)
	pipe.Expire(Ctx, key, MessageCacheTTL)
	_, err = pipe.Exec(Ctx)
	return err
}

func (c *ChatCache) GetRecentMessages(userID, peerID uint, limit int64) ([]*model.ChatMessageDTO, error) {
	if !isConnected {
		return nil, fmt.Errorf("redis not connected")
	}
	key := c.getMessageKey(userID, peerID)
	data, err := Client.LRange(Ctx, key, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	messages := make([]*model.ChatMessageDTO, 0, len(data))
	for _, d := range data {
		var msg model.ChatMessageDTO
		if err := json.Unmarshal([]byte(d), &msg); err == nil {
			messages = append(messages, &msg)
		}
	}
	return messages, nil
}

func (c *ChatCache) SetUserOnline(userID uint) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	key := fmt.Sprintf("%s%d", OnlineUsersKey, userID)
	return Client.Set(Ctx, key, "1", OnlineStatusTTL).Err()
}

func (c *ChatCache) SetUserOffline(userID uint) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	key := fmt.Sprintf("%s%d", OnlineUsersKey, userID)
	return Client.Del(Ctx, key).Err()
}

func (c *ChatCache) IsUserOnline(userID uint) (bool, error) {
	if !isConnected {
		return false, fmt.Errorf("redis not connected")
	}
	key := fmt.Sprintf("%s%d", OnlineUsersKey, userID)
	result, err := Client.Exists(Ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

func (c *ChatCache) GetOnlineUsers() ([]uint, error) {
	if !isConnected {
		return nil, fmt.Errorf("redis not connected")
	}
	pattern := fmt.Sprintf("%s*", OnlineUsersKey)
	keys, err := Client.Keys(Ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	var userIDs []uint
	for _, key := range keys {
		var userID uint
		if _, err := fmt.Sscanf(key, OnlineUsersKey+"%d", &userID); err == nil {
			userIDs = append(userIDs, userID)
		}
	}
	return userIDs, nil
}

func (c *ChatCache) PublishMessage(userID uint, msg *model.ChatMessageDTO) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	channel := fmt.Sprintf("chat:channel:%d", userID)
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return Client.Publish(Ctx, channel, data).Err()
}

func (c *ChatCache) SubscribeUserMessages(userID uint) *redis.PubSub {
	channel := fmt.Sprintf("chat:channel:%d", userID)
	return Client.Subscribe(Ctx, channel)
}

func (c *ChatCache) DeleteChatCache(userID1, userID2 uint) error {
	if !isConnected {
		return fmt.Errorf("redis not connected")
	}
	key := c.getMessageKey(userID1, userID2)
	return Client.Del(Ctx, key).Err()
}
