package tracker

import (
	"bytes"
	"encoding/binary"
	trackerReq "github.com/HunterXuan/bt/app/controller/request/tracker"
	"github.com/HunterXuan/bt/app/domain/model"
	trackerUtil "github.com/HunterXuan/bt/app/infra/util/tracker"
	"github.com/anacrolix/torrent/bencode"
	"math"
	"net"
	"time"
)

type PeerItem struct {
	IP     string `bencode:"ip"`
	Port   uint32 `bencode:"port"`
	PeerID string `bencode:"peer id,omitempty"`
}

type AnnounceResult struct {
	Interval      uint32      `bencode:"interval"`   // 请求时间间隔
	SeederCount   uint64      `bencode:"complete"`   // 当前做种数量
	SnatcherCount uint64      `bencode:"downloaded"` // 已完成下载总数
	LeecherCount  uint64      `bencode:"incomplete"` // 正在下载数量
	Peers         []*PeerItem `bencode:"peers"`      // 同伴
}

type AnnounceCompactResult struct {
	Interval      uint32 `bencode:"interval"`   // 请求时间间隔
	SeederCount   uint64 `bencode:"complete"`   // 当前做种数量
	SnatcherCount uint64 `bencode:"downloaded"` // 已完成下载总数
	LeecherCount  uint64 `bencode:"incomplete"` // 正在下载数量
	Peers         string `bencode:"peers"`      // IPv4 同伴
	Peers6        string `bencode:"peers6"`     // IPv6 同伴
}

type AnnounceResponse struct {
}

func (s AnnounceResponse) Serialize(_ interface{}) interface{} {
	return nil
}

func (s AnnounceResponse) Paginate(_ interface{}) interface{} {
	return nil
}

func (s AnnounceResponse) BEncode(torrent *model.Torrent, modelSlice interface{}, req *trackerReq.AnnounceRequest) string {
	peerSlice, ok := modelSlice.(model.PeerSlice)
	if !ok {
		return string(bencode.MustMarshal(map[string]string{
			"failure reason": "Bad Torrent",
		}))
	}

	if req.Compact == 1 {
		peers, peers6 := genCompactPeers(peerSlice)
		compactResult := &AnnounceCompactResult{
			Interval:      genInterval(torrent.CreatedAt),
			SeederCount:   torrent.SeederCount,
			SnatcherCount: torrent.SnatcherCount,
			LeecherCount:  torrent.LeecherCount,
			Peers:         peers,
			Peers6:        peers6,
		}
		return string(bencode.MustMarshal(compactResult))
	}

	result := &AnnounceResult{
		Interval:      genInterval(torrent.CreatedAt),
		SeederCount:   torrent.SeederCount,
		SnatcherCount: torrent.SnatcherCount,
		LeecherCount:  torrent.LeecherCount,
		Peers:         genPeers(peerSlice),
	}
	return string(bencode.MustMarshal(result))
}

// 计算请求时间间隔，防止大量请求同时发生
// interval = c/(1+a*e^(-kx)); c = 7200; a = 3; k = 0.0001
// 新种子 30 分钟，旧种子 120 分钟
func genInterval(createdAt int64) uint32 {
	diffSeconds := time.Now().Unix() - createdAt
	if diffSeconds < 0 {
		return 0
	}
	return uint32(math.Round(7200 / (1 + 3*math.Exp(float64(-diffSeconds)/60/10000))))
}

// 生成非压缩的同伴列表
func genPeers(peerSlice model.PeerSlice) []*PeerItem {
	var peers []*PeerItem

	for _, peer := range peerSlice {
		if peer.Ipv4 != "" {
			peers = append(peers, &PeerItem{
				IP:     peer.Ipv4,
				Port:   peer.Port,
				PeerID: trackerUtil.RestoreToByteString(peer.PeerID),
			})
		}

		if peer.Ipv6 != "" {
			peers = append(peers, &PeerItem{
				IP:     peer.Ipv6,
				Port:   peer.Port,
				PeerID: trackerUtil.RestoreToByteString(peer.PeerID),
			})
		}
	}

	return peers
}

// 生成压缩的同伴列表
func genCompactPeers(peerSlice model.PeerSlice) (string, string) {
	var peers []byte
	var peers6 []byte

	buf := new(bytes.Buffer)
	byteOrder := binary.BigEndian

	for _, peer := range peerSlice {
		if ip := net.ParseIP(peer.Ipv4).To4(); ip != nil {
			if buf.Reset(); binary.Write(buf, byteOrder, ip) == nil && binary.Write(buf, byteOrder, uint16(peer.Port)) == nil {
				peers = append(peers, buf.Bytes()...)
			}
		}

		if ip := net.ParseIP(peer.Ipv6).To16(); ip != nil {
			if buf.Reset(); binary.Write(buf, byteOrder, ip) == nil && binary.Write(buf, byteOrder, uint16(peer.Port)) == nil {
				peers6 = append(peers6, buf.Bytes()...)
			}
		}
	}

	return string(peers), string(peers6)
}
