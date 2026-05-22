package ai

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type ChatAgent struct {
	client        AIClient
	aiUserID      uint64
	messageBuffer sync.Map
	config        AgentConfig
	onResponse    ResponseHandler
}

func NewChatAgent(client AIClient, config AgentConfig, handler ResponseHandler) *ChatAgent {
	if config.AiUserID == 0 {
		config.AiUserID = 999999
	}
	if config.BufferSize == 0 {
		config.BufferSize = 5
	}
	if config.BufferTimeout == 0 {
		config.BufferTimeout = 2 * time.Second
	}
	if config.ResponseDelay == 0 {
		config.ResponseDelay = 1 * time.Second
	}
	if config.SystemPrompt == "" {
		config.SystemPrompt = defaultSystemPrompt
	}

	return &ChatAgent{
		client:     client,
		config:     config,
		aiUserID:   config.AiUserID,
		onResponse: handler,
	}
}

// OnMessageReceived ai处理客户端传来的消息，并对信息暂存
func (a *ChatAgent) OnMessageReceived(msg *ChatMessage) {
	if !a.config.Enabled {
		return
	}

	if msg.SenderID == a.aiUserID {
		return
	}

	chatID := a.getChatID(msg.SenderID, msg.ReceiverID)

	var buffer []*ChatMessage
	if existing, ok := a.messageBuffer.Load(chatID); ok {
		buffer = existing.([]*ChatMessage)
	}

	buffer = append(buffer, msg)
	if len(buffer) > a.config.BufferSize {
		buffer = buffer[len(buffer)-a.config.BufferSize:]
	}

	a.messageBuffer.Store(chatID, buffer)

	go a.processMessageAsync(chatID, buffer)
}

// 给ai回复设置点延时，拿到回复后调用回调函数
func (a *ChatAgent) processMessageAsync(chatID string, messages []*ChatMessage) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[Agent] Panic: %v", r)
		}
	}()

	time.Sleep(a.config.ResponseDelay)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := a.generateResponse(ctx, messages)
	if err != nil {
		log.Printf("[Agent] 生成回复失败: %v", err)
		return
	}
	if response == "" {
		log.Println("[Agent] AI选择不回复")
		return
	}
	log.Printf("[Agent] AI准备回复: %.50s...", response)
	aiResponse := &AIResponse{
		Content:    response,
		ChatID:     chatID,
		SenderID:   a.aiUserID,
		ReceiverID: messages[0].ReceiverID,
		Messages:   messages,
	}

	if a.onResponse != nil {
		a.onResponse(aiResponse)
	}

	a.messageBuffer.Delete(chatID)
}

// generateResponse 与ai客户端请求，生成ai回复
func (a *ChatAgent) generateResponse(ctx context.Context, messages []*ChatMessage) (string, error) {
	history := make([]Message, 0, len(messages)+1)

	for _, msg := range messages {
		role := "user"
		if msg.SenderID == a.aiUserID || msg.IsAI {
			role = "assistant"
		}

		history = append(history, Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	response, err := a.client.ChatWithHistory(ctx, history)
	if err != nil {
		return "", fmt.Errorf("调用AI失败: %w", err)
	}

	return response, nil
}

func (a *ChatAgent) getChatID(userID1, userID2 uint64) string {
	if userID1 < userID2 {
		return fmt.Sprintf("%d:%d", userID1, userID2)
	}
	return fmt.Sprintf("%d:%d", userID2, userID1)
}
