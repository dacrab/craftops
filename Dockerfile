# Multi-stage Docker build for Minecraft Mod Manager (Go)
FROM golang:1.21-alpine AS builder

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

# Build the application
RUN go build -trimpath -ldflags "-X main.Version=2.0.0 -s -w" -o craftops ./cmd/craftops

# Production stage
FROM alpine:latest

# Set environment variables
ENV MINECRAFT_MOD_MANAGER_CONFIG_FILE="/config/config.toml"

# Install runtime dependencies
RUN apk add --no-cache \
    screen \
    openjdk17-jre-headless \
    curl \
    ca-certificates \
    tzdata \
    && addgroup -g 1000 minecraft \
    && adduser -u 1000 -G minecraft -s /bin/sh -D minecraft

# Copy application from builder
COPY --from=builder /app/craftops /usr/local/bin/

# Create directories
RUN mkdir -p /minecraft/server /minecraft/mods /minecraft/backups /config /logs \
    && chown -R minecraft:minecraft /minecraft /config /logs

# Copy default config
COPY conf.toml /config/config.toml.example

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
      org.opencontainers.image.version="2.0.0"