package xgdnpay

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NotifyHandler[T any] struct {
	client   *Client
	maxDelay int64
	handler  func(*T) error
	verify   func(*T, string, int64) error
}

type NotifyHandlerOption[T any] func(*NotifyHandler[T])

func WithNotifyMaxDelay[T any](delay int64) NotifyHandlerOption[T] {
	return func(h *NotifyHandler[T]) {
		h.maxDelay = delay
	}
}

func NewNotifyHandler(client *Client, handler func(*NotifyRequest) error, opts ...NotifyHandlerOption[NotifyRequest]) *NotifyHandler[NotifyRequest] {
	h := &NotifyHandler[NotifyRequest]{
		client:   client,
		maxDelay: DefaultMaxDelay,
		handler:  handler,
		verify:   VerifyNotifyWithSecret,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func NewRefundNotifyHandler(client *Client, handler func(*RefundNotifyRequest) error, opts ...NotifyHandlerOption[RefundNotifyRequest]) *NotifyHandler[RefundNotifyRequest] {
	h := &NotifyHandler[RefundNotifyRequest]{
		client:   client,
		maxDelay: DefaultMaxDelay,
		handler:  handler,
		verify:   verifyRefundNotifyWrapper,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func verifyRefundNotifyWrapper(req *RefundNotifyRequest, secret string, maxDelay int64) error {
	return VerifyRefundNotifyWithSecret(req, secret, maxDelay)
}

func (h *NotifyHandler[T]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeNotifyError(w, NotifyRespCodeMethod, NotifyRespMsgMethodDenied)
		return
	}

	var req T
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeNotifyError(w, NotifyRespCodeBadRequest, NotifyRespMsgBadFormat)
		return
	}

	if err := h.verify(&req, h.client.getSecret(), h.maxDelay); err != nil {
		writeNotifyError(w, NotifyRespCodeUnauthorized, err.Error())
		return
	}

	if err := h.handler(&req); err != nil {
		writeNotifyError(w, NotifyRespCodeInternal, err.Error())
		return
	}

	writeNotifySuccess(w)
}

func VerifyNotifyWithSecret(req *NotifyRequest, appSecret string, maxDelay int64) error {
	params := map[string]string{
		"app_id":         req.AppID,
		"order_no":       req.OrderNo,
		"out_order_no":   req.OutOrderNo,
		"amount":         formatAmount(req.Amount),
		"title":          req.Title,
		"pay_type":       req.PayType,
		"status":         formatInt(req.Status),
		"transaction_id": req.TransactionID,
		"paid_at":        req.PaidAt,
		"timestamp":      req.Timestamp,
		"nonce":          req.Nonce,
		"sign":           req.Sign,
	}
	return VerifySign(params, appSecret, maxDelay)
}

func VerifyRefundNotifyWithSecret(req *RefundNotifyRequest, appSecret string, maxDelay int64) error {
	params := map[string]string{
		"refund_no":      req.RefundNo,
		"order_no":       req.OrderNo,
		"transaction_id": req.TransactionID,
		"amount":         formatAmount(req.Amount),
		"status":         req.Status,
		"success_time":   req.SuccessTime,
		"timestamp":      req.Timestamp,
		"sign":           req.Sign,
	}
	return VerifySign(params, appSecret, maxDelay)
}

func VerifyNotify(req *NotifyRequest, appSecret string, maxDelay int64) error {
	return VerifyNotifyWithSecret(req, appSecret, maxDelay)
}

func VerifyRefundNotify(req *RefundNotifyRequest, appSecret string, maxDelay int64) error {
	return VerifyRefundNotifyWithSecret(req, appSecret, maxDelay)
}

func writeNotifySuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"code":    NotifyRespCodeSuccess,
		"message": NotifyRespMsgSuccess,
	})
}

func writeNotifyError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"code":    code,
		"message": message,
	})
}

func formatAmount(amount float64) string {
	return fmt.Sprintf("%.2f", amount)
}

func formatInt(v int) string {
	return fmt.Sprintf("%d", v)
}
