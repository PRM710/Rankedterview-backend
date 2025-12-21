# Quick fix script for unused imports
Write-Host "Fixing unused imports..." -ForegroundColor Yellow

# Run go mod tidy
go mod tidy

# Try to build and capture errors
$errors = go build cmd/server/main.go 2>&1

# If errors contain "imported and not used", list them
if ($errors -match "imported and not used") {
    Write-Host "Found unused imports. Running go fmt..." -ForegroundColor Yellow
    go fmt ./...
}

Write-Host "Build attempt:" -ForegroundColor Cyan
go run cmd/server/main.go
