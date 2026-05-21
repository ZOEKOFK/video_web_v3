package social

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/api_gateway/my_jwt"
	commonpb "github.com/ZOEKOFK/video_web_v3/app/pb/common"
	socialpb "github.com/ZOEKOFK/video_web_v3/app/pb/social"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChatWSClient struct {
	conn       *websocket.Conn
	userID     uint64
	send       chan []byte
	mutex      sync.Mutex
	closed     bool
	lastActive time.Time
}

type WSMessage struct {
	Type       string `json:"type"`
	SenderID   uint64 `json:"sender_id"`
	ReceiverID uint64 `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

type WSResponse struct {
	Type    string      `json:"type"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

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

func (c *ChatWSClient) ReadPump() {
	defer func() {
		chatClients.Delete(c.userID)
		decrementOnline()
		c.Close()
		log.Printf("[WS] User %d disconnected, online: %d", c.userID, GetOnlineCount())
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		c.lastActive = time.Now()
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WS] User %d read error: %v", c.userID, err)
			}
			break
		}
		c.lastActive = time.Now()
		c.handleMessage(msg)
	}
}

func (c *ChatWSClient) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(msg)

			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte("\n"))
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
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

func (c *ChatWSClient) handleMessage(msg []byte) {
	var wsMsg WSMessage
	if err := json.Unmarshal(msg, &wsMsg); err != nil {
		log.Printf("[WS] User %d invalid message format: %v", c.userID, err)
		c.sendError("invalid message format")
		return
	}

	switch wsMsg.Type {
	case "chat":
		if wsMsg.Content == "" {
			c.sendError("content is empty")
			return
		}
		if wsMsg.ReceiverID == 0 {
			c.sendError("receiver_id is required")
			return
		}
		c.handleChatMessage(&wsMsg)
	case "ping":
		c.sendJSON(WSResponse{Type: "pong", Code: 0, Message: "ok"})
	default:
		c.sendError(fmt.Sprintf("unknown message type: %s", wsMsg.Type))
	}
}

func (c *ChatWSClient) handleChatMessage(wsMsg *WSMessage) {
	ctx := context.Background()
	ctxWithUserID := client.WithUserID(ctx, int64(c.userID))

	req := &socialpb.SendMessageRequest{
		SenderId:   int64(c.userID),
		ReceiverId: int64(wsMsg.ReceiverID),
		Content:    wsMsg.Content,
		Type:       0,
	}

	resp, err := client.ChatServiceClient.SendMessage(ctxWithUserID, req)
	if err != nil {
		log.Printf("[WS] User %d send message failed: %v", c.userID, err)
		c.sendError("failed to send message")
		return
	}

	if resp.Code != commonpb.ErrorCode_SUCCESS {
		log.Printf("[WS] User %d send message error: %s", c.userID, resp.Message)
		c.sendJSON(WSResponse{
			Type:    "error",
			Code:    int(resp.Code),
			Message: resp.Message,
		})
		return
	}

	c.sendJSON(WSResponse{
		Type:    "ack",
		Code:    0,
		Message: "send message success",
	})

	if peerClient, ok := chatClients.Load(wsMsg.ReceiverID); ok {
		peerWS := peerClient.(*ChatWSClient)
		// 在转发前设置正确的发送者ID
		wsMsg.SenderID = c.userID
		peerWS.sendMessage(wsMsg)
		log.Printf("[WS] Message from user %d to user %d delivered", c.userID, wsMsg.ReceiverID)
	} else {
		log.Printf("[WS] User %d is offline, message stored in DB", wsMsg.ReceiverID)
	}
}

func (c *ChatWSClient) sendMessage(msg *WSMessage) {
	msg.Type = "message"
	// 不再覆盖SenderID，保留在handleChatMessage中已设置的值
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

func WsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil || cookie.Value == "" {
		http.Error(w, `{"type":"error","code":-1,"message":"token required"}`, http.StatusUnauthorized)
		return
	}
	token := cookie.Value

	userID, err := my_jwt.GetUserIDFromTokenString(token)
	if err != nil {
		log.Printf("[WS] Invalid token: %v", err)
		http.Error(w, `{"type":"error","code":-1,"message":"invalid token"}`, http.StatusUnauthorized)
		return
	}

	if _, exists := chatClients.Load(userID); exists {
		http.Error(w, `{"type":"error","code":-2,"message":"already connected"}`, http.StatusConflict)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS] Upgrade failed for user %d: %v", userID, err)
		return
	}

	ws := &ChatWSClient{
		conn:       conn,
		userID:     userID,
		send:       make(chan []byte, 256),
		lastActive: time.Now(),
	}

	chatClients.Store(userID, ws)
	incrementOnline()
	log.Printf("[WS] User %d connected, online: %d", userID, GetOnlineCount())
	go ws.WritePump()
	go ws.pushUnreadMessages()
	ws.ReadPump()
}
