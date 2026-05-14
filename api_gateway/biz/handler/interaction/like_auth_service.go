package interaction

import (
	"context"
	"log"
	"net/http"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/api_gateway/my_jwt"
	commonpb "github.com/ZOEKOFK/video_web_v3/app/pb/common"
	interactionpb "github.com/ZOEKOFK/video_web_v3/app/pb/interaction"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// LikeAction .
// @router /api/likes [POST]
func LikeAction(ctx context.Context, c *app.RequestContext) {
	userID, err := my_jwt.GetUserIDFromToken(ctx, c)
	if err != nil {
		log.Printf("[LikeAction] token提取用户 ID 失败: %v", err)
		c.JSON(consts.StatusUnauthorized, client.FailResponse("invalid token", commonpb.ErrorCode_USER_NOT_LOGIN))
		return
	}
	var req interactionpb.LikeRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[LikeAction] 参数绑定失败: %v", err)
		c.JSON(consts.StatusBadRequest, client.FailResponse(err.Error(), commonpb.ErrorCode_PARAM_ERROR))
		return
	}
	ctxWithUserID := client.WithUserID(ctx, userID)
	resp, err := client.LikeAuthServiceClient.LikeAction(ctxWithUserID, &req)
	if err != nil {
		log.Printf("[LikeAction] gRPC调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, client.FailResponse(err.Error(), commonpb.ErrorCode_PROGRESS_ERROR))
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// GetLikeList .
// @router /api/likes/list [GET]
func GetLikeList(ctx context.Context, c *app.RequestContext) {
	userID, err := my_jwt.GetUserIDFromToken(ctx, c)
	if err != nil {
		log.Printf("[GetLikeList] token提取用户 ID 失败: %v", err)
		c.JSON(consts.StatusUnauthorized, client.FailResponse("invalid token", commonpb.ErrorCode_USER_NOT_LOGIN))
		return
	}
	var req interactionpb.LikeListRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[GetLikeList] 参数绑定失败: %v", err)
		c.JSON(consts.StatusBadRequest, client.FailResponse(err.Error(), commonpb.ErrorCode_PARAM_ERROR))
		return
	}
	ctxWithUserID := client.WithUserID(ctx, userID)
	resp, err := client.LikeAuthServiceClient.GetLikeList(ctxWithUserID, &req)
	if err != nil {
		log.Printf("[GetLikeList] gRPC调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, client.FailResponse(err.Error(), commonpb.ErrorCode_PROGRESS_ERROR))
		return
	}
	c.JSON(consts.StatusOK, resp)
}
