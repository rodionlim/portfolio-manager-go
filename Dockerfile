# Build stage for frontend
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web/ui

# Copy package files
COPY web/ui/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY web/ui/ ./

# Build frontend
RUN npm run build

# Build stage for backend
FROM golang:1.21-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git make bash gzip

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install swag for API documentation generation - use the same version as in CI
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.4

# Copy source code
COPY . .

# Copy built frontend assets from the frontend builder
COPY --from=frontend-builder /app/web/ui/dist ./web/ui/dist

# Generate swagger documentation
RUN swag init --quiet -g cmd/portfolio/main.go

# Compress assets and build the application with embedded assets
RUN make assets-compress build BUILTIN_ASSETS=1

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and wget for health check
RUN apk update && apk add --no-cache ca-certificates tzdata wget

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S portfolio && \
    adduser -u 1001 -S portfolio -G portfolio

# Copy binary from builder stage
COPY --from=backend-builder /app/portfolio-manager .

# Copy default config for reference
COPY --from=backend-builder /app/config.yaml ./config.yaml.example

# Create directories for data and logs
RUN mkdir -p /app/data /app/logs && \
    chown -R portfolio:portfolio /app

# Switch to non-root user
USER portfolio

# Expose ports
EXPOSE 8080 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

# Run the application
CMD ["./portfolio-manager", "-config", "/app/config/config.yaml"]