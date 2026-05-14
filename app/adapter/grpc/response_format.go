package grpc

import "github.com/ZOEKOFK/video_web_v3/app/pb/common"

func SuccessResponse(keyWord string, response *common.Data) *common.CommonResponse {
	return &common.CommonResponse{
		Code:    common.ErrorCode_SUCCESS,
		Message: keyWord + " success",
		Data:    response,
	}
}

func FailResponse(keyWord string, code common.ErrorCode, err error) *common.CommonResponse {
	return &common.CommonResponse{
		Code:    code,
		Message: keyWord + ": " + err.Error(),
		Data:    nil,
	}
}
