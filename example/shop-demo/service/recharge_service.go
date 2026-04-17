package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"shop-demo/config"
	"shop-demo/model"
	"shop-demo/repo"
	"strconv"
	"strings"
	"time"

	xgdnpay "github.com/skylark8866/paysdk"
	"github.com/skylark8866/paysdk/sse"

	"gorm.io/gorm"
)

type RechargeService struct {
	repo   *repo.Repository
	client *xgdnpay.Client
	cfg    *config.Config
	sseHub *sse.Hub
}

func NewRechargeService(repo *repo.Repository, client *xgdnpay.Client, cfg *config.Config) *RechargeService {
	return &RechargeService{
		repo:   repo,
		client: client,
		cfg:    cfg,
	}
}

func (s *RechargeService) SetSSEHub(hub *sse.Hub) {
	s.sseHub = hub
}

type CreateOrderResult struct {
	OrderNo     string
	PayOrderNo  string
	PayURL      string
	CodeURL     string
	PayAmount   float64
	BonusAmount float64
}

func (s *RechargeService) CreateOrder(ctx context.Context, userID uint64, username, packageID string) (*CreateOrderResult, error) {
	var pkg *model.RechargePackage

	// 检查是否是自定义金额 (格式: custom_xxx)
	if strings.HasPrefix(packageID, "custom_") {
		amountStr := strings.TrimPrefix(packageID, "custom_")
		if amount, err := strconv.ParseFloat(amountStr, 64); err == nil && amount >= 1 && amount <= 10000 {
			pkg = model.NewCustomPackage(amount)
		}
	} else {
		pkg = model.GetPackageByID(packageID)
	}

	if pkg == nil {
		return nil, errors.New("套餐不存在或金额无效")
	}

	orderNo := fmt.Sprintf("RCH_%d", time.Now().UnixNano())

	order := &model.RechargeOrder{
		OrderNo:     orderNo,
		UserID:      userID,
		Username:    username,
		PackageID:   packageID,
		PayAmount:   pkg.PayAmount,
		BonusAmount: pkg.BonusAmount,
		Status:      model.OrderStatusPending,
	}

	if err := s.repo.CreateOrder(order); err != nil {
		return nil, errors.New("创建订单失败")
	}

	payOrder, err := s.client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
		OutOrderNo: orderNo,
		Amount:     pkg.PayAmount,
		Title:      fmt.Sprintf("充值-%s", pkg.Name),
		PayType:    xgdnpay.PayTypeNative,
	})
	if err != nil {
		return nil, fmt.Errorf("创建支付订单失败: %w", err)
	}

	return &CreateOrderResult{
		OrderNo:     orderNo,
		PayOrderNo:  payOrder.OrderNo,
		PayURL:      payOrder.PayURL,
		CodeURL:     payOrder.CodeURL,
		PayAmount:   pkg.PayAmount,
		BonusAmount: pkg.BonusAmount,
	}, nil
}

func (s *RechargeService) GetOrder(orderNo string) (*model.RechargeOrder, error) {
	return s.repo.GetOrderByOrderNo(orderNo)
}

func (s *RechargeService) HandlePaymentCallback(outOrderNo string, status int) error {
	if status != xgdnpay.OrderStatusPaid {
		return nil
	}

	order, err := s.repo.GetOrderByOrderNo(outOrderNo)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("订单不存在: %s", outOrderNo)
		}
		return err
	}

	if order.IsPaid() {
		return nil
	}

	if err := s.repo.ProcessPayment(outOrderNo, time.Now()); err != nil {
		return err
	}

	if s.sseHub != nil {
		msg := xgdnpay.NewPayNotifyMessage(outOrderNo, order.PayAmount, "paid").
			SetPayType("wechat")
		data := sse.FormatEvent("pay_notify", mustMarshal(msg))
		s.sseHub.Broadcast(outOrderNo, data)
	}

	return nil
}

func (s *RechargeService) GetUserOrders(userID uint64, limit int) ([]model.RechargeOrder, error) {
	return s.repo.GetUserOrders(userID, limit)
}

func (s *RechargeService) GetUserBalanceLogs(userID uint64, limit int) ([]model.BalanceLog, error) {
	return s.repo.GetUserBalanceLogs(userID, limit)
}

func mustMarshal(v interface{}) []byte {
	data, _ := json.Marshal(v)
	return data
}
