package job

import (
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/infra/db"
	"github.com/HunterXuan/bt/app/infra/dht"
	"github.com/HunterXuan/bt/app/infra/util/tracker"
	"log"
	"math/rand"
	"time"
)

type DHT struct{}

type HotTorrentItem struct {
	ID       uint64 `json:"id"`
	InfoHash string `json:"info_hash"`
	Ip       string `json:"ip"`
	Port     uint32 `json:"port"`
}

func (d *DHT) Run() {
	log.Println("DHT start collecting torrents info")

	for _, item := range getHotTorrentsAndPeers() {
		dht.DHT.Request([]byte(tracker.RestoreToByteString(item.InfoHash)), item.Ip, int(item.Port))
	}

	log.Println("DHT finish collecting torrents info")
}

func getHotTorrentsAndPeers() []HotTorrentItem {
	var torrents []model.Torrent
	findRes := db.DB.Model(&model.Torrent{}).
		Select("id, info_hash").
		Where("meta_info = ?", "").
		Order("leecher_count desc").
		Limit(100).
		Find(&torrents)
	if findRes.Error != nil {
		log.Println("DHT getHotTorrentsAndPeers Err:", findRes.Error)
	} else {
		log.Println("DHT getHotTorrentsAndPeers Torrent Count:", len(torrents))
	}

	var torrentIdSlice []uint64
	for _, torrent := range torrents {
		torrentIdSlice = append(torrentIdSlice, torrent.ID)
	}

	var peers []model.Peer
	findRes = db.DB.Model(&model.Peer{}).
		Select("ipv4, ipv6, port").
		Where("torrent_id IN ? AND is_connectable = ?", torrentIdSlice, 1).
		Find(&peers)
	if findRes.Error != nil {
		log.Println("DHT getHotTorrentsAndPeers Err:", findRes.Error)
	} else {
		log.Println("DHT getHotTorrentsAndPeers Peer Count:", len(peers))
	}

	return groupTorrentAndPeer(torrents, peers)
}

func groupTorrentAndPeer(torrents []model.Torrent, peers []model.Peer) []HotTorrentItem {
	rand.Seed(time.Now().UnixNano())

	groupTorrents := make(map[uint64][]model.Peer)
	for _, peer := range peers {
		if peerGroup, ok := groupTorrents[peer.TorrentID]; ok {
			groupTorrents[peer.TorrentID] = append(peerGroup, peer)
		} else {
			groupTorrents[peer.TorrentID] = []model.Peer{peer}
		}
	}

	var result []HotTorrentItem
	for _, torrent := range torrents {
		if peers, ok := groupTorrents[torrent.ID]; ok {
			peer := peers[rand.Intn(len(peers))]

			ip := peer.Ipv4
			if ip == "" && peer.Ipv6 != "" {
				ip = peer.Ipv6
			}
			result = append(result, HotTorrentItem{
				ID:       torrent.ID,
				InfoHash: torrent.InfoHash,
				Ip:       ip,
				Port:     peer.Port,
			})
		}
	}

	return result
}
