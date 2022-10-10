package constants

import "time"

const (
	TorrentInfoTemplate                 = "torrent:%s:info"
	TorrentSeederCountKeyTemplate       = "torrent:%s:seeder_count"
	TorrentLeecherCountCountKeyTemplate = "torrent:%s:leecher_count"
	TorrentSnatcherCountKeyTemplate     = "torrent:%s:snatcher_count"
	TorrentMetaInfoKeyTemplate          = "torrent:%s:meta_info"
	TorrentPeerKeyTemplate              = "torrent:%s:peer:%s"

	TorrentPeerSearchKeyTemplate = "torrent:%s:peer:*"

	TorrentHotSetKey      = "hot"
	TorrentHotSetCapacity = 1000

	TrackerStatsCacheKey   = "tracker:stats:cache"
	TrackerTorrentStatsKey = "tracker:stats:torrent"
	TrackerPeerStatsKey    = "tracker:stats:peer"
	TrackerTrafficStatsKey = "tracker:stats:traffic"

	TrackerTorrentCountPattern = "torrent:*:info"
	TrackerPeerCountPattern    = "torrent:*:peer:*"

	TorrentExpiration = time.Hour * 24
	PeerExpiration    = time.Hour * 6
)
