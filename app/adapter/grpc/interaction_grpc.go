package grpc

import (
	"context"
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	"github.com/ZOEKOFK/video_web_v3/app/pb/interaction"
	"github.com/ZOEKOFK/video_web_v3/app/usecase"
)

type InteractionGrpc struct {
	interaction.UnimplementedLikeAuthServiceServer
	interaction.UnimplementedCommentAuthServiceServer
	interaction.UnimplementedCommentPublicServiceServer
	interactionUsecase usecase.InteractionUseCase
}

func NewInteractionGrpc(interactionUsecase usecase.InteractionUseCase) *InteractionGrpc {
	return &InteractionGrpc{
		interactionUsecase: interactionUsecase,
	}
}

func (s *InteractionGrpc) LikeAction(ctx context.Context, in *interaction.LikeRequest) (*common.CommonResponse, error) {
	log.Printf("[LikeAction] target_id=%d, type=%d, status=%v", in.TargetId, in.Type, in.Status)
	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("LikeAction", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}
	err = s.interactionUsecase.LikeAction(uint(userID), uint(in.TargetId), int(in.Type), in.Status)
	if err != nil {
		return FailResponse("LikeAction", common.ErrorCode_PROGRESS_ERROR, err), nil
	}
	return SuccessResponse("LikeAction", nil), nil
}

func (s *InteractionGrpc) GetLikeList(ctx context.Context, in *interaction.LikeListRequest) (*common.CommonResponse, error) {
	log.Printf("[GetLikeList] target_id=%d, type=%d, page=%d, pageSize=%d", in.TargetId, in.Type, in.Page, in.PageSize)

	users, count, err := s.interactionUsecase.GetLikeList(uint(in.TargetId), int(in.Type), int(in.Page), int(in.PageSize))
	if err != nil {
		return FailResponse("GetLikeList", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		UserList:   model.UserListToPb(users),
		TotalCount: int32(count),
	}
	return SuccessResponse("GetLikeList", data), nil
}

func (s *InteractionGrpc) CreateComment(ctx context.Context, in *interaction.CreateCommentRequest) (*common.CommonResponse, error) {
	log.Printf("[CreateComment] video_id=%d, parent_id=%d", in.VideoId, in.ParentId)
	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("CreateComment", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}
	comment, err := s.interactionUsecase.CreateComment(uint(userID), uint(in.VideoId), uint(in.ParentId), in.Content)
	if err != nil {
		return FailResponse("CreateComment", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		CommentInfo: model.CommentToPb(comment),
	}
	return SuccessResponse("CreateComment", data), nil
}

func (s *InteractionGrpc) DeleteComment(ctx context.Context, in *interaction.DeleteCommentRequest) (*common.CommonResponse, error) {
	log.Printf("[DeleteComment] comment_id=%d", in.CommentId)
	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("DeleteComment", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}
	err = s.interactionUsecase.DeleteComment(uint(in.CommentId), uint(userID))
	if err != nil {
		return FailResponse("DeleteComment", common.ErrorCode_OPERATION_FORBIDDEN, err), nil
	}
	return SuccessResponse("DeleteComment", nil), nil
}

func (s *InteractionGrpc) GetCommentList(ctx context.Context, in *interaction.CommentListRequest) (*common.CommonResponse, error) {
	log.Printf("[GetCommentList] video_id=%d, page=%d, pageSize=%d", in.VideoId, in.Page, in.PageSize)
	comments, total, err := s.interactionUsecase.GetCommentList(uint(in.VideoId), int(in.Page), int(in.PageSize))
	if err != nil {
		return FailResponse("GetCommentList", common.ErrorCode_PROGRESS_ERROR, err), nil
	}
	data := &common.Data{
		CommentList: model.CommentListToPb(comments),
		TotalCount:  int32(total),
	}
	return SuccessResponse("GetCommentList", data), nil
}
