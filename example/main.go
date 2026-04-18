package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	xgdnpay "github.com/skylark8866/paysdk"
)

func main() {
	client := xgdnpay.NewClient(
		"app_your_app_id",
		"your_app_secret",
		xgdnpay.WithTimeout(30*time.Second),
		xgdnpay.WithBaseURL("http://127.0.0.1:8093"),
	)

	ctx := context.Background()

	fmt.Println("=== 创建扫码支付订单 ===")
	order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
		Amount:  0.01,
		Title:   "测试商品",
		PayType: xgdnpay.PayTypeNative,
	})
	if err != nil {
		if sdkErr, ok := err.(*xgdnpay.SDKError); ok {
			fmt.Printf("创建订单失败: [%d] %s\n", sdkErr.Code, sdkErr.Message)
		} else {
			fmt.Printf("创建订单失败: %v\n", err)
		}
		return
	}
	fmt.Printf("订单号: %s\n", order.OrderNo)
	fmt.Printf("支付链接: %s\n", order.PayURL)
	fmt.Printf("二维码内容: %s\n", order.CodeURL)

	fmt.Println("\n=== 查询订单状态 ===")
	queryResult, err := client.QueryOrder(ctx, order.OrderNo)
	if err != nil {
		fmt.Printf("查询订单失败: %v\n", err)
		return
	}
	fmt.Printf("订单状态: %d (%s)\n", queryResult.Status, xgdnpay.OrderStatus(queryResult.Status).Text())
	fmt.Printf("订单金额: %.2f\n", queryResult.Amount)

	fmt.Println("\n=== 轮询等待支付（最多等待5分钟）===")
	err = client.WaitForPayment(ctx, order.OrderNo, 2*time.Second, 5*time.Minute)
	if err != nil {
		fmt.Printf("等待支付失败: %v\n", err)
		return
	}
	fmt.Println("支付成功！")

	fmt.Println("\n=== 申请退款 ===")
	refund, err := client.CreateRefund(ctx, &xgdnpay.RefundRequest{
		OrderNo: order.OrderNo,
		Amount:  0.01,
		Reason:  "用户申请退款",
	})
	if err != nil {
		fmt.Printf("创建退款失败: %v\n", err)
		return
	}
	fmt.Printf("退款单号: %s\n", refund.RefundNo)
	fmt.Printf("退款状态: %d (%s)\n", refund.Status, xgdnpay.RefundStatus(refund.Status).Text())

	fmt.Println("\n=== 查询退款状态 ===")
	refundQuery, err := client.QueryRefund(ctx, refund.RefundNo)
	if err != nil {
		fmt.Printf("查询退款失败: %v\n", err)
		return
	}
	fmt.Printf("退款状态: %d (%s)\n", refundQuery.Status, xgdnpay.RefundStatus(refundQuery.Status).Text())
	fmt.Printf("退款金额: %.2f\n", refundQuery.RefundAmount)
}

func ExampleNotifyHandler() {
	client := xgdnpay.NewClient("app_id", "app_secret")

	notifyHandler := xgdnpay.NewNotifyHandler(client, func(req *xgdnpay.NotifyRequest) error {
		fmt.Printf("收到支付回调: 订单号=%s, 金额=%.2f\n", req.OrderNo, req.Amount)

		if xgdnpay.OrderStatus(req.Status) == xgdnpay.OrderStatusPaid {
			fmt.Println("支付成功，处理业务逻辑...")
		}

		return nil
	})

	http.Handle("/api/pay/notify", notifyHandler)
	fmt.Println("回调服务启动在 :8080")
	http.ListenAndServe(":8080", nil)
}

func ExampleRefundNotifyHandler() {
	client := xgdnpay.NewClient("app_id", "app_secret")

	refundNotifyHandler := xgdnpay.NewRefundNotifyHandler(client, func(req *xgdnpay.RefundNotifyRequest) error {
		fmt.Printf("收到退款回调: 退款单号=%s, 订单号=%s, 金额=%.2f, 状态=%s\n",
			req.RefundNo, req.OrderNo, req.Amount, req.Status)

		if req.Status == "SUCCESS" {
			fmt.Println("退款成功，处理业务逻辑...")
		}

		return nil
	})

	http.Handle("/api/refund/notify", refundNotifyHandler)
	fmt.Println("退款回调服务启动在 :8080")
	http.ListenAndServe(":8080", nil)
}

func ExampleManualVerify() {
	appSecret := "your_app_secret"

	notifyReq := &xgdnpay.NotifyRequest{
		OrderNo:       "PAY_xxx",
		OutOrderNo:    "BUSINESS_001",
		TransactionID: "WX_TRANS_001",
		Amount:        0.01,
		Status:        int(xgdnpay.OrderStatusPaid),
		PaidAt:        "2024-01-01 12:00:00",
		Timestamp:     fmt.Sprintf("%d", time.Now().Unix()),
		Sign:          "xxx",
	}

	if err := xgdnpay.VerifyNotify(notifyReq, appSecret, 300); err != nil {
		fmt.Printf("验签失败: %v\n", err)
		return
	}

	fmt.Println("验签成功，处理业务逻辑...")
}

func ExampleContext() {
	client := xgdnpay.NewClient("app_id", "app_secret")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
		Amount:  0.01,
		Title:   "测试商品",
		PayType: xgdnpay.PayTypeNative,
	})

	if err != nil {
		if err == context.DeadlineExceeded {
			fmt.Println("请求超时")
			return
		}
		fmt.Printf("创建订单失败: %v\n", err)
		return
	}

	fmt.Printf("订单号: %s\n", order.OrderNo)
}

func ExampleErrorHandling() {
	client := xgdnpay.NewClient("app_id", "app_secret")

	order, err := client.CreateOrder(context.Background(), &xgdnpay.CreateOrderRequest{
		Amount:  0.01,
		Title:   "测试商品",
		PayType: xgdnpay.PayTypeNative,
	})

	if err != nil {
		switch {
		case err == xgdnpay.ErrInvalidParam:
			fmt.Println("参数错误")
		case err == xgdnpay.ErrTimeout:
			fmt.Println("请求超时")
		case err == xgdnpay.ErrRequestFailed:
			fmt.Println("网络请求失败")
		default:
			if sdkErr, ok := err.(*xgdnpay.SDKError); ok {
				fmt.Printf("业务错误: [%d] %s\n", sdkErr.Code, sdkErr.Message)
			} else {
				fmt.Printf("未知错误: %v\n", err)
			}
		}
		return
	}

	fmt.Printf("订单号: %s\n", order.OrderNo)
}
