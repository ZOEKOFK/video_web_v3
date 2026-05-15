package mysql

import (
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/jinzhu/gorm"
)

type InteractionRepositoryImpl struct {
	db *gorm.DB
}

func NewInteractionRepository(db *gorm.DB) *InteractionRepositoryImpl {
	return &InteractionRepositoryImpl{db: db}
}

func (r *InteractionRepositoryImpl) CreateComment(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *InteractionRepositoryImpl) DeleteComment(commentID, userID uint) error {
	result := r.db.Where("id = ? AND user_id = ?", commentID, userID).Delete(&model.Comment{})
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	if err := result.Error; err != nil {
		return err
	}
	r.db.Where("parent_id = ?", commentID).Delete(&model.Comment{})
	return nil
}

func (r *InteractionRepositoryImpl) GetCommentByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Where("id = ?", id).First(&comment).Error
	return &comment, err
}

func (r *InteractionRepositoryImpl) GetCommentList(videoID uint, page, pageSize int) ([]*model.Comment, error) {
	var comments []*model.Comment
	offset := (page - 1) * pageSize
	err := r.db.Where("video_id = ? AND parent_id = 0", videoID).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&comments).Error
	return comments, err
}

func (r *InteractionRepositoryImpl) GetCommentCount(videoID uint) (int, error) {
	var count int
	err := r.db.Model(&model.Comment{}).Where("video_id = ? AND parent_id = 0", videoID).Count(&count).Error
	return count, err
}

func (r *InteractionRepositoryImpl) GetCommentReplies(parentID uint) ([]*model.Comment, error) {
	var replies []*model.Comment
	err := r.db.Where("parent_id = ?", parentID).
		Order("created_at ASC").
		Find(&replies).Error
	return replies, err
}

func (r *InteractionRepositoryImpl) GetCommentsWithUsers(comments []*model.Comment) ([]*model.Comment, error) {
	if len(comments) == 0 {
		return comments, nil
	}
	userIDs := make([]uint, 0, len(comments))
	for _, c := range comments {
		userIDs = append(userIDs, c.UserID)
	}
	var users []*model.Users
	if err := r.db.Where("id IN (?)", userIDs).Find(&users).Error; err != nil {
		return nil, err
	}
	userMap := make(map[uint]*model.Users)
	for _, u := range users {
		userMap[u.ID] = u
	}
	for _, c := range comments {
		if u, ok := userMap[c.UserID]; ok {
			c.User = u
		}
	}
	return comments, nil
}

func (r *InteractionRepositoryImpl) LikeAction(userID, targetID uint, likeType int, status bool) error {
	var existing model.Like
	err := r.db.Unscoped().Where("user_id = ? AND target_id = ? AND type = ?", userID, targetID, likeType).First(&existing).Error
	if err == nil {
		if !existing.DeletedAt.IsZero() {
			return r.db.Unscoped().Model(&existing).Update("deleted_at", nil).Error
		}
		if !status {
			return r.db.Delete(&existing).Error
		}
		return nil
	}
	if !status {
		return nil
	}
	like := &model.Like{
		UserID:   userID,
		TargetID: targetID,
		Type:     likeType,
	}
	return r.db.Create(like).Error
}

func (r *InteractionRepositoryImpl) IsLiked(userID, targetID uint, likeType int) (bool, error) {
	var count int
	err := r.db.Model(&model.Like{}).
		Where("user_id = ? AND target_id = ? AND type = ?", userID, targetID, likeType).
		Count(&count).Error
	return count > 0, err
}

func (r *InteractionRepositoryImpl) GetLikeList(targetID uint, likeType int, page, pageSize int) ([]*model.Users, error) {
	var users []*model.Users
	offset := (page - 1) * pageSize
	err := r.db.Table("users").
		Select("users.*").
		Joins("INNER JOIN likes ON likes.user_id = users.id").
		Where("likes.target_id = ? AND likes.type = ? AND likes.deleted_at IS NULL", targetID, likeType).
		Order("likes.created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users).Error
	return users, err
}

func (r *InteractionRepositoryImpl) GetLikeCount(targetID uint, likeType int) (int64, error) {
	var count int64
	err := r.db.Model(&model.Like{}).
		Where("target_id = ? AND type = ? AND deleted_at IS NULL", targetID, likeType).
		Count(&count).Error
	return count, err
}

func (r *InteractionRepositoryImpl) IncrementCommentLikes(commentID uint) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", commentID).UpdateColumn("likes", gorm.Expr("likes + 1")).Error
}

func (r *InteractionRepositoryImpl) DecrementCommentLikes(commentID uint) error {
	return r.db.Model(&model.Comment{}).Where("id = ? AND likes > 0", commentID).UpdateColumn("likes", gorm.Expr("likes - 1")).Error
}
