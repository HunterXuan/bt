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

	val, err := db.RDB.Get(ctx, constants.StatsCacheKey).Result()
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

	_, err = db.RDB.Set(ctx, constants.StatsCacheKey, bytes, 0).Result()
	if err != nil {
		log.Println("setStatsToCache Err:", err)
	}

	return err
}

func getTorrentCount(ctx context.Context) uint64 {
	realCount, err := db.RDB.ZCard(ctx, constants.ActiveTorrentSetKey).Uint64()
	if err != nil {
		cacheCount, _ := db.RDB.HGet(ctx, constants.StatsKey, constants.StatsTorrentCountKey).Uint64()
		return cacheCount
	}

	db.RDB.HSet(ctx, constants.StatsKey, constants.StatsTorrentCountKey, realCount)

	return realCount
}

func getPeerCount(ctx context.Context) uint64 {
	realCount, err := db.RDB.ZCard(ctx, constants.ActivePeerSetKey).Uint64()
	if err != nil {
		cacheCount, _ := db.RDB.HGet(ctx, constants.StatsKey, constants.StatsPeerCountKey).Uint64()
		return cacheCount
	}

	db.RDB.HSet(ctx, constants.StatsKey, constants.StatsPeerCountKey, realCount)

	return realCount
}

func getTrafficCount(ctx context.Context) uint64 {
	val, _ := db.RDB.HGet(ctx, constants.StatsKey, constants.StatsTrafficCountKey).Uint64()
	return val
}

func getHotStats(ctx context.Context) []statsResp.HotTorrentItem {
	hotInfoHashes, err := db.RDB.ZRange(ctx, constants.ActiveTorrentSetKey, 0, constants.TorrentHotCapacity).Result()
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
