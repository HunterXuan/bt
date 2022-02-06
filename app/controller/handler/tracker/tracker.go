package tracker

import (
	"github.com/HunterXuan/bt/app/controller/request"
	"github.com/HunterXuan/bt/app/controller/request/tracker"
	trackerResp "github.com/HunterXuan/bt/app/controller/response/tracker"
	trackerService "github.com/HunterXuan/bt/app/domain/service/tracker"
	"github.com/HunterXuan/bt/app/infra/util/http"
	"github.com/anacrolix/torrent/bencode"
	"github.com/gin-gonic/gin"
	"log"
)

func Announce(ctx *gin.Context) {
	// 使用 AnnounceRequest 校验请求
	req := &tracker.AnnounceRequest{}
	if _, err := request.Validator.Validate(
		ctx,
		req,
	); err != nil {
		log.Println("Announce (Validate):", err)
		ctx.String(err.Status, packFailReason(err.Error()))
		return
	}

	torrent, peerSlice, err := trackerService.DealWithClientReport(ctx, req)
	if err != nil {
		log.Println("Announce (DealWithClientReport):", err)
		ctx.String(err.Status, packFailReason(err.Error()))
		return
	}
	resp := &trackerResp.AnnounceResponse{}
	ctx.String(http.StatusOK, resp.BEncode(torrent, peerSlice, req))
}

func Scrape(ctx *gin.Context) {
	// 使用 ScrapeRequest 校验请求
	req := &tracker.ScrapeRequest{}
	if _, err := request.Validator.Validate(
		ctx,
		req,
	); err != nil {
		log.Println("Scrape (Validate):", err)
		ctx.String(err.Status, packFailReason(err.Error()))
		return
	}

	// 查找符合条件的种子
	torrentSlice, _ := trackerService.GenScrapeResult(ctx, req)
	resp := &trackerResp.ScrapeResponse{}
	ctx.String(http.StatusOK, resp.BEncode(torrentSlice))
}

// packFailReason BEncode错误原因
func packFailReason(reason string) string {
	return string(bencode.MustMarshal(map[string]string{
		"failure reason": reason,
	}))
}
