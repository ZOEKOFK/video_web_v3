package mysql

import (
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"

	"github.com/jinzhu/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetUserInfoByID(id string) (model.Users, error) {
	user := model.Users{}
	r.db.Table("users").Where("id=?", id).First(&user)
	return user, nil
}

func (r *userRepository) Save(users *model.Users) error {
	return r.db.Table(users.TableName()).Create(&users).Error
}

func (r *userRepository) GetUserInfoByUserName(nickname string) (*model.Users, error) {
	user := model.Users{}
	r.db.Table("users").Where("username=?", nickname).First(&user)
	return &user, nil
}

func (r *userRepository) Update(id, avatarURL string) error {
	return r.db.Table("users").Where("id = ?", id).Update("avatar_url", avatarURL).Error
}

func (r *userRepository) GetMfaSecret(userID uint) (string, error) {
	var secret struct {
		MfaSecret string
	}
	err := r.db.Table("users").Select("mfa_secret").Where("id = ?", userID).First(&secret).Error
	return secret.MfaSecret, err
}

func (r *userRepository) UpdateMfaSecret(userID uint, secret string) error {
	return r.db.Table("users").Where("id = ?", userID).Update("mfa_secret", secret).Error
}

func (r *userRepository) EnableMfa(userID uint) error {
	return r.db.Table("users").Where("id = ?", userID).Update("mfa_enabled", true).Error
}

// func (r *userRepository) DisableMfa(userID uint) error {
// 	return r.db.Table("users").Where("id = ?", userID).Update("mfa_enabled", false).Error
// }
