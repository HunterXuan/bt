package model

import "time"

type Torrent struct {
	Base
	InfoHash      string `gorm:"size:40;not null;unique"`
	SeederCount   uint64 `gorm:"not null"`
	LeecherCount  uint64 `gorm:"not null"`
	SnatcherCount uint64 `gorm:"not null"`
	LastActiveAt  time.Time
}

type TorrentSlice []Torrent
