<#
.SYNOPSIS
    Remove o pdftoinsert instalado pelo scripts/install.ps1.

.DESCRIPTION
    - Remove a pasta "%USERPROFILE%\.pdftoinsert".
    - Remove essa pasta do PATH do usuário, caso esteja presente.

.EXAMPLE
    ./scripts/uninstall.ps1
#>

[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

$installRoot = Join-Path $env:USERPROFILE ".pdftoinsert"
$installDir = Join-Path $installRoot "bin"

if (Test-Path $installRoot) {
    Write-Host "==> Removendo $installRoot..." -ForegroundColor Cyan
    Remove-Item -Recurse -Force $installRoot
} else {
    Write-Host "==> Nada para remover em $installRoot." -ForegroundColor Yellow
}

$currentUserPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathEntries = $currentUserPath -split ";" | Where-Object { $_ -ne "" -and $_ -ne $installDir }
$newPath = $pathEntries -join ";"

if ($newPath -ne $currentUserPath) {
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "==> '$installDir' removido do PATH do usuário." -ForegroundColor Green
} else {
    Write-Host "==> '$installDir' não estava no PATH do usuário." -ForegroundColor Yellow
}

Write-Host "Desinstalação concluída."
