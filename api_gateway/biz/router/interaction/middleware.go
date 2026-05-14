package interaction

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

func _commentsMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _createcommentMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _deletecommentMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _likesMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _likeactionMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _getlikelistMw() []app.HandlerFunc {
	return []app.HandlerFunc{my_jwt.AuthMiddleware.MiddlewareFunc()}
}

func _videosMw() []app.HandlerFunc {
	return nil
}

func _video_idMw() []app.HandlerFunc {
	return nil
}

func _getcommentlistMw() []app.HandlerFunc {
	return nil
}
