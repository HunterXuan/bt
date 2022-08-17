package stats

import (
	"encoding/json"
	statsReq "github.com/HunterXuan/bt/app/controller/request/stats"
	statsResp "github.com/HunterXuan/bt/app/controller/response/stats"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/infra/cache"
	"github.com/HunterXuan/bt/app/infra/db"
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/gin-gonic/gin"
	"log"
	"time"
)

// GetAllStats 获取统计数据
func GetAllStats(ctx *gin.Context, req *statsReq.AllStatsRequest) (*statsResp.AllStatsResponse, *customError.CustomError) {
	if stats, err := getStatsFromCache(ctx); err != nil {
		return stats, nil
	}

	stats := &statsResp.AllStatsResponse{
		Index: statsResp.IndexItem{
			Torrent: getTorrentIndexStats(),
			Peer:    getPeerIndexStats(),
			Traffic: getTrafficIndexStats(),
		},
		Hot: getHotStats(),
	}

	_ = setStatsToCache(ctx, stats)

	return stats, nil
}

func getStatsFromCache(ctx *gin.Context) (*statsResp.AllStatsResponse, error) {
	var stats *statsResp.AllStatsResponse

	val, err := cache.RDB.Get(ctx, genStatsCacheKey()).Result()
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

func setStatsToCache(ctx *gin.Context, stats *statsResp.AllStatsResponse) error {
	bytes, err := json.Marshal(stats)
	if err != nil {
		log.Println("setStatsToCache Err:", err)
		return err
	}

	_, err = cache.RDB.SetEX(ctx, genStatsCacheKey(), bytes, time.Hour).Result()
	if err != nil {
		log.Println("setStatsToCache Err:", err)
	}

	return err
}

func genStatsCacheKey() string {
	return "STATS_CACHE"
}

func getTorrentIndexStats() statsResp.TorrentStats {
	var totalCount, activeCount, deadCount int64
	db.DB.Model(&model.Torrent{}).Count(&totalCount)
	db.DB.Model(&model.Torrent{}).Where("last_active_at > ?", time.Now().Add(-24*time.Hour)).Count(&activeCount)
	db.DB.Model(&model.Torrent{}).Where("last_active_at < ?", time.Now().Add(-72*time.Hour)).Count(&deadCount)

	return statsResp.TorrentStats{
		Total:  uint64(totalCount),
		Active: uint64(activeCount),
		Dead:   uint64(deadCount),
	}
}

func getPeerIndexStats() statsResp.PeerStats {
	var totalCount, seederCount int64
	db.DB.Model(&model.Peer{}).Count(&totalCount)
	db.DB.Model(&model.Peer{}).Where("is_seeder = ?", 1).Count(&seederCount)

	return statsResp.PeerStats{
		Total:   uint64(totalCount),
		Seeder:  uint64(seederCount),
		Leacher: uint64(totalCount - seederCount),
	}
}

func getTrafficIndexStats() statsResp.TrafficStats {
	var trafficStats statsResp.TrafficStats
	takeRes := db.DB.Model(&model.Peer{}).
		Select("sum(uploaded_bytes) AS upload, sum(downloaded_bytes) AS download").
		Take(&trafficStats)
	if takeRes.Error != nil {
		log.Println("getTrafficIndexStats Err:", takeRes.Error)
	}

	trafficStats.Total = trafficStats.Upload + trafficStats.Download

	return trafficStats
}

func getHotStats() []statsResp.HotTorrentItem {
	var torrents []statsResp.HotTorrentItem
	findRes := db.DB.Model(&model.Torrent{}).
		Select("info_hash, seeder_count, leecher_count, snatcher_count").
		Order("leecher_count desc").
		Limit(100).
		Find(&torrents)
	if findRes.Error != nil {
		log.Println("getHotStats Err:", findRes.Error)
	}

	return torrents
}
