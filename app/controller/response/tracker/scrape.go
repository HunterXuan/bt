package tracker

import (
	"github.com/HunterXuan/bt/app/domain/model"
	trackerUtil "github.com/HunterXuan/bt/app/infra/util/tracker"
	"github.com/anacrolix/torrent/bencode"
)

type ScrapeItem struct {
	SeederCount   uint64 `bencode:"complete"`   // 当前做种数量
	SnatcherCount uint64 `bencode:"downloaded"` // 已完成下载总数
	LeecherCount  uint64 `bencode:"incomplete"` // 正在下载数量
}

type ScrapeResult struct {
	Files map[string]*ScrapeItem `bencode:"files"`
}

type ScrapeResponse struct {
}

func (s ScrapeResponse) Serialize(singleModel interface{}) interface{} {
	return nil
}

func (s ScrapeResponse) Paginate(modelSlice interface{}) interface{} {
	return nil
}

func (s ScrapeResponse) BEncode(modelSlice interface{}) string {
	torrentSlice, ok := modelSlice.(model.TorrentSlice)
	if !ok {
		return string(bencode.MustMarshal(map[string]string{
			"failure reason": "Bad Torrent",
		}))
	}

	scrapeResult := &ScrapeResult{
		Files: make(map[string]*ScrapeItem),
	}
	for _, torrent := range torrentSlice {
		scrapeResult.Files[trackerUtil.RestoreToByteString(torrent.InfoHash)] = &ScrapeItem{
			SeederCount:   torrent.SeederCount,
			SnatcherCount: torrent.SnatcherCount,
			LeecherCount:  torrent.LeecherCount,
		}
	}

	return string(bencode.MustMarshal(scrapeResult))
}
