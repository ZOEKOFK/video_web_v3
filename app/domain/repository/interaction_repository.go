package repository

import "github.com/ZOEKOFK/video_web_v3/app/domain/model"

type InteractionRepository interface {
	CreateComment(comment *model.Comment) error
	DeleteComment(commentID, userID uint) error
	GetCommentByID(id uint) (*model.Comment, error)
	GetCommentList(videoID uint, page, pageSize int) ([]*model.Comment, error)
	GetCommentCount(videoID uint) (int, error)
	GetCommentReplies(parentID uint) ([]*model.Comment, error)
	GetCommentsWithUsers(comments []*model.Comment) ([]*model.Comment, error)

	LikeAction(userID, targetID uint, likeType int, status bool) error
	IsLiked(userID, targetID uint, likeType int) (bool, error)
	GetLikeList(targetID uint, likeType int, page, pageSize int) ([]*model.Users, error)
	GetLikeCount(targetID uint, likeType int) (int64, error)

	IncrementCommentLikes(commentID uint) error
	DecrementCommentLikes(commentID uint) error
}
