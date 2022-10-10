package stats

import (
	"context"
	"encoding/json"
	statsReq "github.com/HunterXuan/bt/app/controller/request/stats"
	statsResp "github.com/HunterXuan/bt/app/controller/response/stats"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/domain/service"
	"github.com/HunterXuan/bt/app/infra/constants"
	"github.com/HunterXuan/bt/app/infra/db"
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/gin-gonic/gin"
	"log"
)

// GetAllStats 获取统计数据
func GetAllStats(ctx *gin.Context, req *statsReq.AllStatsRequest) (*statsResp.AllStatsResponse, *customError.CustomError) {
	if stats, err := getStatsFromCache(ctx); stats != nil && err == nil {
		return stats, nil
	}

	return nil, customError.NewBadRequestError("STATS__INVALID_PARAMS")
}

// UpdateStatsCache 更新统计数据缓存
func UpdateStatsCache() error {
	ctx := context.Background()
	stats := &statsResp.AllStatsResponse{
		Index: statsResp.IndexItem{
			Torrent: getTorrentCount(ctx),
			Peer:    getPeerCount(ctx),
			Traffic: getTrafficCount(ctx),
		},
		Hot: getHotStats(ctx),
	}

	return setStatsToCache(context.Background(), stats)
}

func getStatsFromCache(ctx *gin.Context) (*statsResp.AllStatsResponse, error) {
	var stats *statsResp.AllStatsResponse

	val, err := db.RDB.Get(ctx, constants.TrackerStatsCacheKey).Result()
	if err != nil {
		log.Println("getStatsFromCache Err:", err)
		return nil, err
	}

	if err := json.Unmarshal([]byte(val), &stats); err != nil {
		log.Println("getTorrentFromCache Err:", err)
		return nil, err
	}

	return stats, nil
}

func setStatsToCache(ctx context.Context, stats *statsResp.AllStatsResponse) error {
	bytes, err := json.Marshal(stats)
	if err != nil {
		log.Println("setStatsToCache Err:", err)
		return err
	}

	_, err = db.RDB.Set(ctx, constants.TrackerStatsCacheKey, bytes, 0).Result()
	if err != nil {
		log.Println("setStatsToCache Err:", err)
	}

	return err
}

func getTorrentCount(ctx context.Context) uint64 {
	keys, err := db.RDB.Keys(ctx, constants.TrackerTorrentCountPattern).Result()
	if err != nil {
		cacheCount, _ := db.RDB.Get(ctx, constants.TrackerTorrentStatsKey).Uint64()
		return cacheCount
	}

	realCount := uint64(len(keys))
	db.RDB.Set(ctx, constants.TrackerTorrentStatsKey, realCount, 0)

	return realCount
}

func getPeerCount(ctx context.Context) uint64 {
	keys, err := db.RDB.Keys(ctx, constants.TrackerPeerCountPattern).Result()
	if err != nil {
		cacheCount, _ := db.RDB.Get(ctx, constants.TrackerPeerStatsKey).Uint64()
		return cacheCount
	}

	realCount := uint64(len(keys))
	db.RDB.Set(ctx, constants.TrackerPeerStatsKey, realCount, 0)

	return realCount
}

func getTrafficCount(ctx context.Context) uint64 {
	val, _ := db.RDB.Get(ctx, constants.TrackerTrafficStatsKey).Uint64()
	return val
}

func getHotStats(ctx context.Context) []statsResp.HotTorrentItem {
	hotInfoHashes, err := db.RDB.ZRange(ctx, constants.TorrentHotSetKey, 0, constants.TorrentHotSetCapacity).Result()
	if err != nil {
		return nil
	}

	var torrents []statsResp.HotTorrentItem
	for _, infoHash := range hotInfoHashes {
		torrentStr, err := db.RDB.Get(ctx, service.GenTorrentInfoKey(infoHash)).Result()
		if err != nil {
			continue
		}

		var torrent model.Torrent
		if err := json.Unmarshal([]byte(torrentStr), &torrent); err == nil {
			torrents = append(torrents, statsResp.HotTorrentItem{
				InfoHash:      infoHash,
				SeederCount:   torrent.SeederCount,
				LeecherCount:  torrent.LeecherCount,
				SnatcherCount: torrent.SnatcherCount,
				MetaInfo:      torrent.MetaInfo,
			})
		}
	}

	return torrents
}
