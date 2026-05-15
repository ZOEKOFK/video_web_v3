package users

import (
	"context"
	"log"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/app/pb/common"
	userspb "github.com/ZOEKOFK/video_web_v3/app/pb/users"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// RefreshSession .
// @router /api/sessions/refresh [POST]
func RefreshSession(ctx context.Context, c *app.RequestContext) {
	var req userspb.RefreshTokenRequest
	err := c.BindJSON(&req)
	if err != nil {
		log.Println("[RefreshSession] 绑定请求失败:", err)
		c.JSON(consts.StatusOK, client.FailResponse("Invalid request", common.ErrorCode_PARAM_ERROR))
		return
	}
	resp, err := client.UserSessionServiceClient.RefreshSession(ctx, &req)
	if err != nil {
		log.Println("[RefreshSession] gRPC 调用失败:", err)
		c.JSON(consts.StatusOK, client.FailResponse("Refresh failed", common.ErrorCode_PROGRESS_ERROR))
		return
	}

	log.Println("[RefreshSession] gRPC 响应:", resp)
	c.JSON(consts.StatusOK, resp)
}

// Logout .
// @router /api/sessions [DELETE]
func Logout(ctx context.Context, c *app.RequestContext) {
	var req userspb.RefreshTokenRequest
	if err := c.Bind(&req); err != nil {
		log.Println("[Logout] 绑定请求失败:", err)
		c.JSON(consts.StatusOK, client.FailResponse("Invalid request", common.ErrorCode_PARAM_ERROR))
		return
	}

	log.Println("[Logout] 调用 gRPC 服务")
	resp, err := client.UserSessionServiceClient.Logout(ctx, &req)
	if err != nil {
		log.Println("[Logout] gRPC 调用失败:", err)
		c.JSON(consts.StatusOK, client.FailResponse("Logout failed", common.ErrorCode_PROGRESS_ERROR))
		return
	}

	log.Println("[Logout] gRPC 响应:", resp)
	c.JSON(consts.StatusOK, resp)
}
