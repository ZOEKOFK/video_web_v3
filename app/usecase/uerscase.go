package usecase

import (
	"fmt"
	"log"

	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
	"github.com/ZOEKOFK/video_web_v3/app/domain/service_logic"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	"github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"github.com/redis/go-redis/v9"
)

type UserUseCase interface {
	UploadAvatar(request *users.UploadAvatarRequest) (*common.CommonResponse, error)
	RegisterUser(request *users.UserRegisterRequest) (common.ErrorCode, error)
	LoginUser(username, password string, remember bool) (*service_logic.LoginResponse, error)
	RefreshToken(refreshToken string, remember bool) (*service_logic.LoginResponse, error)
	Logout(refreshToken string) error
	GetUserInfo(id string) (*model.Users, error)
	ValidateUserID(userID int64) (*model.Users, error)
}

type userUsecase struct {
	repo    repository.UserRepository
	service *service_logic.UsersServiceLogic
	redis   *redis.Client
}

func NewUserUsecase(repo repository.UserRepository, service *service_logic.UsersServiceLogic, redis *redis.Client) UserUseCase {
	return &userUsecase{
		repo:    repo,
		service: service,
		redis:   redis,
	}
}

func (u *userUsecase) UploadAvatar(request *users.UploadAvatarRequest) (*common.CommonResponse, error) {
	return &common.CommonResponse{}, nil
}

func (u *userUsecase) RegisterUser(req *users.UserRegisterRequest) (common.ErrorCode, error) {
	err := u.service.CheckRegisterInfo(req)
	if err != nil {
		return common.ErrorCode_PARAM_ERROR, err
	}
	encryptedPassword, err := u.service.EncryptPassword(req.Password)
	if err != nil {
		return common.ErrorCode_PROGRESS_ERROR, err
	}
	user := &model.Users{
		Username: req.Username,
		Password: encryptedPassword,
		Nickname: req.Nickname,
	}
	err = u.repo.Save(user)
	if err != nil {
		return common.ErrorCode_PARAM_ERROR, err
	}
	return common.ErrorCode_SUCCESS, nil
}

func (u *userUsecase) LoginUser(username, password string, remember bool) (*service_logic.LoginResponse, error) {
	err := u.service.CheckLoginInfo(username, password)
	if err != nil {
		return nil, err
	}
	user, err := u.repo.GetUserInfoByUserName(username)
	loginResp, err := u.service.BuildLoginResponse(user, remember)
	if err != nil {
		return nil, err
	}
	log.Println("[LoginUser] 登录成功")
	return loginResp, nil
}

func (u *userUsecase) RefreshToken(refreshToken string, remember bool) (*service_logic.LoginResponse, error) {
	log.Println("[RefreshToken] 验证 Refresh Token")
	userID, jti, err := u.service.VerifyRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 删除旧的 refresh token
	log.Println("[RefreshToken] 删除旧 Token")
	u.service.DeleteRefreshToken(userID, jti)

	// 获取用户信息
	user, err := u.GetUserInfo(fmt.Sprintf("%d", userID))
	if err != nil {
		return nil, err
	}

	log.Println("[RefreshToken] 生成新 Token")
	loginResp, err := u.service.BuildLoginResponse(user, remember)
	if err != nil {
		return nil, err
	}

	return loginResp, nil
}

func (u *userUsecase) Logout(refreshToken string) error {
	userID, jti, _ := u.service.VerifyRefreshToken(refreshToken)
	if userID > 0 && jti != "" {
		u.service.DeleteRefreshToken(userID, jti)
	}
	return nil
}

func (u *userUsecase) GetUserInfo(id string) (*model.Users, error) {
	log.Println(id)
	err := u.service.CheckID(id)
	if err != nil {
		fmt.Println("error:", err)
		return nil, err
	}
	user, err := u.repo.GetUserInfoByID(id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ValidateUserID 验证用户 ID 是否有效
func (u *userUsecase) ValidateUserID(userID int64) (*model.Users, error) {
	return u.service.ValidateUserID(userID)
}