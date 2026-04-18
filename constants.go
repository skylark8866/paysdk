package xgdnpay

const (
	DefaultBaseURL = "https://pay.xgdn.net"

	ContentTypeJSON = "application/json"
	DateTimeFormat  = "2006-01-02 15:04:05"
	AmountFormat    = "%.2f"
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
	FieldAppID         = "app_id"
	FieldOrderNo       = "order_no"
	FieldOutOrderNo    = "out_order_no"
	FieldAmount        = "amount"
	FieldTitle         = "title"
	FieldPayType       = "pay_type"
	FieldOpenID        = "openid"
	FieldReturnURL     = "return_url"
	FieldNotifyURL     = "notify_url"
	FieldExtra         = "extra"
	FieldRefundNo      = "refund_no"
	FieldReason        = "reason"
	FieldStatus        = "status"
	FieldTransactionID = "transaction_id"
	FieldPaidAt        = "paid_at"
	FieldTimestamp     = "timestamp"
	FieldNonce         = "nonce"
	FieldSign          = "sign"
	FieldData          = "data"
	FieldAppSecret     = "app_secret"
	FieldSuccessTime   = "success_time"
	FieldRefundAmount  = "refund_amount"
	FieldRefundReason  = "refund_reason"
	FieldCreatedAt     = "created_at"
	FieldOrderAmount   = "order_amount"
	FieldTotalRefunded = "total_refunded"
	FieldRemaining     = "remaining_amount"
	FieldCanRefund     = "can_refund"
	FieldMessage       = "message"
	FieldCode          = "code"
	FieldPayURL        = "pay_url"
	FieldCodeURL       = "code_url"
	FieldError         = "error"
	FieldContentType   = "Content-Type"
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
	ErrMsgMarshalData       = "marshal data failed"
	ErrMsgSortJSON          = "sort json failed"
	ErrMsgOrderClosed       = "订单已关闭"
	ErrMsgOrderRefunded     = "订单已退款"
	ErrMsgInvalidPayType    = "不支持的支付类型"
	ErrMsgOutOrderNoTooLong = "商户订单号长度不能超过64字符"
	ErrMsgOutOrderNoFormat  = "商户订单号只能包含字母、数字、下划线、中划线"
)

const (
	MinOutOrderNoLength = 1
	MaxOutOrderNoLength = 64
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
