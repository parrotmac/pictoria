# Pictoria photo upload application

# Default command - show available commands
default:
    @just --list

# Development
# ----------

# Run both backend and frontend in development mode (recommended)
# dev:
#     @./scripts/dev.sh

# Run only the backend
run-backend:
    go run .

# Run only the frontend
run-frontend:
    cd frontend && npm run dev

# Run backend with auto-reload (requires air)
dev-backend:
    air || (echo "Installing air..." && go install github.com/cosmtrek/air@latest && air)

# Build both frontend and backend
build:
    @echo "Building frontend..."
    cd frontend && npm run build
    @echo "Building backend..."
    go build -o pictoria .
    @echo "✓ Build complete"

# Build only backend
build-backend:
    go build -o pictoria .

# Build only frontend  
build-frontend:
    cd frontend && npm run build

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
    docker compose up -d --build
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

# Open in browser (frontend dev server)
open location="localhost:5173":
    open http://{{location}} || xdg-open http://{{location}} || echo "Please open http://{{location}} in your browser"

# Open backend directly
open-backend:
    just open localhost:8080

# Check if server is running
health:
    @curl -s http://localhost:8080/api/photos > /dev/null && echo "✓ Server is running" || echo "✗ Server is not running"

# Show disk usage
disk-usage:
    @echo "Uploads: $$(du -sh uploads/ 2>/dev/null || echo '0B')"
    @echo "Photos: $$(ls uploads/ 2>/dev/null | wc -l | tr -d ' ') files"

# Maintenance
# -----------

# Setup directories and install dependencies
setup:
    mkdir -p uploads static backups
    @echo "✓ Directories created"
    @echo "Installing backend dependencies..."
    go mod download
    go mod tidy
    @echo "Installing frontend dependencies..."
    cd frontend && npm install
    @echo "✓ Setup complete"

# Install dependencies
deps:
    go mod download
    go mod tidy
    cd frontend && npm install

# Clean build artifacts
clean:
    rm -f pictoria
    rm -f coverage.*
    rm -rf frontend/dist
    rm -rf frontend/node_modules/.vite

# Reset everything (WARNING: deletes all data)
reset: clean
    rm -f data.json
    rm -rf uploads/*
    @echo "✓ Application reset complete"

# Backup data and uploads
backup:
    @mkdir -p backups
    @tar -czf backups/backup-$(date +%Y%m%d-%H%M%S).tar.gz storage.json uploads/
    @echo "✓ Backup created in backups/"

# Database
# --------

# Run database migrations
db-migrate:
    go run cmd/migrate/main.go

# Import data from JSON to PostgreSQL
db-import:
    go run cmd/migrate/main.go -import

# Run PostgreSQL locally for development
db-local:
    docker run --name pictoria-postgres -e POSTGRES_USER=pictoria -e POSTGRES_PASSWORD=pictoria_secret -e POSTGRES_DB=pictoria -p 5432:5432 -d postgres:16-alpine

# Connect to PostgreSQL CLI
db-connect:
    docker exec -it pictoria-postgres psql -U pictoria -d pictoria

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

# Frontend Commands
# -----------------

# Preview production build
preview:
    cd frontend && npm run preview

# Type check frontend
typecheck:
    cd frontend && npm run build-only

# Lint frontend code
lint-frontend:
    cd frontend && npm run lint

# Quick Commands
# --------------

# # Start both servers and open browser
# start: 
#     @just dev &
#     @sleep 3
#     @just open

# Start with Docker and open browser  
docker-start: docker-up open

# Full Docker rebuild and start
docker-fresh: docker-clean docker-build docker-start

# First time setup
# init: setup dev