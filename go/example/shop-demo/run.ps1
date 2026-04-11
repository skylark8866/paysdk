$env:XGDN_APP_ID="your_app_id"
$env:XGDN_APP_SECRET="your_app_secret"
$env:XGDN_BASE_URL="https://pay.xgdn.net"
$env:PORT="8080"

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  XGDN 支付演示商城" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "环境配置:" -ForegroundColor Yellow
Write-Host "  App ID:     $env:XGDN_APP_ID" -ForegroundColor Green
Write-Host "  Base URL:   $env:XGDN_BASE_URL" -ForegroundColor Green
Write-Host "  Port:       $env:PORT" -ForegroundColor Green
Write-Host ""
Write-Host "访问地址: http://localhost:$env:PORT" -ForegroundColor Cyan
Write-Host ""
Write-Host "按 Ctrl+C 停止服务" -ForegroundColor Yellow
Write-Host ""

go run .
