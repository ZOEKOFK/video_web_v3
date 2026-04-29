package client

import "github.com/ZOEKOFK/video_web_v3/app/pb/common"

//因gateway网关执行过程的错误

func FailResponse(msg string, code common.ErrorCode) *common.CommonResponse {
	return &common.CommonResponse{
		Code:    code,
		Message: msg,
		Data:    nil,
	}
}
