package job

import (
	"encoding/json"
	"fmt"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/infra/config"
	"github.com/HunterXuan/bt/app/infra/db"
	"github.com/HunterXuan/bt/app/infra/dht"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"log"
	"time"
)

type DHT struct{}

func (d *DHT) Run() {
	log.Println("DHT start collecting torrents info")

	for _, item := range getHotTorrents() {
		infoHash := item.InfoHash

		go func() {
			log.Println("DHT waiting to process torrent with info_hash", infoHash)

			dht.WorkingInfoHashes <- infoHash
			defer func() {
				if len(dht.WorkingInfoHashes) > 0 {
					<-dht.WorkingInfoHashes
				}
			}()

			log.Println("DHT start to process torrent with info_hash", infoHash)

			t, err := dht.DHT.AddMagnet(fmt.Sprintf("magnet:?xt=urn:btih:%v&tr=http://%v", infoHash, config.Config.GetString("APP_LISTEN_ADDR")))
			if err != nil {
				log.Println("DHT add magnet err:", err)
				return
			}

			tc := time.NewTimer(time.Minute)
			select {
			case <-t.GotInfo():
				break
			case <-tc.C:
				log.Println("DHT get info timeout")
				return
			}

			metaInfo := t.Metainfo()
			t.Drop()

			var info metainfo.Info
			if err := bencode.Unmarshal(metaInfo.InfoBytes, &info); err != nil {
				log.Println("DHT unmarshal info err:", err)
				return
			}

			if jsonInfo, err := json.Marshal(metaInfo.InfoBytes); err != nil {
				log.Println("DHT marshal info err:", err)
			} else if err := db.DB.Model(&model.Torrent{}).
				Where("info_hash = ?", infoHash).
				Updates(map[string]interface{}{"meta_info": string(jsonInfo)}).Error; err != nil {
				log.Println("DHT update info err:", err)
			} else {
				log.Println("DHT update info success:", jsonInfo)
			}
		}()
	}

	log.Println("DHT finish collecting torrents info")
}

func getHotTorrents() []model.Torrent {
	var torrents []model.Torrent
	findRes := db.DB.Model(&model.Torrent{}).
		Select("id, info_hash").
		Where("meta_info = ?", "").
		Order("leecher_count desc").
		Limit(8).
		Find(&torrents)
	if findRes.Error != nil {
		log.Println("DHT getHotTorrents Err:", findRes.Error)
	} else {
		log.Println("DHT getHotTorrents Torrent Count:", len(torrents))
	}

	return torrents
}
