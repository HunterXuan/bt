package job

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/HunterXuan/bt/app/domain/model"
	"github.com/HunterXuan/bt/app/domain/service"
	"github.com/HunterXuan/bt/app/infra/config"
	"github.com/HunterXuan/bt/app/infra/constants"
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

			tc := time.NewTimer(time.Minute)
			select {
			case dht.WorkingInfoHashes <- infoHash:
				break
			case <-tc.C:
				log.Println("DHT add worker timeout:", infoHash)
				return
			}

			defer func() {
				if len(dht.WorkingInfoHashes) > 0 {
					<-dht.WorkingInfoHashes
				}
			}()

			log.Println("DHT start to process torrent with info_hash", infoHash)

			t, err := dht.DHT.AddMagnet(fmt.Sprintf("magnet:?xt=urn:btih:%v&tr=http://%v", infoHash, config.Config.GetString("APP_LISTEN_ADDR")))
			if err != nil {
				log.Println("DHT add magnet err:", infoHash, err)
				return
			}

			tc = time.NewTimer(time.Minute)
			select {
			case <-t.GotInfo():
				break
			case <-tc.C:
				log.Println("DHT get info timeout", infoHash)
				return
			}

			metaInfo := t.Metainfo()
			t.Drop()

			var info metainfo.Info
			if err := bencode.Unmarshal(metaInfo.InfoBytes, &info); err != nil {
				log.Println("DHT unmarshal info err:", infoHash, err)
				return
			}

			if jsonInfo, err := json.Marshal(info); err != nil {
				log.Println("DHT marshal info err:", infoHash, err)
			} else if err := db.RDB.HSet(context.Background(), service.GenTorrentInfoKey(infoHash), constants.TorrentMetaInfoKey, jsonInfo, 0); err != nil {
				log.Println("DHT update info err:", infoHash, err)
			} else {
				log.Println("DHT update info success:", infoHash)
			}
		}()
	}

	log.Println("DHT finish collecting torrents info")
}

func getHotTorrents() model.TorrentSlice {
	ctx := context.Background()
	limit := 10

	hotInfoHashes, err := db.RDB.ZRange(ctx, constants.ActiveTorrentSetKey, 0, constants.TorrentHotCapacity*5).Result()
	if err != nil {
		return nil
	}

	var torrents model.TorrentSlice
	for _, infoHash := range hotInfoHashes {
		if len(torrents) > limit {
			break
		}

		torrentInfo, err := db.RDB.HGetAll(ctx, service.GenTorrentInfoKey(infoHash)).Result()
		if err != nil {
			continue
		}

		if len(torrentInfo[constants.TorrentMetaInfoKey]) == 0 {
			torrents = append(torrents, model.Torrent{
				InfoHash: infoHash,
			})
		}
	}

	return torrents
}
