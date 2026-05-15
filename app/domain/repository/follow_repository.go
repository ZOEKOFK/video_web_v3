package repository

import "github.com/ZOEKOFK/video_web_v3/app/domain/model"

type FollowRepository interface {
	Follow(userID, followID uint) error
	Unfollow(userID, followID uint) error
	IsFollowing(userID, followID uint) (bool, error)
	GetFollowList(userID uint, page, pageSize int) ([]*model.Users, error)
	GetFollowerList(userID uint, page, pageSize int) ([]*model.Users, error)
	GetFriendList(userID uint, page, pageSize int) ([]*model.Follow, error)
	GetFollowCount(userID uint) (int, error)
	GetFollowerCount(userID uint) (int, error)
	GetUsersFromFollows(follows []*model.Follow, isFollower bool) ([]*model.Users, error)
}
