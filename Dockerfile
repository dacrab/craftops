# Multi-stage Docker build for CraftOps (Go)
FROM golang:1.24-alpine AS builder

# Set environment variables
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set work directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=2.0.1

# Build the application
RUN go build -trimpath -ldflags "-X craftops/internal/cli.Version=${VERSION} -s -w" -o craftops ./cmd/craftops

# Production stage
FROM alpine:latest

# No extra environment variables required

# Install runtime dependencies
RUN apk add --no-cache \
    screen \
    openjdk17-jre-headless \
    ca-certificates \
    tzdata \
    && addgroup -g 1000 minecraft \
    && adduser -u 1000 -G minecraft -s /bin/sh -D minecraft

# Copy application from builder
COPY --from=builder /app/craftops /usr/local/bin/

# Create directories
RUN mkdir -p /minecraft/server /minecraft/mods /minecraft/backups /config /logs \
    && chown -R minecraft:minecraft /minecraft /config /logs

# No default config copied; generate via `craftops init-config`

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD craftops health-check || exit 1

# Switch to non-root user
USER minecraft
WORKDIR /minecraft

# Expose volumes
VOLUME ["/minecraft/server", "/minecraft/mods", "/minecraft/backups", "/config", "/logs"]

# Default command
ENTRYPOINT ["craftops"]
CMD ["--help"]

# Labels for metadata
LABEL org.opencontainers.image.title="CraftOps" \
      org.opencontainers.image.description="Modern Minecraft server operations and mod management tool built with Go" \
      org.opencontainers.image.vendor="dacrab" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/dacrab/craftops" \
      org.opencontainers.image.version="${VERSION}"