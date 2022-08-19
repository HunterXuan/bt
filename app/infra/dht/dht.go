package dht

import (
	"encoding/json"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/infra/db"
	trackerUtil "github.com/HunterXuan/bt/app/infra/util/tracker"
	"github.com/anacrolix/torrent/bencode"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/shiyanhui/dht"
	"log"
)

var DHT *dht.Wire

func InitDHT() {
	log.Println("DHT Initializing...")

	DHT = dht.NewWire(1024*8, 1024, 128)

	go func() {
		for resp := range DHT.Response() {
			log.Println("dht response:", resp.InfoHash)

			var info metainfo.Info
			if err := bencode.Unmarshal(resp.MetadataInfo, info); err != nil {
				log.Println("dht unmarshal info err:", err)
				continue
			}

			if jsonInfo, err := json.Marshal(info); err != nil {
				log.Println("dht marshal info err:", err)
			} else if err := db.DB.Model(&model.Torrent{}).
				Where("info_hash = ?", trackerUtil.RestoreToHexString(string(resp.InfoHash))).
				Updates(map[string]interface{}{"meta_info": string(jsonInfo)}).Error; err != nil {
				log.Println("dht update info err:", err)
			} else {
				log.Println("dht update info success:", jsonInfo)
			}
		}
	}()

	go DHT.Run()

	log.Println("DHT Initialized!")
}
