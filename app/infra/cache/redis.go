package cache

import (
	"fmt"
	"github.com/HunterXuan/bt/app/infra/config"
	"github.com/go-redis/redis/v8"
	"log"
)

var RDB *redis.Client

func InitRedisClient() {
	log.Println("RDB Initializing...")

	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%v:%v", config.Config.GetString("REDIS_HOST"), config.Config.GetString("REDIS_PORT")),
		Password: config.Config.GetString("REDIS_PASS"),
		DB:       config.Config.GetInt("REDIS_DB"),
	})

	log.Println("RDB Initialized!")
}
