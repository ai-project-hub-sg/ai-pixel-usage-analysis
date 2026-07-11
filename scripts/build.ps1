param(
  [ValidateSet("test", "web-build", "build", "build-linux", "verify")]
  [string]$Task = "build"
)

$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent $PSScriptRoot

function Build-Web {
  Push-Location (Join-Path $Root "web")
  try { npm run build } finally { Pop-Location }
}

switch ($Task) {
  "test" { go test ./... }
  "web-build" { Build-Web }
  "build" {
    Build-Web
    go build -o (Join-Path $Root "dist/ai-pixel-usage-analysis.exe") ./cmd/ai-pixel-usage-analysis
  }
  "build-linux" {
    Build-Web
    $oldCGO, $oldOS, $oldArch = $env:CGO_ENABLED, $env:GOOS, $env:GOARCH
    try {
      $env:CGO_ENABLED = "0"; $env:GOOS = "linux"; $env:GOARCH = "amd64"
      go build -o (Join-Path $Root "dist/ai-pixel-usage-analysis-linux-amd64") ./cmd/ai-pixel-usage-analysis
    } finally {
      $env:CGO_ENABLED, $env:GOOS, $env:GOARCH = $oldCGO, $oldOS, $oldArch
    }
  }
  "verify" {
    Push-Location (Join-Path $Root "web")
    try { npm test -- --run; npm run build } finally { Pop-Location }
    go test ./...
    go vet ./...
  }
}
