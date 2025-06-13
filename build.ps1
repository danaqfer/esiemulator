# ESI Emulator PowerShell Build Script
param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

# Variables
$BinaryName = "esi-emulator"
$BuildDir = "build"
$MainFile = "main.go"
$ExamplesFile = "cmd/examples/main.go"

# Ensure Go workspaces don't interfere
$env:GOWORK = "off"

function Build {
    Write-Host "🔨 Building ESI Emulator..." -ForegroundColor Green
    if (-not (Test-Path $BuildDir)) {
        New-Item -ItemType Directory -Path $BuildDir -Force | Out-Null
    }
    go build -o "$BuildDir\$BinaryName.exe" $MainFile
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Build complete: $BuildDir\$BinaryName.exe" -ForegroundColor Green
    } else {
        Write-Host "❌ Build failed" -ForegroundColor Red
    }
}

function Run {
    Write-Host "🚀 Running ESI Emulator in development mode..." -ForegroundColor Green
    go run $MainFile -mode development -debug
}

function RunFastly {
    Write-Host "🚀 Running ESI Emulator in Fastly mode..." -ForegroundColor Green
    go run $MainFile -mode fastly -debug
}

function RunAkamai {
    Write-Host "🚀 Running ESI Emulator in Akamai mode..." -ForegroundColor Green
    go run $MainFile -mode akamai -debug
}

function RunW3C {
    Write-Host "🚀 Running ESI Emulator in W3C mode..." -ForegroundColor Green
    go run $MainFile -mode w3c -debug
}

function Examples {
    Write-Host "📚 Running examples..." -ForegroundColor Green
    go run $ExamplesFile
}

function Deps {
    Write-Host "📦 Installing dependencies..." -ForegroundColor Green
    go mod tidy
    go mod download
}

function Test {
    Write-Host "🧪 Running tests..." -ForegroundColor Green
    go test -v ./...
}

function TestCoverage {
    Write-Host "🧪 Running tests with coverage..." -ForegroundColor Green
    go test -v -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    Write-Host "📊 Coverage report generated: coverage.html" -ForegroundColor Green
}

function Format {
    Write-Host "✨ Formatting code..." -ForegroundColor Green
    go fmt ./...
}

function Clean {
    Write-Host "🧹 Cleaning build artifacts..." -ForegroundColor Green
    if (Test-Path $BuildDir) {
        Remove-Item -Recurse -Force $BuildDir
    }
    if (Test-Path "coverage.out") {
        Remove-Item "coverage.out"
    }
    if (Test-Path "coverage.html") {
        Remove-Item "coverage.html"
    }
}

function RunBinary {
    Build
    if ($LASTEXITCODE -eq 0) {
        Write-Host "🚀 Running built binary..." -ForegroundColor Green
        & ".\$BuildDir\$BinaryName.exe"
    }
}

function ShowHelp {
    Write-Host "ESI Emulator - Available Commands:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  build          Build the application" -ForegroundColor White
    Write-Host "  run            Run in development mode" -ForegroundColor White
    Write-Host "  run-fastly     Run in Fastly mode" -ForegroundColor White
    Write-Host "  run-akamai     Run in Akamai mode" -ForegroundColor White
    Write-Host "  run-w3c        Run in W3C mode" -ForegroundColor White
    Write-Host "  examples       Run example programs" -ForegroundColor White
    Write-Host "  deps           Install dependencies" -ForegroundColor White
    Write-Host "  test           Run tests" -ForegroundColor White
    Write-Host "  test-coverage  Run tests with coverage" -ForegroundColor White
    Write-Host "  format         Format code" -ForegroundColor White
    Write-Host "  clean          Clean build artifacts" -ForegroundColor White
    Write-Host "  run-binary     Build and run binary" -ForegroundColor White
    Write-Host "  help           Show this help" -ForegroundColor White
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 build" -ForegroundColor Gray
    Write-Host "  .\build.ps1 run-fastly" -ForegroundColor Gray
    Write-Host "  .\build.ps1 examples" -ForegroundColor Gray
    Write-Host "  `$env:ESI_MODE='akamai'; `$env:PORT='8080'; .\build.ps1 run" -ForegroundColor Gray
}

# Execute command
switch ($Command.ToLower()) {
    "build" { Build }
    "run" { Run }
    "run-fastly" { RunFastly }
    "run-akamai" { RunAkamai }
    "run-w3c" { RunW3C }
    "examples" { Examples }
    "deps" { Deps }
    "test" { Test }
    "test-coverage" { TestCoverage }
    "format" { Format }
    "clean" { Clean }
    "run-binary" { RunBinary }
    "help" { ShowHelp }
    default { 
        Write-Host "❌ Unknown command: $Command" -ForegroundColor Red
        Write-Host "Run '.\build.ps1 help' for available commands" -ForegroundColor Yellow
    }
} 