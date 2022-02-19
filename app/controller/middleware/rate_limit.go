package middleware

import (
	"github.com/gin-gonic/gin"
	limiter "github.com/julianshen/gin-limiter"
	"time"
)

func RateLimit() gin.HandlerFunc {
	return limiter.NewRateLimiter(time.Second, 5, func(ctx *gin.Context) (string, error) {
		key := ctx.Request.Form.Get("info_hash")
		if key != "" {
			return key, nil
		}

		return "ANONYMOUS", nil
	}).Middleware()
}
