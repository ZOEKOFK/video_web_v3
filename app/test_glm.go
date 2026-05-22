package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	ai "github.com/ZOEKOFK/video_web_v3/app/adapter/ai"
)

func main() {

	separator := strings.Repeat("=", 60)

	fmt.Println(separator)
	fmt.Println("🤖 GLM-4-Flash AI客户端测试程序")
	fmt.Println(separator)

	config := ai.DefaultGLMConfig()

	config.APIKey = "aef9731452ac4f3dbbd6cd6b8aaa34ed.G2bjjyGhL1l8Qmnx"

	if config.APIKey == "" || config.APIKey == "noting" {
		log.Fatal("❌ 请先在代码中填入你的智谱API Key！")
		return
	}

	fmt.Println("\n📋 配置信息:")
	fmt.Printf("   API地址: %s\n", config.BaseURL)
	fmt.Printf("   模型:    %s\n", config.Model)
	if len(config.APIKey) > 12 {
		fmt.Printf("   API Key: %s...%s\n", config.APIKey[:8], config.APIKey[len(config.APIKey)-4:])
	} else {
		fmt.Println("   API Key: (太短，无法显示)")
	}

	fmt.Println("\n🔧 正在连接GLM-4-Flash...")
	client := ai.NewGLMClient(*config)
	fmt.Println("✅ 客户端创建成功!")

	ctx := context.Background()

	testCases := []struct {
		name    string
		message string
	}{
		{"简单问候", "你好"},
		{"数学问题", "1+1等于几？"},
		{"自我介绍", "用一句话介绍你自己"},
		{"中文能力", "请用中文写一首关于春天的诗（四句）"},
	}

	fmt.Println("\n" + separator)
	fmt.Println("🚀 开始测试")
	fmt.Println(separator)

	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d/%d: %s ---\n", i+1, len(testCases), tc.name)
		fmt.Printf("💬 你说: \"%s\"\n", tc.message)

		fmt.Println("⏳ 正在等待GLM-4-Flash回复...")
		startTime := time.Now()

		response, err := client.Chat(ctx, tc.message)

		duration := time.Since(startTime)

		if err != nil {
			log.Printf("❌ 错误: %v\n", err)
		} else {
			fmt.Printf("✅ AI回复 (%.2f秒): \"%s\"\n", duration.Seconds(), response)
		}
	}

	fmt.Println("\n" + separator)
	fmt.Println("🎉 测试完成!")
	fmt.Println(separator)
	fmt.Println("\n💡 提示：如果所有测试都通过了，说明你的AI客户端工作正常！")
	fmt.Println("   接下来我们可以把它集成到聊天系统中。")
}
