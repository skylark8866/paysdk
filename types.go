package xgdnpay

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/skylark8866/paysdk/sse"
)

type PayType string

const (
	PayTypeNative PayType = "native"
	PayTypeJSAPI  PayType = "jsapi"
)

var validPayTypes = map[PayType]bool{
	PayTypeNative: true,
	PayTypeJSAPI:  true,
}

func (t PayType) IsValid() bool {
	return validPayTypes[t]
}

func (t PayType) String() string {
	return string(t)
}

func NormalizePayType(raw string) PayType {
	lowered := PayType(strings.ToLower(raw))
	if lowered.IsValid() {
		return lowered
	}
	return PayType(strings.ToLower(raw))
}

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
	return fmt.Sprintf("%d", int(s))
}

func (s OrderStatus) Text() string {
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
		return fmt.Sprintf("未知状态(%d)", int(s))
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

func (s RefundStatus) String() string {
	return fmt.Sprintf("%d", int(s))
}

func (s RefundStatus) Text() string {
	switch s {
	case RefundStatusProcessing:
		return "退款中"
	case RefundStatusSuccess:
		return "退款成功"
	case RefundStatusClosed:
		return "已关闭"
	case RefundStatusFailed:
		return "退款失败"
	case RefundStatusAbnormal:
		return "退款异常"
	default:
		return fmt.Sprintf("未知状态(%d)", int(s))
	}
}

type SDKError struct {
	Code    int
	Message string
}

func (e *SDKError) Error() string {
	return e.Message
}

func (e *SDKError) Is(err error) bool {
	target, ok := err.(*SDKError)
	if !ok {
		return false
	}
	return e.Code == target.Code
}

func NewSDKError(code int, message string) *SDKError {
	return &SDKError{Code: code, Message: message}
}

var (
	ErrInvalidParam   = &SDKError{Code: -1, Message: "参数错误"}
	ErrRequestFailed  = &SDKError{Code: -2, Message: "请求失败"}
	ErrSignFailed     = &SDKError{Code: -3, Message: "签名验证失败"}
	ErrTimeout        = &SDKError{Code: -4, Message: "请求超时"}
	ErrOrderNotFound  = &SDKError{Code: -5, Message: "订单不存在"}
	ErrRefundNotFound = &SDKError{Code: -6, Message: "退款单不存在"}
)

type CreateOrderRequest struct {
	OutOrderNo string                 `json:"out_order_no"`
	Amount     float64                `json:"amount"`
	Title      string                 `json:"title"`
	PayType    PayType                `json:"pay_type"`
	OpenID     string                 `json:"openid,omitempty"`
	ReturnURL  string                 `json:"return_url,omitempty"`
	NotifyURL  string                 `json:"notify_url,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

type CreateOrderResponse struct {
	OrderNo string `json:"order_no"`
	PayURL  string `json:"pay_url"`
	CodeURL string `json:"code_url"`
}

type QueryOrderResponse struct {
	OrderNo       string  `json:"order_no"`
	Status        int     `json:"status"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id,omitempty"`
	PaidAt        string  `json:"paid_at,omitempty"`
}

type CheckStatusResponse struct {
	Status    int    `json:"status"`
	PaidAt    string `json:"paid_at,omitempty"`
	ReturnURL string `json:"return_url,omitempty"`
}

type RefundRequest struct {
	OrderNo   string  `json:"order_no"`
	RefundNo  string  `json:"refund_no"`
	Amount    float64 `json:"amount"`
	Reason    string  `json:"reason"`
	NotifyURL string  `json:"notify_url,omitempty"`
}

type RefundResponse struct {
	ID           uint64  `json:"id"`
	RefundNo     string  `json:"refund_no"`
	OrderNo      string  `json:"order_no"`
	RefundAmount float64 `json:"refund_amount"`
	RefundReason string  `json:"refund_reason"`
	Status       int     `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

type OrderRefundInfo struct {
	OrderNo         string  `json:"order_no"`
	OrderAmount     float64 `json:"order_amount"`
	TotalRefunded   float64 `json:"total_refunded"`
	RemainingAmount float64 `json:"remaining_amount"`
	CanRefund       bool    `json:"can_refund"`
	Message         string  `json:"message,omitempty"`
}

type NotifyRequest struct {
	AppID         string  `json:"app_id"`
	OrderNo       string  `json:"order_no"`
	OutOrderNo    string  `json:"out_order_no"`
	Amount        float64 `json:"amount"`
	Title         string  `json:"title"`
	PayType       string  `json:"pay_type"`
	Status        int     `json:"status"`
	TransactionID string  `json:"transaction_id"`
	PaidAt        string  `json:"paid_at"`
	Timestamp     string  `json:"timestamp"`
	Nonce         string  `json:"nonce"`
	Sign          string  `json:"sign"`
}

type RefundNotifyRequest struct {
	RefundNo      string  `json:"refund_no"`
	OrderNo       string  `json:"order_no"`
	TransactionID string  `json:"transaction_id"`
	Amount        float64 `json:"amount"`
	Status        string  `json:"status"`
	SuccessTime   string  `json:"success_time"`
	Timestamp     string  `json:"timestamp"`
	Sign          string  `json:"sign"`
}

type SignedRequest struct {
	AppID     string          `json:"app_id"`
	Timestamp string          `json:"timestamp"`
	Nonce     string          `json:"nonce"`
	Data      json.RawMessage `json:"data"`
	Sign      string          `json:"sign"`
}

type apiResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type PayNotifyMessage struct {
	OrderNo     string     `json:"order_no"`
	OutOrderNo  string     `json:"out_order_no"`
	Amount      float64    `json:"amount"`
	Status      PayStatus  `json:"status"`
	PaidAt      string     `json:"paid_at"`
	PayType     PayChannel `json:"pay_type"`
	Transaction string     `json:"transaction_id,omitempty"`
}

func NewPayNotifyMessage(orderNo string, amount float64, status PayStatus) *PayNotifyMessage {
	return &PayNotifyMessage{
		OrderNo: orderNo,
		Amount:  amount,
		Status:  status,
		PaidAt:  time.Now().Format(DateTimeFormat),
	}
}

func (m *PayNotifyMessage) SetOutOrderNo(no string) *PayNotifyMessage {
	m.OutOrderNo = no
	return m
}

func (m *PayNotifyMessage) SetPayType(payType PayChannel) *PayNotifyMessage {
	m.PayType = payType
	return m
}

func (m *PayNotifyMessage) SetTransaction(id string) *PayNotifyMessage {
	m.Transaction = id
	return m
}

func (m *PayNotifyMessage) EventName() sse.EventName {
	return sse.EventPayNotify
}

func (m *PayNotifyMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}
