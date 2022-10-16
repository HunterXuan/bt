package service

import (
	"fmt"
	"github.com/HunterXuan/bt/app/infra/constants"
)

func GenTorrentInfoKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentInfoKeyTemplate, infoHash)
}

func GenPeerKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentPeerSetKeyTemplate, infoHash)
}
