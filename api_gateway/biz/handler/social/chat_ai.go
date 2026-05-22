package social

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/app/adapter/ai"
	commonpb "github.com/ZOEKOFK/video_web_v3/app/pb/common"
	socialpb "github.com/ZOEKOFK/video_web_v3/app/pb/social"
)

var (
	ChatAgent  *ai.ChatAgent
	initOnce   sync.Once
	initFailed bool
)

func InitializeChatAI(config ai.Config, agentConfig ai.AgentConfig) {
	if config.APIKey == "" {
		log.Println("[AI] API Key为空，跳过AI初始化")
		return
	}

	ChatAgent = ai.InitializeAgent(config, agentConfig, HandleAIResponse)

	if ChatAgent != nil {
		log.Println("[AI] 聊天AI功能已启用！")
	} else {
		log.Println("[AI] 聊天AI功能未启用")
	}
}

func ensureChatAgent() bool {
	if ChatAgent != nil {
		return true
	}
	if initFailed {
		return false
	}
	var success bool
	initOnce.Do(func() {
		config := ai.LoadConfigFromEnv()
		if config == nil || config.APIKey == "" {
			log.Println("[AI] 懒加载失败: AI_API_KEY 未设置或为空")
			initFailed = true
			success = false
			return
		}
		agentConfig := ai.NewDefaultAgentConfig()
		ChatAgent = ai.InitializeAgent(*config, agentConfig, HandleAIResponse)
		if ChatAgent != nil {
			log.Println("[AI] 懒加载成功: ChatAgent 已初始化")
			success = true
		} else {
			log.Println("[AI] 懒加载失败: InitializeAgent 返回 nil")
			initFailed = true
			success = false
		}
	})
	return success
}

// ForwardToAI 收到的内容封装传给ai处理
func ForwardToAI(senderID, receiverID uint64, content string) {
	if !ensureChatAgent() {
		return
	}

	aiMessage := &ai.ChatMessage{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		Timestamp:  time.Now().Unix(),
		IsAI:       false,
	}
	go ChatAgent.OnMessageReceived(aiMessage)
}

// HandleAIResponse 对ai的回复进行处理
func HandleAIResponse(response *ai.AIResponse) {
	if response == nil || response.Content == "" {
		return
	}

	ctx := context.Background()

	req := &socialpb.SendMessageRequest{
		SenderId:   int64(response.SenderID),
		ReceiverId: int64(response.ReceiverID),
		Content:    response.Content,
		Type:       0,
	}

	resp, err := client.ChatServiceClient.SendMessage(ctx, req)
	if err != nil {
		log.Printf("[AI-WS] 存储AI回复失败: %v", err)
		return
	}

	if resp.Code != commonpb.ErrorCode_SUCCESS {
		log.Printf("[AI-WS] 存储AI回复错误: code=%d, msg=%s", resp.Code, resp.Message)
		return
	}

	log.Printf("[AI-WS] AI回复已存储: Sender=%d, Receiver=%d", response.SenderID, response.ReceiverID)

	wsMsg := &WSMessage{
		Type:       "message",
		SenderID:   response.SenderID,
		ReceiverID: response.ReceiverID,
		Content:    response.Content,
		Timestamp:  time.Now().Unix(),
	}

	targets := resolveTargets(response)
	SendToUsers(targets, wsMsg)
}

func resolveTargets(response *ai.AIResponse) []uint64 {
	targets := []uint64{response.ReceiverID}

	if len(response.Messages) > 0 {
		originalSender := response.Messages[0].SenderID
		if originalSender != response.SenderID && originalSender != response.ReceiverID {
			found := false
			for _, id := range targets {
				if id == originalSender {
					found = true
					break
				}
			}
			if !found {
				targets = append(targets, originalSender)
			}
		}
	}
	return targets
}
