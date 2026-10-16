param(
    [ValidateSet('all', 'windows', 'linux', 'macos-amd64', 'macos-arm64')]
    [string]$Target = 'all',
    [string]$OutputDirectory = 'build/dist'
)

$ErrorActionPreference = 'Stop'

if ($PSVersionTable.PSVersion.Major -lt 7) {
    throw 'PowerShell 7 veya üzeri gereklidir.'
}

$projectRoot = Split-Path -Parent $PSScriptRoot
if ([System.IO.Path]::IsPathRooted($OutputDirectory)) {
    $outputRoot = $OutputDirectory
} else {
    $outputRoot = Join-Path $projectRoot $OutputDirectory
}

$targets = @(
    [pscustomobject]@{
        Id = 'windows'
        Name = 'windows-amd64'
        GOOS = 'windows'
        GOARCH = 'amd64'
        Extension = '.exe'
    },
    [pscustomobject]@{
        Id = 'linux'
        Name = 'linux-amd64'
        GOOS = 'linux'
        GOARCH = 'amd64'
        Extension = ''
    },
    [pscustomobject]@{
        Id = 'macos-amd64'
        Name = 'macos-amd64'
        GOOS = 'darwin'
        GOARCH = 'amd64'
        Extension = ''
    },
    [pscustomobject]@{
        Id = 'macos-arm64'
        Name = 'macos-arm64'
        GOOS = 'darwin'
        GOARCH = 'arm64'
        Extension = ''
    }
)

if ($Target -eq 'all') {
    $selectedTargets = $targets
} else {
    $selectedTargets = @($targets | Where-Object { $_.Id -eq $Target })
}

$null = Get-Command go -ErrorAction Stop
$originalGOOS = $env:GOOS
$originalGOARCH = $env:GOARCH
$originalCGOEnabled = $env:CGO_ENABLED
$assetsCopied = $false

try {
    Push-Location $projectRoot
    New-Item -ItemType Directory -Path $outputRoot -Force | Out-Null

    foreach ($buildTarget in $selectedTargets) {
        $packageRoot = Join-Path $outputRoot $buildTarget.Name
        New-Item -ItemType Directory -Path $packageRoot -Force | Out-Null

        $env:GOOS = $buildTarget.GOOS
        $env:GOARCH = $buildTarget.GOARCH
        $env:CGO_ENABLED = '0'

        $binaryPath = Join-Path $packageRoot "mapp-game$($buildTarget.Extension)"
        Write-Host "Derleniyor: $($buildTarget.Name)"
        & go build -trimpath -o $binaryPath ./cmd/game
        if ($LASTEXITCODE -ne 0) {
            throw "Go derlemesi başarısız oldu: $($buildTarget.Name) (çıkış kodu: $LASTEXITCODE)"
        }

        if (-not $assetsCopied) {
            Copy-Item -Path (Join-Path $projectRoot 'assets') -Destination $outputRoot -Recurse -Force
            $assetsCopied = $true
        }
        Write-Host "Derleme hazır: $binaryPath"
    }
    Write-Host "Çalıştırılabilir dosyalar ve ortak assets klasörü: $outputRoot"
} finally {
    if ((Get-Location).Path -eq $projectRoot) {
        Pop-Location
    }
    $env:GOOS = $originalGOOS
    $env:GOARCH = $originalGOARCH
    $env:CGO_ENABLED = $originalCGOEnabled
}
