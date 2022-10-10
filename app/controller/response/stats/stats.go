package stats

type AllStatsResponse struct {
	Index IndexItem        `json:"index"`
	Hot   []HotTorrentItem `json:"hot"`
}

type IndexItem struct {
	Torrent uint64 `json:"torrent"`
	Peer    uint64 `json:"peer"`
	Traffic uint64 `json:"traffic"`
}

type HotTorrentItem struct {
	InfoHash      string `json:"info_hash"`
	SeederCount   uint64 `json:"seeder_count"`
	LeecherCount  uint64 `json:"leecher_count"`
	SnatcherCount uint64 `json:"snatcher_count"`
	MetaInfo      string `json:"meta_info"`
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
