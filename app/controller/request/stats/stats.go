package stats

import (
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/gin-gonic/gin"
)

type AllStatsRequest struct{}

func (request *AllStatsRequest) ValidateRequest(ctx *gin.Context) *customError.CustomError {
	//
	return nil
}

func (request *AllStatsRequest) ValidateAuth(ctx *gin.Context) *customError.CustomError {
	return nil
}
