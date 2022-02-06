package stats

type AllStatsResponse struct {
	Index IndexItem        `json:"index"`
	Hot   []HotTorrentItem `json:"hot"`
}

type IndexItem struct {
	Torrent TorrentStats `json:"torrent"`
	Peer    PeerStats    `json:"peer"`
	Traffic TrafficStats `json:"traffic"`
}

type TorrentStats struct {
	Total  uint64 `json:"total"`
	Active uint64 `json:"active"`
	Dead   uint64 `json:"dead"`
}

type PeerStats struct {
	Total   uint64 `json:"total"`
	Seeder  uint64 `json:"seeder"`
	Leacher uint64 `json:"leecher"`
}

type TrafficStats struct {
	Total    uint64 `json:"total"`
	Upload   uint64 `json:"upload"`
	Download uint64 `json:"download"`
}

type HotTorrentItem struct {
	InfoHash      string `json:"info_hash"`
	SeederCount   uint64 `json:"seeder_count"`
	LeecherCount  uint64 `json:"leecher_count"`
	SnatcherCount uint64 `json:"snatcher_count"`
}

func (s AllStatsResponse) Serialize(singleModel interface{}) interface{} {
	allStats, ok := singleModel.(*AllStatsResponse)
	if !ok {
		return nil
	}

	return *allStats
}

func (s AllStatsResponse) Paginate(modelSlice interface{}) interface{} {
	return nil
}
