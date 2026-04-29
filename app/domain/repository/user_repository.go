package repository

import "github.com/ZOEKOFK/video_web_v3/app/domain/model"

type UserRepository interface {
	GetUserInfoByID(id string) (model.Users, error)
	Save(*model.Users) error
	GetUserInfoByUserName(nickname string) (*model.Users, error)
}
