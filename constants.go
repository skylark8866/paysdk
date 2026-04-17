package xgdnpay

const (
	DefaultBaseURL = "https://pay.xgdn.net"

	ContentTypeJSON = "application/json"
)

const (
	PathOrderCreate = "/api/v1/order/create"
	PathOrderQuery  = "/api/v1/order/query"
	PathOrderCheck  = "/api/v1/order/check"
	PathOrderClose  = "/api/v1/order/close"
	PathRefund      = "/api/v1/refund"
	PathRefundQuery = "/api/v1/refund/query"
	PathRefundOrder = "/api/v1/refund/order"
	PathRefundInfo  = "/api/v1/refund/info"
)

const (
	ErrMsgAmountRequired    = "金额必须大于0"
	ErrMsgTitleRequired     = "商品标题不能为空"
	ErrMsgOpenIDRequired    = "JSAPI 支付必须提供 openid"
	ErrMsgOrderNoRequired   = "订单号不能为空"
	ErrMsgRefundRequired    = "退款金额必须大于0"
	ErrMsgRefundDefault     = "用户申请退款"
	ErrMsgSignNotFound      = "签名不存在"
	ErrMsgTimestampNotFound = "时间戳不存在"
	ErrMsgTimestampInvalid  = "时间戳格式错误"
	ErrMsgRequestExpired    = "请求已过期"
	ErrMsgSignVerifyFail    = "签名验证失败"
	ErrMsgParseNotifyFail   = "parse notify failed"
	ErrMsgParseRefundFail   = "parse refund notify failed"
)

const (
	NotifyRespCodeSuccess      = 0
	NotifyRespCodeMethod       = 405
	NotifyRespCodeBadRequest   = 400
	NotifyRespCodeUnauthorized = 401
	NotifyRespCodeInternal     = 500

	NotifyRespMsgMethodDenied = "方法不允许"
	NotifyRespMsgBadFormat    = "请求格式错误"
	NotifyRespMsgSuccess      = "成功"
)

const (
	OrderNoPrefix  = "ORD_"
	RefundNoPrefix = "REF_"

	OrderNoRandomDigits = 6
)

const (
	DefaultMaxDelay int64 = 300
)
