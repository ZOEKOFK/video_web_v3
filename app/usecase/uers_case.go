package usecase

import (
	"fmt"
	"log"
	"time"

	my_minio "github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/minio"
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/domain/repository"
	"github.com/ZOEKOFK/video_web_v3/app/domain/service_logic"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	"github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
)

type UserUseCase interface {
	UploadAvatar(request *users.UploadAvatarRequest) (*model.Users, error)
	RegisterUser(request *users.UserRegisterRequest) (common.ErrorCode, error)
	LoginUser(username, password string, remember bool) (*service_logic.LoginResponse, error)
	RefreshToken(refreshToken string, remember bool) (*service_logic.LoginResponse, error)
	Logout(refreshToken string) error
	GetUserInfo(id string) (*model.Users, error)
	GetMFACode(userID uint) (secret string, err error)
	BindMFA(userID uint, code string) error
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

func (u *userUsecase) UploadAvatar(request *users.UploadAvatarRequest) (*model.Users, error) {
	err := u.service.CheckID(request.UserId)
	if err != nil {
		return nil, err
	}
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], request.FileExtension)
	minioFilePath := "/picture/" + filename
	SqlFilePath := my_minio.BucketName + filename
	oldUser, err := u.repo.GetUserInfoByID(request.UserId)
	err = my_minio.Delete(oldUser.Avatarurl)
	if err != nil {
		return nil, err
	}
	err = my_minio.Save(minioFilePath, request.File)
	if err != nil {
		return nil, err
	}
	err = u.repo.Update(request.UserId, SqlFilePath)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	user, err := u.repo.GetUserInfoByID(request.UserId)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return &user, nil
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

	log.Println("[RefreshToken] 删除旧 Token")
	u.service.DeleteRefreshToken(userID, jti)

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

func (u *userUsecase) GetMFACode(userID uint) (secret string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "VideoWeb",
		AccountName: fmt.Sprintf("user:%d", userID),
	})
	if err != nil {
		return "", fmt.Errorf("生成 MFA 密钥失败: %w", err)
	}

	if err = u.repo.UpdateMfaSecret(userID, key.Secret()); err != nil {
		return "", fmt.Errorf("保存 MFA 密钥失败: %w", err)
	}

	log.Printf("[GetMFACode] userID=%d 已生成 MFA 密钥", userID)
	return key.Secret(), nil
}

func (u *userUsecase) BindMFA(userID uint, code string) error {
	secret, err := u.repo.GetMfaSecret(userID)
	if err != nil || secret == "" {
		return fmt.Errorf("未找到 MFA 密钥，请先获取密钥")
	}

	valid := totp.Validate(code, secret)
	if !valid {
		return fmt.Errorf("验证码无效或已过期")
	}

	if err := u.repo.EnableMfa(userID); err != nil {
		return fmt.Errorf("启用 MFA 失败: %w", err)
	}

	log.Printf("[BindMFA] userID=%d 已成功绑定 MFA", userID)
	return nil
}
