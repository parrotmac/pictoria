# Pictoria photo upload application

# Default command - show available commands
default:
    @just --list

# Development
# ----------

# Run the application locally
run:
    go run main.go

# Run in development mode with auto-reload (requires air)
dev:
    air || (echo "Installing air..." && go install github.com/cosmtrek/air@latest && air)

# Build the application
build:
    go build -o pictoria main.go

# Run tests
test:
    go test ./...

# Format code
fmt:
    go fmt ./...

# Run linter
lint:
    golangci-lint run || go vet ./...

# Docker
# ------

# Build Docker image
docker-build:
    docker compose build

# Run with Docker
docker-up:
    docker compose up -d
    @echo "Pictoria is running at http://localhost:8080"
    @echo "View logs: just docker-logs"

# Stop Docker containers
docker-down:
    docker compose down

# View Docker logs
docker-logs:
    docker compose logs -f

# Restart Docker containers
docker-restart:
    docker compose restart

# Remove Docker volumes (WARNING: deletes all data)
docker-clean:
    docker compose down -v

# Utilities
# ---------

# Open in browser
open:
    @open http://localhost:8080 || xdg-open http://localhost:8080 || echo "Please open http://localhost:8080 in your browser"

# Check if server is running
health:
    @curl -s http://localhost:8080/api/photos > /dev/null && echo "✓ Server is running" || echo "✗ Server is not running"

# Show disk usage
disk-usage:
    @echo "Uploads: $$(du -sh uploads/ 2>/dev/null || echo '0B')"
    @echo "Photos: $$(ls uploads/ 2>/dev/null | wc -l | tr -d ' ') files"

# Maintenance
# -----------

# Setup directories
setup:
    mkdir -p uploads static backups
    @echo "✓ Directories created"

# Install dependencies
deps:
    go mod download
    go mod tidy

# Clean build artifacts
clean:
    rm -f pictoria
    rm -f coverage.*

# Reset everything (WARNING: deletes all data)
reset: clean
    rm -f data.json
    rm -rf uploads/*
    @echo "✓ Application reset complete"

# Backup data and uploads
backup:
    @mkdir -p backups
    @tar -czf backups/backup-$(date +%Y%m%d-%H%M%S).tar.gz data.json uploads/
    @echo "✓ Backup created in backups/"

# Deployment
# ----------

# Build for production
build-prod:
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pictoria .

# Build for multiple platforms
build-all:
    @mkdir -p dist
    GOOS=darwin GOARCH=amd64 go build -o dist/pictoria-darwin-amd64 main.go
    GOOS=darwin GOARCH=arm64 go build -o dist/pictoria-darwin-arm64 main.go
    GOOS=linux GOARCH=amd64 go build -o dist/pictoria-linux-amd64 main.go
    GOOS=windows GOARCH=amd64 go build -o dist/pictoria-windows-amd64.exe main.go
    @echo "✓ Built for all platforms"

# Quick Commands
# --------------

# Start locally and open browser
start: run open

# Start with Docker and open browser  
docker-start: docker-up open

# Full Docker rebuild and start
docker-fresh: docker-clean docker-build docker-start