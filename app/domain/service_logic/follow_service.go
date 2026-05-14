package service_logic

import (
	"errors"

	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
)

type FollowServiceLogic struct {
	repo repository.FollowRepository
}

func NewFollowServiceLogic(repo repository.FollowRepository) *FollowServiceLogic {
	return &FollowServiceLogic{repo: repo}
}

func (s *FollowServiceLogic) CheckSelfFollow(userID, followID uint) error {
	if userID == followID {
		return errors.New("cannot follow/unfollow yourself")
	}
	return nil
}

func (s *FollowServiceLogic) ValidatePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return page, pageSize
}
