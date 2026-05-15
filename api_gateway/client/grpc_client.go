package client

import (
	"context"
	"log"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client/consul"
	interactionpb "github.com/ZOEKOFK/video_web_v3/app/pb/interaction"
	socialpb "github.com/ZOEKOFK/video_web_v3/app/pb/social"
	userspb "github.com/ZOEKOFK/video_web_v3/app/pb/users"
	videospb "github.com/ZOEKOFK/video_web_v3/app/pb/videos"
)

var discoverClient *consul.DiscoverClient

var UserAuthServiceClient userspb.UserAuthServiceClient
var UserPublicServiceClient userspb.UserPublicServiceClient
var UserSessionServiceClient userspb.SessionServiceClient

var VideoAuthServiceClient videospb.VideoAuthServiceClient
var VideoPublicServiceClient videospb.VideoPublicServiceClient

var FollowAuthServiceClient socialpb.FollowAuthServiceClient
var FollowPublicClient socialpb.FollowPublicServiceClient

var LikeAuthServiceClient interactionpb.LikeAuthServiceClient
var CommentAuthServiceClient interactionpb.CommentAuthServiceClient
var CommentPublicServiceClient interactionpb.CommentPublicServiceClient

func InitGRPCClient() {
	var err error
	discoverClient, err = consul.NewDiscoverClient(consul.ConsulDefaultAddr)
	if err != nil {
		log.Printf("无法连接 Consul: %v，将使用默认地址", err)
	} else {
		if err := discoverClient.Ping(); err != nil {
			log.Printf("Consul 健康检查失败: %v", err)
			discoverClient = nil
		} else {
			log.Println("网关已连接到 Consul")
		}
	}
	initUserGRPCClient()
	initVideoGRPCClient()
	initSocialGRPCClient()
	initInteractionGRPCClient()
}

func resolveAddr(serviceName string) string {
	if discoverClient == nil {
		return consul.DefaultAddr(serviceName)
	}
	addr, err := discoverClient.DiscoverOneService(serviceName)
	if err != nil {
		log.Printf("从 Consul 发现服务 [%s] 失败: %v", serviceName, err)
		return consul.DefaultAddr(serviceName)
	}
	log.Printf("✅ 从 Consul 发现服务 [%s]: %s", serviceName, addr)
	return addr
}

func initUserGRPCClient() {
	addr := resolveAddr(consul.UserServiceName)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接用户 gRPC 服务失败: %v", err)
	}

	UserAuthServiceClient = userspb.NewUserAuthServiceClient(conn)
	UserPublicServiceClient = userspb.NewUserPublicServiceClient(conn)
	UserSessionServiceClient = userspb.NewSessionServiceClient(conn)

	log.Printf("✅ 用户 gRPC 客户端已连接到 %s", addr)
}

func initVideoGRPCClient() {
	addr := resolveAddr(consul.VideoServiceName)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接视频 gRPC 服务失败: %v", err)
	}

	VideoAuthServiceClient = videospb.NewVideoAuthServiceClient(conn)
	VideoPublicServiceClient = videospb.NewVideoPublicServiceClient(conn)

	log.Printf("✅ 视频 gRPC 客户端已连接到 %s", addr)
}

func initSocialGRPCClient() {
	addr := resolveAddr(consul.SocialServiceName)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接社交 gRPC 服务失败: %v", err)
	}

	FollowAuthServiceClient = socialpb.NewFollowAuthServiceClient(conn)
	tmp := socialpb.NewFollowPublicServiceClient(conn)
	FollowPublicClient = tmp

	log.Printf("✅ 社交 gRPC 客户端已连接到 %s", addr)
}

func initInteractionGRPCClient() {
	addr := resolveAddr(consul.InteractionServiceName)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接互动 gRPC 服务失败: %v", err)
	}

	LikeAuthServiceClient = interactionpb.NewLikeAuthServiceClient(conn)
	CommentAuthServiceClient = interactionpb.NewCommentAuthServiceClient(conn)
	CommentPublicServiceClient = interactionpb.NewCommentPublicServiceClient(conn)

	log.Printf("✅ 互动 gRPC 客户端已连接到 %s", addr)
}

const UserIDMetadataKey = "user-id"

func WithUserID(ctx context.Context, userID int64) context.Context {
	md := metadata.Pairs(UserIDMetadataKey, strconv.FormatInt(userID, 10))
	return metadata.NewOutgoingContext(ctx, md)
}
