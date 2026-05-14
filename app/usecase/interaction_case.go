package usecase

import (
	"errors"
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
)

type InteractionUseCase interface {
	CreateComment(userID, videoID, parentID uint, content string) (*model.Comment, error)
	DeleteComment(commentID, userID uint) error
	GetCommentList(videoID uint, page, pageSize int) ([]*model.Comment, int, error)
	LikeAction(userID, targetID uint, likeType int, status bool) error
	GetLikeList(targetID uint, likeType int, page, pageSize int) ([]*model.Users, int64, error)
}

type interactionUsecase struct {
	repo repository.InteractionRepository
}

func NewInteractionUsecase(repo repository.InteractionRepository) InteractionUseCase {
	return &interactionUsecase{repo: repo}
}

func (u *interactionUsecase) CreateComment(userID, videoID, parentID uint, content string) (*model.Comment, error) {
	if content == "" {
		return nil, errors.New("comment content cannot be empty")
	}
	if len(content) > 500 {
		return nil, errors.New("comment content too long")
	}
	comment := &model.Comment{
		UserID:   userID,
		VideoID:  videoID,
		ParentID: parentID,
		Content:  content,
	}
	if err := u.repo.CreateComment(comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (u *interactionUsecase) DeleteComment(commentID, userID uint) error {
	comment, err := u.repo.GetCommentByID(commentID)
	if err != nil {
		return errors.New("comment not found")
	}
	if comment.UserID != userID {
		return errors.New("cannot delete other user's comment")
	}
	return u.repo.DeleteComment(commentID, userID)
}

func (u *interactionUsecase) GetCommentList(videoID uint, page, pageSize int) ([]*model.Comment, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	comments, err := u.repo.GetCommentList(videoID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	total, err := u.repo.GetCommentCount(videoID)

	if err != nil {
		log.Printf("获取评论总数失败: %v", err)
	}
	for _, c := range comments {
		replies, err := u.repo.GetCommentReplies(c.ID)
		if err != nil {
			log.Printf("获取评论 %d 的回复失败: %v", c.ID, err)
			continue
		}
		if len(replies) > 0 {
			repliesWithUsers, err := u.repo.GetCommentsWithUsers(replies)
			if err != nil {
				log.Printf("填充回复用户信息失败: %v", err)
				for _, reply := range replies {
					c.Replies = append(c.Replies, *reply)
				}
			} else {
				for _, reply := range repliesWithUsers {
					c.Replies = append(c.Replies, *reply)
				}
			}
		}
	}
	commentsWithUsers, err := u.repo.GetCommentsWithUsers(comments)
	if err != nil {
		log.Printf("填充用户信息失败: %v", err)
		return comments, total, nil
	}
	return commentsWithUsers, total, nil
}

func (u *interactionUsecase) LikeAction(userID, targetID uint, likeType int, status bool) error {
	log.Println("[LikeAction]", userID, targetID, status)
	if userID == 0 {
		return errors.New("user not login")
	}
	if likeType != 1 && likeType != 2 {
		return errors.New("invalid like type")
	}
	isLiked, err := u.repo.IsLiked(userID, targetID, likeType)
	if err != nil {
		return err
	}
	if status && isLiked {
		return errors.New("already liked")
	}
	if !status && !isLiked {
		return errors.New("not liked yet")
	}
	if err := u.repo.LikeAction(userID, targetID, likeType, status); err != nil {
		return err
	}
	if likeType == 2 {
		if status {
			return u.repo.IncrementCommentLikes(targetID)
		}
		return u.repo.DecrementCommentLikes(targetID)
	}
	return nil
}

func (u *interactionUsecase) GetLikeList(targetID uint, likeType int, page, pageSize int) ([]*model.Users, int64, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	users, err := u.repo.GetLikeList(targetID, likeType, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	count, err := u.repo.GetLikeCount(targetID, likeType)
	if err != nil {
		log.Printf("获取点赞数失败: %v", err)
	}
	return users, count, nil
}
