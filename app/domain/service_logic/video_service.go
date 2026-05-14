package service_logic

import (
	"errors"
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
)

var supportedVideoFormats = map[string]bool{
	"mp4": true, "avi": true, "mov": true, "wmv": true,
	"flv": true, "mkv": true, "webm": true,
}

type VideosServiceLogic struct {
	repo repository.VideosRepository
}

func NewVideosServiceLogic(repo repository.VideosRepository) *VideosServiceLogic {
	return &VideosServiceLogic{repo: repo}
}

func (s *VideosServiceLogic) CheckVideoInfo(title string) error {
	if title == "" {
		return errors.New("title is required")
	}
	if len(title) > 100 {
		return errors.New("title is too long")
	}
	return nil
}

func (s *VideosServiceLogic) ValidateVideoFormat(ext string) bool {
	ext = trimDot(ext)
	return supportedVideoFormats[ext]
}

func (s *VideosServiceLogic) CreateVideo(userID uint, title, description, videoUrl string) (*model.Videos, error) {
	err := s.CheckVideoInfo(title)
	if err != nil {
		return nil, err
	}
	video := &model.Videos{
		UserID:      userID,
		Title:       title,
		Description: description,
		VideoUrl:    videoUrl,
	}
	err = s.repo.Save(video)
	if err != nil {
		log.Printf("[CreateVideo] 保存视频失败: %v", err)
		return nil, err
	}
	return video, nil
}

func (s *VideosServiceLogic) GetVideoInfo(id uint) (*model.Videos, error) {
	video, err := s.repo.GetByID(id)
	if err != nil {
		log.Printf("[GetVideoInfo] 获取视频失败: %v", err)
		return nil, errors.New("video not found")
	}
	return video, nil
}

func (s *VideosServiceLogic) GetUserVideos(userID uint, page, pageSize int) ([]*model.Videos, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	videos, err := s.repo.GetByUserID(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return videos, nil
}

func (s *VideosServiceLogic) SearchVideos(keyword string, page, pageSize int, sort string) ([]*model.Videos, error) {
	if keyword == "" {
		return nil, errors.New("keyword is required")
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	videos, err := s.repo.Search(keyword, page, pageSize, sort)
	if err != nil {
		log.Printf("[SearchVideos] 搜索视频失败: %v", err)
		return nil, err
	}
	return videos, nil
}

func (s *VideosServiceLogic) GetHotVideos(limit int, videoType string, page int) ([]*model.Videos, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	videos, err := s.repo.GetHotVideos(limit, videoType, page)
	if err != nil {
		log.Printf("[GetHotVideos] 获取热门视频失败: %v", err)
		return nil, err
	}
	return videos, nil
}

func (s *VideosServiceLogic) IncrementViews(videoID uint) error {
	video, err := s.repo.GetByID(videoID)
	if err != nil {
		return err
	}
	video.Views++
	return s.repo.Update(video)
}

func (s *VideosServiceLogic) IncrementLikes(videoID uint) error {
	video, err := s.repo.GetByID(videoID)
	if err != nil {
		return err
	}
	video.Likes++
	return s.repo.Update(video)
}

func trimDot(s string) string {
	if len(s) > 0 && s[0] == '.' {
		return s[1:]
	}
	return s
}
