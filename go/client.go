package xgdnpay

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	appID     string
	appSecret string
	baseURL   string
	client    *resty.Client
}

type ClientOption func(*Client)

func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.SetTimeout(timeout)
	}
}

func NewClient(appID, appSecret string, opts ...ClientOption) *Client {
	c := &Client{
		appID:     appID,
		appSecret: appSecret,
		baseURL:   "https://pay.xgdn.net",
		client:    resty.New(),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.client.SetBaseURL(c.baseURL)
	c.client.SetHeader("Content-Type", "application/json")

	return c
}

func (c *Client) AppID() string {
	return c.appID
}

func (c *Client) AppSecret() string {
	return c.appSecret
}

func checkResponse[T any](resp *apiResponse[T]) error {
	if resp.Code != 0 {
		return NewSDKError(resp.Code, resp.Message)
	}
	return nil
}

func generateOrderNo() string {
	return fmt.Sprintf("ORD_%d_%04d", time.Now().UnixNano()/1e6, rand.Intn(10000))
}

func generateRefundNo() string {
	return fmt.Sprintf("REF_%d_%04d", time.Now().UnixNano()/1e6, rand.Intn(10000))
}

func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	if req == nil {
		return nil, ErrInvalidParam
	}

	if req.Amount <= 0 {
		return nil, NewSDKError(-1, "金额必须大于0")
	}

	if req.Title == "" {
		return nil, NewSDKError(-1, "商品标题不能为空")
	}

	outOrderNo := req.OutOrderNo
	if outOrderNo == "" {
		outOrderNo = generateOrderNo()
	}

	payType := NormalizePayType(req.PayType)
	if payType == "" {
		payType = PayTypeNative
	}

	if payType == PayTypeJSAPI && req.OpenID == "" {
		return nil, NewSDKError(-1, "JSAPI 支付必须提供 openid")
	}

	data := map[string]interface{}{
		"out_order_no": outOrderNo,
		"amount":       req.Amount,
		"title":        req.Title,
		"pay_type":     payType,
	}

	if req.OpenID != "" {
		data["openid"] = req.OpenID
	}
	if req.ReturnURL != "" {
		data["return_url"] = req.ReturnURL
	}
	if req.NotifyURL != "" {
		data["notify_url"] = req.NotifyURL
	}
	if req.Extra != nil {
		data["extra"] = req.Extra
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[CreateOrderResponse]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/order/create")

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	if err := checkResponse(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) QueryOrder(ctx context.Context, orderNo string) (*QueryOrderResponse, error) {
	if orderNo == "" {
		return nil, ErrInvalidParam
	}

	data := map[string]interface{}{
		"order_no": orderNo,
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[QueryOrderResponse]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/order/query")

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	if err := checkResponse(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) CheckStatus(ctx context.Context, orderNo string) (*CheckStatusResponse, error) {
	if orderNo == "" {
		return nil, ErrInvalidParam
	}

	data := map[string]interface{}{
		"order_no": orderNo,
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[CheckStatusResponse]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/order/check")

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	if err := checkResponse(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) CloseOrder(ctx context.Context, orderNo string) error {
	if orderNo == "" {
		return ErrInvalidParam
	}

	data := map[string]interface{}{
		"order_no": orderNo,
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[interface{}]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/order/close")

	if err != nil {
		return fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	return checkResponse(&result)
}

func (c *Client) WaitForPayment(ctx context.Context, orderNo string, interval, timeout time.Duration) error {
	if orderNo == "" {
		return ErrInvalidParam
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return ErrTimeout
			}
			return ctx.Err()
		case <-ticker.C:
			result, err := c.CheckStatus(ctx, orderNo)
			if err != nil {
				continue
			}

			switch result.Status {
			case OrderStatusPaid:
				return nil
			case OrderStatusClosed:
				return NewSDKError(-10, "订单已关闭")
			case OrderStatusRefunded:
				return NewSDKError(-11, "订单已退款")
			}
		}
	}
}

func (c *Client) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	if req == nil {
		return nil, ErrInvalidParam
	}

	if req.OrderNo == "" {
		return nil, NewSDKError(-1, "订单号不能为空")
	}

	if req.Amount <= 0 {
		return nil, NewSDKError(-1, "退款金额必须大于0")
	}

	refundNo := req.RefundNo
	if refundNo == "" {
		refundNo = generateRefundNo()
	}

	reason := req.Reason
	if reason == "" {
		reason = "用户申请退款"
	}

	data := map[string]interface{}{
		"order_no":  req.OrderNo,
		"refund_no": refundNo,
		"amount":    req.Amount,
		"reason":    reason,
	}

	if req.NotifyURL != "" {
		data["notify_url"] = req.NotifyURL
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[RefundResponse]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/refund")

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	if err := checkResponse(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) QueryRefund(ctx context.Context, refundNo string) (*RefundResponse, error) {
	if refundNo == "" {
		return nil, ErrInvalidParam
	}

	data := map[string]interface{}{
		"refund_no": refundNo,
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[RefundResponse]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/refund/query")

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	if err := checkResponse(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) GetRefundsByOrderNo(ctx context.Context, orderNo string) ([]RefundResponse, error) {
	if orderNo == "" {
		return nil, ErrInvalidParam
	}

	data := map[string]interface{}{
		"order_no": orderNo,
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[[]RefundResponse]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/refund/order")

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	if err := checkResponse(&result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (c *Client) GetOrderRefundInfo(ctx context.Context, orderNo string) (*OrderRefundInfo, error) {
	if orderNo == "" {
		return nil, ErrInvalidParam
	}

	data := map[string]interface{}{
		"order_no": orderNo,
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[OrderRefundInfo]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post("/api/v1/refund/info")

	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRequestFailed, err)
	}

	if err := checkResponse(&result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (c *Client) ParseNotify(body []byte) (*NotifyRequest, error) {
	var req NotifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("parse notify failed: %w", err)
	}
	return &req, nil
}

func (c *Client) ParseRefundNotify(body []byte) (*RefundNotifyRequest, error) {
	var req RefundNotifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("parse refund notify failed: %w", err)
	}
	return &req, nil
}
