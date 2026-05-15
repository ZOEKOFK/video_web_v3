package social

import (
	"github.com/ZOEKOFK/video_web_v3/api_gateway/my_jwt"
	"github.com/cloudwego/hertz/pkg/app"
)

func rootMw() []app.HandlerFunc {
	return nil
}

func _apiMw() []app.HandlerFunc {
	return nil
}

func _followactionMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _usersMw() []app.HandlerFunc {
	return nil
}

func _getfriendlistMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _user_idMw() []app.HandlerFunc {
	return nil
}

func _getfollowerlistMw() []app.HandlerFunc {
	return nil
}

func _getfollowlistMw() []app.HandlerFunc {
	return nil
}
