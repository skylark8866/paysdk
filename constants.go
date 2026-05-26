package xgdnpay

import "github.com/skylark8866/paysdk/protocol"

const (
	DefaultBaseURL = "https://pay.xgdn.net"

	ContentTypeJSON = protocol.ContentTypeJSON
	DateTimeFormat  = protocol.DateTimeFormat
)

const (
	PathOrderCreate  = protocol.PathOrderCreate
	PathOrderQuery   = protocol.PathOrderQuery
	PathOrderCheck   = protocol.PathOrderCheck
	PathOrderClose   = protocol.PathOrderClose
	PathRefund       = protocol.PathRefund
	PathRefundQuery  = protocol.PathRefundQuery
	PathRefundOrder  = protocol.PathRefundOrder
	PathRefundInfo   = protocol.PathRefundInfo
)

const (
	FieldAppID         = protocol.FieldAppID
	FieldOrderNo       = protocol.FieldOrderNo
	FieldOutOrderNo    = protocol.FieldOutOrderNo
	FieldAmount        = protocol.FieldAmount
	FieldTitle         = protocol.FieldTitle
	FieldPayType       = protocol.FieldPayType
	FieldOpenID        = protocol.FieldOpenID
	FieldReturnURL     = protocol.FieldReturnURL
	FieldNotifyURL     = protocol.FieldNotifyURL
	FieldExtra         = protocol.FieldExtra
	FieldRefundNo      = protocol.FieldRefundNo
	FieldReason        = protocol.FieldReason
	FieldStatus        = protocol.FieldStatus
	FieldTransactionID = protocol.FieldTransactionID
	FieldPaidAt        = protocol.FieldPaidAt
	FieldTimestamp     = protocol.FieldTimestamp
	FieldNonce         = protocol.FieldNonce
	FieldSign          = protocol.FieldSign
	FieldData          = protocol.FieldData
	FieldAppSecret     = protocol.FieldAppSecret
	FieldSuccessTime   = protocol.FieldSuccessTime
	FieldRefundAmount  = protocol.FieldRefundAmount
	FieldRefundReason  = protocol.FieldRefundReason
	FieldCreatedAt     = protocol.FieldCreatedAt
	FieldOrderAmount   = protocol.FieldOrderAmount
	FieldTotalRefunded = protocol.FieldTotalRefunded
	FieldRemaining     = protocol.FieldRemaining
	FieldCanRefund     = protocol.FieldCanRefund
	FieldMessage       = protocol.FieldMessage
	FieldCode          = protocol.FieldCode
	FieldPayURL        = protocol.FieldPayURL
	FieldCodeURL       = protocol.FieldCodeURL
	FieldError         = protocol.FieldError
	FieldContentType   = protocol.FieldContentType
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
	MinOutOrderNoLength = protocol.MinOutOrderNoLength
	MaxOutOrderNoLength = protocol.MaxOutOrderNoLength
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
	OrderNoPrefix  = protocol.OrderNoPrefix
	RefundNoPrefix = protocol.RefundNoPrefix

	OrderNoRandomDigits = protocol.OrderNoRandomDigits
)

const (
	DefaultMaxDelay = protocol.DefaultMaxDelay
)
