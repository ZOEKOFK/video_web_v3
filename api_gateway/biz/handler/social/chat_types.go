package social

import (
	"net/http"
	"sync"
	"time"

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

// ChatWSClient websocket 建立连接的用户结构体
type ChatWSClient struct {
	conn       *websocket.Conn
	userID     uint64
	send       chan []byte
	mutex      sync.Mutex
	closed     bool
	lastActive time.Time
}

// WSMessage 客户端发来的信息
type WSMessage struct {
	Type       string `json:"type"`
	SenderID   uint64 `json:"sender_id"`
	ReceiverID uint64 `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
}

// WSResponse 返回给客户端的信息
type WSResponse struct {
	Type    string      `json:"type"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
