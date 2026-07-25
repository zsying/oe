$AppName = "oe"

function Build-Windows {
    Write-Host "Building for Windows amd64..." -ForegroundColor Cyan
    $env:GOOS = "windows"; $env:GOARCH = "amd64"
    go build -o "$AppName.exe" .
    if ($LASTEXITCODE -ne 0) { Write-Host "Build failed" -ForegroundColor Red; exit 1 }
    Write-Host "Build successful: $AppName.exe" -ForegroundColor Green
}

function Build-Linux {
    Write-Host "Building for Linux amd64..." -ForegroundColor Cyan
    $env:GOOS = "linux"; $env:GOARCH = "amd64"
    go build -o "$AppName-linux-amd64" .
    if ($LASTEXITCODE -ne 0) { Write-Host "Build failed" -ForegroundColor Red; exit 1 }
    Write-Host "Build successful: $AppName-linux-amd64" -ForegroundColor Green
}

function Build-Mac {
    Write-Host "Building for macOS amd64..." -ForegroundColor Cyan
    $env:GOOS = "darwin"; $env:GOARCH = "amd64"
    go build -o "$AppName-darwin-amd64" .
    if ($LASTEXITCODE -ne 0) { Write-Host "Build failed" -ForegroundColor Red; exit 1 }
    Write-Host "Build successful: $AppName-darwin-amd64" -ForegroundColor Green

    Write-Host "Building for macOS arm64..." -ForegroundColor Cyan
    $env:GOOS = "darwin"; $env:GOARCH = "arm64"
    go build -o "$AppName-darwin-arm64" .
    if ($LASTEXITCODE -ne 0) { Write-Host "Build failed" -ForegroundColor Red; exit 1 }
    Write-Host "Build successful: $AppName-darwin-arm64" -ForegroundColor Green
}

function Run-App {
    Write-Host "Building and running $AppName..." -ForegroundColor Cyan
    $env:GOOS = "windows"; $env:GOARCH = "amd64"
    go build -o "$AppName.exe" .
    if ($LASTEXITCODE -ne 0) { Write-Host "Build failed" -ForegroundColor Red; exit 1 }
    Write-Host "Running $AppName..." -ForegroundColor Green
    & ".\$AppName.exe"
}

# Parse args manually
$command = if ($args.Count -ge 1) { $args[0] } else { "" }
$flag = if ($args.Count -ge 2) { $args[1] } else { "" }

switch ($command.ToLower()) {
    "build" {
        switch ($flag.ToLower()) {
            "-linux" { Build-Linux }
            "-mac"   { Build-Mac }
            default  { Build-Windows }
        }
    }
    "" { Run-App }
    default {
        Write-Host "Usage:" -ForegroundColor Yellow
        Write-Host "  .\run.ps1              - Build and run" -ForegroundColor White
        Write-Host "  .\run.ps1 build        - Build for Windows" -ForegroundColor White
        Write-Host "  .\run.ps1 build -linux - Build for Linux amd64" -ForegroundColor White
        Write-Host "  .\run.ps1 build -mac   - Build for macOS" -ForegroundColor White
    }
}
