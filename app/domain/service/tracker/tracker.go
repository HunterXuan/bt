package tracker

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	trackerReq "github.com/HunterXuan/bt/app/controller/request/tracker"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/infra/cache"
	"github.com/HunterXuan/bt/app/infra/db"
	"github.com/HunterXuan/bt/app/infra/util/convert"
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net"
	"time"
)

// GenScrapeResult 生成Scrape结果
func GenScrapeResult(ctx *gin.Context, req *trackerReq.ScrapeRequest) (model.TorrentSlice, *customError.CustomError) {
	torrentSlice, err := getOrCreateTorrentByInfoHashSlice(req.InfoHashSlice)
	if err != nil {
		return nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	return torrentSlice, nil
}

// DealWithClientReport 处理客户端上报的请求
func DealWithClientReport(ctx *gin.Context, req *trackerReq.AnnounceRequest) (*model.Torrent, model.PeerSlice, *customError.CustomError) {
	// 查询种子
	var torrent *model.Torrent
	torrent = getTorrentFromCache(ctx, req.InfoHash)
	if torrent == nil {
		torrent, err := getOrCreateTorrentByInfoHash(req.InfoHash)
		if torrent == nil || err != nil {
			return nil, nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
		}
	}
	_ = setTorrentToCache(ctx, req.InfoHash, torrent)

	// 更新数据
	if err := updateData(torrent, req); err != nil {
		return nil, nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	// 查找同伴列表
	peerSlice, err := retrievePeerList(ctx, torrent, req)
	if err != nil {
		return nil, nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	return torrent, peerSlice, nil
}

func getOrCreateTorrentByInfoHashSlice(infoHashSlice []string) (model.TorrentSlice, error) {
	var torrentSlice model.TorrentSlice
	findResult := db.DB.Where("info_hash IN ?", infoHashSlice).Find(&torrentSlice)

	if !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
		return nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	if len(torrentSlice) < len(infoHashSlice) {
		var foundInfoHash map[string]bool
		for _, torrent := range torrentSlice {
			foundInfoHash[torrent.InfoHash] = true
		}

		var missingTorrentSlice model.TorrentSlice
		for _, infoHash := range infoHashSlice {
			if _, ok := foundInfoHash[infoHash]; !ok {
				missingTorrentSlice = append(missingTorrentSlice, model.Torrent{
					InfoHash:     infoHash,
					LastActiveAt: time.Now(),
				})
			}
		}

		if len(missingTorrentSlice) > 0 {
			db.DB.CreateInBatches(missingTorrentSlice, len(missingTorrentSlice))
		}
	}

	db.DB.Where("info_hash IN ?", infoHashSlice).Find(&torrentSlice)

	return torrentSlice, nil
}

// getOrCreateTorrent 查询或创建
func getOrCreateTorrentByInfoHash(infoHash string) (*model.Torrent, error) {
	var torrent *model.Torrent
	takeResult := db.DB.Where("info_hash = ?", infoHash).Take(&torrent)
	if errors.Is(takeResult.Error, gorm.ErrRecordNotFound) {
		torrent.InfoHash = infoHash
		torrent.LastActiveAt = time.Now()
		if createResult := db.DB.Create(&torrent); createResult.Error != nil {
			return nil, createResult.Error
		}

		return torrent, nil
	}

	if takeResult.Error != nil {
		return nil, takeResult.Error
	}

	return torrent, takeResult.Error
}

func getPeerByTorrentIDAndPeerID(torrentID uint64, peerID string) (*model.Peer, error) {
	var peer *model.Peer
	takeResult := db.DB.Where("torrent_id = ? AND peer_id = ?", torrentID, peerID).Take(&peer)
	if takeResult.Error != nil {
		if errors.Is(takeResult.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, takeResult.Error
	}

	return peer, nil
}

// 更新数据
func updateData(torrent *model.Torrent, req *trackerReq.AnnounceRequest) error {
	peer, err := getPeerByTorrentIDAndPeerID(torrent.ID, req.PeerID)
	if err != nil {
		return err
	}

	if peer != nil {
		if req.Event == "stopped" {
			if err := handleStoppedEvent(peer, req); err != nil {
				return err
			}
		} else {
			if err := handleNormalEvent(peer, req); err != nil {
				return err
			}
		}
	} else {
		if err := handleNewPeer(torrent, req); err != nil {
			return err
		}
	}

	return nil
}

// handle stopped
func handleStoppedEvent(peer *model.Peer, req *trackerReq.AnnounceRequest) error {
	deleteResult := db.DB.Delete(&peer)
	if deleteResult.Error != nil {
		return deleteResult.Error
	}

	updateTorrentStats(peer.TorrentID, 0)

	return nil
}

// handle normal event
func handleNormalEvent(peer *model.Peer, req *trackerReq.AnnounceRequest) error {
	peer.Ipv4 = req.IPv4
	peer.Ipv6 = req.IPv6
	peer.UploadedBytes = req.UploadedBytes
	peer.DownloadedBytes = req.DownloadedBytes
	peer.LeftBytes = req.LeftBytes
	peer.IsSeeder = convert.ParseBoolToUint8(req.LeftBytes == 0)
	peer.Agent = req.Agent
	if req.Event == "completed" {
		peer.FinishedAt = sql.NullTime{Time: time.Now()}
		updateTorrentStats(peer.TorrentID, 1)
	}
	if saveResult := db.DB.Save(&peer); saveResult.Error != nil {
		return saveResult.Error
	}

	return nil
}

// handle new peer
func handleNewPeer(torrent *model.Torrent, req *trackerReq.AnnounceRequest) error {
	currentPeer := &model.Peer{
		TorrentID:       torrent.ID,
		PeerID:          req.PeerID,
		Ipv4:            req.IPv4,
		Ipv6:            req.IPv6,
		Port:            req.Port,
		UploadedBytes:   req.UploadedBytes,
		DownloadedBytes: req.DownloadedBytes,
		LeftBytes:       req.LeftBytes,
		IsSeeder:        convert.ParseBoolToUint8(req.LeftBytes == 0),
		IsConnectable:   convert.ParseBoolToUint8(checkConnectable(req.IPv4, req.IPv6, req.Port)),
		Agent:           req.Agent,
	}

	if createResult := db.DB.Create(&currentPeer); createResult.Error != nil {
		return createResult.Error
	}

	updateTorrentStats(torrent.ID, 0)

	return nil
}

// 更新种子数据
func updateTorrentStats(id uint64, snatcherCount int8) {
	updateValues := map[string]interface{}{
		"snatcher_count": gorm.Expr("snatcher_count + 1 * ?", snatcherCount),
		"last_active_at": time.Now(),
	}

	rand.Seed(time.Now().UnixNano())
	if rand.Int() > 0 {
		var peerAggResult *model.PeerAggResult
		aggResult := db.DB.Model(&model.Peer{}).
			Select("torrent_id, count(*) AS peer_count, sum(is_seeder) AS seeder_count").
			Where("torrent_id = ?", id).
			Group("torrent_id").
			Take(&peerAggResult)
		if aggResult.Error != nil {
			log.Println("peerAggResult Err:", aggResult.Error)
		} else {
			updateValues["seeder_count"] = peerAggResult.SeederCount
			updateValues["leecher_count"] = peerAggResult.PeerCount - peerAggResult.SeederCount
		}
	}

	updateResult := db.DB.Model(&model.Torrent{}).
		Where("id = ?", id).
		Updates(updateValues)
	if updateResult.Error != nil {
		log.Println("addSeederCount Err:", updateResult.Error)
	}
}

// 检查连通性
func checkConnectable(ipv4 string, ipv6 string, port uint32) bool {
	connectable := false

	if ipv6 != "" {
		if _, err := net.DialTimeout(
			"tcp6",
			fmt.Sprintf("[%v]:%v", ipv6, port),
			5*time.Second,
		); err == nil {
			connectable = true
		}
	}

	if !connectable && ipv4 != "" {
		if _, err := net.DialTimeout(
			"tcp4",
			fmt.Sprintf("%v:%v", ipv4, port),
			5*time.Second,
		); err == nil {
			connectable = true
		}
	}

	return connectable
}

// 获取同伴列表
// 如果当前用户为做种者，只返回其它非做种者
func retrievePeerList(ctx *gin.Context, torrent *model.Torrent, req *trackerReq.AnnounceRequest) (model.PeerSlice, error) {
	columns := []string{"ipv4", "ipv6", "port"}
	if req.Compact != 1 || req.NoPeerID != 1 {
		columns = append(columns, "peer_id")
	}

	isSeeder := req.LeftBytes == 0
	if isSeeder {
		var peerSlice model.PeerSlice
		findResult := db.DB.Where(
			"torrent_id = ? AND peer_id <> ? AND is_seeder <> ?",
			torrent.ID,
			req.PeerID,
			convert.ParseBoolToUint8(true),
		).Order("peer_id").Limit(calPeerLimitCount(req.NumWanted)).Find(&peerSlice)

		return peerSlice, findResult.Error
	}

	var peerSlice model.PeerSlice
	findResult := db.DB.Where(
		"torrent_id = ? AND peer_id <> ?",
		torrent.ID,
		req.PeerID,
	).Order("peer_id").Limit(calPeerLimitCount(req.NumWanted)).Find(&peerSlice)

	return peerSlice, findResult.Error
}

// 获取同伴数量
func calPeerLimitCount(numWanted uint8) int {
	if numWanted > 0 && numWanted < 50 {
		return int(numWanted)
	}

	return 50
}

func setTorrentToCache(ctx *gin.Context, infoHash string, torrent *model.Torrent) error {
	bytes, err := json.Marshal(torrent)
	if err != nil {
		return nil
	}

	_, err = cache.RDB.SetEX(ctx, genTorrentCacheKey(infoHash), bytes, time.Hour).Result()
	return err
}

func getTorrentFromCache(ctx *gin.Context, infoHash string) *model.Torrent {
	var torrent model.Torrent
	val, err := cache.RDB.Get(ctx, genTorrentCacheKey(infoHash)).Result()
	if err != nil {
		log.Println("getTorrentFromCache Err:", err)
		return nil
	}

	if err := json.Unmarshal([]byte(val), &torrent); err != nil {
		log.Println("getTorrentFromCache Err:", err)
		return nil
	}

	return &torrent
}

func genTorrentCacheKey(infoHash string) string {
	return fmt.Sprintf("TORRENT_WITH_HASH_%v", infoHash)
}
