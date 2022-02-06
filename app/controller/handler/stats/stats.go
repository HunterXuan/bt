package stats

import (
	"github.com/HunterXuan/bt/app/controller/request"
	"github.com/HunterXuan/bt/app/controller/request/stats"
	"github.com/HunterXuan/bt/app/controller/response"
	statsService "github.com/HunterXuan/bt/app/domain/service/stats"
	"github.com/HunterXuan/bt/app/infra/util/http"
	"github.com/gin-gonic/gin"
	"log"
)

func GetAllStats(ctx *gin.Context) {
	// 使用 AnnounceRequest 校验请求
	req := &stats.AllStatsRequest{}
	if _, err := request.Validator.Validate(
		ctx,
		req,
	); err != nil {
		log.Println("GetAllStats (Validate):", err)
		ctx.JSON(err.Status, err.Serialize())
		return
	}

	resp, err := statsService.GetAllStats(ctx, req)
	if err != nil {
		log.Println("GetAllStats (GetAllStats):", err)
		ctx.JSON(err.Status, err.Serialize())
		return
	}

	ctx.JSON(http.StatusOK, response.RespSerializer.Serialize(resp, resp))
}
