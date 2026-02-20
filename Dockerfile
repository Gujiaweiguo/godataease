# DataEase Go Backend Docker Image (Production Ready)
# Build: docker build -t dataease:latest .
# Run: docker run -p 8080:8080 dataease:latest

# Build stage
FROM golang:1.21-alpine AS builder

# Set Go proxy for China network
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make upx

# Copy go mod files first for better cache
COPY apps/backend-go/go.mod apps/backend-go/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY apps/backend-go/ ./

# Build with optimizations and strip debug info
RUN go build -ldflags="-s -w -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" -o dataease-backend ./cmd/api

# Compress binary with UPX (reduces size by ~30%)
RUN upx --best --lzma dataease-backend 2>/dev/null || true

# Final stage - minimal runtime image
FROM alpine:3.19

# Labels for production
LABEL maintainer="DataEase Team"
LABEL version="2.0"
LABEL description="DataEase Go Backend - Production Ready"

WORKDIR /opt/module/dataease2.0

# Install minimal runtime dependencies
RUN apk add --no-cache --no-cache ca-certificates tzdata curl && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata

# Create non-root user for security
RUN addgroup -g 1000 dataease && \
    adduser -u 1000 -G dataease -s /bin/sh -D dataease

# Create directories with proper permissions
RUN mkdir -p /opt/module/dataease2.0/configs \
    /opt/module/dataease2.0/data \
    /opt/module/dataease2.0/logs && \
    chown -R dataease:dataease /opt/module/dataease2.0

# Copy binary from builder
COPY --from=builder --chown=dataease:dataease /build/dataease-backend /opt/module/dataease2.0/

# Copy default config
COPY --chown=dataease:dataease apps/backend-go/configs/config.example.yaml /opt/module/dataease2.0/configs/config.yaml

# Switch to non-root user
USER dataease

# Expose port
EXPOSE 8080

# Health check with proper timeout
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -fsS http://localhost:8080/health || exit 1

# Set environment variables
ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

# Run the binary
CMD ["/opt/module/dataease2.0/dataease-backend"]
