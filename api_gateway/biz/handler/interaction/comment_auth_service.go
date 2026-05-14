package interaction

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	"github.com/ZOEKOFK/video_web_v3/api_gateway/my_jwt"
	commonpb "github.com/ZOEKOFK/video_web_v3/app/pb/common"
	interactionpb "github.com/ZOEKOFK/video_web_v3/app/pb/interaction"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// CreateComment .
// @router /api/comments [POST]
func CreateComment(ctx context.Context, c *app.RequestContext) {
	userID, err := my_jwt.GetUserIDFromToken(ctx, c)
	if err != nil {
		log.Printf("[CreateComment] token提取用户 ID 失败: %v", err)
		c.JSON(consts.StatusUnauthorized, client.FailResponse("invalid token", commonpb.ErrorCode_USER_NOT_LOGIN))
		return
	}
	var req interactionpb.CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[CreateComment] 参数绑定失败: %v", err)
		c.JSON(consts.StatusBadRequest, client.FailResponse(err.Error(), commonpb.ErrorCode_PARAM_ERROR))
		return
	}
	ctxWithUserID := client.WithUserID(ctx, userID)
	resp, err := client.CommentAuthServiceClient.CreateComment(ctxWithUserID, &req)
	if err != nil {
		log.Printf("[CreateComment] gRPC调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, client.FailResponse(err.Error(), commonpb.ErrorCode_PROGRESS_ERROR))
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// DeleteComment .
// @router /api/comments/:comment_id [DELETE]
func DeleteComment(ctx context.Context, c *app.RequestContext) {
	userID, err := my_jwt.GetUserIDFromToken(ctx, c)
	if err != nil {
		log.Printf("[DeleteComment] token提取用户 ID 失败: %v", err)
		c.JSON(consts.StatusUnauthorized, client.FailResponse("invalid token", commonpb.ErrorCode_USER_NOT_LOGIN))
		return
	}
	commentIDStr := c.Param("comment_id")
	var req interactionpb.DeleteCommentRequest
	req.CommentId, _ = parseCommentID(commentIDStr)
	ctxWithUserID := client.WithUserID(ctx, userID)
	resp, err := client.CommentAuthServiceClient.DeleteComment(ctxWithUserID, &req)
	if err != nil {
		log.Printf("[DeleteComment] gRPC调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, client.FailResponse(err.Error(), commonpb.ErrorCode_PROGRESS_ERROR))
		return
	}
	c.JSON(consts.StatusOK, resp)
}

func parseCommentID(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
