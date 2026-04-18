# XGDN Pay SDK 开发手册

> 版本：v1.4.2  
> 更新日期：2026-04-18  
> 仓库：https://github.com/skylark8866/paysdk

## 目录

- [概述](#概述)
- [快速开始](#快速开始)
- [类型安全体系](#类型安全体系)
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
3. **最小化硬编码** - 所有常量统一管理，零硬编码字符串
4. **类型安全优先** - 支付类型、状态、渠道等全部使用自定义类型，编译期拦截非法值
5. **泛型优先** - 能用泛型的地方都用泛型，减少重复代码

### 目录结构

```
paysdk/
├── client.go                 # 客户端核心
├── constants.go              # 常量定义（字段名、错误信息、路径等）
├── notify.go                 # 回调处理（泛型实现）
├── sign.go                   # 签名与验签
├── types.go                  # 类型定义（自定义类型 + 请求/响应结构）
├── sse/                      # SSE 实时通知
│   ├── client.go             # SSE 客户端
│   ├── constants.go          # SSE 常量 + EventName 类型
│   ├── gin.go               # Gin 适配器
│   ├── handler.go           # 标准库适配器
│   ├── hub.go               # SSE Hub
│   ├── message.go           # 消息格式化（使用 EventName 类型）
│   └── response.go          # 响应处理
├── example/                  # 使用示例（不随 go get 发布）
│   ├── main.go              # 命令行基础示例
│   ├── shop-demo/           # Web 商城示例
│   ├── desktop/             # 桌面应用
│   └── dev-test/            # 开发调试工具
├── .gitignore
├── LICENSE                   # MIT 许可证
└── DEVELOPMENT.md            # 开发手册（本文件）
```

---

## 快速开始

### 安装

```bash
go get github.com/skylark8866/paysdk@v1.4.2
```

### 30 秒上手

```go
package main

import (
    "context"
    "fmt"

    xgdnpay "github.com/skylark8866/paysdk"
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

## 类型安全体系

SDK 对支付领域中的固定概念全部使用自定义类型，从编译期拦截非法值，杜绝硬编码字符串。

### 类型总览

| 类型 | 底层类型 | 用途 | 可选值 |
|------|---------|------|--------|
| `PayType` | `string` | 支付方式 | `PayTypeNative`, `PayTypeJSAPI` |
| `PayStatus` | `string` | 支付状态 | `PayStatusPaid`, `PayStatusPending`, `PayStatusClosed` |
| `PayChannel` | `string` | 支付渠道 | `PayChannelWechat`, `PayChannelAlipay` |
| `OrderStatus` | `int` | 订单状态 | `OrderStatusPending(0)`, `OrderStatusPaid(1)`, `OrderStatusClosed(2)`, `OrderStatusRefunded(3)` |
| `RefundStatus` | `int` | 退款状态 | `RefundStatusProcessing(0)`, `RefundStatusSuccess(1)`, `RefundStatusClosed(2)`, `RefundStatusFailed(3)`, `RefundStatusAbnormal(4)` |
| `EventName` | `string` | SSE 事件名 | `EventConnected`, `EventPayNotify`, `EventRefundNotify`, `EventKeepAlive` |

### 验证方法

每个自定义类型都提供 `IsValid()` 方法，用于运行时校验：

```go
payType := xgdnpay.PayType("native")
if payType.IsValid() {
    // 合法值
}

event := sse.EventPayNotify
if event.IsValid() {
    // 合法事件名
}
```

SDK 内部在关键入口处自动调用 `IsValid()`，例如 `CreateOrder` 会验证 `PayType`：

```go
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount:  0.01,
    Title:   "测试商品",
    PayType: xgdnpay.PayType("invalid"),  // 运行时返回错误
})
// err: "不支持的支付类型: invalid"
```

### 状态文本方法

`OrderStatus` 和 `RefundStatus` 提供 `Text()` 方法获取中文描述：

```go
status := xgdnpay.OrderStatusPaid
fmt.Println(status.Text())  // "已支付"

refundStatus := xgdnpay.RefundStatusSuccess
fmt.Println(refundStatus.Text())  // "退款成功"
```

当 API 返回 `int` 类型状态时，需要类型转换：

```go
result, _ := client.QueryOrder(ctx, orderNo)
text := xgdnpay.OrderStatus(result.Status).Text()  // "已支付"
```

### SSE 事件名类型

SSE 包的 `EventName` 类型确保事件名不会被拼错：

```go
// ✅ 正确：使用 EventName 常量
data := sse.FormatEvent(sse.EventPayNotify, msgBytes)

// ❌ 编译错误：FormatEvent 只接受 EventName 类型
data := sse.FormatEvent("pay_notify", msgBytes)
```

### 支付通知消息

`PayNotifyMessage` 使用 `PayStatus` 和 `PayChannel` 类型：

```go
msg := xgdnpay.NewPayNotifyMessage(orderNo, amount, xgdnpay.PayStatusPaid).
    SetPayType(xgdnpay.PayChannelWechat).
    SetOutOrderNo(outOrderNo).
    SetTransaction(transactionID)
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
// 扫码支付（最简写法，PayType 默认为 PayTypeNative）
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
refund, err := client.CreateRefund(ctx, &xgdnpay.RefundRequest{
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
// OrderStatus 常量
const (
    OrderStatusPending  OrderStatus = 0  // 待支付
    OrderStatusPaid     OrderStatus = 1  // 已支付
    OrderStatusClosed   OrderStatus = 2  // 已关闭
    OrderStatusRefunded OrderStatus = 3  // 已退款
)

// 获取状态中文文本
text := xgdnpay.OrderStatus(status).Text()  // "待支付"、"已支付" 等

// 退款状态
const (
    RefundStatusProcessing RefundStatus = 0  // 退款中
    RefundStatusSuccess    RefundStatus = 1  // 退款成功
    RefundStatusClosed     RefundStatus = 2  // 已关闭
    RefundStatusFailed     RefundStatus = 3  // 退款失败
    RefundStatusAbnormal   RefundStatus = 4  // 退款异常
)

text := xgdnpay.RefundStatus(status).Text()  // "退款中"、"退款成功" 等
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
    "github.com/skylark8866/paysdk/sse"
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

#### 失效连接自动清理

Hub 内置失效连接检测机制，通过心跳检测异常断开的客户端：

**工作原理：**
1. Hub 每隔 `keepAlive` 间隔向所有客户端发送心跳消息
2. 如果客户端的发送缓冲区已满（说明 handler 端已停止消费），心跳发送失败
3. 发送失败的客户端被标记为"失效"，立即从 Hub 中移除并关闭

**为什么缓冲区满意味着连接失效？**

正常情况下，handler 的 for 循环会持续消费 `client.Send` channel。当客户端异常断开（网络中断、浏览器崩溃）时：
- TCP 连接进入半打开状态，服务端操作系统不知道连接已断开
- `Context.Done()` 不会立即触发（等待 TCP 超时，可能 2-10 分钟）
- handler 的 for 循环阻塞在 `client.Send <- msg`，无法继续消费
- 缓冲区逐渐填满，最终导致心跳发送失败

通过心跳检测，Hub 可以在 `keepAlive` 间隔内发现并清理这些"僵尸连接"。

**建议配置：**
- 生产环境建议 `keepAlive` 设置为 10-30 秒
- 缓冲区大小 `bufferSize` 不宜过大，否则会延迟失效检测

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
// 使用类型安全的方式格式化并发送事件
msg := xgdnpay.NewPayNotifyMessage(orderNo, amount, xgdnpay.PayStatusPaid).
    SetPayType(xgdnpay.PayChannelWechat)
data := sse.FormatEvent(sse.EventPayNotify, mustMarshal(msg))
sseHub.Broadcast(orderNo, data)

// 或使用 JSON 广播
sseHub.BroadcastJSON(channel, map[string]any{
    "event": sse.EventPayNotify,
    "data":  msg,
})
```

**关键：** `sse.FormatEvent` 和 `Message.SetEvent` 只接受 `sse.EventName` 类型，不接受裸字符串。可用的事件名常量：

| 常量 | 值 | 用途 |
|------|-----|------|
| `sse.EventConnected` | `"connected"` | 连接确认 |
| `sse.EventPayNotify` | `"pay_notify"` | 支付通知 |
| `sse.EventRefundNotify` | `"refund_notify"` | 退款通知 |
| `sse.EventKeepAlive` | `"keep_alive"` | 心跳保活 |

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
    if xgdnpay.OrderStatus(req.Status) == xgdnpay.OrderStatusPaid {
        fmt.Printf("订单 %s 支付成功，金额: %.2f\n", req.OutOrderNo, req.Amount)
    }
    return nil
})

// 退款回调
handler := xgdnpay.NewRefundNotifyHandler(client, func(req *xgdnpay.RefundNotifyRequest) error {
    fmt.Printf("退款 %s 状态: %s\n", req.RefundNo, req.Status)
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

### 2. 使用类型安全常量

SDK 提供了完整的自定义类型体系，务必使用常量而非硬编码字符串：

```go
// ❌ 错误做法：硬编码字符串
data := sse.FormatEvent("pay_notify", msgBytes)
msg := xgdnpay.NewPayNotifyMessage(orderNo, amount, "paid").SetPayType("wechat")

// ✅ 正确做法：使用类型安全常量
data := sse.FormatEvent(sse.EventPayNotify, msgBytes)
msg := xgdnpay.NewPayNotifyMessage(orderNo, amount, xgdnpay.PayStatusPaid).SetPayType(xgdnpay.PayChannelWechat)
```

### 3. 订单号管理

SDK 会自动生成订单号，但建议使用自己的订单号：

```go
// ✅ 推荐：使用自己的订单号
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    OutOrderNo: "MY_ORDER_" + time.Now().Format("20060102150405"),
    Amount:     0.01,
    Title:      "商品名称",
})
```

**OutOrderNo 格式要求：**
- 长度：1-64 字符（空值允许，会自动生成）
- 字符集：只允许字母、数字、下划线（_）、中划线（-）
- 不允许中文、空格、特殊符号

```go
// ✅ 合法
"ORDER_001"
"order-2026-04-18-001"
"ORD123456"

// ❌ 非法（会返回错误）
"订单001"      // 包含中文
"order 001"    // 包含空格
"order@001"    // 包含特殊符号
"6"            // 过于简单，容易重复
```

SDK 会在 `CreateOrder` 时自动验证格式，也可手动调用：

```go
if err := xgdnpay.ValidateOutOrderNo(outOrderNo); err != nil {
    // 格式错误
}
```

### 4. SSE Channel 命名

使用订单号作为 channel，确保唯一性：

```go
// ✅ 正确：使用订单号
sseHub.Broadcast(orderNo, data)

// ❌ 错误：使用用户 ID（可能导致消息混乱）
sseHub.Broadcast(userID, data)
```

### 5. SSE 端点认证

SSE 端点必须添加认证，防止未授权订阅：

```go
// ✅ 正确：添加认证中间件
r.GET("/api/events/:channel", authMiddleware, sseHub.GinHandler(sse.WithConnectMessage()))

// ❌ 危险：无认证，任何人都能订阅
r.GET("/api/events/:channel", sseHub.GinHandler(sse.WithConnectMessage()))
```

认证中间件对 SSE 请求应返回 HTTP 401，而非 HTTP 200 + JSON body，否则 `EventSource` 无法正确触发 `onerror`。

### 6. SSE 连接数限制

Hub 内置连接数保护，防止资源耗尽：

```go
sseHub := sse.NewHub(
    sse.WithMaxClients(500),     // 最大总连接数
    sse.WithMaxPerChannel(5),    // 每 channel 最大连接数
)
```

超过限制时，`TrySubscribe` 返回错误，Gin/Handler 返回 HTTP 503。

### 7. 回调幂等性

回调可能重复发送，确保处理逻辑幂等：

```go
func handleCallback(req *xgdnpay.NotifyRequest) error {
    order, _ := db.GetOrder(req.OutOrderNo)
    if xgdnpay.OrderStatus(order.Status) == xgdnpay.OrderStatusPaid {
        return nil  // 已处理，直接返回成功
    }

    // 处理支付成功逻辑
    return db.UpdateOrderStatus(req.OutOrderNo, int(xgdnpay.OrderStatusPaid))
}
```

### 8. 超时设置

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

order, err := client.CreateOrder(ctx, req)
```

---

## 示例说明

示例位于仓库的 `example/` 目录下，不会随 `go get` 发布。如需运行示例，请克隆完整仓库：

```bash
git clone https://github.com/skylark8866/paysdk.git
cd paysdk/example/<example-name>
```

### shop-demo - Web 商城示例

完整的 Web 商城示例，包含：
- 用户注册/登录
- 充值套餐选择
- 二维码支付
- SSE 实时通知
- 支付回调处理

```bash
cd example/shop-demo
go build -o shop-demo-server.exe main.go embed.go
./shop-demo-server.exe
```

访问 http://localhost:8081 体验完整流程。

### desktop - 桌面应用

使用 Gio 框架的桌面 GUI 工具：

```bash
cd example/desktop
go run main.go
```

### dev-test - 开发调试工具

命令行调试工具，用于测试 API 接口：

```bash
cd example/dev-test
go run main.go
```

### main.go - 基础命令行示例

最简化的命令行示例，展示核心 API 用法：

```bash
cd example
go run main.go
```

---

## 常见问题

### Q: 如何安装 SDK？

A: 使用 Go 模块安装：
```bash
go get github.com/skylark8866/paysdk@v1.4.0
```

### Q: SSE 连接建立后收不到事件？

A: 确保使用的是事件格式 `event: name\ndata: {}\n\n`，而不是注释格式 `: comment\n\n`。

### Q: 回调签名验证失败？

A: 检查以下几点：
1. AppSecret 是否正确
2. 时间戳是否在有效期内（默认 300 秒）
3. 签名算法是否正确（SHA-256）

### Q: 如何处理网络超时？

A: 使用 context 设置超时：

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
```

### Q: StatusText / RefundStatusText 函数找不到了？

A: v1.4.0 移除了这两个函数，改用类型方法：

```go
// 旧写法（v1.3.0 及之前）
text := xgdnpay.StatusText(status)
text := xgdnpay.RefundStatusText(status)

// 新写法（v1.4.0+）
text := xgdnpay.OrderStatus(status).Text()
text := xgdnpay.RefundStatus(status).Text()
```

### Q: FormatEvent 不能传字符串了？

A: v1.4.0 将事件名改为 `sse.EventName` 类型，必须使用常量：

```go
// 旧写法（v1.3.0 及之前）
data := sse.FormatEvent("pay_notify", msgBytes)

// 新写法（v1.4.0+）
data := sse.FormatEvent(sse.EventPayNotify, msgBytes)
```

### Q: 示例代码如何运行？

A: 示例不在 Go 模块中发布，需要克隆仓库：
```bash
git clone https://github.com/skylark8866/paysdk.git
cd paysdk/example/<example-name>
go run main.go
```

---

## 更新日志

### v1.4.2 (2026-04-18)

- **OutOrderNo 格式验证**：SDK 端验证商户订单号格式，长度限制 64 字符，只允许字母、数字、下划线、中划线
- **后端重复订单号检查**：创建订单前检查 OutOrderNo 是否已存在，返回友好错误提示
- 新增 `ValidateOutOrderNo()` 函数用于验证商户订单号
- 修复用户传入简单订单号（如 "6"）重复时返回 500 错误的问题

### v1.4.1 (2026-04-18)

- **SSE 失效连接自动清理**：Hub 通过心跳检测异常断开的客户端，缓冲区满时自动移除并关闭连接
- 新增 `cleanStaleClients()` 方法，在心跳发送失败时清理僵尸连接
- 解决客户端异常断开（网络中断、浏览器崩溃）后连接残留问题

### v1.4.0 (2026-04-18) ⚠️ 破坏性变更

- **类型安全体系**：新增 `PayType`、`PayStatus`、`PayChannel`、`OrderStatus`、`RefundStatus` 自定义类型，每个类型提供 `IsValid()` 验证方法
- **SSE EventName 类型**：`sse.FormatEvent` 和 `Message.SetEvent` 参数从 `string` 改为 `sse.EventName`，编译期拦截非法事件名
- **移除 `StatusText` / `RefundStatusText`**：替换为 `OrderStatus.Text()` / `RefundStatus.Text()` 方法
- **PayNotifyMessage 类型安全**：`status` 字段改用 `PayStatus`，`pay_type` 字段改用 `PayChannel`
- **消除全部硬编码**：SDK 核心代码零硬编码字符串，所有字段名、错误信息、路径统一由常量管理
- **CreateOrder 自动验证 PayType**：传入非法支付类型时返回错误

**迁移指南：**
```go
// 1. StatusText → OrderStatus.Text()
xgdnpay.StatusText(status)           → xgdnpay.OrderStatus(status).Text()
xgdnpay.RefundStatusText(status)     → xgdnpay.RefundStatus(status).Text()

// 2. SSE 事件名使用 EventName 常量
sse.FormatEvent("pay_notify", data)  → sse.FormatEvent(sse.EventPayNotify, data)
sse.NewMessage().SetEvent("pay_notify") → sse.NewMessage().SetEvent(sse.EventPayNotify)

// 3. PayNotifyMessage 使用类型安全常量
NewPayNotifyMessage(no, amt, "paid")           → NewPayNotifyMessage(no, amt, xgdnpay.PayStatusPaid)
msg.SetPayType("wechat")                       → msg.SetPayType(xgdnpay.PayChannelWechat)

// 4. OrderStatus 比较需要类型转换（API 返回的 Status 是 int）
if req.Status == xgdnpay.OrderStatusPaid       → if xgdnpay.OrderStatus(req.Status) == xgdnpay.OrderStatusPaid
```

### v1.3.0 (2026-04-18) ⚠️ 破坏性变更

- **模块路径迁移**：从 `xgdn-pay` 迁移至 `github.com/skylark8866/paysdk`
- **目录结构重构**：SDK 代码从 `sdk/go/` 提升至仓库根目录，符合 Go 标准模块布局
- **所有 import 路径更新**：请将项目中所有 `"xgdn-pay"` 替换为 `"github.com/skylark8866/paysdk"`
- **移除 JS SDK**：前端代码简单，直接使用浏览器原生 EventSource API 即可
- **移除废弃示例**：删除依赖 JS SDK 的 `sse-test` 和 `sdk-integration-test`
- **添加 MIT LICENSE**
- **支持 `go get` 直接引用**：`go get github.com/skylark8866/paysdk@v1.3.0`

### v1.2.0 (2026-04-17)

- SSE Hub 添加连接数限制（WithMaxClients、WithMaxPerChannel）
- SSE Hub 添加 TrySubscribe 方法，超限时返回错误
- SSE 端点添加 JWT 认证保护
- 认证中间件区分 SSE 请求，返回正确的 HTTP 状态码
- 前端 SSE 添加指数退避重连机制
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
