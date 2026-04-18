package xgdnpay

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
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
		baseURL:   DefaultBaseURL,
		client:    resty.New(),
	}

	for _, opt := range opts {
		opt(c)
	}

	c.client.SetBaseURL(c.baseURL)
	c.client.SetHeader(FieldContentType, ContentTypeJSON)

	return c
}

func (c *Client) AppID() string {
	return c.appID
}

func (c *Client) getSecret() string {
	return c.appSecret
}

func checkResponse[T any](resp *apiResponse[T]) error {
	if resp.Code != 0 {
		return NewSDKError(resp.Code, resp.Message)
	}
	return nil
}

func generateOrderNo() string {
	return fmt.Sprintf("%s%d_%s", OrderNoPrefix, time.Now().UnixMilli(), randomHex(OrderNoRandomDigits))
}

func generateRefundNo() string {
	return fmt.Sprintf("%s%d_%s", RefundNoPrefix, time.Now().UnixMilli(), randomHex(OrderNoRandomDigits))
}

func randomHex(n int) string {
	max := new(big.Int).Exp(big.NewInt(16), big.NewInt(int64(n)), nil)
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		return fmt.Sprintf("%0*x", n, time.Now().UnixNano()%int64(max.Int64()))
	}
	return fmt.Sprintf("%0*x", n, r)
}

func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	if req == nil {
		return nil, ErrInvalidParam
	}

	if req.Amount <= 0 {
		return nil, NewSDKError(-1, ErrMsgAmountRequired)
	}

	if req.Title == "" {
		return nil, NewSDKError(-1, ErrMsgTitleRequired)
	}

	outOrderNo := req.OutOrderNo
	if outOrderNo == "" {
		outOrderNo = generateOrderNo()
	}

	payType := req.PayType
	if payType == "" {
		payType = PayTypeNative
	}

	if !payType.IsValid() {
		return nil, NewSDKError(-1, ErrMsgInvalidPayType+": "+payType.String())
	}

	if payType == PayTypeJSAPI && req.OpenID == "" {
		return nil, NewSDKError(-1, ErrMsgOpenIDRequired)
	}

	data := map[string]interface{}{
		FieldOutOrderNo: outOrderNo,
		FieldAmount:     req.Amount,
		FieldTitle:      req.Title,
		FieldPayType:    payType,
	}

	if req.OpenID != "" {
		data[FieldOpenID] = req.OpenID
	}
	if req.ReturnURL != "" {
		data[FieldReturnURL] = req.ReturnURL
	}
	if req.NotifyURL != "" {
		data[FieldNotifyURL] = req.NotifyURL
	}
	if req.Extra != nil {
		data[FieldExtra] = req.Extra
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
		Post(PathOrderCreate)

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
		FieldOrderNo: orderNo,
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
		Post(PathOrderQuery)

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
		FieldOrderNo: orderNo,
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
		Post(PathOrderCheck)

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
		FieldOrderNo: orderNo,
	}

	signedReq, err := c.buildSignedRequest(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSignFailed, err)
	}

	var result apiResponse[struct{}]
	_, err = c.client.R().
		SetContext(ctx).
		SetBody(signedReq).
		SetResult(&result).
		Post(PathOrderClose)

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

			switch OrderStatus(result.Status) {
			case OrderStatusPaid:
				return nil
			case OrderStatusClosed:
				return NewSDKError(-10, ErrMsgOrderClosed)
			case OrderStatusRefunded:
				return NewSDKError(-11, ErrMsgOrderRefunded)
			}
		}
	}
}

func (c *Client) CreateRefund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	if req == nil {
		return nil, ErrInvalidParam
	}

	if req.OrderNo == "" {
		return nil, NewSDKError(-1, ErrMsgOrderNoRequired)
	}

	if req.Amount <= 0 {
		return nil, NewSDKError(-1, ErrMsgRefundRequired)
	}

	refundNo := req.RefundNo
	if refundNo == "" {
		refundNo = generateRefundNo()
	}

	reason := req.Reason
	if reason == "" {
		reason = ErrMsgRefundDefault
	}

	data := map[string]interface{}{
		FieldOrderNo:  req.OrderNo,
		FieldRefundNo: refundNo,
		FieldAmount:   req.Amount,
		FieldReason:   reason,
	}

	if req.NotifyURL != "" {
		data[FieldNotifyURL] = req.NotifyURL
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
		Post(PathRefund)

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
		FieldRefundNo: refundNo,
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
		Post(PathRefundQuery)

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
		FieldOrderNo: orderNo,
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
		Post(PathRefundOrder)

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
		FieldOrderNo: orderNo,
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
		Post(PathRefundInfo)

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
		return nil, fmt.Errorf("%s: %w", ErrMsgParseNotifyFail, err)
	}
	return &req, nil
}

func (c *Client) ParseRefundNotify(body []byte) (*RefundNotifyRequest, error) {
	var req RefundNotifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("%s: %w", ErrMsgParseRefundFail, err)
	}
	return &req, nil
}

func (c *Client) VerifyNotify(req *NotifyRequest) error {
	return VerifyNotifyWithSecret(req, c.appSecret, DefaultMaxDelay)
}

func (c *Client) VerifyRefundNotify(req *RefundNotifyRequest) error {
	return VerifyRefundNotifyWithSecret(req, c.appSecret, DefaultMaxDelay)
}
