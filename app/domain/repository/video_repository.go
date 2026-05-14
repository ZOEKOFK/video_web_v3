package repository

import "github.com/ZOEKOFK/video_web_v3/app/domain/model"

type VideosRepository interface {
	Save(video *model.Videos) error
	GetByID(id uint) (*model.Videos, error)
	GetByUserID(userID uint, page, pageSize int) ([]*model.Videos, error)
	Search(keyword string, page, pageSize int, sort string) ([]*model.Videos, error)
	GetHotVideos(limit int, videoType string, page int) ([]*model.Videos, error)
	Update(video *model.Videos) error
	Delete(id uint) error
}
