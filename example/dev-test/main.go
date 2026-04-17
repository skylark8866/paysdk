package main

import (
	"context"
	"fmt"
	"time"

	xgdnpay "github.com/skylark8866/paysdk"
)

func main() {
	fmt.Println("=== XGDN Pay SDK 开发模式测试 ===")
	fmt.Println()

	client := xgdnpay.NewClient(
		"app_your_app_id",
		"your_app_secret",
		xgdnpay.WithBaseURL("http://localhost:8093"),
		xgdnpay.WithTimeout(30*time.Second),
	)

	ctx := context.Background()

	fmt.Println("=== 1. 创建扫码支付订单 ===")
	order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
		OutOrderNo: fmt.Sprintf("SDK_TEST_%d", time.Now().Unix()),
		Amount:     0.01,
		Title:      "SDK开发测试",
		PayType:    xgdnpay.PayTypeNative,
		ReturnURL:  "http://localhost:3001/result",
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
	fmt.Println()

	fmt.Println("=== 2. 查询订单状态 ===")
	queryResult, err := client.QueryOrder(ctx, order.OrderNo)
	if err != nil {
		fmt.Printf("查询订单失败: %v\n", err)
	} else {
		fmt.Printf("订单状态: %d (0=待支付, 1=已支付, 2=已关闭, 3=已退款)\n", queryResult.Status)
		fmt.Printf("订单金额: %.2f\n", queryResult.Amount)
	}
	fmt.Println()

	fmt.Println("=== 3. 检查订单状态 ===")
	checkResult, err := client.CheckStatus(ctx, order.OrderNo)
	if err != nil {
		fmt.Printf("检查状态失败: %v\n", err)
	} else {
		fmt.Printf("状态: %d\n", checkResult.Status)
		fmt.Printf("返回URL: %s\n", checkResult.ReturnURL)
	}
	fmt.Println()

	fmt.Println("=== 4. 获取订单退款信息 ===")
	refundInfo, err := client.GetOrderRefundInfo(ctx, order.OrderNo)
	if err != nil {
		fmt.Printf("获取退款信息失败: %v\n", err)
	} else {
		fmt.Printf("订单金额: %.2f\n", refundInfo.OrderAmount)
		fmt.Printf("已退款: %.2f\n", refundInfo.TotalRefunded)
		fmt.Printf("剩余可退: %.2f\n", refundInfo.RemainingAmount)
		fmt.Printf("可退款: %v\n", refundInfo.CanRefund)
		if refundInfo.Message != "" {
			fmt.Printf("消息: %s\n", refundInfo.Message)
		}
	}
	fmt.Println()

	fmt.Println("=== 测试完成 ===")
	fmt.Println()
	fmt.Println("提示: 请使用以下链接进行支付测试:")
	fmt.Printf("支付页面: http://localhost:3001/pay/%s\n", order.OrderNo)
	fmt.Println()
	fmt.Println("支付完成后，可以运行退款测试...")
}
