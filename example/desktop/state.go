package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"sync"
	"time"

	xgdnpay "github.com/skylark8866/paysdk"

	"github.com/skip2/go-qrcode"
)

type AppState struct {
	mu sync.RWMutex

	Config *Config

	Client *xgdnpay.Client

	OrderNo     string
	OutOrderNo  string
	Amount      float64
	Title       string
	PayType     string
	CodeURL     string
	QRCodeImage image.Image

	Status        string
	IsPolling     bool
	IsPaid        bool
	IsRefunded    bool
	TransactionID string

	RefundNo     string
	RefundStatus string

	Err error

	cancelPolling context.CancelFunc
}

type Config struct {
	AppID     string
	AppSecret string
	BaseURL   string
}

func NewState(cfg *Config) *AppState {
	client := xgdnpay.NewClient(
		cfg.AppID,
		cfg.AppSecret,
		xgdnpay.WithBaseURL(cfg.BaseURL),
		xgdnpay.WithTimeout(30*time.Second),
	)

	return &AppState{
		Config:  cfg,
		Client:  client,
		Amount:  0.01,
		Title:   "测试商品",
		PayType: xgdnpay.PayTypeNative,
		Status:  "就绪",
	}
}

func (s *AppState) SetStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status = status
}

func (s *AppState) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Err = err
	s.Status = fmt.Sprintf("错误: %v", err)
	s.IsPolling = false
}

func (s *AppState) CreateOrder() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.OutOrderNo = fmt.Sprintf("TEST_%d", time.Now().Unix())
	s.Status = "创建订单中..."
	s.Err = nil
	s.IsPaid = false
	s.IsRefunded = false
	s.TransactionID = ""
	s.CodeURL = ""
	s.QRCodeImage = nil

	if s.cancelPolling != nil {
		s.cancelPolling()
		s.cancelPolling = nil
	}

	req := &xgdnpay.CreateOrderRequest{
		OutOrderNo: s.OutOrderNo,
		Amount:     s.Amount,
		Title:      s.Title,
		PayType:    s.PayType,
	}

	order, err := s.Client.CreateOrder(context.Background(), req)
	if err != nil {
		s.Err = err
		s.Status = fmt.Sprintf("创建订单失败: %v", err)
		return
	}

	s.OrderNo = order.OrderNo
	s.CodeURL = order.CodeURL

	if s.CodeURL != "" {
		png, err := qrcode.Encode(s.CodeURL, qrcode.Medium, 256)
		if err != nil {
			s.Err = err
			s.Status = fmt.Sprintf("生成二维码失败: %v", err)
			return
		}

		img, _, err := image.Decode(bytes.NewReader(png))
		if err != nil {
			s.Err = err
			s.Status = fmt.Sprintf("解码二维码失败: %v", err)
			return
		}

		s.QRCodeImage = img
	}

	s.Status = "等待支付..."
	s.IsPolling = true

	go s.startPolling()
}

func (s *AppState) startPolling() {
	ctx, cancel := context.WithCancel(context.Background())

	s.mu.Lock()
	s.cancelPolling = cancel
	s.mu.Unlock()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := s.Client.CheckStatus(ctx, s.OrderNo)
			if err != nil {
				continue
			}

			s.mu.Lock()
			if result.Status == xgdnpay.OrderStatusPaid {
				s.IsPaid = true
				s.IsPolling = false
				s.Status = "支付成功！"
				s.TransactionID = result.PaidAt
				s.mu.Unlock()
				return
			}

			if result.Status == xgdnpay.OrderStatusClosed {
				s.Status = "订单已关闭"
				s.IsPolling = false
				s.mu.Unlock()
				return
			}

			if result.Status == xgdnpay.OrderStatusRefunded {
				s.IsRefunded = true
				s.Status = "订单已退款"
				s.IsPolling = false
				s.mu.Unlock()
				return
			}
			s.mu.Unlock()
		}
	}
}

func (s *AppState) QueryOrder() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.OrderNo == "" {
		s.Status = "请先创建订单"
		return
	}

	s.Status = "查询中..."

	result, err := s.Client.QueryOrder(context.Background(), s.OrderNo)
	if err != nil {
		s.Err = err
		s.Status = fmt.Sprintf("查询失败: %v", err)
		return
	}

	statusText := "未知"
	switch result.Status {
	case xgdnpay.OrderStatusPending:
		statusText = "待支付"
	case xgdnpay.OrderStatusPaid:
		statusText = "已支付"
		s.IsPaid = true
	case xgdnpay.OrderStatusClosed:
		statusText = "已关闭"
	case xgdnpay.OrderStatusRefunded:
		statusText = "已退款"
		s.IsRefunded = true
	}

	s.Status = fmt.Sprintf("状态: %s, 金额: %.2f", statusText, result.Amount)
}

func (s *AppState) Refund() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.OrderNo == "" {
		s.Status = "请先创建订单"
		return
	}

	if !s.IsPaid {
		s.Status = "订单未支付，无法退款"
		return
	}

	s.Status = "申请退款中..."
	s.RefundNo = fmt.Sprintf("REFUND_%d", time.Now().Unix())

	req := &xgdnpay.RefundRequest{
		OrderNo:  s.OrderNo,
		RefundNo: s.RefundNo,
		Amount:   s.Amount,
		Reason:   "用户申请退款",
	}

	refund, err := s.Client.CreateRefund(context.Background(), req)
	if err != nil {
		s.Err = err
		s.Status = fmt.Sprintf("退款失败: %v", err)
		return
	}

	s.IsRefunded = true
	s.RefundStatus = "退款成功"
	s.Status = fmt.Sprintf("退款成功，退款单号: %s", refund.RefundNo)
}

func (s *AppState) GetStatus() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.Status
}

func (s *AppState) GetQRCode() image.Image {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.QRCodeImage
}

func (s *AppState) GetOrderInfo() (orderNo, outOrderNo, codeURL string, isPaid, isRefunded bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.OrderNo, s.OutOrderNo, s.CodeURL, s.IsPaid, s.IsRefunded
}

func (s *AppState) IsPollingNow() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.IsPolling
}

func (s *AppState) StopPolling() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancelPolling != nil {
		s.cancelPolling()
		s.cancelPolling = nil
	}
	s.IsPolling = false
}
