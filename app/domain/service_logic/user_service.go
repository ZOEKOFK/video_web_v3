package service_logic

import (
	"errors"

	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
	"github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"golang.org/x/crypto/bcrypt"
)

type UsersServiceLogic struct {
	repo repository.UserRepository
}

func NewUsersServiceLogic(repo repository.UserRepository) *UsersServiceLogic {
	return &UsersServiceLogic{repo: repo}
}

func (u *UsersServiceLogic) CheckID(id string) error {

	if id == "" || id == "0" {
		return errors.New("id is invalid")
	}
	return nil
}

func (u *UsersServiceLogic) CheckLoginInfo(username string, password string) error {
	if username == "" || password == "" {
		return errors.New("username or password is empty")
	}
	if len(password) < 8 {
		return errors.New("password is too short")
	}
	user, err := u.repo.GetUserInfoByUserName(username)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		err = errors.New("password is wrong")
		return err
	}
	return nil
}

func (u *UsersServiceLogic) CheckRegisterInfo(request *users.UserRegisterRequest) error {
	if request.Username == "" || request.Password == "" {
		return errors.New("username or password is empty")
	}
	if len(request.Password) < 8 {
		return errors.New("password is too short")
	}
	user, err := u.repo.GetUserInfoByUserName(request.Username)
	if err != nil {
		return err
	}
	if user.Username != "" {
		return errors.New("username is exist")
	}
	return nil
}

func (u *UsersServiceLogic) EncryptPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		err := errors.New("encrypt password fail:" + err.Error())
		return "", err
	}
	return string(hashedPassword), nil
}
