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
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
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

	addToHotSet(ctx, req.InfoHash)

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

	torrentStr, _ := db.RDB.Get(ctx, service.GenTorrentInfoKey(infoHash)).Result()
	if len(torrentStr) > 0 {
		err := json.Unmarshal([]byte(torrentStr), &torrent)
		if err != nil {
			return nil, err
		}
	} else {
		// perhaps new torrent
		db.RDB.IncrBy(ctx, constants.TrackerTorrentStatsKey, 1)
		torrent = &model.Torrent{
			InfoHash:      infoHash,
			SeederCount:   0,
			LeecherCount:  0,
			SnatcherCount: 0,
			MetaInfo:      "",
			CreatedAt:     time.Now().Unix(),
		}
	}

	rand.Seed(time.Now().Unix())
	if len(torrentStr) == 0 || rand.Intn(1000) > 500 {
		seederCount, _ := db.RDB.Get(ctx, service.GenTorrentSeederCountKey(infoHash)).Uint64()
		leecherCount, _ := db.RDB.Get(ctx, service.GenTorrentLeecherCountKey(infoHash)).Uint64()
		snatcherCount, _ := db.RDB.Get(ctx, service.GenTorrentSnatcherCountKey(infoHash)).Uint64()
		metaInfoStr, _ := db.RDB.Get(ctx, service.GenTorrentMetaInfoKey(infoHash)).Result()
		torrent.SeederCount = seederCount
		torrent.LeecherCount = leecherCount
		torrent.SnatcherCount = snatcherCount
		torrent.MetaInfo = metaInfoStr

		torrentStr, err := json.Marshal(torrent)
		if err != nil {
			return nil, err
		}

		db.RDB.Set(ctx, service.GenTorrentInfoKey(infoHash), torrentStr, constants.TorrentExpiration)
	}

	return torrent, nil
}

func getPeerByInfoHashAndPeerID(ctx context.Context, infoHash, peerID string) (*model.Peer, error) {
	var peer *model.Peer
	peerStr, _ := db.RDB.Get(ctx, service.GenPeerKey(infoHash, peerID)).Result()
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

	return nil
}

func updateTrackerStats(ctx context.Context, req *trackerReq.AnnounceRequest, peer *model.Peer) {
	var trafficBytesIncr int64
	if peer == nil {
		trafficBytesIncr = int64(req.UploadedBytes + req.DownloadedBytes)
	} else {
		trafficBytesIncr = int64(req.UploadedBytes + req.DownloadedBytes - peer.UploadedBytes - peer.DownloadedBytes)
	}

	if peer == nil {
		db.RDB.IncrBy(ctx, constants.TrackerPeerStatsKey, 1)
	} else if req.Event == "stopped" {
		db.RDB.DecrBy(ctx, constants.TrackerPeerStatsKey, 1)
	}

	db.RDB.IncrBy(ctx, constants.TrackerTrafficStatsKey, trafficBytesIncr)
}

func addToHotSet(ctx context.Context, infoHash string) {
	db.RDB.ZAdd(ctx, constants.TorrentHotSetKey, &redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: infoHash,
	})

	count, err := db.RDB.ZCard(ctx, constants.TorrentHotSetKey).Result()
	if err != nil && count > constants.TorrentHotSetCapacity {
		db.RDB.ZRemRangeByRank(ctx, constants.TorrentHotSetKey, 0, 0)
	}
}

// handle stopped
func handleStoppedEvent(ctx context.Context, peer *model.Peer, _ *trackerReq.AnnounceRequest) error {
	_, err := db.RDB.Del(ctx, service.GenPeerKey(peer.InfoHash, peer.PeerID)).Result()
	if err != nil {
		return err
	}

	updateTorrentStats(ctx, peer.InfoHash, 0)

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
	if req.Event == "completed" {
		peer.FinishedAt = time.Now().Unix()
		updateTorrentStats(ctx, peer.InfoHash, 1)
	}

	peerStr, err := json.Marshal(peer)
	if err != nil {
		return nil
	}

	_, err = db.RDB.Set(ctx, service.GenPeerKey(peer.InfoHash, peer.PeerID), peerStr, constants.PeerExpiration).Result()

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
	}

	currentPeerStr, err := json.Marshal(currentPeer)
	if err != nil {
		return nil
	}

	_, err = db.RDB.SetEX(ctx, service.GenPeerKey(req.InfoHash, req.PeerID), currentPeerStr, constants.PeerExpiration).Result()
	if err != nil {
		return err
	}

	updateTorrentStats(ctx, req.InfoHash, 0)

	return nil
}

// 更新种子数据
func updateTorrentStats(ctx context.Context, infoHash string, snatcherCountIncr int64) {
	if snatcherCountIncr != 0 {
		_, err := db.RDB.IncrBy(ctx, service.GenTorrentSnatcherCountKey(infoHash), snatcherCountIncr).Result()
		if err != nil {
			log.Println("snatcherCountIncr Err:", err)
		}
	}

	rand.Seed(time.Now().UnixNano())
	if rand.Int() > 0 {
		var seederCount, leecherCount int64

		keys, _ := db.RDB.Keys(ctx, service.GenPeerSearchPattern(infoHash)).Result()
		for _, key := range keys {
			peerStr, _ := db.RDB.Get(ctx, key).Result()
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

		db.RDB.Set(ctx, service.GenTorrentSeederCountKey(infoHash), seederCount, constants.TorrentExpiration)
		db.RDB.Set(ctx, service.GenTorrentLeecherCountKey(infoHash), leecherCount, constants.TorrentExpiration)
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
func retrievePeerList(ctx *gin.Context, req *trackerReq.AnnounceRequest) (model.PeerSlice, error) {
	var peers model.PeerSlice

	isSeeder := req.LeftBytes == 0

	peerLimitCount := calPeerLimitCount(req.NumWanted)

	keys, _ := db.RDB.Keys(ctx, service.GenPeerSearchPattern(req.InfoHash)).Result()
	if len(keys) > 0 {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(keys), func(i, j int) {
			keys[i], keys[j] = keys[j], keys[i]
		})
	}

	for _, key := range keys {
		peerStr, _ := db.RDB.Get(ctx, key).Result()
		if len(peerStr) <= 0 {
			continue
		}

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
