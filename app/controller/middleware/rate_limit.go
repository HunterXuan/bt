package middleware

import (
	"github.com/HunterXuan/bt/app/infra/config"
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/gin-gonic/gin"
	limiter "github.com/julianshen/gin-limiter"
	"time"
)

func RateLimit() gin.HandlerFunc {
	return limiter.NewRateLimiter(time.Second, config.Config.GetInt64("RATE_LIMITING"), func(ctx *gin.Context) (string, error) {
		key := ctx.ClientIP()
		if key != "" {
			return key, nil
		}

		return "", customError.NewBadRequestError("DEFAULT__RATE_LIMIT")
	}).Middleware()
}
