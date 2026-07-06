$ErrorActionPreference = "Stop"

$Repo = "rodionlim/portfolio-manager-go"
$InstallDir = if ($env:PORTFOLIO_MANAGER_HOME) { $env:PORTFOLIO_MANAGER_HOME } else { Join-Path $HOME "portfolio-manager" }
$Binary = Join-Path $InstallDir "portfolio-manager.exe"
$Url = "https://github.com/$Repo/releases/latest/download/portfolio-manager.exe"

Write-Host "Installing Portfolio Manager into $InstallDir"
New-Item -ItemType Directory -Force $InstallDir | Out-Null
Invoke-WebRequest -Uri $Url -OutFile $Binary
Unblock-File $Binary

Write-Host ""
Write-Host "Portfolio Manager installed:"
Write-Host "  $Binary"
Write-Host ""
Write-Host "Start it with:"
Write-Host "  cd `"$InstallDir`""
Write-Host "  .\portfolio-manager.exe"
Write-Host ""
Write-Host "Default URLs:"
Write-Host "  Backend: http://localhost:8080"
Write-Host "  MCP:     http://localhost:8081/mcp"
