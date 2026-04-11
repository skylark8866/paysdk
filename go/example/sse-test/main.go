package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const htmlTemplate = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>XGDN Pay JS SDK 测试</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; max-width: 800px; margin: 0 auto; }
        .log { background: #f5f5f5; padding: 10px; margin: 10px 0; border-radius: 4px; font-family: monospace; font-size: 14px; }
        .success { color: green; }
        .error { color: red; }
        .info { color: #333; }
        button { padding: 10px 20px; font-size: 16px; cursor: pointer; margin: 5px; }
        .sdk-info { background: #e8f4f8; padding: 15px; border-radius: 8px; margin: 15px 0; }
        code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>XGDN Pay JS SDK 测试</h1>
    
    <div class="sdk-info">
        <p>后端地址: <code>{{.BackendURL}}</code></p>
        <p>当前页面: <code id="origin"></code></p>
        <p>SDK 版本: <code>1.0.0</code></p>
    </div>

    <div>
        <button onclick="testSDK()">SDK 方式测试</button>
        <button onclick="testRawSSE()">原生 SSE 测试</button>
        <button onclick="testReconnect()">重连测试（模拟断线）</button>
        <button onclick="clearLog()">清空日志</button>
    </div>

    <div id="logs"></div>

    <script src="/xgdn-pay.js"></script>
    <script>
        document.getElementById('origin').textContent = window.location.origin;
        
        var baseURL = '{{.BackendURL}}';
        var currentWatcher = null;

        function log(msg, type) {
            type = type || 'info';
            var div = document.createElement('div');
            div.className = 'log ' + type;
            div.textContent = '[' + new Date().toLocaleTimeString() + '] ' + msg;
            document.getElementById('logs').prepend(div);
        }

        function clearLog() {
            document.getElementById('logs').innerHTML = '';
        }

        function testSDK() {
            if (currentWatcher) {
                currentWatcher.close();
            }

            var orderNo = 'TEST_' + Date.now();
            log('SDK.watch() 订单号: ' + orderNo);

            currentWatcher = XGDNPay.watch(orderNo, {
                baseURL: baseURL,
                timeout: 60000,
                onConnected: function () {
                    log('✅ SDK: SSE 连接已建立', 'success');
                },
                onPaid: function (data) {
                    log('✅ SDK: 支付成功! ' + JSON.stringify(data), 'success');
                },
                onTimeout: function () {
                    log('⏰ SDK: 支付超时', 'error');
                },
                onError: function (err) {
                    log('❌ SDK: 错误 ' + err, 'error');
                }
            });

            log('watcher 对象已创建，可调用 watcher.close() 关闭');
        }

        function testRawSSE() {
            var orderNo = 'TEST_' + Date.now();
            var url = baseURL + '/api/v1/sse/subscribe/' + orderNo;
            log('原生 SSE 连接: ' + url);

            try {
                var es = new EventSource(url);
                es.onopen = function () { log('✅ 原生: 连接已打开', 'success'); };
                es.addEventListener('connected', function (e) { log('✅ 原生: connected = ' + e.data, 'success'); });
                es.addEventListener('message', function (e) { log('📨 原生: message = ' + e.data, 'info'); });
                es.onerror = function () { log('❌ 原生: 连接错误', 'error'); es.close(); };
            } catch (err) {
                log('❌ 原生: 创建失败 ' + err.message, 'error');
            }
        }

        function testReconnect() {
            if (currentWatcher) {
                currentWatcher.close();
            }

            var orderNo = 'TEST_' + Date.now();
            log('重连测试: 订单号 ' + orderNo + '（5秒后手动断开，观察重连）');

            currentWatcher = XGDNPay.watch(orderNo, {
                baseURL: baseURL,
                timeout: 120000,
                onConnected: function () {
                    log('✅ 重连测试: SSE 连接已建立', 'success');
                },
                onPaid: function (data) {
                    log('✅ 重连测试: 支付成功! ' + JSON.stringify(data), 'success');
                },
                onTimeout: function () {
                    log('⏰ 重连测试: 超时', 'error');
                },
                onError: function (err) {
                    log('❌ 重连测试: 错误 ' + err, 'error');
                }
            });

            setTimeout(function () {
                if (currentWatcher && currentWatcher._eventSource) {
                    log('🔧 模拟网络断线（触发 onerror），观察自动重连...', 'info');
                    var es = currentWatcher._eventSource;
                    if (es.onerror) {
                        es.onerror(new Event('error'));
                    }
                }
            }, 5000);
        }
    </script>
</body>
</html>`

func main() {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "https://pay.xgdn.net"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	tmpl := template.Must(template.New("index").Parse(htmlTemplate))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"BackendURL": backendURL,
		}
		tmpl.Execute(w, data)
	})

	jsPath := os.Getenv("JS_SDK_PATH")
	if jsPath == "" {
		cwd, _ := os.Getwd()
		jsPath = filepath.Join(cwd, "..", "..", "..", "js", "xgdn-pay.js")
	}
	jsPath = strings.ReplaceAll(jsPath, `\`, "/")

	jsContent, err := os.ReadFile(jsPath)
	if err != nil {
		fmt.Printf("警告: 无法读取 JS SDK 文件: %v\n", err)
	}

	http.HandleFunc("/xgdn-pay.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(jsContent)
	})

	fmt.Printf("SSE 测试服务启动在 http://localhost:%s\n", port)
	fmt.Printf("测试后端: %s\n", backendURL)
	fmt.Printf("JS SDK 路径: %s\n", jsPath)
	http.ListenAndServe(":"+port, nil)
}
