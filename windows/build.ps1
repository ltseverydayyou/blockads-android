$ErrorActionPreference = "Stop"

$repo = Split-Path -Parent $PSScriptRoot
$core = Join-Path $PSScriptRoot "core"
$desktop = Join-Path $PSScriptRoot "desktop"
$backendResource = Join-Path $desktop "src\main\resources\backend"
$dist = Join-Path $PSScriptRoot "dist"
$go = "C:\Program Files\Go\bin\go.exe"
if (-not (Test-Path $go)) { $go = "go" }
$gradle = Join-Path $repo "gradlew.bat"

New-Item -ItemType Directory -Force $backendResource | Out-Null
New-Item -ItemType Directory -Force $dist | Out-Null
Get-ChildItem $dist -File -ErrorAction SilentlyContinue | Remove-Item -Force

Push-Location $core
& $go test ./...
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
& $go build -trimpath -ldflags "-s -w -H=windowsgui" -o (Join-Path $backendResource "BlockAdsCore.exe") .\cmd\blockadscore
if ($LASTEXITCODE -ne 0) { Pop-Location; exit $LASTEXITCODE }
Pop-Location

& $gradle -p $desktop clean createDistributable packageExe
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$installer = Get-ChildItem (Join-Path $desktop "build\compose\binaries\main\exe") -Filter "*.exe" | Select-Object -First 1
if (-not $installer) { throw "Compose installer was not produced" }
$setupOut = Join-Path $dist "BlockAds-Windows-Setup-v1.3.exe"
Copy-Item $installer.FullName $setupOut -Force

$appDir = Join-Path $desktop "build\compose\binaries\main\app\BlockAds"
if (-not (Test-Path $appDir)) { throw "Compose portable app image was not produced" }
$portableOut = Join-Path $dist "BlockAds-Windows-Portable-v1.3.zip"
Compress-Archive -Path (Join-Path $appDir "*") -DestinationPath $portableOut -CompressionLevel Optimal -Force

Write-Host "Built:"
Get-Item $setupOut, $portableOut | Select-Object Name, Length, LastWriteTime
Write-Host "SHA256:"
Get-FileHash $setupOut, $portableOut -Algorithm SHA256 | Select-Object Path, Hash
