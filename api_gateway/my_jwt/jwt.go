package my_jwt

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/hertz-contrib/jwt"
)

var (
	identityKey    = "ID"
	AuthMiddleware *jwt.HertzJWTMiddleware
	jwtSecret      = []byte("this-is-a-secret")
)

const (
	accessTokenTTL  = 15 * time.Minute
	tokenTypeAccess = "access"
)

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

func InitJWT() error {
	var err error

	AuthMiddleware, err = jwt.New(&jwt.HertzJWTMiddleware{
		Realm:         "test zone",
		Key:           jwtSecret,
		Timeout:       accessTokenTTL,
		MaxRefresh:    accessTokenTTL,
		IdentityKey:   identityKey,
		TokenLookup:   "header: Authorization, query: token",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,

		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*common.CommonResponse); ok {
				return jwt.MapClaims{
					identityKey:  v.Data.UserInfo.Id,
					"permission": "all",
					"typ":        tokenTypeAccess,
				}
			}
			return jwt.MapClaims{}
		},

		LoginResponse: func(ctx context.Context, c *app.RequestContext, code int, message string, expire time.Time) {
		},

		Authenticator: func(ctx context.Context, c *app.RequestContext) (interface{}, error) {
			return nil, nil
		},

		IdentityHandler: func(ctx context.Context, c *app.RequestContext) interface{} {
			claims := jwt.ExtractClaims(ctx, c)
			if permission, ok := claims["permission"].(string); ok {
				return permission
			}
			return ""
		},

		Authorizator: func(data interface{}, ctx context.Context, c *app.RequestContext) bool {
			permission, ok := data.(string)
			if !ok {
				return false
			}
			return permission == "all"
		},

		Unauthorized: func(ctx context.Context, c *app.RequestContext, code int, message string) {
			c.JSON(401, client.FailResponse(message, common.ErrorCode_USER_NOT_LOGIN))
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func GetUserIDFromToken(ctx context.Context, c *app.RequestContext) (int64, error) {
	claims := jwt.ExtractClaims(ctx, c)
	idValue, ok := claims[identityKey]
	if !ok {
		return 0, ErrUserIDMissing
	}
	userID, ok := parseUserID(idValue)
	if !ok {
		return 0, ErrInvalidUserID
	}
	return userID, nil
}

func CompareUserID(tokenID int64, userID string) bool {
	i, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		log.Println("string to int64 转化失败")
	}
	return i == tokenID
}

var (
	ErrUserIDMissing = jwt.ErrFailedAuthentication
	ErrInvalidUserID = jwt.ErrFailedAuthentication
)
