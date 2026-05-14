package repository

import "github.com/ZOEKOFK/video_web_v3/app/domain/model"

type UserRepository interface {
	GetUserInfoByID(id string) (model.Users, error)
	Save(*model.Users) error
	GetUserInfoByUserName(nickname string) (*model.Users, error)
	Update(id, avatar string) error
	GetMfaSecret(userID uint) (string, error)
	UpdateMfaSecret(userID uint, secret string) error
	EnableMfa(userID uint) error
}
