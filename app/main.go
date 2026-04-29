package main

import (
	"fmt"
	"log"
	"net"

	"github.com/ZOEKOFK/video_web_v3/app/adapter/grpc"
	"github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/mysql"
	"github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/redis"
	"github.com/ZOEKOFK/video_web_v3/app/domain/service_logic"
	"github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"github.com/ZOEKOFK/video_web_v3/app/usecase"

	orgin "google.golang.org/grpc"
)

func main() {
	// 初始化数据库连接
	db, err := mysql.InitDB()
	if err != nil {
		log.Printf(err.Error())
		panic(err)
	}
	defer db.Close()
	//初始化redis
	redisClient := redis.InitRedis()
	// 初始化依赖
	repo := mysql.NewUserRepository(db)
	service := service_logic.NewUsersServiceLogic(repo)
	userUseCase := usecase.NewUserUsecase(repo, service, redisClient)

	// 创建 gRPC 服务端
	server := orgin.NewServer()

	// 注册 UserAuthService 服务（只包含 GetUserInfo 方法）
	usersGrpc := grpc.NewUsersGrpc(userUseCase)
	users.RegisterUserAuthServiceServer(server, usersGrpc)
	users.RegisterUserPublicServiceServer(server, usersGrpc)
	users.RegisterSessionServiceServer(server, usersGrpc)
	// 启动 gRPC 服务器
	addr := ":50051"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("启动 gRPC 服务器失败: %v", err)
	}

	fmt.Printf("gRPC 服务器正在监听 %s...\n", addr)
	if err := server.Serve(l); err != nil {
		log.Fatalf("gRPC 服务器启动失败: %v", err)
	}
}
