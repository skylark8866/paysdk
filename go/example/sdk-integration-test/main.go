package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	xgdnpay "xgdn-pay"
)

var (
	client *xgdnpay.Client
	jsSDK  []byte
)

func init() {
	appID := os.Getenv("XGDN_APP_ID")
	appSecret := os.Getenv("XGDN_APP_SECRET")
	baseURL := os.Getenv("XGDN_BASE_URL")

	if appID == "" {
		appID = "app_your_app_id"
	}
	if appSecret == "" {
		appSecret = "your_app_secret"
	}
	if baseURL == "" {
		baseURL = "https://pay.xgdn.net"
	}

	client = xgdnpay.NewClient(appID, appSecret,
		xgdnpay.WithBaseURL(baseURL),
	)

	cwd, _ := os.Getwd()
	jsPath := filepath.Join(cwd, "..", "..", "..", "js", "xgdn-pay.js")
	jsPath = strings.ReplaceAll(jsPath, `\`, "/")

	var err error
	jsSDK, err = os.ReadFile(jsPath)
	if err != nil {
		log.Printf("警告: 无法读取 JS SDK 文件: %v\n", err)
	}

	fmt.Println("=====================================")
	fmt.Println("  XGDN Pay SDK 集成测试 Demo")
	fmt.Println("=====================================")
	fmt.Printf("  AppID:    %s\n", appID)
	fmt.Printf("  BaseURL:  %s\n", baseURL)
	fmt.Printf("  JS SDK:   %s\n", jsPath)
	fmt.Println("=====================================")
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", serveIndex)
	mux.HandleFunc("/xgdn-pay.js", serveJSSDK)
	mux.HandleFunc("/api/order/create", handleCreateOrder)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8099"
	}

	fmt.Printf("\n服务启动: http://localhost:%s\n\n", port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

func serveJSSDK(w http.ResponseWriter, r *http.Request) {
	if len(jsSDK) == 0 {
		http.Error(w, "JS SDK 未加载", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(jsSDK)
}

type APIResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		writeJSON(w, APIResponse{Code: 405, Msg: "方法不允许"})
		return
	}

	var req struct {
		Amount  float64 `json:"amount"`
		Title   string  `json:"title"`
		PayType string  `json:"payType"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, APIResponse{Code: 400, Msg: "参数错误: " + err.Error()})
		return
	}

	payType := xgdnpay.NormalizePayType(req.PayType)
	if payType == "" {
		payType = xgdnpay.PayTypeNative
	}

	resp, err := client.CreateOrder(r.Context(), &xgdnpay.CreateOrderRequest{
		Amount:  req.Amount,
		Title:   req.Title,
		PayType: payType,
	})
	if err != nil {
		writeJSON(w, APIResponse{Code: 500, Msg: "创建订单失败: " + err.Error()})
		return
	}

	writeJSON(w, APIResponse{Code: 0, Msg: "成功", Data: map[string]interface{}{
		"orderNo": resp.OrderNo,
		"payUrl":  resp.PayURL,
		"codeUrl": resp.CodeURL,
		"amount":  req.Amount,
		"payType": payType,
	}})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	json.NewEncoder(w).Encode(data)
}

var indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>XGDN Pay SDK 集成测试</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
               background: #f5f7fa; color: #333; line-height: 1.6; }
        .container { max-width: 900px; margin: 0 auto; padding: 20px; }
        h1 { text-align: center; color: #2c3e50; margin-bottom: 30px; font-size: 24px; }
        h2 { color: #34495e; margin: 20px 0 15px; font-size: 18px; border-left: 4px solid #3498db; padding-left: 10px; }

        .card { background: white; border-radius: 12px; padding: 25px; margin-bottom: 20px;
                box-shadow: 0 2px 12px rgba(0,0,0,0.08); }
        
        .form-row { display: flex; gap: 15px; margin-bottom: 15px; flex-wrap: wrap; }
        .form-group { flex: 1; min-width: 200px; }
        .form-group label { display: block; font-weight: 600; margin-bottom: 6px; color: #555; font-size: 14px; }
        .form-group input, .form-group select {
            width: 100%; padding: 10px 12px; border: 1px solid #ddd; border-radius: 8px;
            font-size: 14px; transition: border-color 0.3s;
        }
        .form-group input:focus, .form-group select:focus { outline: none; border-color: #3498db; }

        button { padding: 12px 24px; border: none; border-radius: 8px; cursor: pointer;
                 font-size: 14px; font-weight: 600; transition: all 0.3s; }
        .btn-primary { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .btn-primary:hover { transform: translateY(-2px); box-shadow: 0 4px 15px rgba(102,126,234,0.4); }
        .btn-warning { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); color: white; }
        .btn-info { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); color: white; }
        button:disabled { opacity: 0.6; cursor: not-allowed; transform: none !important; }

        .qr-container { text-align: center; padding: 20px; min-height: 280px; display: flex;
                       align-items: center; justify-content: center; background: #fafafa;
                       border-radius: 10px; border: 2px dashed #ddd; }
        .qr-container img { border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.1); }

        .log-area { background: #1e1e1e; color: #00ff88; padding: 15px; border-radius: 8px;
                    font-family: 'Consolas', 'Monaco', monospace; font-size: 13px;
                    max-height: 300px; overflow-y: auto; white-space: pre-wrap; word-break: break-all; }
        .log-area .error { color: #ff6b6b; }
        .log-area .success { color: #69db7c; }
        .log-area .info { color: #74c0fc; }
        .log-area .warn { color: #ffd43b; }

        .result-panel { display: none; margin-top: 20px; padding: 20px; background: #f0fff4;
                       border-radius: 10px; border: 1px solid #9ae6b4; }
        .result-panel.show { display: block; animation: fadeIn 0.5s ease; }
        @keyframes fadeIn { from { opacity: 0; transform: translateY(10px); } to { opacity: 1; transform: translateY(0); } }

        .grid-2 { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }
        @media (max-width: 768px) { .grid-2 { grid-template-columns: 1fr; } }

        .api-section { background: #f8f9fa; padding: 15px; border-radius: 8px; margin-top: 15px; }
        .api-section h4 { color: #495057; margin-bottom: 10px; }
        pre { background: #282c34; color: #abb2bf; padding: 15px; border-radius: 8px; overflow-x: auto;
              font-size: 13px; line-height: 1.5; }
    </style>
</head>
<body>
<div class="container">
    <h1>XGDN Pay SDK 集成测试</h1>

    <div class="card">
        <h2>1. 创建订单（Go 后端 SDK）</h2>
        <div class="form-row">
            <div class="form-group">
                <label>金额（元）</label>
                <input type="number" id="amount" value="0.01" step="0.01" min="0.01">
            </div>
            <div class="form-group">
                <label>商品名称</label>
                <input type="text" id="title" value="测试商品">
            </div>
            <div class="form-group">
                <label>支付方式</label>
                <select id="payType">
                    <option value="native">扫码支付（native）</option>
                    <option value="jsapi">公众号/小程序（jsapi）</option>
                </select>
            </div>
        </div>
        <div style="display:flex; gap:10px;">
            <button class="btn-primary" onclick="createOrder()">创建订单</button>
        </div>

        <div id="orderResult" class="api-section" style="display:none;">
            <h4>订单信息</h4>
            <pre id="orderInfo"></pre>
        </div>
    </div>

    <div class="card">
        <h2>2. 支付流程（前端 JS SDK）</h2>
        <div class="grid-2">
            <div>
                <h4 style="color:#555;margin-bottom:10px;">二维码支付</h4>
                <div id="qrContainer" class="qr-container">
                    <span style="color:#999;">请先创建订单</span>
                </div>
                <div style="margin-top:15px;display:flex;gap:10px;flex-wrap:wrap;">
                    <button class="btn-info" id="startPayBtn" onclick="startPayment()" disabled>开始支付</button>
                    <button class="btn-warning" id="closePayBtn" onclick="closePayment()" disabled>关闭支付</button>
                </div>
            </div>
            <div>
                <h4 style="color:#555;margin-bottom:10px;">实时日志</h4>
                <div id="logArea" class="log-area">等待操作...\n</div>
            </div>
        </div>

        <div id="payResult" class="result-panel">
            <h3 style="color:#065f46;">支付成功！</h3>
            <pre id="payResultInfo"></pre>
        </div>
    </div>

    <div class="card">
        <h2>3. SDK 使用说明</h2>
        <div class="grid-2">
            <div class="api-section">
                <h4>Go 后端 SDK</h4>
                <pre><code>// 初始化客户端
client := xgdnpay.NewClient(appID, appSecret)

// 创建订单（OutOrderNo 可选，不填自动生成）
resp, err := client.CreateOrder(ctx, &xgdnpay.CreateOrderRequest{
    Amount:  0.01,
    Title:   "测试商品",
    PayType: xgdnpay.PayTypeNative, // 默认 native
})

// 查询订单
order, err := client.QueryOrder(ctx, orderNo)

// 退款（RefundNo 可选，不填自动生成）
refund, err := client.CreateRefund(ctx, &xgdnpay.RefundRequest{
    OrderNo: orderNo,
    Amount:  0.01,
    Reason:  "用户申请退款",
})</code></pre>
            </div>
            <div class="api-section">
                <h4>JavaScript 前端 SDK</h4>
                <pre><code>// CDN 引入
&lt;script src="xgdn-pay.js"&gt;&lt;/script&gt;

// 初始化
const pay = new XGDNPay({
    appID: 'your_app_id',
    baseURL: 'https://pay.xgdn.net'
})

// 监听支付状态
const watcher = XGDNPay.watch(orderNo, {
    baseURL: 'https://pay.xgdn.net',
    onPaid: (data) => console.log('已支付', data),
    onTimeout: () => console.log('超时'),
})

// 一键支付（推荐）
const payment = pay.createPayment(orderNo, {
    container: '#qrcode',
    onPaid: (data) => console.log('已支付', data),
})</code></pre>
            </div>
        </div>
    </div>
</div>

<script src="/xgdn-pay.js"></script>
<script>
let currentOrderNo = null;
let currentPayment = null;
let currentPayInfo = null;

function log(msg, type) {
    const area = document.getElementById('logArea');
    const time = new Date().toLocaleTimeString();
    const cls = type || '';
    area.innerHTML += '<span class="' + cls + '">[' + time + '] ' + msg + '</span>\n';
    area.scrollTop = area.scrollHeight;
}

async function createOrder() {
    const amount = parseFloat(document.getElementById('amount').value);
    const title = document.getElementById('title').value;
    const payType = document.getElementById('payType').value;

    log('正在创建订单...', 'info');

    try {
        const resp = await fetch('/api/order/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ amount, title, payType })
        });
        const result = await resp.json();

        if (result.code === 0) {
            currentOrderNo = result.data.orderNo;
            currentPayInfo = result.data;
            log('订单创建成功: ' + currentOrderNo, 'success');
            
            document.getElementById('orderResult').style.display = 'block';
            document.getElementById('orderInfo').textContent = JSON.stringify(result.data, null, 2);
            
            enablePayButtons();
        } else {
            log('创建失败: ' + result.msg, 'error');
        }
    } catch (err) {
        log('请求异常: ' + err.message, 'error');
    }
}

function enablePayButtons() {
    document.getElementById('startPayBtn').disabled = false;
    document.getElementById('closePayBtn').disabled = false;
}

function startPayment() {
    if (!currentOrderNo) {
        log('请先创建订单', 'warn');
        return;
    }

    closePayment();

    log('开始支付流程，订单号: ' + currentOrderNo, 'info');

    var codeUrl = currentPayInfo.codeUrl;
    
    if (codeUrl) {
        log('使用后端返回的支付信息', 'success');
        log('支付链接: ' + codeUrl, 'info');
        
        var qrContainer = document.querySelector('#qrContainer');
        if (qrContainer) {
            qrContainer.innerHTML = '';
            var qrURL = 'https://api.qrserver.com/v1/create-qr-code/?size=220x220&data=' + encodeURIComponent(codeUrl);
            var img = document.createElement('img');
            img.src = qrURL;
            img.style.width = '220px';
            img.style.height = '220px';
            img.alt = '支付二维码';
            img.onload = function() {
                log('二维码加载完成', 'success');
            };
            img.onerror = function() {
                log('二维码加载失败', 'error');
            };
            qrContainer.appendChild(img);
        }
        
        log('支付准备就绪', 'success');
        log('支付类型: ' + (currentPayInfo.payType || 'native'), 'info');
        log('金额: ' + currentPayInfo.amount, 'info');
    } else {
        log('未获取到支付链接', 'warn');
    }

    currentPayment = XGDNPay.watch(currentOrderNo, {
        baseURL: 'https://pay.xgdn.net',
        timeout: 120000,

        onConnected: function() {
            log('SSE 连接已建立，等待支付...', 'info');
        },

        onPaid: function(data) {
            log('支付成功!', 'success');
            log('交易单号: ' + (data.transaction_id || 'N/A'), 'info');
            
            document.getElementById('payResult').classList.add('show');
            document.getElementById('payResultInfo').textContent = JSON.stringify(data, null, 2);
        },

        onTimeout: function() {
            log('支付超时', 'warn');
        },

        onError: function(err) {
            log('SSE错误: ' + err, 'error');
        }
    });
}

function closePayment() {
    if (currentPayment) {
        currentPayment.close();
        currentPayment = null;
        log('支付流程已关闭', 'info');
    }
    
    var qrContainer = document.querySelector('#qrContainer');
    if (qrContainer) {
        qrContainer.innerHTML = '<span style="color:#999;">请先创建订单</span>';
    }
}
</script>
</body>
</html>
`
