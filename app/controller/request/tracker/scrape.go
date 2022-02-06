package tracker

import (
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	trackerUtil "github.com/HunterXuan/bt/app/infra/util/tracker"
	"github.com/gin-gonic/gin"
)

type ScrapeRequest struct {
	InfoHashSlice []string `form:"info_hash" binding:"required"` // 唯一哈希码
}

func (request *ScrapeRequest) ValidateRequest(ctx *gin.Context) *customError.CustomError {
	// TODO: 客户端白名单校验

	// 校验参数格式
	if err := ctx.ShouldBind(request); err != nil {
		return customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	{
		// 转换 InfoHash
		var infoHashSlice []string
		for _, infoHash := range request.InfoHashSlice {
			infoHashSlice = append(infoHashSlice, trackerUtil.RestoreToHexString(infoHash))
		}
		request.InfoHashSlice = infoHashSlice
	}

	return nil
}

func (request *ScrapeRequest) ValidateAuth(ctx *gin.Context) *customError.CustomError {
	return nil
}
