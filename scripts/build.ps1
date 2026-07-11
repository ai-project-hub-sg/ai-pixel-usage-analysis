param(
  [ValidateSet("test", "web-build", "build", "build-linux", "verify")]
  [string]$Task = "build"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot
$Dist = Join-Path $Root "dist"
New-Item -ItemType Directory -Force $Dist | Out-Null

function Assert-Exit([string]$Name) {
  if ($LASTEXITCODE -ne 0) { throw "$Name failed with exit code $LASTEXITCODE" }
}

function Build-Web {
  Push-Location (Join-Path $Root "web")
  try { npm run build; Assert-Exit "frontend build" } finally { Pop-Location }
}

function Build-Native([bool]$BuildFrontend = $true) {
  if ($BuildFrontend) { Build-Web }
  go build -o (Join-Path $Dist "ai-pixel-usage-analysis.exe") ./cmd/ai-pixel-usage-analysis
  Assert-Exit "Windows build"
}

function Build-Linux([bool]$BuildFrontend = $true) {
  if ($BuildFrontend) { Build-Web }
  $oldCGO, $oldOS, $oldArch = $env:CGO_ENABLED, $env:GOOS, $env:GOARCH
  try {
    $env:CGO_ENABLED = "0"; $env:GOOS = "linux"; $env:GOARCH = "amd64"
    go build -o (Join-Path $Dist "ai-pixel-usage-analysis-linux-amd64") ./cmd/ai-pixel-usage-analysis
    Assert-Exit "Linux build"
  } finally {
    $env:CGO_ENABLED, $env:GOOS, $env:GOARCH = $oldCGO, $oldOS, $oldArch
  }
}

function Test-NativeHealth {
  $temporaryRoot = [IO.Path]::GetFullPath([IO.Path]::GetTempPath())
  $temporary = Join-Path $temporaryRoot ("ai-pixel-smoke-" + [Guid]::NewGuid().ToString("N"))
  New-Item -ItemType Directory -Force $temporary | Out-Null
  $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback, 0)
  $listener.Start()
  $port = ([Net.IPEndPoint]$listener.LocalEndpoint).Port
  $listener.Stop()
  $configPath = Join-Path $temporary "config.toml"
  $envPath = Join-Path $temporary ".env"
  $databasePath = Join-Path $temporary "data/smoke.db"
  $stdoutPath = Join-Path $temporary "stdout.log"
  $stderrPath = Join-Path $temporary "stderr.log"
  [IO.File]::WriteAllText($configPath, @"
[server]
host = "127.0.0.1"
port = $port
public_url = "http://127.0.0.1:$port"
secure_cookie = false

[analysis]
timezone = "Asia/Shanghai"
sync_interval = "1m"
sync_overlap = "5m"
preferred_host_probe_interval = "5m"

[auth]
session_ttl = "24h"

[[host]]
url = "https://smoke.invalid/"
priority = 1

[[account]]
id = "smoke"
name = "Smoke"
email_env = "SMOKE_EMAIL"
password_env = "SMOKE_PASSWORD"
enabled = false
"@, [Text.UTF8Encoding]::new($false))
  [IO.File]::WriteAllText($envPath, "SMOKE_EMAIL=smoke@example.invalid`nSMOKE_PASSWORD=not-used`n", [Text.UTF8Encoding]::new($false))
  $process = $null
  try {
    $binary = Join-Path $Dist "ai-pixel-usage-analysis.exe"
    $process = Start-Process -FilePath $binary -ArgumentList @("-config", $configPath, "-env", $envPath, "-database", $databasePath) -PassThru -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath
    $deadline = [DateTime]::UtcNow.AddSeconds(30)
    do {
      if ($process.HasExited) {
        $details = if (Test-Path $stderrPath) { Get-Content -Raw $stderrPath } else { "no stderr" }
        throw "native health smoke exited early: $details"
      }
      try {
        $response = Invoke-WebRequest -UseBasicParsing -Uri "http://127.0.0.1:$port/health/ready" -TimeoutSec 2
        if ($response.StatusCode -eq 200 -and $response.Content -match 'ready') { return }
      } catch {
        Start-Sleep -Milliseconds 100
      }
    } while ([DateTime]::UtcNow -lt $deadline)
    throw "native health smoke timed out"
  } finally {
    if ($null -ne $process -and -not $process.HasExited) { Stop-Process -Id $process.Id -Force; $process.WaitForExit() }
    $resolved = [IO.Path]::GetFullPath($temporary)
    if (-not $resolved.StartsWith($temporaryRoot, [StringComparison]::OrdinalIgnoreCase)) { throw "refusing to remove non-temporary path $resolved" }
    Remove-Item -LiteralPath $resolved -Recurse -Force
  }
}

switch ($Task) {
  "test" {
    Push-Location (Join-Path $Root "web")
    try { npm run typecheck; Assert-Exit "frontend typecheck"; npm test -- --run; Assert-Exit "frontend tests" } finally { Pop-Location }
    go test ./...; Assert-Exit "Go tests"
  }
  "web-build" { Build-Web }
  "build" { Build-Native }
  "build-linux" { Build-Linux }
  "verify" {
    Push-Location (Join-Path $Root "web")
    try { npm run typecheck; Assert-Exit "frontend typecheck"; npm test -- --run; Assert-Exit "frontend tests"; npm run build; Assert-Exit "frontend build" } finally { Pop-Location }
    go test ./...; Assert-Exit "Go tests"
    Build-Native $false
    Test-NativeHealth
    Write-Host "native health smoke passed"
    go vet ./...; Assert-Exit "Go vet"
    Build-Linux $false
  }
}
