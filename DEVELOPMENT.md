# XGDN Pay SDK 开发手册

> 版本：v1.6.0 | 仓库：https://github.com/skylark8866/paysdk

## 目录

- [快速开始](#快速开始)
- [无站点支付（推荐）](#无站点支付推荐)
- [传统回调模式](#传统回调模式)
- [API 参考](#api-参考)
- [类型速查](#类型速查)
- [前端集成](#前端集成)
- [常见问题](#常见问题)
- [更新日志](#更新日志)

---

## 快速开始

### 安装

```bash
go get github.com/skylark8866/paysdk@v1.6.0
```

### 配置凭证

创建 `config.yaml`（与程序同目录）：

```yaml
payment:
  app_id: "your_app_id"
  app_secret: "your_app_secret"
```

也可通过环境变量覆盖：`XGDN_PAY_APP_ID`、`XGDN_PAY_APP_SECRET`、`XGDN_PAY_BASE_URL`

### 30 秒上手

```go
package main

import (
    "context"
    "fmt"

    xgdnpay "github.com/skylark8866/paysdk"
)

func main() {
    client := xgdnpay.NewClient()

    order, err := client.CreateOrder(context.Background(), &xgdnpay.CreateOrderRequest{
        Amount: 100,
        Title:  "测试商品",
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("订单号:", order.OrderNo)
    fmt.Println("二维码内容:", order.CodeURL)
}
```

---

## 无站点支付（推荐）

传统支付依赖 HTTP 回调，站点必须公网可达。**无站点支付**通过 SSE 订阅替代 HTTP 回调，站点无需暴露任何端口，本地开发 / 内网 / 生产环境均可用。

```
传统模式：支付网关 ──HTTP POST──► 站点后端（需要公网可达）
无站点模式：站点后端 ──SSE 订阅──► 支付网关 ──SSE 推送──► 站点后端（无需公网可达）
```

### 3 步集成

**第 1 步：初始化**

```go
client := xgdnpay.NewClient()
paySSE := xgdnpay.NewPaySSE()
defer paySSE.Shutdown()
```

**第 2 步：注册 SSE 端点 + 创建订单时订阅**

```go
// 注册 SSE 端点（供前端 EventSource 连接，必须加认证中间件）
r.GET("/api/events/:channel", authMiddleware, paySSE.Hub().GinHandler(sse.WithConnectMessage()))

// 创建订单后订阅支付通知
payOrder, _ := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    OutOrderNo: orderNo,
    Amount:     100,
    Title:      "充值-10元套餐",
})

paySSE.Subscribe(payOrder.OrderNo, func(event *sse.PayNotifyEvent) {
    if event.Status != "paid" {
        return
    }
    processPayment(orderNo)

    msg := xgdnpay.NewPayNotifyMessage(orderNo, amount, xgdnpay.PayStatusPaid).
        SetPayType(xgdnpay.PayChannelWechat)
    paySSE.Hub().BroadcastMessage(orderNo, msg)
})
```

**第 3 步：前端监听**

```javascript
var es = new EventSource('/api/events/' + orderNo);
es.addEventListener('pay_notify', function(e) {
    var data = JSON.parse(e.data);
    if (data.status === 'paid') {
        es.close();
        alert('支付成功！');
    }
});
```

### 注意事项

1. **PayOrderNo 必须存储**：`Subscribe` 使用支付网关订单号（`payOrder.OrderNo`），必须通过 `PayOrderNo` 字段关联本地订单
2. **幂等保护**：回调可能重复触发，已支付订单不可重复处理
3. **优雅关闭**：务必 `defer paySSE.Shutdown()` 避免 goroutine 泄漏
4. **自动重连**：Subscriber 内置指数退避重连（1s → 2s → 4s → ... → 60s），断线不丢通知

### 完整示例

参考 `example/shop-demo-no-server/`，完整无站点支付商城，包含用户注册/登录、充值、二维码支付、SSE 实时通知。

---

## 传统回调模式

站点可公网可达时，可使用 HTTP 回调 + SSE Hub 广播：

```go
sseHub := sse.NewHub()
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go sseHub.Run(ctx)

// 创建订单时传入 NotifyURL
order, _ := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount:    100,
    Title:     "商品名称",
    NotifyURL: "https://yoursite.com/api/callback/pay",
})

// 回调处理中向 Hub 广播
func HandleCallback(c *gin.Context) {
    var req xgdnpay.NotifyRequest
    c.ShouldBindJSON(&req)
    // 验签 + 业务处理...
    msg := xgdnpay.NewPayNotifyMessage(req.OutOrderNo, req.Amount, xgdnpay.PayStatusPaid)
    sseHub.BroadcastMessage(req.OutOrderNo, msg)
}
```

---

## API 参考

### 客户端

```go
// 自动加载凭证（推荐）
client := xgdnpay.NewClient()

// 手动覆盖
client := xgdnpay.NewClient(
    xgdnpay.WithAppConfig("app_id", "app_secret"),
    xgdnpay.WithBaseURL("https://pay.xgdn.net"),
    xgdnpay.WithTimeout(30 * time.Second),
)
```

凭证加载优先级：环境变量 > `WithAppConfig` > config.yaml > config.local.yaml

### 订单操作

```go
// 创建订单
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    OutOrderNo: "YOUR_ORDER_001",  // 可选，不填自动生成
    Amount:     100,               // 必填，单位：分
    Title:      "商品名称",         // 必填
    PayType:    xgdnpay.PayTypeNative,  // 可选，默认 native
    ReturnURL:  "https://yoursite.com/success",  // 可选
    NotifyURL:  "https://yoursite.com/callback", // 可选
})

// JSAPI 支付（微信公众号/小程序）
order, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount:  100,
    Title:   "商品名称",
    PayType: xgdnpay.PayTypeJSAPI,
    OpenID:  "user_openid",  // JSAPI 必填
})

// 查询订单
order, err := client.QueryOrder(ctx, "ORD_xxx")

// 关闭订单
err := client.CloseOrder(ctx, "ORD_xxx")
```

**OutOrderNo 格式要求**：1-64 字符，只允许字母、数字、下划线、中划线。SDK 自动验证，也可手动调用 `xgdnpay.ValidateOutOrderNo(no)`。

### 退款操作

```go
// 创建退款
refund, err := client.CreateRefund(ctx, &xgdnpay.RefundRequest{
    OrderNo:  "ORD_xxx",
    Amount:   100,
    Reason:   "用户申请退款",  // 可选，默认"用户申请退款"
    RefundNo: "REF_xxx",       // 可选，不填自动生成
})

// 查询退款
refund, err := client.QueryRefund(ctx, "REF_xxx")
```

### 回调处理

```go
// 支付回调（泛型处理器）
handler := xgdnpay.NewNotifyHandler(client, func(req *xgdnpay.NotifyRequest) error {
    if xgdnpay.OrderStatus(req.Status) == xgdnpay.OrderStatusPaid {
        fmt.Printf("订单 %s 支付成功\n", req.OutOrderNo)
    }
    return nil
})

// 退款回调
handler := xgdnpay.NewRefundNotifyHandler(client, func(req *xgdnpay.RefundNotifyRequest) error {
    fmt.Printf("退款 %s 状态: %s\n", req.RefundNo, req.Status)
    return nil
})
```

回调可能重复发送，处理逻辑必须幂等。

### SSE 配置

#### PaySSE（推荐）

```go
paySSE := xgdnpay.NewPaySSE(
    xgdnpay.WithSSEBaseURL("https://pay.xgdn.net"),
    xgdnpay.WithSSEHubOpts(
        sse.WithKeepAlive(10*time.Second),
        sse.WithMaxClients(500),
    ),
)
defer paySSE.Shutdown()

paySSE.Subscribe(payOrderNo, callback)   // 订阅支付网关通知
paySSE.Hub()                             // 获取 Hub（广播/注册端点）
paySSE.Hub().BroadcastMessage(ch, msg)   // 向前端广播
paySSE.Hub().GinHandler(sse.WithConnectMessage())  // Gin 端点
```

#### 单独使用 Hub

```go
sseHub := sse.NewHub(
    sse.WithKeepAlive(10*time.Second),   // 心跳间隔，默认 15 秒
    sse.WithMaxClients(500),             // 最大总连接数，默认 1000
    sse.WithMaxPerChannel(5),            // 每 channel 最大连接数，默认 10
    sse.WithHubBufferSize(512),          // 客户端缓冲区大小，默认 256
)
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
go sseHub.Run(ctx)
```

#### 单独使用 Subscriber

```go
subscriber := sse.NewSubscriber(sse.WithBaseURL("https://pay.xgdn.net"))
defer subscriber.Close()

subscriber.Subscribe(payOrderNo, func(event *sse.PayNotifyEvent) {
    // event.Status: "paid" / "refunded" / ...
    // event.OrderNo: 支付网关订单号
    // event.TransactionID: 微信交易号
})
subscriber.Unsubscribe(payOrderNo)
subscriber.IsSubscribed(payOrderNo)
subscriber.SubCount()
```

#### 标准库适配器

```go
http.Handle("/api/events/", sseHub.Handler(
    sse.WithHandlerChannelFunc(func(r *http.Request) string {
        return strings.TrimPrefix(r.URL.Path, "/api/events/")
    }),
))
```

#### SSE 端点认证

SSE 端点**必须**添加认证中间件。`EventSource` API 不支持自定义 Header，认证应通过 Cookie 实现：

```go
r.GET("/api/events/:channel", authMiddleware, sseHub.GinHandler(sse.WithConnectMessage()))
```

认证中间件对 SSE 请求应返回 HTTP 401（而非 200 + JSON），这样 `EventSource` 才能正确触发 `onerror`：

```go
func writeAuthError(c *gin.Context, message string) {
    accept := c.GetHeader("Accept")
    if strings.Contains(accept, "text/event-stream") {
        c.JSON(http.StatusUnauthorized, gin.H{"error": message})
        return
    }
    c.JSON(http.StatusOK, gin.H{"code": 401, "message": message})
}
```

#### 广播消息

```go
// 推荐：一行搞定
msg := xgdnpay.NewPayNotifyMessage(orderNo, amount, xgdnpay.PayStatusPaid).
    SetPayType(xgdnpay.PayChannelWechat).
    SetOutOrderNo(outOrderNo).
    SetTransaction(transactionID)
sseHub.BroadcastMessage(orderNo, msg)

// 底层 API
data := sse.FormatEvent(sse.EventPayNotify, jsonData)
sseHub.Broadcast(orderNo, data)
```

#### Gin Handler 选项

```go
sseHub.GinHandler(
    sse.WithChannelParam("order_id"),           // 自定义 channel 参数名
    sse.WithConnectMessage(),                     // 发送 connected 事件（推荐）
    sse.WithBeforeSubscribe(func(c *gin.Context, ch string) error {
        return nil  // 订阅前验证
    }),
)
```

### 错误处理

SDK 错误均为 `*xgdnpay.SDKError`，可通过 `errors.As` 判断：

```go
order, err := client.CreateOrder(ctx, req)
if err != nil {
    var sdkErr *xgdnpay.SDKError
    if errors.As(err, &sdkErr) {
        // sdkErr.Code: -1 参数错误, -2 请求失败, -3 签名失败, -4 超时, -5 订单不存在
    }
    return err
}
```

---

## 类型速查

| 类型 | 常量 | 说明 |
|------|------|------|
| `PayType` | `PayTypeNative` `PayTypeJSAPI` `PayTypeH5` | 支付方式 |
| `PayStatus` | `PayStatusPaid` `PayStatusPending` `PayStatusClosed` | 支付状态 |
| `PayChannel` | `PayChannelWechat` `PayChannelAlipay` | 支付渠道 |
| `OrderStatus` | `OrderStatusPending(0)` `OrderStatusPaid(1)` `OrderStatusClosed(2)` `OrderStatusRefunded(3)` | 订单状态，`.Text()` 获取中文 |
| `RefundStatus` | `RefundStatusProcessing(0)` `RefundStatusSuccess(1)` `RefundStatusClosed(2)` `RefundStatusFailed(3)` `RefundStatusAbnormal(4)` | 退款状态，`.Text()` 获取中文 |
| `sse.EventName` | `EventConnected` `EventPayNotify` `EventRefundNotify` `EventKeepAlive` | SSE 事件名 |

所有类型均提供 `IsValid()` 运行时校验。API 返回 `int` 状态时需类型转换：`xgdnpay.OrderStatus(status).Text()`

---

## 前端集成

使用浏览器原生 `EventSource` API，配合指数退避重连：

```javascript
function connectSSE(orderNo, retryCount) {
    retryCount = retryCount || 0;
    var es = new EventSource('/api/events/' + orderNo);

    es.addEventListener('connected', function() { retryCount = 0; });

    es.addEventListener('pay_notify', function(e) {
        var data = JSON.parse(e.data);
        if (data.status === 'paid') {
            es.close();
            onPaid(data);
        }
    });

    es.onerror = function() {
        es.close();
        if (retryCount >= 10) { onFallback(); return; }
        var delay = Math.min(1000 * Math.pow(2, retryCount), 30000);
        setTimeout(function() { connectSSE(orderNo, retryCount + 1); }, delay);
    };
}

connectSSE(orderNo);
```

重连策略：1s → 2s → 4s → ... → 30s，最多 10 次，超限降级为轮询。

---

## 常见问题

**Q: 无站点支付和传统回调怎么选？**

优先无站点支付（`PaySSE`）。传统回调仅在需要审计日志或已有回调基础设施时使用。两种模式可叠加。

**Q: SSE 订阅断开后支付结果会丢失吗？**

不会。Subscriber 内置自动重连（指数退避 1s → 60s），支付网关保留推送能力，重连后正常接收。

**Q: SSE 端点收不到事件？**

确保使用事件格式 `event: name\ndata: {}\n\n`，而非注释格式 `: comment\n\n`。使用 `sse.WithConnectMessage()` 可自动发送 `connected` 事件。

**Q: 回调签名验证失败？**

检查：1) AppSecret 是否正确；2) 时间戳是否在 300 秒有效期内；3) 签名算法是否为 SHA-256。

**Q: 示例代码如何运行？**

示例不随 `go get` 发布，需克隆仓库：`git clone https://github.com/skylark8866/paysdk.git && cd paysdk/example/<name>`

---

## 更新日志

### v1.6.0 (2026-05-26)

- **无站点支付模式**：新增 `PaySSE` 统一入口，3 步完成集成
- **SSE Subscriber**：向支付网关订阅支付通知，内置指数退避自动重连
- **客户端自动加载凭证**：`NewClient()` 从 config.yaml / 环境变量自动加载
- **标准库适配器**：SSE Hub 支持 `net/http`（`Handler()` 方法）
- **shop-demo-no-server 示例**：完整的无站点支付商城

### v1.5.0 (2026-04-19)

- 新增 `SSEMessage` 接口，`BroadcastMessage` 一行完成广播
- `PayNotifyMessage` 自动绑定 `EventPayNotify` 事件名

### v1.4.0 (2026-04-18) ⚠️ 破坏性变更

- 类型安全体系：`PayType`/`PayStatus`/`PayChannel`/`OrderStatus`/`RefundStatus`/`EventName`
- `StatusText` → `OrderStatus(status).Text()`，`RefundStatusText` → `RefundStatus(status).Text()`
- SSE 事件名从 `string` 改为 `sse.EventName` 类型

### v1.3.0 (2026-04-18) ⚠️ 破坏性变更

- 模块路径从 `xgdn-pay` 迁移至 `github.com/skylark8866/paysdk`
