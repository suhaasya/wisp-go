# Install wisp to %LOCALAPPDATA%\Programs\wisp (override with $Env:WISP_INSTALL_DIR).
$ErrorActionPreference = "Stop"

$Repo = if ($Env:WISP_INSTALL_REPO) { $Env:WISP_INSTALL_REPO } else { "suhaasya/wisp-go" }
$Version = if ($Env:WISP_VERSION) { $Env:WISP_VERSION } else { "latest" }
$Base = if ($Env:WISP_INSTALL_BASE) { $Env:WISP_INSTALL_BASE } else { "https://github.com" }
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$PubKey = if ($Env:WISP_PUBLIC_KEY) { $Env:WISP_PUBLIC_KEY } else { Join-Path $ScriptDir "..\release\minisign.pub" }

function Require($cmd) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
    throw "install.ps1: missing required command: $cmd"
  }
}
Require curl
Require minisign

$arch = "amd64"
if ($Version -eq "latest") {
  $rel = curl -fsSL "https://api.github.com/repos/$Repo/releases/latest" | ConvertFrom-Json
  $tag = $rel.tag_name
} else {
  $tag = $Version
}
$ver = $tag -replace '^v', ''
$archive = "wisp_${ver}_windows_${arch}.zip"
$baseUrl = "$Base/$Repo/releases/download/$tag"

$tmpdir = New-Item -ItemType Directory -Path (Join-Path $env:TEMP ([guid]::NewGuid().ToString()))
try {
  curl -fsSL -o (Join-Path $tmpdir "checksums.txt") "$baseUrl/checksums.txt"
  curl -fsSL -o (Join-Path $tmpdir "checksums.txt.minisig") "$baseUrl/checksums.txt.minisig"
  & minisign -Vqm (Join-Path $tmpdir "checksums.txt") -x (Join-Path $tmpdir "checksums.txt.minisig") -p $PubKey

  $archivePath = Join-Path $tmpdir $archive
  curl -fsSL -o $archivePath "$baseUrl/$archive"

  $checksums = Get-Content (Join-Path $tmpdir "checksums.txt")
  $line = $checksums | Where-Object { $_ -match "  $([regex]::Escape($archive))$" } | Select-Object -First 1
  if (-not $line) { throw "install.ps1: no checksum line for $archive" }
  $want = ($line -split '\s+')[0]
  $got = (Get-FileHash $archivePath -Algorithm SHA256).Hash.ToLower()
  if ($want -ne $got) { throw "install.ps1: checksum mismatch" }

  Expand-Archive -Path $archivePath -DestinationPath $tmpdir -Force
  $bin = Join-Path $tmpdir "wisp.exe"

  $installDir = if ($Env:WISP_INSTALL_DIR) { $Env:WISP_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA "Programs\wisp" }
  New-Item -ItemType Directory -Force -Path $installDir | Out-Null
  Copy-Item $bin (Join-Path $installDir "wisp.exe") -Force
  Write-Host "Installed wisp $ver to $installDir\wisp.exe"
} finally {
  Remove-Item -Recurse -Force $tmpdir
}
