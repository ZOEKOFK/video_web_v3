package social

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/api_gateway/my_jwt"
	commonpb "github.com/ZOEKOFK/video_web_v3/app/pb/common"
	socialpb "github.com/ZOEKOFK/video_web_v3/app/pb/social"
)

// ReadPump 从客户端读入消息，配置连接参数，持续接收消息
func (c *ChatWSClient) ReadPump() {
	defer func() {
		UnregisterClient(c.userID)
		c.Close()
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

// WritePump 做心跳检测，监听通道，有消息就写回客户端
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

// handleMessage 反序列化然后给根据类型给不同函数处理
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
		c.sendError("unknown message type: " + wsMsg.Type)
	}
}

// handleChatMessage chat类型的消息处理：储存传来的消息，发送ack,把消息丢给ai
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

	wsMsg.SenderID = c.userID
	if !SendToUser(wsMsg.ReceiverID, wsMsg) {
		log.Printf("[WS] User %d is offline, message stored in DB", wsMsg.ReceiverID)
	}

	ForwardToAI(c.userID, wsMsg.ReceiverID, wsMsg.Content)
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

	if IsUserOnline(userID) {
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

	RegisterClient(userID, ws)
	go ws.WritePump()
	go ws.pushUnreadMessages()
	ws.ReadPump()
}
