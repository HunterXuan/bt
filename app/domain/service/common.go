package service

import (
	"fmt"
	"github.com/HunterXuan/bt/app/infra/constants"
)

func GenTorrentInfoKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentInfoTemplate, infoHash)
}

func GenTorrentSeederCountKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentSeederCountKeyTemplate, infoHash)
}

func GenTorrentLeecherCountKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentLeecherCountCountKeyTemplate, infoHash)
}

func GenTorrentSnatcherCountKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentSnatcherCountKeyTemplate, infoHash)
}

func GenTorrentMetaInfoKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentMetaInfoKeyTemplate, infoHash)
}

func GenPeerKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentPeerKeyTemplate, infoHash)
}
