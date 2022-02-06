package model

import (
	"database/sql"
)

type Peer struct {
	Base
	TorrentID       uint64 `gorm:"index;not null"`
	PeerID          string `gorm:"index;not null"`
	Agent           string `gorm:"size:255;not null"`
	Ipv4            string `gorm:"size:255;not null"`
	Ipv6            string `gorm:"size:255;not null"`
	Port            uint32 `gorm:"not null"`
	UploadedBytes   uint64 `gorm:"not null"`
	DownloadedBytes uint64 `gorm:"not null"`
	LeftBytes       uint64 `gorm:"not null"`
	IsSeeder        uint8  `gorm:"not null"`
	IsConnectable   uint8  `gorm:"not null"`
	FinishedAt      sql.NullTime
}

type PeerSlice []Peer

type PeerAggResult struct {
	TorrentID   uint64
	SeederCount uint64
	PeerCount   uint64
}
