package mysql

import (
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/jinzhu/gorm"
)

type FollowRepositoryImpl struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) *FollowRepositoryImpl {
	return &FollowRepositoryImpl{db: db}
}

func (r *FollowRepositoryImpl) Follow(userID, followID uint) error {
	var existing model.Follow
	err := r.db.Unscoped().Where("follower_id = ? AND followed_id = ?", userID, followID).First(&existing).Error
	if err == nil {
		if !existing.DeletedAt.IsZero() {
			return r.db.Unscoped().Model(&existing).Update("deleted_at", nil).Error
		}
		return nil
	}
	follow := &model.Follow{
		FollowerID: userID,
		FollowedID: followID,
	}
	return r.db.Create(follow).Error
}

func (r *FollowRepositoryImpl) Unfollow(userID, followID uint) error {
	return r.db.Where("follower_id = ? AND followed_id = ?", userID, followID).Delete(&model.Follow{}).Error
}

func (r *FollowRepositoryImpl) IsFollowing(userID, followID uint) (bool, error) {
	var count int
	err := r.db.Model(&model.Follow{}).
		Where("follower_id = ? AND followed_id = ?", userID, followID).
		Count(&count).Error
	return count > 0, err
}

func (r *FollowRepositoryImpl) GetFollowList(userID uint, page, pageSize int) ([]*model.Users, error) {
	var users []*model.Users
	offset := (page - 1) * pageSize
	err := r.db.Table("users").
		Select("users.*").
		Joins("INNER JOIN follows ON follows.followed_id = users.id").
		Where("follows.follower_id = ? AND follows.deleted_at IS NULL", userID).
		Order("follows.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error
	return users, err
}

func (r *FollowRepositoryImpl) GetFollowerList(userID uint, page, pageSize int) ([]*model.Users, error) {
	var users []*model.Users
	offset := (page - 1) * pageSize
	err := r.db.Table("users").
		Select("users.*").
		Joins("INNER JOIN follows ON follows.follower_id = users.id").
		Where("follows.followed_id = ? AND follows.deleted_at IS NULL", userID).
		Order("follows.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error
	return users, err
}

func (r *FollowRepositoryImpl) GetFriendList(userID uint, page, pageSize int) ([]*model.Follow, error) {
	var follows []*model.Follow
	offset := (page - 1) * pageSize
	err := r.db.Unscoped().Table("follows f1").
		Select("f1.*").
		Joins("INNER JOIN follows f2 ON f1.followed_id = f2.follower_id AND f2.followed_id = f1.follower_id").
		Where("f1.follower_id = ? AND f1.deleted_at IS NULL AND f2.deleted_at IS NULL", userID).
		Order("f1.created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&follows).Error
	return follows, err
}

func (r *FollowRepositoryImpl) GetUsersFromFollows(follows []*model.Follow, isFollower bool) ([]*model.Users, error) {
	if len(follows) == 0 {
		return nil, nil
	}
	userIDs := make([]uint, 0, len(follows))
	for _, f := range follows {
		if isFollower {
			userIDs = append(userIDs, f.FollowerID)
		} else {
			userIDs = append(userIDs, f.FollowedID)
		}
	}
	var users []*model.Users
	err := r.db.Where("id IN (?)", userIDs).Find(&users).Error
	return users, err
}

func (r *FollowRepositoryImpl) GetFollowCount(userID uint) (int, error) {
	var count int
	err := r.db.Model(&model.Follow{}).Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *FollowRepositoryImpl) GetFollowerCount(userID uint) (int, error) {
	var count int
	err := r.db.Model(&model.Follow{}).Where("followed_id = ?", userID).Count(&count).Error
	return count, err
}
