package main

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/biz/handler/social"
	"github.com/ZOEKOFK/video_web_v3/api_gateway/biz/router"
	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/api_gateway/my_jwt"
	"github.com/ZOEKOFK/video_web_v3/app/adapter/ai"
)

func main() {
	err := my_jwt.InitJWT()
	if err != nil {
		log.Fatalf("init jwt fail: %v", err)
	}

	client.InitGRPCClient()

	aiConfig := ai.LoadConfigFromEnv()
	if aiConfig != nil {
		social.InitializeChatAI(*aiConfig, ai.NewDefaultAgentConfig())
	} else {
		log.Println("[AI] 未配置AI_API_KEY，AI功能已禁用")
	}

	h := server.Default(server.WithHostPorts(":8888"), server.WithMaxRequestBodySize(500*1024*1024))
	router.GeneratedRegister(h)
	customizedRegister(h)
	h.Spin()
}
