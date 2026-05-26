package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"shop-demo/model"
	"shop-demo/repo"
	"strconv"
	"strings"
	"time"

	xgdnpay "github.com/skylark8866/paysdk"
	"github.com/skylark8866/paysdk/sse"
)

type RechargeService struct {
	repo   *repo.Repository
	client *xgdnpay.Client
	paySSE *xgdnpay.PaySSE
}

func NewRechargeService(repo *repo.Repository, client *xgdnpay.Client, paySSE *xgdnpay.PaySSE) *RechargeService {
	return &RechargeService{
		repo:   repo,
		client: client,
		paySSE: paySSE,
	}
}

type CreateOrderResult struct {
	OrderNo     string
	PayOrderNo  string
	PayURL      string
	CodeURL     string
	PayAmount   int64
	BonusAmount int64
}

func (s *RechargeService) CreateOrder(ctx context.Context, userID uint64, username, packageID string) (*CreateOrderResult, error) {
	var pkg *model.RechargePackage

	if strings.HasPrefix(packageID, "custom_") {
		amountStr := strings.TrimPrefix(packageID, "custom_")
		if amount, err := strconv.ParseInt(amountStr, 10, 64); err == nil && amount >= 1 && amount <= 10000 {
			pkg = model.NewCustomPackage(amount)
		}
	} else {
		pkg = model.GetPackageByID(packageID)
	}

	if pkg == nil {
		return nil, errors.New("套餐不存在或金额无效")
	}

	orderNo := fmt.Sprintf("RCH_%d", time.Now().UnixNano())

	payOrder, err := s.client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
		OutOrderNo: orderNo,
		Amount:     pkg.PayAmount,
		Title:      fmt.Sprintf("充值-%s", pkg.Name),
		PayType:    xgdnpay.PayTypeNative,
	})
	if err != nil {
		return nil, fmt.Errorf("创建支付订单失败: %w", err)
	}

	order := &model.RechargeOrder{
		OrderNo:     orderNo,
		PayOrderNo:  payOrder.OrderNo,
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

	s.subscribePayment(payOrder.OrderNo, orderNo)

	return &CreateOrderResult{
		OrderNo:     orderNo,
		PayOrderNo:  payOrder.OrderNo,
		PayURL:      payOrder.PayURL,
		CodeURL:     payOrder.CodeURL,
		PayAmount:   pkg.PayAmount,
		BonusAmount: pkg.BonusAmount,
	}, nil
}

func (s *RechargeService) subscribePayment(payOrderNo, outOrderNo string) {
	log.Printf("[subscribePayment] 订阅后端支付通知: payOrderNo=%s, outOrderNo=%s", payOrderNo, outOrderNo)

	s.paySSE.Subscribe(payOrderNo, func(event *sse.PayNotifyEvent) {
		log.Printf("[subscribePayment] 收到后端通知: status=%s, orderNo=%s", event.Status, event.OrderNo)

		if event.Status != "paid" {
			log.Printf("[subscribePayment] 状态不是 paid，忽略")
			return
		}

		order, err := s.repo.GetOrderByPayOrderNo(payOrderNo)
		if err != nil {
			log.Printf("[subscribePayment] 查询订单失败: %v", err)
			return
		}

		if order.IsPaid() {
			log.Printf("[subscribePayment] 订单已支付，跳过")
			return
		}

		if err := s.repo.ProcessPayment(outOrderNo, time.Now()); err != nil {
			log.Printf("[subscribePayment] 处理支付失败: %v", err)
			return
		}

		log.Printf("[subscribePayment] 支付处理成功，向前端广播: outOrderNo=%s", outOrderNo)

		msg := xgdnpay.NewPayNotifyMessage(outOrderNo, order.PayAmount, xgdnpay.PayStatusPaid).
			SetPayType(xgdnpay.PayChannelWechat)
		s.paySSE.Hub().BroadcastMessage(outOrderNo, msg)
		log.Printf("[subscribePayment] 前端广播完成")
	})
}

func (s *RechargeService) GetOrder(orderNo string) (*model.RechargeOrder, error) {
	return s.repo.GetOrderByOrderNo(orderNo)
}

func (s *RechargeService) GetUserOrders(userID uint64, limit int) ([]model.RechargeOrder, error) {
	return s.repo.GetUserOrders(userID, limit)
}

func (s *RechargeService) GetUserBalanceLogs(userID uint64, limit int) ([]model.BalanceLog, error) {
	return s.repo.GetUserBalanceLogs(userID, limit)
}
