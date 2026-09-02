$ErrorActionPreference = "Stop"
$repo = Split-Path -Parent $PSScriptRoot
$core = Join-Path $PSScriptRoot "core"
$cmd = Join-Path $core "cmd\blockads"
$dist = Join-Path $PSScriptRoot "dist"
$go = "C:\Program Files\Go\bin\go.exe"
if (-not (Test-Path $go)) { $go = "go" }
New-Item -ItemType Directory -Force $dist | Out-Null
Push-Location $cmd
& $go run github.com/akavel/rsrc@v0.10.2 -manifest blockads.manifest -o rsrc_windows.syso
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Pop-Location
Push-Location $core
& $go test ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
& $go build -ldflags "-H=windowsgui -s -w" -o (Join-Path $dist "BlockAds.exe") .\cmd\blockads
$code = $LASTEXITCODE
Pop-Location
exit $code
