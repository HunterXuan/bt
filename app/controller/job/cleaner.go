package job

import (
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/infra/db"
	"log"
	"time"
)

type Cleaner struct{}

const ActiveTorrentTtl = 24 * time.Hour
const ActivePeerTtl = 6 * time.Hour

func (cleaner *Cleaner) Run() {
	log.Println("Cleaner start working")

	log.Println("Cleaner cleaning dead torrents")
	if deleteRes := db.DB.Unscoped().Where("updated_at < ?", time.Now().Add(-ActiveTorrentTtl)).Delete(model.Torrent{}); deleteRes.Error != nil {
		log.Println("Cleaner (err):", deleteRes.Error)
	}

	log.Println("Cleaner cleaning dead peers")
	if deleteRes := db.DB.Unscoped().Where("updated_at < ?", time.Now().Add(-ActivePeerTtl)).Delete(model.Peer{}); deleteRes.Error != nil {
		log.Println("Cleaner (err):", deleteRes.Error)
	}

	log.Println("Cleaner finish working")
}
