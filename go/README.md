# XGDN Pay Go SDK

XGDN 支付平台 Go 语言 SDK，对内复杂对外简单，三步完成支付接入。

## 安装

```bash
go get xgdn-pay
```

## 30 秒上手

```go
package main

import (
    "context"
    "fmt"
    "time"

    xgdnpay "xgdn-pay"
)

func main() {
    // 1. 创建客户端
    client := xgdnpay.NewClient("your_app_id", "your_app_secret")

    // 2. 创建订单（OutOrderNo 可选，不填自动生成；PayType 默认 native）
    order, err := client.CreateOrder(context.Background(), &xgdnpay.CreateOrderRequest{
        Amount: 0.01,
        Title:  "测试商品",
    })
    if err != nil {
        panic(err)
    }

    // 3. 用 order.CodeURL 生成二维码给用户扫码，或用 order.PayURL 跳转支付
    fmt.Println("订单号:", order.OrderNo)
    fmt.Println("二维码内容:", order.CodeURL)
}
```

## 初始化客户端

```go
client := xgdnpay.NewClient(
    "your_app_id",      // 必填：平台分配的 AppID
    "your_app_secret",  // 必填：平台分配的 AppSecret
    xgdnpay.WithBaseURL("https://pay.xgdn.net"),  // 可选：默认生产环境
    xgdnpay.WithTimeout(30 * time.Second),         // 可选：请求超时
)
```

开发环境指定本地后端：

```go
client := xgdnpay.NewClient(appID, appSecret,
    xgdnpay.WithBaseURL("http://localhost:8093"),
)
```

## 订单操作

### 创建订单

SDK 会自动处理以下事项，你不需要操心：

- `OutOrderNo`：不填则自动生成（格式 `ORD_时间戳_随机数`）
- `PayType`：不填则默认 `native`（扫码支付）
- `Reason`（退款时）：不填则默认"用户申请退款"
- `RefundNo`（退款时）：不填则自动生成

```go
// 扫码支付（最简写法）
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount: 0.01,
    Title:  "商品名称",
})

// 扫码支付（完整参数）
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    OutOrderNo: "YOUR_ORDER_001",  // 可选，不填自动生成
    Amount:     0.01,
    Title:      "商品名称",
    PayType:    xgdnpay.PayTypeNative,  // 可选，默认 native
    ReturnURL:  "https://yoursite.com/success",  // 可选，支付成功跳转
    NotifyURL:  "https://yoursite.com/callback", // 可选，支付回调地址
})

// JSAPI 支付（微信公众号/小程序内）
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount:  0.01,
    Title:   "商品名称",
    PayType: xgdnpay.PayTypeJSAPI,
    OpenID:  "user_openid",  // JSAPI 必填
})
```

返回值：

| 字段 | 类型 | 说明 |
|------|------|------|
| `OrderNo` | string | 平台订单号 |
| `PayURL` | string | 支付跳转链接（JSAPI 用） |
| `CodeURL` | string | 二维码内容（NATIVE 用） |

### 查询订单

```go
result, err := client.QueryOrder(ctx, "PAY_xxx")
fmt.Println("状态:", xgdnpay.StatusText(result.Status))  // 如 "已支付"
fmt.Println("金额:", result.Amount)
```

### 检查支付状态

轻量级接口，只返回状态相关信息：

```go
status, err := client.CheckStatus(ctx, "PAY_xxx")
fmt.Println("状态:", status.Status)
fmt.Println("支付时间:", status.PaidAt)
fmt.Println("跳转地址:", status.ReturnURL)
```

### 关闭订单

```go
err := client.CloseOrder(ctx, "PAY_xxx")
```

### 轮询等待支付

适用于桌面软件、命令行工具等没有前端回调的场景：

```go
// 每 2 秒查询一次，最多等 5 分钟
err := client.WaitForPayment(ctx, "PAY_xxx", 2*time.Second, 5*time.Minute)
if err != nil {
    if err == xgdnpay.ErrTimeout {
        fmt.Println("支付超时")
    }
    return
}
fmt.Println("支付成功！")
```

## 退款操作

### 申请退款

```go
// 最简写法（RefundNo 自动生成）
refund, err := client.CreateRefund(ctx, &xgdnpay.RefundRequest{
    OrderNo: "PAY_xxx",
    Amount:  0.01,
})

// 完整参数
refund, err := client.CreateRefund(ctx, &xgdnpay.RefundRequest{
    OrderNo:   "PAY_xxx",
    RefundNo:  "REFUND_001",    // 可选，不填自动生成
    Amount:    0.01,
    Reason:    "用户申请退款",   // 可选，不填默认此值
    NotifyURL: "https://yoursite.com/refund-callback",  // 可选
})
```

### 查询退款

```go
refund, err := client.QueryRefund(ctx, "REFUND_001")
fmt.Println("状态:", xgdnpay.RefundStatusText(refund.Status))  // 如 "退款成功"
```

### 查询订单退款信息

```go
info, err := client.GetOrderRefundInfo(ctx, "PAY_xxx")
fmt.Println("订单金额:", info.OrderAmount)
fmt.Println("已退款:", info.TotalRefunded)
fmt.Println("剩余可退:", info.RemainingAmount)
fmt.Println("可退款:", info.CanRefund)
```

## 回调通知

### 方式一：HTTP Handler（推荐）

最简用法，SDK 自动处理验签、解析、响应：

```go
// 支付回调
http.Handle("/api/pay/notify", xgdnpay.NewNotifyHandler(client, func(req *xgdnpay.NotifyRequest) error {
    if req.Status == xgdnpay.OrderStatusPaid {
        fmt.Printf("支付成功: %s 金额 %.2f\n", req.OrderNo, req.Amount)
        // 处理你的业务逻辑...
    }
    return nil  // 返回 nil 表示处理成功
}))

// 退款回调
http.Handle("/api/refund/notify", xgdnpay.NewRefundNotifyHandler(client, func(req *xgdnpay.RefundNotifyRequest) error {
    if req.Status == "SUCCESS" {
        fmt.Printf("退款成功: %s\n", req.RefundNo)
    }
    return nil
}))

http.ListenAndServe(":8080", nil)
```

自定义最大延迟时间（防重放攻击）：

```go
handler := xgdnpay.NewNotifyHandler(client, callback,
    xgdnpay.WithMaxDelay(600),  // 允许 600 秒内的请求，默认 300
)
```

### 方式二：手动验签

适合需要自定义 HTTP 处理逻辑的场景：

```go
func handleNotify(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)

    req, err := xgdnpay.ParseNotifyRequest(body)
    if err != nil {
        http.Error(w, "解析失败", 400)
        return
    }

    if err := xgdnpay.VerifyNotify(req, appSecret, 300); err != nil {
        http.Error(w, "验签失败", 401)
        return
    }

    // 验签成功，处理业务逻辑
    if req.Status == xgdnpay.OrderStatusPaid {
        // ...
    }

    // 必须返回成功响应
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "成功"})
}
```

## 错误处理

SDK 提供两种错误判断方式：

```go
order, err := client.CreateOrder(ctx, req)
if err != nil {
    // 方式一：预定义错误（适用于需要区分错误类型的场景）
    switch {
    case err == xgdnpay.ErrInvalidParam:
        fmt.Println("参数错误，请检查请求参数")
    case err == xgdnpay.ErrTimeout:
        fmt.Println("请求超时，请稍后重试")
    case err == xgdnpay.ErrRequestFailed:
        fmt.Println("网络请求失败，请检查网络连接")
    case err == xgdnpay.ErrSignFailed:
        fmt.Println("签名失败，请检查 AppSecret")
    default:
        // 方式二：SDKError（包含服务端返回的错误码和消息）
        if sdkErr, ok := err.(*xgdnpay.SDKError); ok {
            fmt.Printf("业务错误 [%d]: %s\n", sdkErr.Code, sdkErr.Message)
        }
    }
    return
}
```

## Context 支持

所有 API 都支持 `context.Context`，可以控制超时和取消：

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

order, err := client.CreateOrder(ctx, req)
if err == context.DeadlineExceeded {
    fmt.Println("请求超时")
}
```

## 常量参考

### 支付类型

| 常量 | 值 | 说明 |
|------|------|------|
| `PayTypeNative` | `"native"` | 扫码支付（PC 网站、桌面软件） |
| `PayTypeJSAPI` | `"jsapi"` | 微信内支付（公众号、小程序） |

### 订单状态

| 常量 | 值 | `StatusText()` 返回 |
|------|------|------|
| `OrderStatusPending` | `0` | 待支付 |
| `OrderStatusPaid` | `1` | 已支付 |
| `OrderStatusClosed` | `2` | 已关闭 |
| `OrderStatusRefunded` | `3` | 已退款 |

### 退款状态

| 常量 | 值 | `RefundStatusText()` 返回 |
|------|------|------|
| `RefundStatusProcessing` | `0` | 退款中 |
| `RefundStatusSuccess` | `1` | 退款成功 |
| `RefundStatusClosed` | `2` | 已关闭 |
| `RefundStatusFailed` | `3` | 退款失败 |
| `RefundStatusAbnormal` | `4` | 退款异常 |

### 工具函数

| 函数 | 说明 |
|------|------|
| `StatusText(status int) string` | 订单状态转中文 |
| `RefundStatusText(status int) string` | 退款状态转中文 |
| `NormalizePayType(raw string) string` | 规范化支付类型（"NATIVE" → "native"） |

### 错误常量

| 常量 | 说明 |
|------|------|
| `ErrInvalidParam` | 参数错误 |
| `ErrRequestFailed` | 请求失败 |
| `ErrSignFailed` | 签名验证失败 |
| `ErrTimeout` | 请求超时 |
| `ErrOrderNotFound` | 订单不存在 |
| `ErrRefundNotFound` | 退款单不存在 |

## API 速查

### Client 方法

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `CreateOrder` | `ctx, *CreateOrderRequest` | `*CreateOrderResponse, error` | 创建订单 |
| `QueryOrder` | `ctx, orderNo` | `*QueryOrderResponse, error` | 查询订单详情 |
| `CheckStatus` | `ctx, orderNo` | `*CheckStatusResponse, error` | 检查支付状态 |
| `CloseOrder` | `ctx, orderNo` | `error` | 关闭订单 |
| `WaitForPayment` | `ctx, orderNo, interval, timeout` | `error` | 轮询等待支付 |
| `CreateRefund` | `ctx, *RefundRequest` | `*RefundResponse, error` | 申请退款 |
| `QueryRefund` | `ctx, refundNo` | `*RefundResponse, error` | 查询退款 |
| `GetRefundsByOrderNo` | `ctx, orderNo` | `[]RefundResponse, error` | 查询订单所有退款 |
| `GetOrderRefundInfo` | `ctx, orderNo` | `*OrderRefundInfo, error` | 获取退款概要 |
| `ParseNotify` | `body []byte` | `*NotifyRequest, error` | 解析支付回调 |
| `ParseRefundNotify` | `body []byte` | `*RefundNotifyRequest, error` | 解析退款回调 |

### 回调处理函数

| 函数 | 说明 |
|------|------|
| `NewNotifyHandler(client, handler, ...opts)` | 创建支付回调 HTTP Handler |
| `NewRefundNotifyHandler(client, handler, ...opts)` | 创建退款回调 HTTP Handler |
| `ParseNotifyRequest(body)` | 解析支付回调 JSON |
| `ParseRefundNotifyRequest(body)` | 解析退款回调 JSON |
| `VerifyNotify(req, appSecret, maxDelay)` | 验证支付回调签名 |
| `VerifyRefundNotify(req, appSecret, maxDelay)` | 验证退款回调签名 |

## 示例项目

| 项目 | 说明 | 运行 |
|------|------|------|
| `example/main.go` | 命令行完整示例 | `go run main.go` |
| `example/shop-demo/` | Web 商城示例 | `cd shop-demo && go run main.go` |
| `example/desktop/` | Gio GUI 桌面工具 | `cd desktop && go run main.go` |
| `example/dev-test/` | 开发调试工具 | `cd dev-test && go run main.go` |

## 开发环境

```go
client := xgdnpay.NewClient(appID, appSecret,
    xgdnpay.WithBaseURL("http://localhost:8093"),
)
```

## License

MIT
