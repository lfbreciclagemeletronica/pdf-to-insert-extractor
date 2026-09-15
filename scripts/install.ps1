<#
.SYNOPSIS
    Compila e instala o pdftoinsert no computador, deixando o comando
    "pdftoinsert" disponível em qualquer terminal (PowerShell/CMD).

.DESCRIPTION
    - Compila o binário a partir do código-fonte deste repositório.
    - Copia o binário para "%USERPROFILE%\.pdftoinsert\bin".
    - Adiciona essa pasta ao PATH do usuário (persistente), caso ainda não esteja lá.

.EXAMPLE
    ./scripts/install.ps1
#>

[CmdletBinding()]
param()

$ErrorActionPreference = "Stop"

# Raiz do projeto (um nível acima da pasta scripts/)
$repoRoot = Split-Path -Parent $PSScriptRoot

# Verifica se o Go está instalado
$goCmd = Get-Command go -ErrorAction SilentlyContinue
if (-not $goCmd) {
    Write-Error "Go not found in PATH. Instale o Go (https://go.dev/dl/) antes de continuar."
    exit 1
}

# Diretório e arquivo de instalação
$installDir = Join-Path $env:USERPROFILE ".pdftoinsert\bin"
$binaryName = "pdftoinsert.exe"
$binaryPath = Join-Path $installDir $binaryName

Write-Host "==> Compilando pdftoinsert..." -ForegroundColor Cyan
Push-Location $repoRoot
try {
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    }

    go build -o $binaryPath .
    if ($LASTEXITCODE -ne 0) {
        throw "Falha ao compilar o projeto."
    }
}
finally {
    Pop-Location
}

Write-Host "==> Projeto instalado em: $binaryPath" -ForegroundColor Green

# Adiciona o diretório de instalação ao PATH do usuário, se necessário
$currentUserPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathEntries = $currentUserPath -split ";" | Where-Object { $_ -ne "" }

if ($pathEntries -notcontains $installDir) {
    Write-Host "==> Adicionando '$installDir' ao PATH do user..." -ForegroundColor Cyan
    $newPath = if ([string]::IsNullOrEmpty($currentUserPath)) {
        $installDir
    } else {
        "$currentUserPath;$installDir"
    }
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")

    # Atualiza o PATH da sessão atual também
    $env:Path = "$env:Path;$installDir"

    Write-Host "==> PATH atualizado. Abra um novo terminal para que tenha efeito." -ForegroundColor Yellow
} else {
    Write-Host "==> '$installDir' ja colocado  no PATH do user." -ForegroundColor Green
}

Write-Host ""
Write-Host "CONCLUIDO! Exemplo de uso:" -ForegroundColor Green
Write-Host "  pdftoinsert -f arquivo.pdf -o output.sql"
