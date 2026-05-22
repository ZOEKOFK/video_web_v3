package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ZOEKOFK/video_web_v3/app/adapter/ai"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║            AI API 测试工具 v2.0                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		fmt.Println("❌ 请设置环境变量 AI_API_KEY")
		fmt.Println("   PowerShell: $env:AI_API_KEY=\"your-key\"")
		fmt.Println("   CMD: set AI_API_KEY=your-key")
		fmt.Println("   Linux/Mac: export AI_API_KEY=\"your-key\"")
		os.Exit(1)
	}

	config := ai.Config{
		APIKey:  apiKey,
		BaseURL: "https://open.bigmodel.cn/api/paas/v4",
		Model:   "glm-4-flash",
	}

	agentConfig := ai.AgentConfig{
		Enabled:       true,
		ResponseDelay: 500 * time.Millisecond,
		SystemPrompt:  "你是一个友好的AI助手，用简洁有趣的方式回答问题。",
		AiUserID:      999999,
		BufferSize:    5,
	}

	fmt.Println("📋 可用测试:")
	fmt.Println("   1. 测试单轮对话 (Chat)")
	fmt.Println("   2. 测试多轮对话 (ChatWithHistory)")
	fmt.Println("   3. 测试完整Agent流程 (含回调)")
	fmt.Println("   4. 运行所有测试")
	fmt.Println("   0. 退出")
	fmt.Print("\n请选择 [1-4, 0]: ")

	var choice int
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		testSingleChat(config)
	case 2:
		testChatWithHistory(config)
	case 3:
		testFullAgent(config, agentConfig)
	case 4:
		fmt.Println("🧪 运行所有测试...")
		testSingleChat(config)
		fmt.Println()
		testChatWithHistory(config)
		fmt.Println()
		testFullAgent(config, agentConfig)
	default:
		fmt.Println("👋 退出")
	}
}

func testSingleChat(config ai.Config) {
	separator := strings.Repeat("=", 60)

	fmt.Println(separator)
	fmt.Println("🧪 测试1: 单轮对话 (Chat)")
	fmt.Println(separator)
	client := ai.NewGLMClient(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	question := "你好，请用一句话介绍你自己"
	fmt.Printf("📤 提问: %s\n", question)

	start := time.Now()
	response, err := client.Chat(ctx, question)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	fmt.Printf("📥 回答 (耗时 %v):\n", elapsed)
	fmt.Printf("   %s\n", response)
	fmt.Println("✅ 单轮对话测试通过!")
}

func testChatWithHistory(config ai.Config) {
	separator := strings.Repeat("=", 60)

	fmt.Println(separator)
	fmt.Println("🧪 测试2: 多轮对话 (ChatWithHistory)")

	client := ai.NewGLMClient(config)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	messages := []ai.Message{
		{Role: "user", Content: "我喜欢编程"},
		{Role: "assistant", Content: "太棒了！编程是一项很有趣的技能。你主要擅长哪种编程语言？"},
		{Role: "user", Content: "我正在学习Go语言"},
	}

	fmt.Println("📜 对话历史:")
	for i, msg := range messages {
		fmt.Printf("   [%d] %s: %s\n", i+1, msg.Role, msg.Content)
	}

	question := "你能给我一些学习Go语言的建议吗？"
	fmt.Printf("\n📤 追问: %s\n", question)

	start := time.Now()
	response, err := client.ChatWithHistory(ctx, messages)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ 错误: %v\n", err)
		return
	}

	fmt.Printf("📥 回答 (耗时 %v):\n", elapsed)
	fmt.Printf("   %s\n", response)
	fmt.Println("✅ 多轮对话测试通过!")
}

func testFullAgent(config ai.Config, agentConfig ai.AgentConfig) {
	separator := strings.Repeat("=", 60)

	fmt.Println(separator)
	fmt.Println("🧪 测试3: 完整Agent流程 (含回调)")
	fmt.Println(separator)

	receivedResponse := make(chan string, 1)

	handler := func(response *ai.AIResponse) {
		fmt.Println("\n📨 回调触发!")
		fmt.Printf("   发送者: %d (AI)\n", response.SenderID)
		fmt.Printf("   接收者: %d\n", response.ReceiverID)
		fmt.Printf("   回复内容: %s\n", response.Content)
		fmt.Printf("   关联消息数: %d\n", len(response.Messages))
		receivedResponse <- response.Content
	}

	agent := ai.NewChatAgent(
		ai.NewGLMClient(config),
		agentConfig,
		handler,
	)

	if agent == nil {
		fmt.Println("❌ Agent初始化失败")
		return
	}

	fmt.Println("✅ Agent初始化成功")
	fmt.Printf("   模型: %s\n", config.Model)
	fmt.Printf("   AI用户ID: %d\n", agentConfig.AiUserID)
	fmt.Printf("   响应延迟: %v\n", agentConfig.ResponseDelay)
	fmt.Printf("   缓冲大小: %d\n", agentConfig.BufferSize)

	testMessages := []*ai.ChatMessage{
		{
			SenderID:   1001,
			ReceiverID: 1002,
			Content:    "你好！今天天气真好啊",
			Timestamp:  time.Now().Unix(),
			IsAI:       false,
		},
		{
			SenderID:   1002,
			ReceiverID: 1001,
			Content:    "是啊，很适合出去走走",
			Timestamp:  time.Now().Unix(),
			IsAI:       false,
		},
	}

	fmt.Printf("\n📤 模拟用户发送 %d 条消息...\n", len(testMessages))
	for i, msg := range testMessages {
		fmt.Printf("   [%d] 用户%d -> 用户%d: %s\n", i+1, msg.SenderID, msg.ReceiverID, msg.Content)
		agent.OnMessageReceived(msg)
	}

	fmt.Println("⏳ 等待AI响应...")

	select {
	case response := <-receivedResponse:
		fmt.Println("\n✅ 完整Agent流程测试通过!")
		fmt.Printf("   AI回复: %s\n", response)
	case <-time.After(10 * time.Second):
		fmt.Println("\n⚠️ 等待超时，测试失败")
	}
}
