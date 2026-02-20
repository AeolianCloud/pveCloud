# 后端 Air 自动热更启动脚本（PowerShell）
# 用法：
#   在 backend 目录执行：.\start-air.ps1

$ErrorActionPreference = "Stop"

# 固定 Air 版本，避免团队环境差异导致行为不一致。
$airVersion = "v1.61.7"

# 检查 air 是否已安装；未安装则自动安装到 GOPATH/bin。
if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
    Write-Host "[air] 未检测到 air，开始安装 github.com/air-verse/air@$airVersion ..."
    go install "github.com/air-verse/air@$airVersion"
}

Write-Host "[air] 使用配置 .air.toml 启动后端自动热更..."
air -c .air.toml
