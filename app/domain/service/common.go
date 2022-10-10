package service

import (
	"fmt"
	"github.com/HunterXuan/bt/app/infra/constants"
)

func GenTorrentPlaceHoldKey(infoHash string) string {
	return fmt.Sprintf(constants.TorrentPlaceHoldTemplate, infoHash)
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

func GenPeerKey(infoHash string, peerID string) string {
	return fmt.Sprintf(constants.TorrentPeerKeyTemplate, infoHash, peerID)
}

func GenPeerSearchPattern(infoHash string) string {
	return fmt.Sprintf(constants.TorrentPeerSearchKeyTemplate, infoHash)
}
