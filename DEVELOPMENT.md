# XGDN Pay SDK 开发手册

> 版本：v1.2.0  
> 更新日期：2026-04-17

## 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [核心功能](#核心功能)
- [SSE 实时通知](#sse-实时通知)
- [回调处理](#回调处理)
- [错误处理](#错误处理)
- [最佳实践](#最佳实践)
- [示例说明](#示例说明)

---

## 概述

XGDN Pay SDK 是 XGDN 支付平台的官方 Go SDK。设计理念是"对内复杂，对外简单"，让开发者用最少的代码完成支付集成。前端代码极其简单，直接使用浏览器原生 EventSource API 即可。

### 设计原则

1. **对内复杂，对外简单** - 内部处理签名、验签、重试等复杂逻辑，对外提供简洁 API
2. **只暴露必要接口** - 不暴露内部实现细节
3. **最小化硬编码** - 所有常量统一管理
4. **泛型优先** - 能用泛型的地方都用泛型，减少重复代码

### 目录结构

```
sdk/
├── go/                        # Go SDK
│   ├── client.go             # 客户端核心
│   ├── constants.go          # 常量定义
│   ├── notify.go             # 回调处理（泛型实现）
│   ├── sign.go               # 签名与验签
│   ├── types.go              # 类型定义
│   ├── sse/                  # SSE 实时通知
│   │   ├── client.go         # SSE 客户端
│   │   ├── constants.go      # SSE 常量
│   │   ├── gin.go            # Gin 适配器
│   │   ├── handler.go        # 标准库适配器
│   │   ├── hub.go            # SSE Hub
│   │   ├── message.go        # 消息格式化
│   │   └── response.go       # 响应处理
│   └── example/              # 使用示例
│       ├── shop-demo/        # Web 商城示例
│       ├── sse-test/         # SSE 测试
│       ├── desktop/          # 桌面应用
│       └── dev-test/         # 开发调试
└── DEVELOPMENT.md            # 开发手册（本文件）
```

---

## 快速开始

### 安装

```bash
go get xgdn-pay
```

### 30 秒上手

```go
package main

import (
    "context"
    "fmt"

    xgdnpay "xgdn-pay"
)

func main() {
    // 1. 创建客户端
    client := xgdnpay.NewClient("your_app_id", "your_app_secret")

    // 2. 创建订单
    order, err := client.CreateOrder(context.Background(), &xgdnpay.CreateOrderRequest{
        Amount: 0.01,
        Title:  "测试商品",
    })
    if err != nil {
        panic(err)
    }

    // 3. 使用 order.CodeURL 生成二维码
    fmt.Println("订单号:", order.OrderNo)
    fmt.Println("二维码内容:", order.CodeURL)
}
```

---

## 核心功能

### 客户端初始化

```go
client := xgdnpay.NewClient(
    "your_app_id",      // 必填：平台分配的 AppID
    "your_app_secret",  // 必填：平台分配的 AppSecret
    xgdnpay.WithBaseURL("https://pay.xgdn.net"),  // 可选：默认生产环境
    xgdnpay.WithTimeout(30 * time.Second),         // 可选：请求超时
)
```

### 订单操作

#### 创建订单

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
    ReturnURL:  "https://yoursite.com/success",  // 可选
    NotifyURL:  "https://yoursite.com/callback", // 可选
})

// JSAPI 支付（微信公众号/小程序）
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount:  0.01,
    Title:   "商品名称",
    PayType: xgdnpay.PayTypeJSAPI,
    OpenID:  "user_openid",  // JSAPI 必填
})
```

#### 查询订单

```go
// 按平台订单号查询
order, err := client.QueryOrder(ctx, "ORD_xxx")

// 按商户订单号查询
order, err := client.QueryOrderByOutOrderNo(ctx, "YOUR_ORDER_001")
```

#### 关闭订单

```go
err := client.CloseOrder(ctx, "ORD_xxx")
```

### 退款操作

#### 创建退款

```go
refund, err := client.CreateRefund(ctx, &xgdnpay.CreateRefundRequest{
    OrderNo:  "ORD_xxx",       // 平台订单号
    Amount:   0.01,            // 退款金额
    Reason:   "用户申请退款",  // 可选，默认"用户申请退款"
    RefundNo: "REF_xxx",       // 可选，不填自动生成
})
```

#### 查询退款

```go
// 按退款单号查询
refund, err := client.QueryRefund(ctx, "REF_xxx")

// 按商户退款单号查询
refund, err := client.QueryRefundByOutRefundNo(ctx, "YOUR_REFUND_001")
```

### 订单状态

```go
const (
    OrderStatusPending  = 0  // 待支付
    OrderStatusPaid     = 1  // 已支付
    OrderStatusClosed   = 2  // 已关闭
    OrderStatusRefunded = 3  // 已退款
)

// 获取状态文本
text := xgdnpay.StatusText(status)  // "待支付"、"已支付" 等
```

---

## SSE 实时通知

SSE（Server-Sent Events）用于实时推送支付状态，无需轮询。

### 架构设计

```
┌─────────────┐     SSE 连接      ┌─────────────┐
│   前端页面   │ ◄─────────────── │   SSE Hub   │
└─────────────┘                   └─────────────┘
                                        ▲
                                        │ 广播
                                        │
┌─────────────┐                   ┌─────────────┐
│  支付平台    │ ──── 回调 ──────► │  后端服务    │
└─────────────┘                   └─────────────┘
```

### 后端集成

#### 1. 创建 SSE Hub

```go
import (
    "xgdn-pay/sse"
)

sseHub := sse.NewHub()
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go sseHub.Run(ctx)
```

Hub 支持以下配置选项：

```go
sseHub := sse.NewHub(
    sse.WithKeepAlive(10*time.Second),   // 心跳间隔，默认 15 秒
    sse.WithMaxClients(500),             // 最大总连接数，默认 1000
    sse.WithMaxPerChannel(5),            // 每 channel 最大连接数，默认 10
    sse.WithHubBufferSize(512),          // 客户端缓冲区大小，默认 256
)
```

#### 2. 注册 SSE 端点（Gin 框架）

```go
// 带连接确认消息（推荐）
r.GET("/api/events/:channel", sseHub.GinHandler(sse.WithConnectMessage()))

// 不带连接确认消息
r.GET("/api/events/:channel", sseHub.GinHandler())
```

**重要：SSE 端点必须添加认证中间件！** 因为 `EventSource` API 不支持自定义 Header，认证应通过 Cookie 实现：

```go
// ✅ 正确：添加认证中间件
r.GET("/api/events/:channel", authMiddleware, sseHub.GinHandler(sse.WithConnectMessage()))

// ❌ 危险：无认证，任何人都能订阅
r.GET("/api/events/:channel", sseHub.GinHandler(sse.WithConnectMessage()))
```

认证中间件需要区分 SSE 请求和普通请求，对 SSE 请求返回 HTTP 401（而非 HTTP 200 + JSON body），这样 `EventSource` 才能正确触发 `onerror`：

```go
func Auth(userService *UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if token == "" {
            writeAuthError(c, "未登录")
            c.Abort()
            return
        }
        // ... 验证 token ...
    }
}

func writeAuthError(c *gin.Context, message string) {
    accept := c.GetHeader("Accept")
    if strings.Contains(accept, "text/event-stream") {
        c.JSON(http.StatusUnauthorized, gin.H{"error": message})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 401, "message": message})
}
```

#### 3. 广播消息

```go
// 格式化并发送事件
msg := xgdnpay.NewPayNotifyMessage(orderNo, amount, "paid").
    SetPayType("wechat")
data := sse.FormatEvent("pay_notify", mustMarshal(msg))
sseHub.Broadcast(orderNo, data)

// 或使用 JSON 广播
sseHub.BroadcastJSON(channel, map[string]any{
    "event": "pay_notify",
    "data":  msg,
})
```

### 前端集成

前端使用浏览器原生 `EventSource` API，配合指数退避重连：

```javascript
function connectSSE(orderNo, retryCount) {
    retryCount = retryCount || 0;
    var eventSource = new EventSource('/api/events/' + orderNo);

    // 连接确认，重置重试计数
    eventSource.addEventListener('connected', function() {
        retryCount = 0;
    });

    // 监听支付通知
    eventSource.addEventListener('pay_notify', function(e) {
        var data = JSON.parse(e.data);
        if (data.status === 'paid') {
            eventSource.close();
            onPaid(data);
        }
    });

    // 指数退避重连：1s → 2s → 4s → ... → 30s，最多 10 次
    eventSource.onerror = function() {
        eventSource.close();
        if (retryCount >= 10) {
            onFallback();  // 降级为轮询
            return;
        }
        var delay = Math.min(1000 * Math.pow(2, retryCount), 30000);
        setTimeout(function() {
            connectSSE(orderNo, retryCount + 1);
        }, delay);
    };
}

// 启动 SSE
connectSSE(orderNo);
```

**重连策略说明：**
- 延迟从 1 秒开始，指数递增（1s → 2s → 4s → 8s → ...）
- 最大延迟 30 秒
- 连接成功后重置延迟
- 超过 10 次重试后降级为 HTTP 轮询

### SSE 事件格式

SDK 使用标准的 SSE 事件格式：

```
event: connected
data: {}

event: pay_notify
data: {"order_no":"ORD_xxx","amount":10.00,"status":"paid"}
```

**重要：** 不要使用注释格式 `: comment\n\n`，它不会触发前端事件监听器。

### Gin Handler 选项

```go
// 自定义 channel 参数名
sseHub.GinHandler(sse.WithChannelParam("order_id"))

// 自定义 channel 获取函数
sseHub.GinHandler(sse.WithChannelFunc(func(c *gin.Context) string {
    return c.Query("order_no")
}))

// 订阅前验证
sseHub.GinHandler(sse.WithBeforeSubscribe(func(c *gin.Context, channel string) error {
    // 验证用户是否有权限订阅此 channel
    return nil
}))

// 连接/断开回调
sseHub.GinHandler(
    sse.WithOnConnect(func(c *gin.Context, channel string) {
        log.Printf("客户端连接: %s", channel)
    }),
    sse.WithOnDisconnect(func(c *gin.Context, channel string) {
        log.Printf("客户端断开: %s", channel)
    }),
)
```

---

## 回调处理

### 泛型回调处理器

SDK 使用泛型实现统一的回调处理器：

```go
// 支付回调
handler := xgdnpay.NewNotifyHandler(client, func(req *xgdnpay.NotifyRequest) error {
    // 处理支付成功逻辑
    fmt.Printf("订单 %s 支付成功，金额: %.2f\n", req.OutOrderNo, req.Amount)
    return nil
})

// 退款回调
handler := xgdnpay.NewNotifyHandler(client, func(req *xgdnpay.RefundNotifyRequest) error {
    // 处理退款成功逻辑
    fmt.Printf("退款 %s 成功\n", req.OutRefundNo)
    return nil
})
```

### 手动验证回调

```go
var req xgdnpay.NotifyRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.String(400, "参数错误")
    return
}

// 验证签名（300 秒超时）
if err := xgdnpay.VerifyNotify(&req, appSecret, 300); err != nil {
    c.String(401, "签名验证失败")
    return
}

// 处理业务逻辑
// ...
```

### 回调数据结构

```go
type NotifyRequest struct {
    AppID         string  `json:"app_id"`
    OrderNo       string  `json:"order_no"`
    OutOrderNo    string  `json:"out_order_no"`
    Amount        float64 `json:"amount"`
    Title         string  `json:"title"`
    PayType       string  `json:"pay_type"`
    Status        int     `json:"status"`
    TransactionID string  `json:"transaction_id"`
    PaidAt        string  `json:"paid_at"`
    Timestamp     string  `json:"timestamp"`
    Nonce         string  `json:"nonce"`
    Sign          string  `json:"sign"`
}
```

---

## 错误处理

### SDK 错误类型

```go
type SDKError struct {
    Code    int
    Message string
}

// 预定义错误
var (
    ErrInvalidParam   = &SDKError{Code: -1, Message: "参数错误"}
    ErrRequestFailed  = &SDKError{Code: -2, Message: "请求失败"}
    ErrSignFailed     = &SDKError{Code: -3, Message: "签名验证失败"}
    ErrTimeout        = &SDKError{Code: -4, Message: "请求超时"}
    ErrOrderNotFound  = &SDKError{Code: -5, Message: "订单不存在"}
    ErrRefundNotFound = &SDKError{Code: -6, Message: "退款单不存在"}
)
```

### 错误处理示例

```go
order, err := client.CreateOrder(ctx, req)
if err != nil {
    var sdkErr *xgdnpay.SDKError
    if errors.As(err, &sdkErr) {
        switch sdkErr.Code {
        case -1:
            // 参数错误
        case -2:
            // 请求失败
        default:
            // 其他错误
        }
    }
    return err
}
```

---

## 最佳实践

### 1. 凭证管理

**不要硬编码凭证！**

```go
// ❌ 错误做法
client := xgdnpay.NewClient("app_xxx", "secret_xxx")

// ✅ 正确做法：使用环境变量
appID := os.Getenv("XGDN_APP_ID")
appSecret := os.Getenv("XGDN_APP_SECRET")
client := xgdnpay.NewClient(appID, appSecret)
```

### 2. 订单号管理

SDK 会自动生成订单号，但建议使用自己的订单号：

```go
// ✅ 推荐：使用自己的订单号
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    OutOrderNo: "MY_ORDER_" + time.Now().Format("20060102150405"),
    Amount:     0.01,
    Title:      "商品名称",
})
```

### 3. SSE Channel 命名

使用订单号作为 channel，确保唯一性：

```go
// ✅ 正确：使用订单号
sseHub.Broadcast(orderNo, data)

// ❌ 错误：使用用户 ID（可能导致消息混乱）
sseHub.Broadcast(userID, data)
```

### 4. SSE 端点认证

SSE 端点必须添加认证，防止未授权订阅：

```go
// ✅ 正确：添加认证中间件
r.GET("/api/events/:channel", authMiddleware, sseHub.GinHandler(sse.WithConnectMessage()))

// ❌ 危险：无认证，任何人都能订阅
r.GET("/api/events/:channel", sseHub.GinHandler(sse.WithConnectMessage()))
```

认证中间件对 SSE 请求应返回 HTTP 401，而非 HTTP 200 + JSON body，否则 `EventSource` 无法正确触发 `onerror`。

### 5. SSE 连接数限制

Hub 内置连接数保护，防止资源耗尽：

```go
sseHub := sse.NewHub(
    sse.WithMaxClients(500),     // 最大总连接数
    sse.WithMaxPerChannel(5),    // 每 channel 最大连接数
)
```

超过限制时，`TrySubscribe` 返回错误，Gin/Handler 返回 HTTP 503。

### 6. 回调幂等性

回调可能重复发送，确保处理逻辑幂等：

```go
func handleCallback(req *xgdnpay.NotifyRequest) error {
    // 检查订单是否已处理
    order, _ := db.GetOrder(req.OutOrderNo)
    if order.Status == OrderStatusPaid {
        return nil  // 已处理，直接返回成功
    }
    
    // 处理支付成功逻辑
    return db.UpdateOrderStatus(req.OutOrderNo, OrderStatusPaid)
}
```

### 7. 超时设置

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

order, err := client.CreateOrder(ctx, req)
```

---

## 示例说明

### shop-demo - Web 商城示例

完整的 Web 商城示例，包含：
- 用户注册/登录
- 充值套餐选择
- 二维码支付
- SSE 实时通知
- 支付回调处理

```bash
cd go/example/shop-demo
go build -o shop-demo-server.exe main.go embed.go
./shop-demo-server.exe
```

访问 http://localhost:8081 体验完整流程。

### sse-test - SSE 测试

简单的 SSE 功能测试：

```bash
cd go/example/sse-test
go run main.go
```

### desktop - 桌面应用

使用 Fyne 框架的桌面 GUI 工具：

```bash
cd go/example/desktop
go run main.go
```

---

## 常见问题

### Q: SSE 连接建立后收不到事件？

A: 确保使用的是事件格式 `event: name\ndata: {}\n\n`，而不是注释格式 `: comment\n\n`。

### Q: 回调签名验证失败？

A: 检查以下几点：
1. AppSecret 是否正确
2. 时间戳是否在有效期内（默认 300 秒）
3. 签名算法是否正确（HMAC-SHA256）

### Q: 如何处理网络超时？

A: 使用 context 设置超时：

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
```

---

## 更新日志

### v1.2.0 (2026-04-17)

- SSE Hub 添加连接数限制（WithMaxClients、WithMaxPerChannel）
- SSE Hub 添加 TrySubscribe 方法，超限时返回错误
- SSE 端点添加 JWT 认证保护
- 认证中间件区分 SSE 请求，返回正确的 HTTP 状态码
- 前端 SSE 添加指数退避重连机制
- 移除 JS SDK（前端代码简单，直接使用 EventSource API）
- 统一开发手册为 DEVELOPMENT.md

### v1.1.0 (2025-04-17)

- 修复 SSE connected 事件格式（从注释格式改为事件格式）
- 统一 SSE 常量管理
- 优化 NotifyHandler 使用泛型实现

### v1.0.0

- 初始版本
- 支持订单创建、查询、关闭
- 支持退款创建、查询
- 支持 SSE 实时通知
- 支持回调处理
