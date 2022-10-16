package model

type Torrent struct {
	InfoHash      string
	SeederCount   uint64
	LeecherCount  uint64
	SnatcherCount uint64
	MetaInfo      string
	CreatedAt     int64
	LastActiveAt  int64
}

type TorrentSlice []Torrent
