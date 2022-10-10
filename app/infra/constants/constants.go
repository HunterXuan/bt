package constants

import "time"

const (
	TorrentPlaceHoldTemplate            = "torrent_%s_place_hold"
	TorrentSeederCountKeyTemplate       = "torrent_%s_seeder_count"
	TorrentLeecherCountCountKeyTemplate = "torrent_%s_leecher_count"
	TorrentSnatcherCountKeyTemplate     = "torrent_%s_snatcher_count"
	TorrentMetaInfoKeyTemplate          = "torrent_%s_meta_info"
	TorrentPeerKeyTemplate              = "torrent_%s_peer_%s"

	TorrentPeerSearchKeyTemplate = "torrent_%s_peer_*"

	TorrentHotSetKey      = "torrent_hot"
	TorrentHotSetCapacity = 1000

	TrackerStatsCacheKey = "tracker_stats_cache"

	TrackerTorrentStatsKey = "tracker_stats_torrent"
	TrackerPeerStatsKey    = "tracker_stats_peer"
	TrackerTrafficStatsKey = "tracker_stats_traffic"

	TrackerTorrentCountPattern = "torrent_*_place_hold"
	TrackerPeerCountPattern    = "torrent_*_peer_*"

	TorrentExpiration = time.Hour * 24
	PeerExpiration    = time.Hour * 6
)
