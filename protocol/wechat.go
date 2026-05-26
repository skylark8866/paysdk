package protocol

const (
	WechatTradeStateSuccess  = "SUCCESS"
	WechatTradeStateNotPay   = "NOTPAY"
	WechatTradeStateClosed   = "CLOSED"
	WechatTradeStateRefund   = "REFUND"
	WechatTradeStatePayError = "PAYERROR"

	WechatRefundStatusSuccess  = "SUCCESS"
	WechatRefundStatusClosed   = "CLOSED"
	WechatRefundStatusProcessing = "PROCESSING"
	WechatRefundStatusAbnormal = "ABNORMAL"

	WechatCallbackSuccess = "SUCCESS"

	WechatCurrencyCNY = "CNY"
)
