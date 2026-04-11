#!/bin/bash

export XGDN_APP_ID="your_app_id"
export XGDN_APP_SECRET="your_app_secret"
export XGDN_BASE_URL="https://pay.xgdn.net"
export PORT="8080"

echo "======================================"
echo "  XGDN 支付演示商城"
echo "======================================"
echo ""
echo "环境配置:"
echo "  App ID:     $XGDN_APP_ID"
echo "  Base URL:   $XGDN_BASE_URL"
echo "  Port:       $PORT"
echo ""
echo "访问地址: http://localhost:$PORT"
echo ""
echo "按 Ctrl+C 停止服务"
echo ""

go run .
