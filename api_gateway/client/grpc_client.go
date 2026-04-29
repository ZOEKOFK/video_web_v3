package client

import (
	"context"
	"log"
	"strconv"

	userspb "github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	UserAuthServiceClient    userspb.UserAuthServiceClient
	UserPublicServiceClient userspb.UserPublicServiceClient
	UserSessionServiceClient userspb.SessionServiceClient
)

const UserIDMetadataKey = "user-id"

func InitGRPCClient() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接 gRPC 服务失败: %v", err)
	}
	UserAuthServiceClient = userspb.NewUserAuthServiceClient(conn)
	UserPublicServiceClient = userspb.NewUserPublicServiceClient(conn)
	UserSessionServiceClient = userspb.NewSessionServiceClient(conn)
}

// WithUserID 将用户 ID 添加到 gRPC Metadata 中
func WithUserID(ctx context.Context, userID int64) context.Context {
	return metadata.AppendToOutgoingContext(ctx, UserIDMetadataKey, strconv.FormatInt(userID, 10))
}