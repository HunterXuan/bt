package error

type BizErrorStruct struct {
	Code    int
	Message string
}

var BizErrorMap = map[string]BizErrorStruct{
	// 1~1000: Default
	"DEFAULT__BAD_REQUEST":    {Code: 400, Message: "请求错误"},
	"DEFAULT__UNAUTHORIZED":   {Code: 401, Message: "需要登录"},
	"DEFAULT__FORBIDDEN":      {Code: 403, Message: "没有权限"},
	"DEFAULT__NOT_FOUND":      {Code: 404, Message: "资源不存在"},
	"DEFAULT__INTERNAL_ERROR": {Code: 500, Message: "内部错误"},

	// 1000~2000: Tracker
	"TRACKER__INVALID_CLIENT": {Code: 1000, Message: "客户端不合法"},
	"TRACKER__INVALID_PARAMS": {Code: 1001, Message: "请求参数不合法"},

	// 2000~3000: Stats
	"STATS__INVALID_PARAMS": {Code: 2000, Message: "请求参数不合法"},
}
