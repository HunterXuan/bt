package middleware

import (
	"github.com/HunterXuan/bt/app/infra/config"
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginMiddleware "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	"time"
)

func RateLimit() gin.HandlerFunc {
	return ginMiddleware.NewMiddleware(limiter.New(memory.NewStore(), limiter.Rate{
		Period: 1 * time.Second,
		Limit:  config.Config.GetInt64("RATE_LIMITING"),
	}))
}
