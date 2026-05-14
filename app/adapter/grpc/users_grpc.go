package grpc

import (
	"context"
	"errors"
	"strconv"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
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
	idStr := in.Id
	user, err := s.userUsecase.GetUserInfo(idStr)
	if err != nil {
		return FailResponse("GetUserInfo", common.ErrorCode_PROGRESS_ERROR, err), nil
	}
	pbUser := model.UserToPb(user)
	data := &common.Data{
		UserInfo: pbUser,
	}
	return SuccessResponse("GetUserInfo", data), nil
}

func (s *UsersGrpc) UploadAvatar(ctx context.Context, in *users.UploadAvatarRequest) (*common.CommonResponse, error) {
	user, err := s.userUsecase.UploadAvatar(in)
	if err != nil {
		return FailResponse("UploadAvatar", common.ErrorCode_PROGRESS_ERROR, err), nil
	}
	dto := model.UserToPb(user)
	data := &common.Data{
		UserInfo: dto,
	}
	return SuccessResponse("UploadAvatar", data), nil
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

func (s *UsersGrpc) GetMFACode(ctx context.Context, in *users.GetMFACodeRequest) (*common.CommonResponse, error) {
	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("GetMFACode", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}

	secret, err := s.userUsecase.GetMFACode(uint(userID))
	if err != nil {
		return FailResponse("GetMFACode", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	mfaInfo := &common.MFA{
		MfaSecret: secret,
	}
	data := &common.Data{
		MfaInfo: mfaInfo,
	}
	return SuccessResponse("GetMFACode", data), nil
}

func (s *UsersGrpc) BindMFA(ctx context.Context, in *users.BindMFARequest) (*common.CommonResponse, error) {
	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("BindMFA", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}

	err = s.userUsecase.BindMFA(uint(userID), in.Code)
	if err != nil {
		return FailResponse("BindMFA", common.ErrorCode_PARAM_ERROR, err), nil
	}

	return SuccessResponse("BindMFA", nil), nil
}

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
