package grpc

import (
	"context"
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	"github.com/ZOEKOFK/video_web_v3/app/pb/videos"
	"github.com/ZOEKOFK/video_web_v3/app/usecase"
)

type VideosGrpc struct {
	videos.UnimplementedVideoPublicServiceServer
	videos.UnimplementedVideoAuthServiceServer
	videoUsecase usecase.VideoUseCase
}

func NewVideosGrpc(videoUsecase usecase.VideoUseCase) *VideosGrpc {
	return &VideosGrpc{
		videoUsecase: videoUsecase,
	}
}

func (s *VideosGrpc) SearchVideos(ctx context.Context, in *videos.SearchVideoRequest) (*common.CommonResponse, error) {
	videosList, err := s.videoUsecase.SearchVideos(in.Keyword, int(in.Page), int(in.PageSize), in.Sort)
	if err != nil {
		return FailResponse("SearchVideos", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		VideoList: model.VideoListToPb(videosList),
	}
	return SuccessResponse("SearchVideos", data), nil
}

func (s *VideosGrpc) GetHotVideos(ctx context.Context, in *videos.HotVideoRequest) (*common.CommonResponse, error) {
	log.Printf("[GetHotVideos] limit=%d", in.Limit)

	videosList, err := s.videoUsecase.GetHotVideos(int(in.Limit), in.Type, int(in.Page))
	if err != nil {
		return FailResponse("GetHotVideos", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		VideoList: model.VideoListToPb(videosList),
	}
	return SuccessResponse("GetHotVideos", data), nil
}

func (s *VideosGrpc) GetUserVideos(ctx context.Context, in *videos.UserVideoListRequest) (*common.CommonResponse, error) {
	videosList, err := s.videoUsecase.GetUserVideos(uint(in.Id), int(in.Page), int(in.PageSize))
	if err != nil {
		return FailResponse("GetUserVideos", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		VideoList: model.VideoListToPb(videosList),
	}
	return SuccessResponse("GetUserVideos", data), nil
}

func (s *VideosGrpc) UploadVideo(ctx context.Context, in *videos.UploadVideoRequest) (*common.CommonResponse, error) {
	log.Printf("[UploadVideo] title=%s", in.Title)

	userID, err := extractUserIDFromMetadata(ctx)
	if err != nil {
		return FailResponse("UploadVideo", common.ErrorCode_USER_NOT_LOGIN, err), nil
	}

	video, err := s.videoUsecase.UploadVideo(uint(userID), in)
	if err != nil {
		return FailResponse("UploadVideo", common.ErrorCode_PROGRESS_ERROR, err), nil
	}

	data := &common.Data{
		VideoInfo: model.VideoToPb(video),
	}
	return SuccessResponse("UploadVideo", data), nil
}
