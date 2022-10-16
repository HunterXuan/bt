package constants

import "time"

const (
	ActiveTorrentSetKey = "active:torrent"
	ActivePeerSetKey    = "active:peer"

	TorrentInfoKeyTemplate    = "torrent:%s:info"
	TorrentPeerSetKeyTemplate = "torrent:%s:peer"

	TorrentInfoHashKey      = "info_hash"
	TorrentSeederCountKey   = "seeder_count"
	TorrentLeecherCountKey  = "leecher_count"
	TorrentSnatcherCountKey = "snatcher_count"
	TorrentMetaInfoKey      = "meta_info"
	TorrentCreatedAtKey     = "created_at"
	TorrentLastActiveAt     = "active_at"

	TorrentHotCapacity = 100

	StatsKey             = "stats"
	StatsTorrentCountKey = "torrent"
	StatsPeerCountKey    = "peer"
	StatsTrafficCountKey = "traffic"
	StatsCacheKey        = "stats:cache"

	TrackerPeerCountPattern = "torrent:*:peer"

	TorrentExpiration = time.Hour * 24
	PeerExpiration    = time.Hour * 6
)
