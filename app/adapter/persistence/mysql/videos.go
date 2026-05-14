package mysql

import (
	"time"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/jinzhu/gorm"
)

type VideosRepositoryImpl struct {
	db *gorm.DB
}

func NewVideosRepository(db *gorm.DB) *VideosRepositoryImpl {
	return &VideosRepositoryImpl{db: db}
}

func (r *VideosRepositoryImpl) Save(video *model.Videos) error {
	return r.db.Create(video).Error
}

func (r *VideosRepositoryImpl) GetByID(id uint) (*model.Videos, error) {
	var video model.Videos
	err := r.db.Where("id = ?", id).First(&video).Error
	return &video, err
}

func (r *VideosRepositoryImpl) GetByUserID(userID uint, page, pageSize int) ([]*model.Videos, error) {
	var videos []*model.Videos
	offset := (page - 1) * pageSize
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&videos).Error
	return videos, err
}

func (r *VideosRepositoryImpl) Search(keyword string, page, pageSize int, sort string) ([]*model.Videos, error) {
	var videos []*model.Videos
	offset := (page - 1) * pageSize

	query := r.db.Where("title LIKE ? OR description LIKE ?", "%"+keyword+"%", "%"+keyword+"%")

	// 排序
	switch sort {
	case "hot":
		query = query.Order("views DESC")
	case "new":
		query = query.Order("created_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	err := query.Offset(offset).Limit(pageSize).Find(&videos).Error
	return videos, err
}

func (r *VideosRepositoryImpl) GetHotVideos(limit int, videoType string, page int) ([]*model.Videos, error) {
	var videos []*model.Videos
	offset := (page - 1) * limit

	query := r.db.Order("views DESC")

	if videoType != "" {
		var startTime time.Time
		now := time.Now()
		switch videoType {
		case "day":
			startTime = now.Add(-24 * time.Hour)
		case "week":
			startTime = now.Add(-7 * 24 * time.Hour)
		case "month":
			startTime = now.Add(-30 * 24 * time.Hour)
		default:
		}

		if !startTime.IsZero() {
			query = query.Where("created_at >= ?", startTime)
		}
	}

	err := query.Offset(offset).Limit(limit).Find(&videos).Error
	return videos, err
}

func (r *VideosRepositoryImpl) Update(video *model.Videos) error {
	return r.db.Save(video).Error
}

func (r *VideosRepositoryImpl) Delete(id uint) error {
	return r.db.Where("id = ?", id).Delete(&model.Videos{}).Error
}
