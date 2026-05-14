package usecase

import (
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
	"github.com/ZOEKOFK/video_web_v3/app/domain/service_logic"
)

type SocialUseCase interface {
	FollowAction(userID, followID uint, status bool) error
	GetFollowList(userID uint, page, pageSize int) ([]*model.Users, error)
	GetFollowerList(userID uint, page, pageSize int) ([]*model.Users, error)
	GetFriendList(userID uint, page, pageSize int) ([]*model.Users, error)
}

type socialUsecase struct {
	repo    repository.FollowRepository
	service *service_logic.FollowServiceLogic
}

func NewSocialUsecase(repo repository.FollowRepository, service *service_logic.FollowServiceLogic) SocialUseCase {
	return &socialUsecase{
		repo:    repo,
		service: service,
	}
}

func (u *socialUsecase) FollowAction(userID, followID uint, status bool) error {
	if err := u.service.CheckSelfFollow(userID, followID); err != nil {
		return err
	}
	if status {
		err := u.repo.Follow(userID, followID)
		if err != nil {
			isFollowing, checkErr := u.repo.IsFollowing(userID, followID)
			if checkErr == nil && isFollowing {
				return nil
			}
			return err
		}
		return nil
	} else {
		return u.repo.Unfollow(userID, followID)
	}
}

func (u *socialUsecase) GetFollowList(userID uint, page, pageSize int) ([]*model.Users, error) {

	page, pageSize = u.service.ValidatePagination(page, pageSize)
	return u.repo.GetFollowList(userID, page, pageSize)
}

func (u *socialUsecase) GetFollowerList(userID uint, page, pageSize int) ([]*model.Users, error) {
	page, pageSize = u.service.ValidatePagination(page, pageSize)
	return u.repo.GetFollowerList(userID, page, pageSize)
}

func (u *socialUsecase) GetFriendList(userID uint, page, pageSize int) ([]*model.Users, error) {
	page, pageSize = u.service.ValidatePagination(page, pageSize)

	follows, err := u.repo.GetFriendList(userID, page, pageSize)
	if err != nil {
		log.Printf("获取好友列表失败: %v", err)
		return nil, err
	}
	return u.repo.GetUsersFromFollows(follows, false)
}
