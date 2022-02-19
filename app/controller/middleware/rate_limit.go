package middleware

import (
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/gin-gonic/gin"
	limiter "github.com/julianshen/gin-limiter"
	"time"
)

func RateLimit() gin.HandlerFunc {
	return limiter.NewRateLimiter(time.Second, 5, func(ctx *gin.Context) (string, error) {
		key := ctx.ClientIP()
		if key != "" {
			return key, nil
		}

		return "", customError.NewBadRequestError("DEFAULT__RATE_LIMIT")
	}).Middleware()
}
