package model

type Peer struct {
	InfoHash        string
	PeerID          string
	Agent           string
	Ipv4            string
	Ipv6            string
	Port            uint32
	UploadedBytes   uint64
	DownloadedBytes uint64
	LeftBytes       uint64
	IsSeeder        bool
	IsConnectable   bool
	FinishedAt      int64
}

type PeerSlice []Peer

type PeerAggResult struct {
	TorrentID   uint64
	SeederCount uint64
	PeerCount   uint64
}
