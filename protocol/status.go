package protocol

type PayStatus string

const (
	PayStatusPaid    PayStatus = "paid"
	PayStatusPending PayStatus = "pending"
	PayStatusClosed  PayStatus = "closed"
)

var validPayStatuses = map[PayStatus]bool{
	PayStatusPaid:    true,
	PayStatusPending: true,
	PayStatusClosed:  true,
}

func (s PayStatus) IsValid() bool {
	return validPayStatuses[s]
}

func (s PayStatus) String() string {
	return string(s)
}

type OrderStatus int

const (
	OrderStatusPending  OrderStatus = 0
	OrderStatusPaid     OrderStatus = 1
	OrderStatusClosed   OrderStatus = 2
	OrderStatusRefunded OrderStatus = 3
)

var validOrderStatuses = map[OrderStatus]bool{
	OrderStatusPending:  true,
	OrderStatusPaid:     true,
	OrderStatusClosed:   true,
	OrderStatusRefunded: true,
}

func (s OrderStatus) IsValid() bool {
	return validOrderStatuses[s]
}

func (s OrderStatus) String() string {
	switch s {
	case OrderStatusPending:
		return "待支付"
	case OrderStatusPaid:
		return "已支付"
	case OrderStatusClosed:
		return "已关闭"
	case OrderStatusRefunded:
		return "已退款"
	default:
		return "未知状态"
	}
}

type RefundStatus int

const (
	RefundStatusProcessing RefundStatus = 0
	RefundStatusSuccess    RefundStatus = 1
	RefundStatusClosed     RefundStatus = 2
	RefundStatusFailed     RefundStatus = 3
	RefundStatusAbnormal   RefundStatus = 4
)

var validRefundStatuses = map[RefundStatus]bool{
	RefundStatusProcessing: true,
	RefundStatusSuccess:    true,
	RefundStatusClosed:     true,
	RefundStatusFailed:     true,
	RefundStatusAbnormal:   true,
}

func (s RefundStatus) IsValid() bool {
	return validRefundStatuses[s]
}

type PayChannel string

const (
	PayChannelWechat PayChannel = "wechat"
	PayChannelAlipay PayChannel = "alipay"
)

var validPayChannels = map[PayChannel]bool{
	PayChannelWechat: true,
	PayChannelAlipay: true,
}

func (ch PayChannel) IsValid() bool {
	return validPayChannels[ch]
}

func (ch PayChannel) String() string {
	return string(ch)
}
