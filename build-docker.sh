#!/bin/bash

# Build script for creating Docker image locally
# This script builds the application locally and creates a Docker image

set -e

echo "Building portfolio-manager Docker image..."

# Clean up any previous builds
echo "Cleaning up previous builds..."
make clean

# Build the UI assets
echo "Building UI assets..."
cd web/ui
npm ci
npm run build
cd ../..

# Compress assets and build the application
echo "Building application with embedded assets..."
export PATH=$PATH:$(go env GOPATH)/bin
make assets-compress build BUILTIN_ASSETS=1

# Create a simple Dockerfile for the pre-built binary
cat > Dockerfile.local << 'EOF'
FROM busybox:glibc

WORKDIR /app

# Copy binary and config
COPY portfolio-manager .
COPY config.yaml ./config.yaml.example

# Create directories for data and logs
RUN mkdir -p /app/data /app/logs

# Expose ports
EXPOSE 8080 8081

# Run the application
CMD ["./portfolio-manager", "-config", "/app/config/config.yaml"]
EOF

# Build the Docker image using the local Dockerfile
echo "Building Docker image..."

# Temporarily rename .dockerignore to allow portfolio-manager binary
if [ -f .dockerignore ]; then
    mv .dockerignore .dockerignore.bak
fi

docker build -f Dockerfile.local -t portfolio-manager .

# Restore .dockerignore
if [ -f .dockerignore.bak ]; then
    mv .dockerignore.bak .dockerignore
fi

# Clean up temporary Dockerfile
rm Dockerfile.local

echo "Docker image 'portfolio-manager' built successfully!"
echo "You can now run it with:"
echo "  docker run -d --name portfolio-manager -p 8080:8080 -p 8081:8081 -v \$(pwd)/config.yaml:/app/config/config.yaml:ro portfolio-manager"