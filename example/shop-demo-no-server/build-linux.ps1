$ErrorActionPreference = "Stop"

$projectDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $projectDir

Write-Host "Building for Linux AMD64..." -ForegroundColor Cyan

$env:GOOS = "linux"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "0"

$output = "shop-demo-server"

go build -ldflags="-s -w" -o $output .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful: $output" -ForegroundColor Green
    
    $fileInfo = Get-Item $output -ErrorAction SilentlyContinue
    if ($fileInfo) {
        $sizeMB = [math]::Round($fileInfo.Length / 1MB, 2)
        Write-Host "File size: $sizeMB MB" -ForegroundColor Gray
    }
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}
