package xgdnpay

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const DefaultMaxDelay int64 = 300

type NotifyHandler struct {
	client   *Client
	maxDelay int64
	handler  func(*NotifyRequest) error
}

type NotifyHandlerOption func(*NotifyHandler)

func WithMaxDelay(delay int64) NotifyHandlerOption {
	return func(h *NotifyHandler) {
		h.maxDelay = delay
	}
}

func NewNotifyHandler(client *Client, handler func(*NotifyRequest) error, opts ...NotifyHandlerOption) *NotifyHandler {
	h := &NotifyHandler{
		client:   client,
		maxDelay:  DefaultMaxDelay,
		handler:   handler,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *NotifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, 405, "方法不允许")
		return
	}

	var req NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, 400, "请求格式错误")
		return
	}

	if err := VerifyNotify(&req, h.client.AppSecret(), h.maxDelay); err != nil {
		h.writeError(w, 401, err.Error())
		return
	}

	if err := h.handler(&req); err != nil {
		h.writeError(w, 500, err.Error())
		return
	}

	h.writeSuccess(w)
}

type RefundNotifyHandler struct {
	client   *Client
	maxDelay int64
	handler  func(*RefundNotifyRequest) error
}

type RefundNotifyHandlerOption func(*RefundNotifyHandler)

func WithRefundMaxDelay(delay int64) RefundNotifyHandlerOption {
	return func(h *RefundNotifyHandler) {
		h.maxDelay = delay
	}
}

func NewRefundNotifyHandler(client *Client, handler func(*RefundNotifyRequest) error, opts ...RefundNotifyHandlerOption) *RefundNotifyHandler {
	h := &RefundNotifyHandler{
		client:   client,
		maxDelay:  DefaultMaxDelay,
		handler:   handler,
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *RefundNotifyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, 405, "方法不允许")
		return
	}

	var req RefundNotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, 400, "请求格式错误")
		return
	}

	if err := VerifyRefundNotify(&req, h.client.AppSecret(), h.maxDelay); err != nil {
		h.writeError(w, 401, err.Error())
		return
	}

	if err := h.handler(&req); err != nil {
		h.writeError(w, 500, err.Error())
		return
	}

	h.writeSuccess(w)
}

func ParseNotifyRequest(body []byte) (*NotifyRequest, error) {
	var req NotifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func ParseRefundNotifyRequest(body []byte) (*RefundNotifyRequest, error) {
	var req RefundNotifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func VerifyNotify(req *NotifyRequest, appSecret string, maxDelay int64) error {
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

func VerifyRefundNotify(req *RefundNotifyRequest, appSecret string, maxDelay int64) error {
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

func (h *NotifyHandler) writeSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"code":    0,
		"message": "成功",
	})
}

func (h *NotifyHandler) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"code":    code,
		"message": message,
	})
}

func (h *RefundNotifyHandler) writeSuccess(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"code":    0,
		"message": "成功",
	})
}

func (h *RefundNotifyHandler) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
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
