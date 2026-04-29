package service_logic

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/ZOEKOFK/video_web_v3/app/adapter/persistence/redis"
	"github.com/ZOEKOFK/video_web_v3/app/domain/model"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	jwtSecret   = []byte("this-is-a-secret")
	identityKey = "ID"
)

const (
	AccessTokenTTL          = 15 * time.Minute
	DefaultRefreshTokenTTL  = 7 * 24 * time.Hour
	RememberRefreshTokenTTL = 30 * 24 * time.Hour
	TokenTypeAccess         = "access"
	TokenTypeRefresh        = "refresh"
)

// LoginResponse 登录响应
type LoginResponse struct {
	UserInfo  *common.User
	TokenInfo *common.Token
}

func (u *UsersServiceLogic) VerifyPassword(username, password string) (*model.Users, error) {
	log.Println("[VerifyPassword] 查询用户:", username)
	user, err := u.repo.GetUserInfoByUserName(username)
	if err != nil {
		log.Println("[VerifyPassword] 用户不存在:", err)
		return nil, errors.New("user not found")
	}
	log.Println("[VerifyPassword] 验证密码")
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Println("[VerifyPassword] 密码错误:", err)
		return nil, errors.New("password error")
	}
	return user, nil
}

// ------------------------------ Token 相关方法 ------------------------------

func parseUserID(value interface{}) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint64:
		return int64(v), true
	case float64:
		return int64(v), true
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, false
		}
		return id, true
	default:
		return 0, false
	}
}

func refreshTokenKey(userID int64, jti string) string {
	return fmt.Sprintf("refresh:token:%d:%s", userID, jti)
}

func refreshTTL(remember bool) time.Duration {
	if remember {
		return RememberRefreshTokenTTL
	}
	return DefaultRefreshTokenTTL
}

// GenerateAccessToken 生成 Access Token
func (u *UsersServiceLogic) GenerateAccessToken(userID int64) (string, time.Time, error) {
	expire := time.Now().Add(AccessTokenTTL)
	claims := jwtv4.MapClaims{
		identityKey:  userID,
		"permission": "all",
		"typ":        TokenTypeAccess,
		"exp":        expire.Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expire, nil
}

// GenerateRefreshToken 生成 Refresh Token
func (u *UsersServiceLogic) GenerateRefreshToken(userID int64, remember bool) (string, string, time.Time, error) {
	expire := time.Now().Add(refreshTTL(remember))
	jti := uuid.NewString()
	claims := jwtv4.MapClaims{
		identityKey: userID,
		"typ":       TokenTypeRefresh,
		"jti":       jti,
		"exp":       expire.Unix(),
		"iat":       time.Now().Unix(),
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", "", time.Time{}, err
	}
	return tokenString, jti, expire, nil
}

// StoreRefreshToken 存储 Refresh Token 到 Redis
func (u *UsersServiceLogic) StoreRefreshToken(userID int64, jti string, expire time.Time) error {
	ttl := time.Until(expire)
	if ttl <= 0 {
		return errors.New("refresh token expired")
	}
	return redis.SetWithExpiration(refreshTokenKey(userID, jti), "1", ttl)
}

// DeleteRefreshToken 从 Redis 删除 Refresh Token
func (u *UsersServiceLogic) DeleteRefreshToken(userID int64, jti string) error {
	return redis.Delete(refreshTokenKey(userID, jti))
}

// VerifyRefreshToken 验证 Refresh Token
func (u *UsersServiceLogic) VerifyRefreshToken(tokenString string) (int64, string, error) {
	userID, jti, _, err := u.parseRefreshToken(tokenString)
	if err != nil {
		return 0, "", err
	}
	exists, err := redis.Exists(refreshTokenKey(userID, jti))
	if err != nil {
		return 0, "", err
	}
	if !exists {
		return 0, "", errors.New("refresh token not found or expired")
	}
	return userID, jti, nil
}

func (u *UsersServiceLogic) parseRefreshToken(tokenString string) (int64, string, time.Time, error) {
	token, err := jwtv4.Parse(tokenString, func(token *jwtv4.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwtv4.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, "", time.Time{}, errors.New("invalid refresh token")
	}
	claims, ok := token.Claims.(jwtv4.MapClaims)
	if !ok {
		return 0, "", time.Time{}, errors.New("invalid token claims")
	}
	tokenType, ok := claims["typ"].(string)
	if !ok || tokenType != TokenTypeRefresh {
		return 0, "", time.Time{}, errors.New("invalid token type")
	}
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return 0, "", time.Time{}, errors.New("invalid token jti")
	}
	userID, ok := parseUserID(claims[identityKey])
	if !ok {
		return 0, "", time.Time{}, errors.New("invalid token user")
	}
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return 0, "", time.Time{}, errors.New("invalid token expire")
	}
	expire := time.Unix(int64(expFloat), 0)
	return userID, jti, expire, nil
}
//func (u *UsersServiceLogic) ParseAccessToken(id string) error {
//
//}

func (u *UsersServiceLogic) BuildLoginResponse(user *model.Users, remember bool) (*LoginResponse, error) {
	userID := int64(user.ID)

	pbUser := &common.User{
		Id:        userID,
		Username:  user.Username,
		AvatarUrl: user.Avatarurl,
		Nickname:  user.Nickname,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	accessToken, accessExpire, err := u.GenerateAccessToken(userID)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshJti, refreshExpire, err := u.GenerateRefreshToken(userID, remember)
	if err != nil {
		return nil, err
	}
	if err = u.StoreRefreshToken(userID, refreshJti, refreshExpire); err != nil {
		return nil, err
	}

	return &LoginResponse{
		UserInfo: pbUser,
		TokenInfo: &common.Token{
			AccessToken:   accessToken,
			AccessExpire:  accessExpire.Format("2006-01-02 15:04:05"),
			RefreshToken:  refreshToken,
			RefreshExpire: refreshExpire.Format("2006-01-02 15:04:05"),
		},
	}, nil
}
