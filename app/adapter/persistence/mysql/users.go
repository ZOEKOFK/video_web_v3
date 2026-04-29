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
