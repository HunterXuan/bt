package tracker

import (
	customError "github.com/HunterXuan/bt/app/infra/util/error"
	trackerUtil "github.com/HunterXuan/bt/app/infra/util/tracker"
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"strings"
)

type AnnounceRequest struct {
	Agent           string `header:"User-Agent"`                 // 客户端
	OriInfoHash     string `form:"info_hash" binding:"required"` // 客户端上报的种子哈希码
	OriPeerID       string `form:"peer_id" binding:"required"`   // 客户端上报的同伴ID
	InfoHash        string ``                                    // 转换过后的哈希码
	PeerID          string ``                                    // 转换过后的同伴ID
	Port            uint32 `form:"port" binding:"min=1,max=65535"`
	DownloadedBytes uint64 `form:"downloaded"`
	UploadedBytes   uint64 `form:"uploaded"`
	LeftBytes       uint64 `form:"left"`
	Event           string `form:"event"`
	IP              string `form:"ip" binding:"omitempty,ip"`
	IPv4            string `form:"ipv4" binding:"omitempty,ip"`
	IPv6            string `form:"ipv6" binding:"omitempty,ip"`
	Compact         uint8  `form:"compact" binding:"omitempty,min=0,max=1"`
	NoPeerID        uint8  `form:"no_peer_id" binding:"omitempty,min=0,max=1"`
	NumWanted       uint8  `form:"numwant" binding:"omitempty"`
}

// ValidateRequest 校验参数合法性
// 通常校验不应该修改 request
// 但为了避免后续反复转换 InfoHash 和 PeerID，此处允许做修改
func (request *AnnounceRequest) ValidateRequest(ctx *gin.Context) *customError.CustomError {
	// TODO: 客户端白名单校验

	// 校验参数格式
	if err := ctx.ShouldBind(request); err != nil {
		log.Println("AnnounceRequest ValidateRequest ShouldBind Err:", err)
		return customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	if err := request.Refine(ctx); err != nil {
		log.Println("AnnounceRequest ValidateRequest Refine Err:", err)
		return customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	{
		// 转换 InfoHash 和 PeerID
		request.InfoHash = trackerUtil.RestoreToHexString(request.OriInfoHash)
		request.PeerID = trackerUtil.RestoreToHexString(request.OriPeerID)
	}

	return nil
}

// ValidateAuth 校验是否有权限
func (request *AnnounceRequest) ValidateAuth(ctx *gin.Context) *customError.CustomError {
	return nil
}

// Refine 修正参数，方便后续处理
func (request *AnnounceRequest) Refine(ctx *gin.Context) *customError.CustomError {
	if err := fillAgent(ctx, request); err != nil {
		return customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	if err := seedIP(ctx, request); err != nil {
		return customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	if err := convertHexCode(request); err != nil {
		return customError.NewBadRequestError("TRACKER__INVALID_PARAMS")
	}

	return nil
}

// fillAgent
func fillAgent(ctx *gin.Context, request *AnnounceRequest) error {
	// Gin Bind 可能失效
	if request.Agent == "" {
		request.Agent = ctx.Request.Header.Get("User-Agent")
	}

	return nil
}

// seedIP 处理 IP
// 优先级：IP > IPv4 > IPv6 > ctx.ip
func seedIP(ctx *gin.Context, request *AnnounceRequest) error {
	// 如果上报的 IPv4 地址有误，则清空
	if ip := net.ParseIP(request.IPv4); ip == nil || !strings.Contains(request.IPv4, ".") {
		request.IPv4 = ""
	}

	// 如果上报的 IPv6 地址有误，则清空
	if ip := net.ParseIP(request.IPv6); ip == nil || !strings.Contains(request.IPv6, ":") {
		request.IPv6 = ""
	}

	// 如果上报的 IP 地址有效，则覆盖对应的 IPv4/IPv6 地址
	if ip := net.ParseIP(request.IP); ip != nil {
		if strings.Contains(request.IP, ":") {
			request.IPv6 = request.IP
		} else {
			request.IPv4 = request.IP
		}
	}

	// 如果均为空，则使用客户端地址填充
	if request.IPv4 == "" && request.IPv6 == "" {
		clientIP := ctx.ClientIP()
		if ip := net.ParseIP(clientIP); ip != nil {
			if strings.Contains(clientIP, ":") {
				request.IPv6 = clientIP
			} else {
				request.IPv4 = clientIP
			}
		}
	}

	return nil
}

// convertHexCode 转换 InfoHash 和 PeerID
func convertHexCode(request *AnnounceRequest) error {
	request.InfoHash = trackerUtil.RestoreToHexString(request.OriInfoHash)
	request.PeerID = trackerUtil.RestoreToHexString(request.OriPeerID)

	return nil
}
