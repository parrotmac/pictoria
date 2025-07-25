# Frontend build stage
FROM node:24-alpine AS frontend-builder

# Set working directory
WORKDIR /app

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
# RUN npm ci
# We need dev dependencies to build the app
RUN npm install --include dev

# Copy frontend source
COPY frontend/ ./

# Build the Vue app
RUN npm run build

# Backend build stage
FROM golang:1.24-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pictoria .

# Final stage
FROM alpine:3.20

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates bash curl vim lsof

WORKDIR /root/

# Copy the binary from backend builder
COPY --from=backend-builder /app/pictoria .

# Copy built Vue app from frontend builder
COPY --from=frontend-builder /app/dist ./frontend/dist

COPY entrypoint.sh /entrypoint.sh

# Create uploads directory
RUN mkdir -p uploads

# Expose port
EXPOSE 8080

# Run the application
CMD ["/entrypoint.sh"]