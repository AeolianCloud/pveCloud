# 启动前台管理后台开发服务
# 依赖：Node.js + npm

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$FrontendDir = Join-Path $ScriptDir "frontend-admin"

Set-Location $FrontendDir

Write-Host "[sp2] 启动前端管理后台（开发模式）..."
Write-Host "[sp2] 工作目录: $FrontendDir"
Write-Host "[sp2] 访问地址: http://localhost:5173"
Write-Host ""

npm run dev
