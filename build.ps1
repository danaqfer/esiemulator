# Edge Computing Emulator Suite Build Script
# PowerShell script for building and running the emulator suite

param(
    [string]$Command = "help",
    [string]$Mode = "esi",
    [string]$ESIMode = "akamai",
    [int]$Port = 3000,
    [switch]$Debug
)

$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
$BuildDir = Join-Path $ProjectRoot "bin"
$MainPath = Join-Path $ProjectRoot "cmd/edge-emulator/main.go"

# Create build directory if it doesn't exist
if (!(Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir | Out-Null
}

function Show-Help {
    Write-Host "Edge Computing Emulator Suite Build Script" -ForegroundColor Green
    Write-Host "=============================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Available Commands:" -ForegroundColor Yellow
    Write-Host "  build         - Build the application"
    Write-Host "  test          - Run all tests"
    Write-Host "  clean         - Clean build artifacts"
    Write-Host "  run           - Run ESI emulator (Akamai mode)"
    Write-Host "  run-fastly    - Run ESI emulator (Fastly mode)"
    Write-Host "  run-w3c       - Run ESI emulator (W3C mode)"
    Write-Host "  run-property-manager - Run Property Manager emulator"
    Write-Host "  examples      - Run example programs"
    Write-Host "  lint          - Run linter checks"
    Write-Host "  format        - Format code"
    Write-Host "  coverage      - Run tests with coverage"
    Write-Host "  help          - Show this help"
    Write-Host ""
    Write-Host "Environment Variables:" -ForegroundColor Yellow
    Write-Host "  EMULATOR_MODE - Set to 'esi' or 'property-manager'"
    Write-Host "  ESI_MODE      - Set to 'fastly', 'akamai', 'w3c', or 'development'"
    Write-Host "  PORT          - Server port (default: 3000)"
    Write-Host "  DEBUG         - Enable debug mode"
    Write-Host ""
    Write-Host "Examples:" -ForegroundColor Yellow
    Write-Host "  .\build.ps1 build"
    Write-Host "  .\build.ps1 run"
    Write-Host "  .\build.ps1 run-property-manager"
    Write-Host "  $env:DEBUG='true'; .\build.ps1 run"
}

function Build-Application {
    Write-Host "Building Edge Computing Emulator Suite..." -ForegroundColor Green
    
    $BuildFlags = @()
    if ($Debug) {
        $BuildFlags += "-ldflags=-X main.debug=true"
    }
    
    $OutputPath = Join-Path $BuildDir "edge-emulator.exe"
    
    Push-Location $ProjectRoot
    try {
        $BuildCmd = "go build -o `"$OutputPath`" $($BuildFlags -join ' ') `"$MainPath`""
        Write-Host "Executing: $BuildCmd" -ForegroundColor Gray
        Invoke-Expression $BuildCmd
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "Build successful! Binary created at: $OutputPath" -ForegroundColor Green
        } else {
            Write-Host "Build failed!" -ForegroundColor Red
            exit 1
        }
    }
    finally {
        Pop-Location
    }
}

function Test-Application {
    Write-Host "Running tests..." -ForegroundColor Green
    
    Push-Location $ProjectRoot
    try {
        $TestCmd = "go test -v ./..."
        Write-Host "Executing: $TestCmd" -ForegroundColor Gray
        Invoke-Expression $TestCmd
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "Tests failed!" -ForegroundColor Red
            exit 1
        }
    }
    finally {
        Pop-Location
    }
}

function Clean-Build {
    Write-Host "Cleaning build artifacts..." -ForegroundColor Green
    
    if (Test-Path $BuildDir) {
        Remove-Item -Path $BuildDir -Recurse -Force
        Write-Host "Build directory cleaned." -ForegroundColor Green
    }
    
    # Clean test artifacts
    $TestFiles = Get-ChildItem -Path $ProjectRoot -Recurse -Include "*.test", "*.out", "coverage.html"
    if ($TestFiles) {
        $TestFiles | Remove-Item -Force
        Write-Host "Test artifacts cleaned." -ForegroundColor Green
    }
}

function Run-ESIEmulator {
    param([string]$ESIMode = "akamai")
    
    Write-Host "Running ESI Emulator in $ESIMode mode..." -ForegroundColor Green
    
    $EnvVars = @{
        "EMULATOR_MODE" = "esi"
        "ESI_MODE" = $ESIMode
        "PORT" = $Port
    }
    
    if ($Debug) {
        $EnvVars["DEBUG"] = "true"
    }
    
    Push-Location $ProjectRoot
    try {
        $RunCmd = "go run `"$MainPath`" -mode=esi -esi-mode=$ESIMode -port=$Port"
        if ($Debug) {
            $RunCmd += " -debug"
        }
        
        Write-Host "Executing: $RunCmd" -ForegroundColor Gray
        Write-Host "Environment: $($EnvVars | ConvertTo-Json)" -ForegroundColor Gray
        
        # Set environment variables
        foreach ($key in $EnvVars.Keys) {
            Set-Item -Path "env:$key" -Value $EnvVars[$key]
        }
        
        Invoke-Expression $RunCmd
    }
    finally {
        Pop-Location
    }
}

function Run-PropertyManagerEmulator {
    Write-Host "Running Property Manager Emulator..." -ForegroundColor Green
    
    $EnvVars = @{
        "EMULATOR_MODE" = "property-manager"
        "PORT" = $Port
    }
    
    if ($Debug) {
        $EnvVars["DEBUG"] = "true"
    }
    
    Push-Location $ProjectRoot
    try {
        $RunCmd = "go run `"$MainPath`" -mode=property-manager -port=$Port"
        if ($Debug) {
            $RunCmd += " -debug"
        }
        
        Write-Host "Executing: $RunCmd" -ForegroundColor Gray
        Write-Host "Environment: $($EnvVars | ConvertTo-Json)" -ForegroundColor Gray
        
        # Set environment variables
        foreach ($key in $EnvVars.Keys) {
            Set-Item -Path "env:$key" -Value $EnvVars[$key]
        }
        
        Invoke-Expression $RunCmd
    }
    finally {
        Pop-Location
    }
}

function Run-Examples {
    Write-Host "Running examples..." -ForegroundColor Green
    
    Push-Location $ProjectRoot
    try {
        $ExamplesDir = Join-Path $ProjectRoot "cmd/examples"
        if (Test-Path $ExamplesDir) {
            Get-ChildItem -Path $ExamplesDir -Filter "*.go" | ForEach-Object {
                Write-Host "Running example: $($_.Name)" -ForegroundColor Yellow
                go run $_.FullName
            }
        } else {
            Write-Host "No examples found in $ExamplesDir" -ForegroundColor Yellow
        }
    }
    finally {
        Pop-Location
    }
}

function Invoke-Lint {
    Write-Host "Running linter checks..." -ForegroundColor Green
    
    Push-Location $ProjectRoot
    try {
        # Check if golangci-lint is installed
        $LintCmd = "golangci-lint run"
        try {
            Invoke-Expression $LintCmd
        }
        catch {
            Write-Host "golangci-lint not found, running go vet instead..." -ForegroundColor Yellow
            go vet ./...
        }
    }
    finally {
        Pop-Location
    }
}

function Format-Code {
    Write-Host "Formatting code..." -ForegroundColor Green
    
    Push-Location $ProjectRoot
    try {
        go fmt ./...
        Write-Host "Code formatting complete." -ForegroundColor Green
    }
    finally {
        Pop-Location
    }
}

function Test-Coverage {
    Write-Host "Running tests with coverage..." -ForegroundColor Green
    
    Push-Location $ProjectRoot
    try {
        $CoverageFile = Join-Path $ProjectRoot "coverage.out"
        $CoverageHTML = Join-Path $ProjectRoot "coverage.html"
        
        go test -v -coverprofile=$CoverageFile ./...
        
        if (Test-Path $CoverageFile) {
            go tool cover -html=$CoverageFile -o=$CoverageHTML
            Write-Host "Coverage report generated: $CoverageHTML" -ForegroundColor Green
        }
    }
    finally {
        Pop-Location
    }
}

# Main execution
switch ($Command.ToLower()) {
    "build" { Build-Application }
    "test" { Test-Application }
    "clean" { Clean-Build }
    "run" { Run-ESIEmulator -ESIMode $ESIMode }
    "run-fastly" { Run-ESIEmulator -ESIMode "fastly" }
    "run-w3c" { Run-ESIEmulator -ESIMode "w3c" }
    "run-property-manager" { Run-PropertyManagerEmulator }
    "examples" { Run-Examples }
    "lint" { Invoke-Lint }
    "format" { Format-Code }
    "coverage" { Test-Coverage }
    "help" { Show-Help }
    default {
        Write-Host "Unknown command: $Command" -ForegroundColor Red
        Show-Help
        exit 1
    }
} 