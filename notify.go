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
		FieldAppID:         req.AppID,
		FieldOrderNo:       req.OrderNo,
		FieldOutOrderNo:    req.OutOrderNo,
		FieldAmount:        formatAmount(req.Amount),
		FieldTitle:         req.Title,
		FieldPayType:       req.PayType,
		FieldStatus:        formatInt(req.Status),
		FieldTransactionID: req.TransactionID,
		FieldPaidAt:        req.PaidAt,
		FieldTimestamp:     req.Timestamp,
		FieldNonce:         req.Nonce,
		FieldSign:          req.Sign,
	}
	return VerifySign(params, appSecret, maxDelay)
}

func VerifyRefundNotifyWithSecret(req *RefundNotifyRequest, appSecret string, maxDelay int64) error {
	params := map[string]string{
		FieldRefundNo:      req.RefundNo,
		FieldOrderNo:       req.OrderNo,
		FieldTransactionID: req.TransactionID,
		FieldAmount:        formatAmount(req.Amount),
		FieldStatus:        req.Status,
		FieldSuccessTime:   req.SuccessTime,
		FieldTimestamp:     req.Timestamp,
		FieldSign:          req.Sign,
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
	w.Header().Set(FieldContentType, ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		FieldCode:    NotifyRespCodeSuccess,
		FieldMessage: NotifyRespMsgSuccess,
	})
}

func writeNotifyError(w http.ResponseWriter, code int, message string) {
	w.Header().Set(FieldContentType, ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		FieldCode:    code,
		FieldMessage: message,
	})
}

func formatAmount(amount float64) string {
	return fmt.Sprintf(AmountFormat, amount)
}

func formatInt(v int) string {
	return fmt.Sprintf("%d", v)
}
