# 启动后端开发服务（热更新）
# 依赖：air (go install github.com/air-verse/air@latest)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$BackendDir = Join-Path $ScriptDir "backend"

Set-Location $BackendDir

# 检查 air 是否安装
if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
    Write-Host "[sp1] air 未安装，正在安装..."
    go install github.com/air-verse/air@latest
}

Write-Host "[sp1] 启动后端（热更新模式）..."
Write-Host "[sp1] 工作目录: $BackendDir"
Write-Host "[sp1] 配置文件: .air.toml"
Write-Host ""

air -c .air.toml
