package job

import (
	"context"
	"fmt"
	"github.com/HunterXuan/bt/app/domain/service"
	"github.com/HunterXuan/bt/app/infra/constants"
	"github.com/HunterXuan/bt/app/infra/db"
	"github.com/go-redis/redis/v8"
	"log"
	"strconv"
	"strings"
	"time"
)

type Cleaner struct{}

const ActiveTorrentTtl = 12 * time.Hour
const ActivePeerTtl = 4 * time.Hour

func (s *Cleaner) Run() {
	log.Println("Cleaner start working")

	ctx := context.Background()

	deadTorrentHashes, err := db.RDB.ZRangeByScore(ctx, constants.ActiveTorrentSetKey, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%v", time.Now().Add(-ActiveTorrentTtl).Unix()),
		Count: 100,
	}).Result()
	if err == nil {
		for _, infoHash := range deadTorrentHashes {
			db.RDB.HDel(ctx, service.GenTorrentInfoKey(infoHash))
			if err := db.RDB.HGetAll(ctx, service.GenTorrentInfoKey(infoHash)).Err(); err == nil || err == redis.Nil {
				db.RDB.ZRem(ctx, constants.ActiveTorrentSetKey, &redis.Z{
					Member: infoHash,
				})

				db.RDB.HDel(ctx, service.GenPeerKey(infoHash))
			}
		}
	}

	deadPeers, err := db.RDB.ZRangeByScore(ctx, constants.ActivePeerSetKey, &redis.ZRangeBy{
		Min:   "-inf",
		Max:   fmt.Sprintf("%v", time.Now().Add(-ActivePeerTtl).Unix()),
		Count: 100,
	}).Result()
	if err == nil {
		for _, deadPeer := range deadPeers {
			parts := strings.Split(deadPeer, ":")
			if len(parts) == 2 {
				db.RDB.HDel(ctx, service.GenPeerKey(parts[0]), parts[1])
			}

			db.RDB.ZRem(ctx, constants.ActivePeerSetKey, &redis.Z{
				Member: deadPeer,
			})
		}
	}

	var usedMemory, totalSystemMemory float64
	memoryInfo, err := db.RDB.Info(ctx, "MEMORY").Result()
	if err == nil {
		lines := strings.Split(memoryInfo, "\n")
		for _, line := range lines {
			parts := strings.Split(line, ":")
			if len(parts) != 2 {
				continue
			}

			if parts[0] == "used_memory" {
				usedMemory, _ = strconv.ParseFloat(parts[1], 64)
			}

			if parts[0] == "total_system_memory" {
				totalSystemMemory, _ = strconv.ParseFloat(parts[1], 64)
			}
		}

		maximumValue := totalSystemMemory * 0.8
		if usedMemory > 0 && totalSystemMemory > 0 && usedMemory > maximumValue {
			log.Printf("since used memory (%v) exceed the maximum value (%v), start flushing...", usedMemory, maximumValue)
			db.RDB.FlushDB(ctx)
		}
	}

	log.Println("Cleaner finish working")
}
