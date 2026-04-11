# XGDN 支付演示商城

这是一个最小化的支付演示项目，展示如何使用 XGDN Pay SDK 集成支付功能。

## 功能特性

- ✅ 商品展示
- ✅ 创建订单
- ✅ 生成支付二维码
- ✅ 实时监听支付状态（SSE）
- ✅ 支付成功跳转
- ✅ 支付回调处理

## 快速开始

### 方式一：使用默认配置（生产环境）

```powershell
# Windows
.\run.ps1

# Linux/Mac
chmod +x run.sh
./run.sh
```

### 方式二：自定义配置

```powershell
# Windows
$env:XGDN_APP_ID="your_app_id"
$env:XGDN_APP_SECRET="your_app_secret"
$env:XGDN_BASE_URL="https://pay.xgdn.net"
$env:PORT="8080"
go run .
```

```bash
# Linux/Mac
export XGDN_APP_ID="your_app_id"
export XGDN_APP_SECRET="your_app_secret"
export XGDN_BASE_URL="https://pay.xgdn.net"
export PORT="8080"
go run .
```

## 环境变量说明

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| XGDN_APP_ID | 应用ID | （必填） |
| XGDN_APP_SECRET | 应用密钥 | （必填） |
| XGDN_BASE_URL | 支付平台地址 | https://pay.xgdn.net |
| PORT | 服务端口 | 8080 |
| PUBLIC_URL | 公网地址（用于回调） | http://localhost:8080 |

## 访问地址

启动后访问：http://localhost:8080

## 支付流程

```
1. 浏览商品列表
2. 点击"立即购买"
3. 系统创建订单并生成二维码
4. 使用微信扫码支付
5. 支付成功后自动跳转到结果页
```

## 技术栈

- **后端**: Go + html/template
- **前端**: 原生 HTML/CSS/JavaScript
- **二维码**: QRCode.js
- **实时通知**: Server-Sent Events (SSE)

## 注意事项

### 关于回调地址

由于演示程序运行在本地，支付平台无法直接回调到 `localhost`。有两种解决方案：

#### 方案一：使用内网穿透工具（推荐）

使用 ngrok、frp 等工具将本地服务映射到公网：

```bash
# 使用 ngrok
ngrok http 8080

# 会得到类似 https://abc123.ngrok.io 的地址
# 然后设置环境变量
export PUBLIC_URL="https://abc123.ngrok.io"
```

#### 方案二：使用轮询模式

当前实现已经包含 SSE 实时监听，即使回调失败，前端也会通过轮询获取支付状态。

## 目录结构

```
shop-demo/
├── main.go              # 主程序
├── go.mod               # Go 模块文件
├── run.ps1              # Windows 运行脚本
├── run.sh               # Linux/Mac 运行脚本
├── README.md            # 本文件
└── templates/           # HTML 模板
    ├── index.html       # 商品列表页
    ├── pay.html         # 支付页
    └── result.html      # 支付结果页
```

## 扩展建议

这个演示项目可以扩展为：

1. **添加数据库** - 使用 SQLite/MySQL 存储订单
2. **添加用户系统** - 用户登录、订单历史
3. **添加商品管理** - 后台管理商品
4. **添加退款功能** - 集成退款 API
5. **添加多种支付方式** - JSAPI、H5 等

## License

MIT
