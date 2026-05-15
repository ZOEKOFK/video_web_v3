package interaction

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/ZOEKOFK/video_web_v3/api_gateway/client"
	commonpb "github.com/ZOEKOFK/video_web_v3/app/pb/common"
	interactionpb "github.com/ZOEKOFK/video_web_v3/app/pb/interaction"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// GetCommentList .
// @router /api/videos/:video_id/comments [GET]
func GetCommentList(ctx context.Context, c *app.RequestContext) {
	videoIDStr := c.Param("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(consts.StatusBadRequest, client.FailResponse("invalid video_id", commonpb.ErrorCode_PARAM_ERROR))
		return
	}
	var req interactionpb.CommentListRequest
	if err := c.BindAndValidate(&req); err != nil {
		log.Printf("[GetCommentList] 参数绑定失败: %v", err)
		c.JSON(consts.StatusBadRequest, client.FailResponse(err.Error(), commonpb.ErrorCode_PARAM_ERROR))
		return
	}
	req.VideoId = videoID
	log.Printf("[GetCommentList] videoID=%d, page=%d, pageSize=%d", req.VideoId, req.Page, req.PageSize)
	resp, err := client.CommentPublicServiceClient.GetCommentList(ctx, &req)
	if err != nil {
		log.Printf("[GetCommentList] gRPC调用失败: %v", err)
		c.JSON(http.StatusInternalServerError, client.FailResponse(err.Error(), commonpb.ErrorCode_PROGRESS_ERROR))
		return
	}
	c.JSON(consts.StatusOK, resp)
}
