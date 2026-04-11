# XGDN Pay SDK

XGDN 支付平台多语言 SDK，帮助第三方软件快速集成支付功能。

## 目录结构

```
sdk/
├── go/                    # Go 后端 SDK
│   ├── client.go         # 客户端核心（订单、退款、查询）
│   ├── sign.go           # 签名与验签
│   ├── notify.go         # 回调通知处理
│   ├── types.go          # 类型定义与常量
│   ├── example/          # 使用示例
│   │   ├── main.go       # 命令行完整示例
│   │   ├── shop-demo/    # Web 商城示例
│   │   ├── desktop/      # 桌面 GUI 工具
│   │   ├── dev-test/     # 开发调试工具
│   │   └── sdk-integration-test/  # 前后端集成测试
│   └── README.md         # Go SDK 文档
├── js/                    # JavaScript 前端 SDK
│   ├── xgdn-pay.js       # SDK 源码
│   └── package.json
└── README.md              # 本文件
```

## 快速选择

| 场景 | 推荐方案 |
|------|----------|
| Go 后端服务 | [Go SDK](./go/README.md) |
| 前端支付监听 | [JS SDK](#javascript-前端-sdk) |
| 桌面软件 | Go SDK + 轮询模式 |
| 全栈应用 | Go SDK（后端）+ JS SDK（前端） |

## 接入流程

```
1. 注册站点 → 2. 获取密钥 → 3. 安装 SDK → 4. 测试验证 → 5. 上线运营
```

1. **注册站点** - 在管理后台创建站点，获取 `app_id` 和 `app_secret`
2. **获取密钥** - 记录 AppID 和 AppSecret
3. **安装 SDK** - 选择对应语言的 SDK
4. **测试验证** - 使用沙箱环境测试
5. **上线运营** - 切换到生产环境

## 配置凭证

SDK 需要配置 AppID 和 AppSecret 才能使用。**请勿将凭证硬编码到代码中或提交到版本控制。**

### 环境变量（推荐）

```bash
# Linux/Mac
export XGDN_APP_ID="your_app_id"
export XGDN_APP_SECRET="your_app_secret"

# Windows PowerShell
$env:XGDN_APP_ID="your_app_id"
$env:XGDN_APP_SECRET="your_app_secret"
```

### 代码中读取

```go
appID := os.Getenv("XGDN_APP_ID")
appSecret := os.Getenv("XGDN_APP_SECRET")
if appID == "" || appSecret == "" {
    log.Fatal("请配置 XGDN_APP_ID 和 XGDN_APP_SECRET 环境变量")
}
client := xgdnpay.NewClient(appID, appSecret)
```

## 典型集成方案

### 后端创建订单 + 前端监听支付（推荐）

```
用户下单 → Go 后端调用 CreateOrder → 返回 CodeURL → 前端展示二维码
                                                    ↓
前端 JS SDK 监听 SSE ← 支付平台推送 ← 用户扫码支付 ←
                                                    ↓
Go 后端收到回调通知 ← 支付平台回调 ← 支付成功 ←
```

**Go 后端：**

```go
client := xgdnpay.NewClient(appID, appSecret)

// 创建订单
order, _ := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount: 0.01,
    Title:  "商品名称",
})
// 返回 order.CodeURL 给前端

// 处理支付回调
http.Handle("/callback", xgdnpay.NewNotifyHandler(client, func(req *xgdnpay.NotifyRequest) error {
    if req.Status == xgdnpay.OrderStatusPaid {
        // 更新订单状态
    }
    return nil
}))
```

**JS 前端：**

```html
<script src="xgdn-pay.js"></script>
<script>
// 监听支付状态
const watcher = XGDNPay.watch(orderNo, {
    baseURL: 'https://pay.xgdn.net',
    onPaid: (data) => {
        alert('支付成功！');
        location.href = '/success';
    },
    onTimeout: () => alert('支付超时'),
});
</script>
```

### 纯后端轮询（桌面/CLI）

```go
order, _ := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount: 0.01,
    Title:  "商品名称",
})
// 展示 order.CodeURL 的二维码
err := client.WaitForPayment(ctx, order.OrderNo, 2*time.Second, 5*time.Minute)
```

## API 端点

| 环境 | 地址 |
|------|------|
| 生产环境 | `https://pay.xgdn.net` |
| 开发环境 | `http://localhost:8093` |

## 支付类型

| 类型 | 值 | 说明 | 适用场景 |
|------|------|------|----------|
| Native | `native` | 扫码支付 | PC 网站、桌面软件 |
| JSAPI | `jsapi` | 微信内支付 | 微信公众号、小程序 |

## JavaScript 前端 SDK

### 引入方式

```html
<!-- CDN -->
<script src="xgdn-pay.js"></script>

<!-- npm -->
npm install xgdn-pay
```

### 初始化

```javascript
const pay = new XGDNPay({
    appID: 'your_app_id',
    baseURL: 'https://pay.xgdn.net'  // 可选，默认生产环境
})
```

### 监听支付状态

```javascript
const watcher = XGDNPay.watch(orderNo, {
    baseURL: 'https://pay.xgdn.net',  // 可选
    timeout: 300000,                   // 可选，超时时间（毫秒），默认 5 分钟

    onConnected: function() {
        console.log('SSE 连接已建立');
    },
    onPaid: function(data) {
        console.log('支付成功', data);
        // data.status === 'paid'
    },
    onTimeout: function() {
        console.log('支付超时');
    },
    onError: function(err) {
        console.log('连接错误', err);
    }
})

// 手动关闭
watcher.close()
```

### 一键支付（推荐）

自动获取支付信息、展示二维码、监听支付状态：

```javascript
const payment = pay.createPayment(orderNo, {
    container: '#qrcode',      // 二维码容器（CSS 选择器或 DOM 元素）
    qrSize: 200,               // 二维码大小，默认 200px
    timeout: 300000,           // 超时时间（毫秒）

    onReady: function(payInfo) {
        console.log('支付信息已获取', payInfo);
    },
    onPaid: function(data) {
        console.log('支付成功', data);
    },
    onTimeout: function() {
        console.log('支付超时');
    },
    onError: function(err) {
        console.log('错误', err);
    }
})

// 启动支付
payment.start()

// 关闭支付
payment.close()
```

### 获取支付信息

```javascript
// 扫码支付
const payInfo = await pay.getPayInfo(orderNo)
// payInfo.code_url  → 二维码内容
// payInfo.pay_url   → 支付跳转链接

// JSAPI 支付
const jsapiInfo = await pay.getJSAPIPayInfo(orderNo, openID)
```

### 常量

```javascript
XGDNPay.PayType = {
    NATIVE: 'native',  // 扫码支付
    JSAPI: 'jsapi'     // 微信内支付
}

XGDNPay.OrderStatus = {
    PENDING: 0,   // 待支付
    PAID: 1,      // 已支付
    CLOSED: 2,    // 已关闭
    REFUNDED: 3   // 已退款
}

XGDNPay.RefundStatus = {
    PROCESSING: 0,  // 退款中
    SUCCESS: 1,     // 退款成功
    CLOSED: 2,      // 已关闭
    FAILED: 3,      // 退款失败
    ABNORMAL: 4     // 退款异常
}
```

### SSE 重连机制

JS SDK 内置自动重连：

- 断线后自动重连，无需手动处理
- 重连延迟从 1 秒开始，指数递增（1s → 2s → 4s → ...），最大 30 秒
- 连接成功后重置延迟
- 超时或手动关闭后不再重连

## License

MIT
