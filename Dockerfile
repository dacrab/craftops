# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X craftops/internal/cli.Version=${VERSION}" -o /craftops ./cmd/craftops

FROM alpine:3.20
# Install runtime dependencies and setup user in one layer
RUN apk add --no-cache screen openjdk17-jre-headless ca-certificates tzdata \
    && adduser -D -u 1000 minecraft \
    && mkdir -p /minecraft/server /minecraft/mods /minecraft/backups /config /logs \
    && chown -R minecraft:minecraft /minecraft /config /logs

USER minecraft
WORKDIR /minecraft
VOLUME ["/minecraft/server", "/minecraft/mods", "/minecraft/backups", "/config", "/logs"]

HEALTHCHECK --interval=60s --timeout=10s CMD ["craftops", "health-check"]
ENTRYPOINT ["craftops"]
CMD ["--help"]

LABEL org.opencontainers.image.source="https://github.com/dacrab/craftops"
