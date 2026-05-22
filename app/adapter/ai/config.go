package ai

import (
	"log"
	"os"
	"time"
)

func LoadConfigFromEnv() *Config {
	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		return nil
	}

	return &Config{
		APIKey:  apiKey,
		BaseURL: envOrDefault("AI_BASE_URL", "https://open.bigmodel.cn/api/paas/v4"),
		Model:   envOrDefault("AI_MODEL", "glm-4-flash"),
	}
}

func NewDefaultAgentConfig() AgentConfig {
	return AgentConfig{
		Enabled:       true,
		ResponseDelay: 1 * time.Second,
		SystemPrompt:  defaultSystemPrompt,
		AiUserID:      999999,
		BufferSize:    5,
		BufferTimeout: 2 * time.Second,
	}
}

func InitializeAgent(config Config, agentConfig AgentConfig, handler ResponseHandler) *ChatAgent {
	if config.APIKey == "" {
		log.Println("[AI] API Key为空，AI功能已禁用")
		return nil
	}

	client := NewGLMClient(config)
	agent := NewChatAgent(client, agentConfig, handler)

	log.Println("[AI] ChatAgent初始化成功！")
	log.Printf("[AI] 模型: %s | AI用户ID: %d", config.Model, agentConfig.AiUserID)

	return agent
}

func envOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

const defaultSystemPrompt = `你是一个友好的聊天助手，正在参与群聊。你的特点：
1. 自然地融入对话，不要过于频繁发言
2. 当被直接提问时优先回答
3. 保持简洁有趣的回复风格
4. 用中文回复`
