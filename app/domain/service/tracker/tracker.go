package tracker

import (
	"context"
	"encoding/json"
	"fmt"
	trackerReq "github.com/HunterXuan/bt/app/controller/request/tracker"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/domain/service"
	"github.com/HunterXuan/bt/app/infra/constants"
	"github.com/HunterXuan/bt/app/infra/db"
	"github.com/HunterXuan/bt/app/infra/util/convert"
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/HunterXuan/bt/app/infra/util/prob"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"math/rand"
	"net"
	"time"
)

// GenScrapeResult 生成Scrape结果
func GenScrapeResult(ctx *gin.Context, req *trackerReq.ScrapeRequest) (model.TorrentSlice, *customError.CustomError) {
	torrentSlice, err := getOrCreateTorrentByInfoHashSlice(ctx, req.InfoHashSlice)
	if err != nil {
		return nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	return torrentSlice, nil
}

// DealWithClientReport 处理客户端上报的请求
func DealWithClientReport(ctx *gin.Context, req *trackerReq.AnnounceRequest) (*model.Torrent, model.PeerSlice, *customError.CustomError) {
	// 查询种子
	torrent, _ := getOrCreateTorrentByInfoHash(ctx, req.InfoHash)

	// 更新数据
	if err := updateData(ctx, req); err != nil {
		return nil, nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	// 查找同伴列表
	peerSlice, err := retrievePeerList(ctx, req)
	if err != nil {
		return nil, nil, customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	addToActiveSet(ctx, req.InfoHash)

	return torrent, peerSlice, nil
}

func getOrCreateTorrentByInfoHashSlice(ctx context.Context, infoHashSlice []string) (model.TorrentSlice, error) {
	var torrentSlice model.TorrentSlice

	for _, infoHash := range infoHashSlice {
		torrent, _ := getOrCreateTorrentByInfoHash(ctx, infoHash)
		torrentSlice = append(torrentSlice, *torrent)
	}

	return torrentSlice, nil
}

// getOrCreateTorrent 查询或创建
func getOrCreateTorrentByInfoHash(ctx context.Context, infoHash string) (*model.Torrent, error) {
	var torrent *model.Torrent

	torrentInfoKey := service.GenTorrentInfoKey(infoHash)

	torrentInfo, _ := db.RDB.HGetAll(ctx, torrentInfoKey).Result()
	if len(torrentInfo) > 0 {
		torrent = &model.Torrent{
			InfoHash:      infoHash,
			SeederCount:   convert.ParseStringToUint64(torrentInfo[constants.TorrentSeederCountKey]),
			LeecherCount:  convert.ParseStringToUint64(torrentInfo[constants.TorrentLeecherCountKey]),
			SnatcherCount: convert.ParseStringToUint64(torrentInfo[constants.TorrentSnatcherCountKey]),
			MetaInfo:      torrentInfo[constants.TorrentMetaInfoKey],
			CreatedAt:     convert.ParseStringToInt64(torrentInfo[constants.TorrentCreatedAtKey]),
			LastActiveAt:  convert.ParseStringToInt64(torrentInfo[constants.TorrentLastActiveAt]),
		}
	} else {
		// perhaps new torrent
		nowTime := time.Now()
		db.RDB.HIncrBy(ctx, constants.StatsKey, constants.StatsTorrentCountKey, 1)
		torrent = &model.Torrent{
			InfoHash:      infoHash,
			SeederCount:   0,
			LeecherCount:  0,
			SnatcherCount: 0,
			MetaInfo:      "",
			CreatedAt:     nowTime.Unix(),
			LastActiveAt:  nowTime.Unix(),
		}

		db.RDB.HSet(ctx, torrentInfoKey, map[string]interface{}{
			constants.TorrentInfoHashKey:      torrent.InfoHash,
			constants.TorrentSeederCountKey:   torrent.SeederCount,
			constants.TorrentLeecherCountKey:  torrent.LeecherCount,
			constants.TorrentSnatcherCountKey: torrent.SnatcherCount,
			constants.TorrentMetaInfoKey:      torrent.MetaInfo,
			constants.TorrentCreatedAtKey:     torrent.CreatedAt,
			constants.TorrentLastActiveAt:     torrent.LastActiveAt,
		})
	}

	return torrent, nil
}

func getPeerByInfoHashAndPeerID(ctx context.Context, infoHash, peerID string) (*model.Peer, error) {
	var peer *model.Peer
	peerStr, _ := db.RDB.HGet(ctx, service.GenPeerKey(infoHash), peerID).Result()
	if len(peerStr) > 0 {
		err := json.Unmarshal([]byte(peerStr), &peer)
		if err != nil {
			return nil, err
		}
	}

	return peer, nil
}

// 更新数据
func updateData(ctx context.Context, req *trackerReq.AnnounceRequest) error {
	peer, err := getPeerByInfoHashAndPeerID(ctx, req.InfoHash, req.PeerID)
	if err != nil {
		return err
	}

	updateTrackerStats(ctx, req, peer)

	if peer != nil {
		if req.Event == "stopped" {
			if err := handleStoppedEvent(ctx, peer, req); err != nil {
				return err
			}
		} else {
			if err := handleNormalEvent(ctx, peer, req); err != nil {
				return err
			}
		}
	} else {
		if err := handleNewPeer(ctx, req); err != nil {
			return err
		}
	}

	cleanDeadPeers(ctx, req.InfoHash)

	return nil
}

func cleanDeadPeers(ctx context.Context, infoHash string) {
	rand.Seed(time.Now().Unix())
	if rand.Intn(1000) > 800 {
		peers, _ := db.RDB.HGetAll(ctx, service.GenPeerKey(infoHash)).Result()
		for _, peerStr := range peers {
			if len(peerStr) > 0 {
				var peer model.Peer
				if err := json.Unmarshal([]byte(peerStr), &peer); err != nil {
					if peer.LastActiveAt+int64(constants.PeerExpiration.Seconds()) < time.Now().Unix() {
						db.RDB.HDel(ctx, service.GenPeerKey(infoHash), peer.PeerID)
					}
				}
			}
		}
	}
}

func updateTrackerStats(ctx context.Context, req *trackerReq.AnnounceRequest, peer *model.Peer) {
	var trafficBytesIncr int64
	if peer == nil {
		trafficBytesIncr = int64(req.UploadedBytes + req.DownloadedBytes)
	} else {
		trafficBytesIncr = int64(req.UploadedBytes + req.DownloadedBytes - peer.UploadedBytes - peer.DownloadedBytes)
	}

	if peer == nil {
		db.RDB.HIncrBy(ctx, constants.StatsKey, constants.StatsPeerCountKey, 1)
	} else if req.Event == "stopped" {
		db.RDB.HIncrBy(ctx, constants.StatsKey, constants.StatsPeerCountKey, -1)
	}

	db.RDB.HIncrBy(ctx, constants.StatsKey, constants.StatsTrafficCountKey, trafficBytesIncr)
}

func addToActiveSet(ctx context.Context, infoHash string) {
	db.RDB.ZAdd(ctx, constants.ActiveTorrentSetKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: infoHash,
	})
}

// handle stopped
func handleStoppedEvent(ctx context.Context, peer *model.Peer, _ *trackerReq.AnnounceRequest) error {
	db.RDB.HDel(ctx, service.GenPeerKey(peer.InfoHash), peer.PeerID)
	db.RDB.ZRem(ctx, constants.ActivePeerSetKey, &redis.Z{
		Member: fmt.Sprintf("%v:%v", peer.InfoHash, peer.PeerID),
	})

	if peer.IsSeeder {
		updateTorrentStats(ctx, peer.InfoHash, 0, -1, 0)
	} else {
		updateTorrentStats(ctx, peer.InfoHash, -1, 0, 0)
	}

	return nil
}

// handle normal event
func handleNormalEvent(ctx context.Context, peer *model.Peer, req *trackerReq.AnnounceRequest) error {
	peer.Ipv4 = req.IPv4
	peer.Ipv6 = req.IPv6
	peer.UploadedBytes = req.UploadedBytes
	peer.DownloadedBytes = req.DownloadedBytes
	peer.LeftBytes = req.LeftBytes
	peer.IsSeeder = req.LeftBytes == 0
	peer.Agent = req.Agent
	peer.LastActiveAt = time.Now().Unix()
	if req.Event == "completed" {
		peer.FinishedAt = time.Now().Unix()
		updateTorrentStats(ctx, peer.InfoHash, -1, 1, 1)
	}

	peerStr, err := json.Marshal(peer)
	if err != nil {
		return nil
	}

	_, err = db.RDB.HSet(ctx, service.GenPeerKey(peer.InfoHash), peer.PeerID, peerStr).Result()
	db.RDB.ZAdd(ctx, constants.ActivePeerSetKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: fmt.Sprintf("%v:%v", peer.InfoHash, peer.PeerID),
	})

	return err
}

// handle new peer
func handleNewPeer(ctx context.Context, req *trackerReq.AnnounceRequest) error {
	currentPeer := &model.Peer{
		InfoHash:        req.InfoHash,
		PeerID:          req.PeerID,
		Ipv4:            req.IPv4,
		Ipv6:            req.IPv6,
		Port:            req.Port,
		UploadedBytes:   req.UploadedBytes,
		DownloadedBytes: req.DownloadedBytes,
		LeftBytes:       req.LeftBytes,
		IsSeeder:        req.LeftBytes == 0,
		IsConnectable:   checkConnectable(req.IPv4, req.IPv6, req.Port),
		Agent:           req.Agent,
		LastActiveAt:    time.Now().Unix(),
	}

	currentPeerStr, err := json.Marshal(currentPeer)
	if err != nil {
		return nil
	}

	db.RDB.HSet(ctx, service.GenPeerKey(req.InfoHash), req.PeerID, currentPeerStr)
	db.RDB.ZAdd(ctx, constants.ActivePeerSetKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: fmt.Sprintf("%v:%v", req.InfoHash, req.PeerID),
	})

	if currentPeer.IsSeeder {
		updateTorrentStats(ctx, req.InfoHash, 0, 1, 0)
	} else {
		updateTorrentStats(ctx, req.InfoHash, 1, 0, 0)
	}

	return nil
}

// 更新种子数据
func updateTorrentStats(ctx context.Context, infoHash string, leecherCountIncr, seederCountIncr, snatcherCountIncr int64) {
	if leecherCountIncr != 0 {
		db.RDB.HIncrBy(ctx, service.GenTorrentInfoKey(infoHash), constants.TorrentLeecherCountKey, leecherCountIncr)
	}
	if seederCountIncr != 0 {
		db.RDB.HIncrBy(ctx, service.GenTorrentInfoKey(infoHash), constants.TorrentSeederCountKey, seederCountIncr)
	}
	if snatcherCountIncr != 0 {
		db.RDB.HIncrBy(ctx, service.GenTorrentInfoKey(infoHash), constants.TorrentSnatcherCountKey, snatcherCountIncr)
	}

	if prob.IfProbGreaterThan(0.8) {
		var seederCount, leecherCount int64

		peers, _ := db.RDB.HGetAll(ctx, service.GenPeerKey(infoHash)).Result()
		for _, peerStr := range peers {
			if len(peerStr) > 0 {
				var peer model.Peer
				if err := json.Unmarshal([]byte(peerStr), &peer); err != nil {
					if peer.IsSeeder {
						seederCount = seederCount + 1
						continue
					}
				}
			}

			leecherCount = leecherCount + 1
		}

		db.RDB.HSet(ctx, service.GenTorrentInfoKey(infoHash), constants.TorrentSeederCountKey, seederCount)
		db.RDB.HSet(ctx, service.GenTorrentInfoKey(infoHash), constants.TorrentLeecherCountKey, leecherCount)
	}
}

// 检查连通性
func checkConnectable(ipv4 string, ipv6 string, port uint32) bool {
	connectable := false

	if ipv6 != "" {
		connectable = !net.ParseIP(ipv6).IsPrivate()
	}

	if !connectable && ipv4 != "" {
		connectable = !net.ParseIP(ipv4).IsPrivate()
	}

	return connectable
}

// 获取同伴列表
// 如果当前用户为做种者，只返回其它非做种者
func retrievePeerList(ctx *gin.Context, req *trackerReq.AnnounceRequest) (model.PeerSlice, error) {
	var peers model.PeerSlice

	isSeeder := req.LeftBytes == 0

	peerLimitCount := calPeerLimitCount(req.NumWanted)

	peersMap, _ := db.RDB.HGetAll(ctx, service.GenPeerKey(req.InfoHash)).Result()

	for _, peerStr := range peersMap {
		var peer model.Peer
		err := json.Unmarshal([]byte(peerStr), &peer)
		if err != nil {
			continue
		}

		if isSeeder && peer.IsSeeder {
			continue
		}

		if len(peers) >= peerLimitCount {
			break
		}

		peers = append(peers, peer)
	}

	return peers, nil
}

// 获取同伴数量
func calPeerLimitCount(numWanted uint8) int {
	if numWanted > 0 && numWanted < 50 {
		return int(numWanted)
	}

	return 50
}
