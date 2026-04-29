package grpc

import (
	"context"
	"errors"
	"log"
	"strconv"

	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	"github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"github.com/ZOEKOFK/video_web_v3/app/usecase"
	"google.golang.org/grpc/metadata"
)

const UserIDMetadataKey = "user-id"

type UsersGrpc struct {
	users.UnimplementedSessionServiceServer
	users.UnimplementedUserPublicServiceServer
	users.UnimplementedUserAuthServiceServer
	userUsecase usecase.UserUseCase
}

func NewUsersGrpc(userUsecase usecase.UserUseCase) *UsersGrpc {
	return &UsersGrpc{
		userUsecase: userUsecase,
	}
}

func (s *UsersGrpc) Register(ctx context.Context, in *users.UserRegisterRequest) (*common.CommonResponse, error) {
	code, err := s.userUsecase.RegisterUser(in)
	if err != nil {
		return FailResponse("Register", code, err), nil
	}
	return SuccessResponse("Register", nil), nil
}

func (s *UsersGrpc) Login(ctx context.Context, in *users.UserLoginRequest) (*common.CommonResponse, error) {
	loginResp, err := s.userUsecase.LoginUser(in.Username, in.Password, in.Remember)
	if err != nil {
		return FailResponse("Login", common.ErrorCode_USER_PASSWORD_ERROR, err), nil
	}
	data := &common.Data{
		UserInfo:  loginResp.UserInfo,
		TokenInfo: loginResp.TokenInfo,
	}
	return SuccessResponse("Login", data), nil
}

func (s *UsersGrpc) GetUserInfo(ctx context.Context, in *common.IDRequest) (*common.CommonResponse, error) {
	// ============================================
	// 从 Metadata 中提取并验证用户 ID 的示例
	// ============================================
	
	// 1️⃣ 从 Metadata 中提取用户 ID
	userID, err := extractUserIDFromMetadata(ctx)
	if err == nil {
		log.Printf("[GetUserInfo] 从 Metadata 中提取到用户 ID: %d", userID)
		
		// 2️⃣ 验证用户 ID 有效性
		_, err = s.userUsecase.ValidateUserID(userID)
		if err != nil {
			log.Printf("[GetUserInfo] 用户 ID 验证失败: %v", err)
			return FailResponse("GetUserInfo", common.ErrorCode_USER_NOT_LOGIN, err), nil
		}
		log.Printf("[GetUserInfo] 用户 ID 验证通过")
	} else {
		log.Printf("[GetUserInfo] 没有从 Metadata 中提取到用户 ID（这是正常的，公开接口不需要认证）")
	}
	
	// 原有逻辑
	idStr := in.Id
	user, err := s.userUsecase.GetUserInfo(idStr)
	if err != nil {
		return FailResponse("GetUserInfo", common.ErrorCode_PROGRESS_ERROR, err), nil
	}
	pbUser := &common.User{
		Id:        int64(user.ID),
		Username:  user.Username,
		AvatarUrl: user.Avatarurl,
		Nickname:  user.Nickname,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	data := &common.Data{
		UserInfo: pbUser,
	}
	return SuccessResponse("GetUserInfo", data), nil
}

func (s *UsersGrpc) UploadAvatar(ctx context.Context, in *users.UploadAvatarRequest) (*common.CommonResponse, error) {
	return &common.CommonResponse{}, nil
}

func (s *UsersGrpc) RefreshSession(ctx context.Context, in *users.RefreshTokenRequest) (*common.CommonResponse, error) {
	loginResp, err := s.userUsecase.RefreshToken(in.RefreshToken, in.Remember)
	if err != nil {
		return FailResponse("RefreshSession", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}
	data := &common.Data{
		UserInfo:  loginResp.UserInfo,
		TokenInfo: loginResp.TokenInfo,
	}
	return SuccessResponse("RefreshSession", data), nil
}

func (s *UsersGrpc) Logout(ctx context.Context, in *users.RefreshTokenRequest) (*common.CommonResponse, error) {
	err := s.userUsecase.Logout(in.RefreshToken)
	if err != nil {
		
	}
	return SuccessResponse("Logout", nil), nil
}

// ============================================
// Metadata 辅助函数
// ============================================

// extractUserIDFromMetadata 从 gRPC Metadata 中提取用户 ID
func extractUserIDFromMetadata(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return 0, errors.New("metadata not found")
	}

	userIDValues := md.Get(UserIDMetadataKey)
	if len(userIDValues) == 0 {
		return 0, errors.New("user id not found in metadata")
	}

	userID, err := strconv.ParseInt(userIDValues[0], 10, 64)
	if err != nil {
		return 0, errors.New("invalid user id format")
	}

	return userID, nil
}