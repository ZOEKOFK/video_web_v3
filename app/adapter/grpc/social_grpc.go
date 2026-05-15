package grpc

import (
	"context"
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	"github.com/ZOEKOFK/video_web_v3/app/pb/social"
	"github.com/ZOEKOFK/video_web_v3/app/usecase"
)

type SocialGrpc struct {
	social.UnimplementedFollowAuthServiceServer
	social.UnimplementedFollowPublicServiceServer
	socialUsecase usecase.SocialUseCase
}

func NewSocialGrpc(socialUsecase usecase.SocialUseCase) *SocialGrpc {
	return &SocialGrpc{
		socialUsecase: socialUsecase,
	}
}

func (s *SocialGrpc) FollowAction(ctx context.Context, in *social.FollowRequest) (*common.CommonResponse, error) {
	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("FollowAction", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}
	err = s.socialUsecase.FollowAction(uint(userID), uint(in.UserId), in.Status)
	if err != nil {
		return FailResponse("FollowAction", common.ErrorCode_PROGRESS_ERROR, err), nil
	}
	return SuccessResponse("FollowAction", nil), nil
}

func (s *SocialGrpc) GetFriendList(ctx context.Context, in *social.FriendListRequest) (*common.CommonResponse, error) {
	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("GetFriendList", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}

	friends, err := s.socialUsecase.GetFriendList(uint(userID), int(in.Page), int(in.PageSize))
	if err != nil {
		return FailResponse("GetFriendList", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		UserList: model.UserListToPb(friends),
	}
	return SuccessResponse("GetFriendList", data), nil
}

func (s *SocialGrpc) GetFollowList(ctx context.Context, in *social.FollowListRequest) (*common.CommonResponse, error) {
	log.Printf("[GetFollowList] page=%d, pageSize=%d", in.Page, in.PageSize)

	users, err := s.socialUsecase.GetFollowList(uint(in.UserId), int(in.Page), int(in.PageSize))
	if err != nil {
		return FailResponse("GetFollowList", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		UserList: model.UserListToPb(users),
	}
	return SuccessResponse("GetFollowList", data), nil
}

func (s *SocialGrpc) GetFollowerList(ctx context.Context, in *social.FollowerListRequest) (*common.CommonResponse, error) {
	log.Printf("[GetFollowerList] page=%d, pageSize=%d", in.Page, in.PageSize)

	users, err := s.socialUsecase.GetFollowerList(uint(in.UserId), int(in.Page), int(in.PageSize))
	if err != nil {
		return FailResponse("GetFollowerList", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		UserList: model.UserListToPb(users),
	}
	return SuccessResponse("GetFollowerList", data), nil
}
