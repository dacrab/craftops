# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X craftops/internal/cli.Version=${VERSION}" -o /craftops ./cmd/craftops

FROM alpine:3.20
RUN apk add --no-cache screen openjdk17-jre-headless ca-certificates tzdata \
    && adduser -D -u 1000 minecraft
COPY --from=builder /craftops /usr/local/bin/
RUN mkdir -p /minecraft/{server,mods,backups} /config /logs \
    && chown -R minecraft:minecraft /minecraft /config /logs
USER minecraft
WORKDIR /minecraft
VOLUME ["/minecraft/server", "/minecraft/mods", "/minecraft/backups", "/config", "/logs"]
HEALTHCHECK --interval=60s --timeout=10s CMD craftops health-check || exit 1
ENTRYPOINT ["craftops"]
CMD ["--help"]
LABEL org.opencontainers.image.source="https://github.com/dacrab/craftops"
