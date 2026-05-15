package main

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/ZOEKOFK/video_web_v3/app/adapter/consul"
	"github.com/ZOEKOFK/video_web_v3/app/adapter/grpc"
	"github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/mysql"
	"github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/redis"
	"github.com/ZOEKOFK/video_web_v3/app/domain/service_logic"
	"github.com/ZOEKOFK/video_web_v3/app/pb/interaction"
	socialpb "github.com/ZOEKOFK/video_web_v3/app/pb/social"
	"github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"github.com/ZOEKOFK/video_web_v3/app/pb/videos"
	"github.com/ZOEKOFK/video_web_v3/app/usecase"

	orgin "google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

var consulClient *consul.ConsulClient

func main() {
	var err error
	consulClient, err = consul.NewConsulClient(consul.ConsulDefaultAddr)
	if err != nil {
		log.Printf("⚠️ 无法连接 Consul: %v，将以单机模式运行", err)
	} else {
		if err := consulClient.Ping(); err != nil {
			log.Printf("⚠️ Consul 健康检查失败: %v", err)
		} else {
			log.Println("✅ 已连接到 Consul")
		}
	}

	consulClient.SetupGracefulShutdown()

	db, err := mysql.InitDB()
	if err != nil {
		log.Printf(err.Error())
		panic(err)
	}
	defer db.Close()

	redisClient := redis.InitRedis()

	userRepo := mysql.NewUserRepository(db)
	userService := service_logic.NewUsersServiceLogic(userRepo)
	userUseCase := usecase.NewUserUsecase(userRepo, userService, redisClient)

	videoRepo := mysql.NewVideosRepository(db)
	videoService := service_logic.NewVideosServiceLogic(videoRepo)
	videoUseCase := usecase.NewVideoUsecase(videoRepo, videoService)

	followRepo := mysql.NewFollowRepository(db)
	followService := service_logic.NewFollowServiceLogic(followRepo)
	socialUseCase := usecase.NewSocialUsecase(followRepo, followService)

	interactionRepo := mysql.NewInteractionRepository(db)
	interactionUseCase := usecase.NewInteractionUsecase(interactionRepo)

	var wg sync.WaitGroup

	wg.Add(4)

	go func() {
		defer wg.Done()
		startUserService(userUseCase)
	}()

	go func() {
		defer wg.Done()
		startVideoService(videoUseCase)
	}()

	go func() {
		defer wg.Done()
		startSocialService(socialUseCase)
	}()
	go func() {
		defer wg.Done()
		startInteractionService(interactionUseCase)
	}()
	wg.Wait()
}

func startUserService(userUseCase usecase.UserUseCase) {
	server := orgin.NewServer(
		orgin.MaxRecvMsgSize(500*1024*1024),
		orgin.MaxSendMsgSize(500*1024*1024),
	)

	usersGrpc := grpc.NewUsersGrpc(userUseCase)
	users.RegisterUserAuthServiceServer(server, usersGrpc)
	users.RegisterUserPublicServiceServer(server, usersGrpc)
	users.RegisterSessionServiceServer(server, usersGrpc)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	addr := fmt.Sprintf(":%d", consul.UserPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("启动用户 gRPC 服务失败: %v", err)
	}

	consulClient.SafeRegister(consul.UserServiceName, consul.ServiceHost, consul.UserPort, []string{"grpc", "user"})

	fmt.Printf("✅ 用户服务 (User Service) 正在监听 %s...\n", addr)
	fmt.Println("   - UserAuthService")
	fmt.Println("   - UserPublicService")
	fmt.Println("   - SessionService")

	if err := server.Serve(l); err != nil {
		log.Fatalf("用户 gRPC 服务启动失败: %v", err)
	}
}

func startInteractionService(interactionUseCase usecase.InteractionUseCase) {
	server := orgin.NewServer(
		orgin.MaxRecvMsgSize(500*1024*1024),
		orgin.MaxSendMsgSize(500*1024*1024),
	)
	interactionGrpc := grpc.NewInteractionGrpc(interactionUseCase)
	interaction.RegisterCommentAuthServiceServer(server, interactionGrpc)
	interaction.RegisterCommentPublicServiceServer(server, interactionGrpc)
	interaction.RegisterLikeAuthServiceServer(server, interactionGrpc)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	addr := fmt.Sprintf(":%d", consul.InteractionPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("启动互动 gRPC 服务失败: %v", err)
	}

	consulClient.SafeRegister(consul.InteractionServiceName, consul.ServiceHost, consul.InteractionPort, []string{"grpc", "interaction"})

	fmt.Printf("✅ 互动服务 (Interaction Service) 正在监听 %s...\n", addr)
	fmt.Println("   - CommentAuthService")
	fmt.Println("   - CommentPublicService")
	fmt.Println("   - LikeAuthService")

	if err := server.Serve(l); err != nil {
		log.Fatalf("互动 gRPC 服务启动失败: %v", err)
	}
}

func startVideoService(videoUseCase usecase.VideoUseCase) {
	server := orgin.NewServer(
		orgin.MaxRecvMsgSize(500*1024*1024),
		orgin.MaxSendMsgSize(500*1024*1024),
	)

	videosGrpc := grpc.NewVideosGrpc(videoUseCase)
	videos.RegisterVideoAuthServiceServer(server, videosGrpc)
	videos.RegisterVideoPublicServiceServer(server, videosGrpc)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	addr := fmt.Sprintf(":%d", consul.VideoPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("启动视频 gRPC 服务失败: %v", err)
	}

	consulClient.SafeRegister(consul.VideoServiceName, consul.ServiceHost, consul.VideoPort, []string{"grpc", "video"})

	fmt.Printf("✅ 视频服务 (Video Service) 正在监听 %s...\n", addr)
	fmt.Println("   - VideoAuthService")
	fmt.Println("   - VideoPublicService")

	if err := server.Serve(l); err != nil {
		log.Fatalf("视频 gRPC 服务启动失败: %v", err)
	}
}

func startSocialService(socialUseCase usecase.SocialUseCase) {
	server := orgin.NewServer(
		orgin.MaxRecvMsgSize(500*1024*1024),
		orgin.MaxSendMsgSize(500*1024*1024),
	)

	socialGrpc := grpc.NewSocialGrpc(socialUseCase)
	socialpb.RegisterFollowAuthServiceServer(server, socialGrpc)
	socialpb.RegisterFollowPublicServiceServer(server, socialGrpc)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	addr := fmt.Sprintf(":%d", consul.SocialPort)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("启动社交 gRPC 服务失败: %v", err)
	}

	consulClient.SafeRegister(consul.SocialServiceName, consul.ServiceHost, consul.SocialPort, []string{"grpc", "social"})

	fmt.Printf("✅ 社交服务 (Social Service) 正在监听 %s...\n", addr)
	fmt.Println("   - FollowAuthService")
	fmt.Println("   - FollowPublicService")

	if err := server.Serve(l); err != nil {
		log.Fatalf("社交 gRPC 服务启动失败: %v", err)
	}
}
