package dht

import (
	"github.com/anacrolix/torrent"
	"log"
)

var (
	DHT               *torrent.Client
	WorkingInfoHashes = make(chan string, 8)
)

func InitDHT() {
	log.Println("DHT Initializing...")

	var err error
	DHT, err = torrent.NewClient(nil)
	if err != nil {
		log.Fatalf("Err: %v", err)
	}

	log.Println("DHT Initialized!")
}
