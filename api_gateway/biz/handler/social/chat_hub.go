package social

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	commonpb "github.com/ZOEKOFK/video_web_v3/app/pb/common"
	socialpb "github.com/ZOEKOFK/video_web_v3/app/pb/social"
)

var chatClients sync.Map
var onlineCount int64
var countMutex sync.RWMutex

func incrementOnline() {
	countMutex.Lock()
	defer countMutex.Unlock()
	onlineCount++
}

func decrementOnline() {
	countMutex.Lock()
	defer countMutex.Unlock()
	onlineCount--
}

func GetOnlineCount() int64 {
	countMutex.RLock()
	defer countMutex.RUnlock()
	return onlineCount
}

func IsUserOnline(userID uint64) bool {
	_, ok := chatClients.Load(userID)
	return ok
}

func RegisterClient(userID uint64, ws *ChatWSClient) {
	chatClients.Store(userID, ws)
	incrementOnline()
	log.Printf("[WS] User %d connected, online: %d", userID, GetOnlineCount())
}

func UnregisterClient(userID uint64) {
	chatClients.Delete(userID)
	decrementOnline()
	log.Printf("[WS] User %d disconnected, online: %d", userID, GetOnlineCount())
}

// SendToUser 把消息发给指定用户的缓冲区
func SendToUser(targetID uint64, msg *WSMessage) bool {
	if peerClient, ok := chatClients.Load(targetID); ok {
		peerWS := peerClient.(*ChatWSClient)
		peerWS.sendMessage(msg)
		return true
	}
	log.Printf("[WS] User %d is offline, message stored in DB", targetID)
	return false
}

// SendToUsers 把消息发给指定的用户们的缓冲区
func SendToUsers(targetIDs []uint64, msg *WSMessage) {
	for _, targetID := range targetIDs {
		msgCopy := *msg
		if !SendToUser(targetID, &msgCopy) {
			log.Printf("[WS] User %d is offline, message stored in DB", targetID)
		}
	}
}

func StoreMessage(senderID, receiverID int64, content string) bool {
	ctx := context.Background()

	req := &socialpb.SendMessageRequest{
		SenderId:   senderID,
		ReceiverId: receiverID,
		Content:    content,
		Type:       0,
	}

	resp, err := client.ChatServiceClient.SendMessage(ctx, req)
	if err != nil {
		log.Printf("[WS] Store message failed: %v", err)
		return false
	}

	if resp.Code != commonpb.ErrorCode_SUCCESS {
		log.Printf("[WS] Store message error: %s", resp.Message)
		return false
	}

	return true
}

func StoreMessageWithCtx(ctx context.Context, senderID, receiverID int64, content string) bool {
	req := &socialpb.SendMessageRequest{
		SenderId:   senderID,
		ReceiverId: receiverID,
		Content:    content,
		Type:       0,
	}

	resp, err := client.ChatServiceClient.SendMessage(ctx, req)
	if err != nil {
		log.Printf("[WS] Store message failed: %v", err)
		return false
	}

	if resp.Code != commonpb.ErrorCode_SUCCESS {
		log.Printf("[WS] Store message error: %s", resp.Message)
		return false
	}

	return true
}

func (c *ChatWSClient) sendMessage(msg *WSMessage) {
	msg.Type = "message"
	msg.Timestamp = time.Now().Unix()
	data, _ := json.Marshal(msg)
	select {
	case c.send <- data:
	default:
		log.Printf("[WS] Send buffer full for user %d, dropping message", c.userID)
	}
}

func (c *ChatWSClient) sendError(errMsg string) {
	c.sendJSON(WSResponse{Type: "error", Code: -1, Message: errMsg})
}

func (c *ChatWSClient) sendJSON(resp WSResponse) {
	data, _ := json.Marshal(resp)
	select {
	case c.send <- data:
	default:
		log.Printf("[WS] Send buffer full for user %d, dropping response", c.userID)
	}
}

func (c *ChatWSClient) write(messageType int, data []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	return c.conn.WriteMessage(messageType, data)
}

func (c *ChatWSClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.closed {
		c.closed = true
		close(c.send)
	}
}

func (c *ChatWSClient) pushUnreadMessages() {
	ctx := context.Background()
	ctxWithUserID := client.WithUserID(ctx, int64(c.userID))

	req := &socialpb.GetUnreadMessagesRequest{
		UserId: int64(c.userID),
	}

	resp, err := client.ChatServiceClient.GetUnreadMessages(ctxWithUserID, req)
	if err != nil {
		log.Printf("[WS] User %d get unread messages failed: %v", c.userID, err)
		return
	}

	if resp.Code != commonpb.ErrorCode_SUCCESS || resp.Data == nil || len(resp.Data.ChatMessageList) == 0 {
		log.Printf("[WS] User %d no unread messages", c.userID)
		return
	}

	log.Printf("[WS] Pushing %d unread messages to user %d", len(resp.Data.ChatMessageList), c.userID)

	for _, msg := range resp.Data.ChatMessageList {
		c.sendJSON(WSResponse{
			Type: "chat",
			Data: msg,
		})
	}
}
