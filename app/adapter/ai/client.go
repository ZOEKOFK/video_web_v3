package ai

import (
	"context"
	"time"
)

type AIClient interface {
	Chat(ctx context.Context, message string) (string, error)
	ChatWithHistory(ctx context.Context, messages []Message) (string, error)
}
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Config LLM客服端的配置
type Config struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
}

// ResponseHandler 回调函数
type ResponseHandler func(response *AIResponse)

type AIResponse struct {
	Content    string
	ChatID     string
	SenderID   uint64
	ReceiverID uint64
	Messages   []*ChatMessage
}

type ChatMessage struct {
	SenderID   uint64 `json:"sender_id"`
	ReceiverID uint64 `json:"receiver_id"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
	IsAI       bool   `json:"is_ai"`
}

type AgentConfig struct {
	Enabled       bool          `json:"enabled"`
	ResponseDelay time.Duration `json:"response_delay"`
	SystemPrompt  string        `json:"system_prompt"`
	AiUserID      uint64        `json:"ai_user_id"`
	BufferSize    int           `json:"buffer_size"`
	BufferTimeout time.Duration `json:"buffer_timeout"`
}
