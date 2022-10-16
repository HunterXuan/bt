package job

import (
	"context"
	"fmt"
	"github.com/HunterXuan/bt/app/domain/service"
	"github.com/HunterXuan/bt/app/infra/constants"
	"github.com/HunterXuan/bt/app/infra/db"
	"github.com/go-redis/redis/v8"
	"log"
	"strings"
	"time"
)

type Cleaner struct{}

const ActiveTorrentTtl = 24 * time.Hour
const ActivePeerTtl = 6 * time.Hour

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

	log.Println("Cleaner finish working")
}
