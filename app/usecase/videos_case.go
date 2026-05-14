package usecase

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	my_minio "github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/minio"
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
	"github.com/ZOEKOFK/video_web_v3/app/domain/service_logic"
	"github.com/ZOEKOFK/video_web_v3/app/pb/videos"
	"github.com/google/uuid"
)

type VideoUseCase interface {
	UploadVideo(userID uint, req *videos.UploadVideoRequest) (*model.Videos, error)
	GetVideoInfo(id string) (*model.Videos, error)
	GetUserVideos(userID uint, page, pageSize int) ([]*model.Videos, error)
	SearchVideos(keyword string, page, pageSize int, sort string) ([]*model.Videos, error)
	GetHotVideos(limit int, videoType string, page int) ([]*model.Videos, error)
	IncrementViews(videoID string) error
	IncrementLikes(videoID string) error
}

type videoUsecase struct {
	repo    repository.VideosRepository
	service *service_logic.VideosServiceLogic
}

func NewVideoUsecase(repo repository.VideosRepository, service *service_logic.VideosServiceLogic) VideoUseCase {
	return &videoUsecase{
		repo:    repo,
		service: service,
	}
}

func (u *videoUsecase) UploadVideo(userID uint, req *videos.UploadVideoRequest) (*model.Videos, error) {
	log.Printf("[UploadVideo] 用户 %d 上传视频: title=%s", userID, req.Title)

	if len(req.File) == 0 {
		return nil, errors.New("video file is empty")
	}

	ext := req.FileExtension
	if ext == "" {
		return nil, errors.New("file extension is required")
	}

	if !u.service.ValidateVideoFormat(ext) {
		return nil, fmt.Errorf("unsupported video format: %s", ext)
	}

	filename := fmt.Sprintf("%d_%s.%s", time.Now().Unix(), uuid.New().String()[:8], ext)
	objectName := "videos/" + filename

	videoURL, err := my_minio.SaveVideo(objectName, req.File, ext)
	if err != nil {
		log.Printf("MinIO 上传失败: %v", err)
		return nil, err
	}

	video, err := u.service.CreateVideo(userID, req.Title, req.Description, videoURL)
	if err != nil {
		log.Printf("保存视频记录失败: %v", err)
		return nil, err
	}
	return video, nil
}

func (u *videoUsecase) GetVideoInfo(id string) (*model.Videos, error) {
	videoID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}
	video, err := u.service.GetVideoInfo(uint(videoID))
	if err != nil {
		return nil, err
	}
	_ = u.service.IncrementViews(uint(videoID))
	return video, nil
}

func (u *videoUsecase) GetUserVideos(userID uint, page, pageSize int) ([]*model.Videos, error) {
	return u.service.GetUserVideos(userID, page, pageSize)
}

func (u *videoUsecase) SearchVideos(keyword string, page, pageSize int, sort string) ([]*model.Videos, error) {
	log.Printf("[SearchVideos] 搜索视频: keyword=%s, page=%d, pageSize=%d, sort=%s", keyword, page, pageSize, sort)
	return u.service.SearchVideos(keyword, page, pageSize, sort)
}

func (u *videoUsecase) GetHotVideos(limit int, videoType string, page int) ([]*model.Videos, error) {
	log.Printf("[GetHotVideos] 获取热门视频: limit=%d, type=%s, page=%d", limit, videoType, page)
	return u.service.GetHotVideos(limit, videoType, page)
}

func (u *videoUsecase) IncrementViews(videoID string) error {
	id, err := strconv.ParseUint(videoID, 10, 32)
	if err != nil {
		return err
	}
	return u.service.IncrementViews(uint(id))
}

func (u *videoUsecase) IncrementLikes(videoID string) error {
	id, err := strconv.ParseUint(videoID, 10, 32)
	if err != nil {
		return err
	}
	return u.service.IncrementLikes(uint(id))
}
