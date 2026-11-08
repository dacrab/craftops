# syntax=docker/dockerfile:1@sha256:4edf897a3ffa55b89f906fc8cc78afdb3f1834cc9c7083565e611a8a7d5fe99e
FROM golang:1.27.3-alpine@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -buildvcs=auto \
    -ldflags "-s -w -X craftops/internal/buildinfo.Version=${VERSION#v}" \
    -o /craftops ./cmd/craftops

FROM alpine:3.24@sha256:3f4d6b32e26703c6082839234f43769d5046b3ac7c910033911d1e71d8cd630a
RUN apk add --no-cache screen openjdk17-jre-headless ca-certificates tzdata \
    && adduser -D -u 1000 minecraft \
    && mkdir -p /minecraft/server /minecraft/mods /minecraft/backups /config /logs \
    && chown -R minecraft:minecraft /minecraft /config /logs

COPY --from=builder /craftops /usr/local/bin/craftops

USER minecraft
WORKDIR /minecraft
VOLUME ["/minecraft/server", "/minecraft/mods", "/minecraft/backups", "/config", "/logs"]

HEALTHCHECK --interval=60s --timeout=10s --start-period=30s CMD ["craftops", "server", "status"]
ENTRYPOINT ["craftops"]
CMD ["--help"]

LABEL org.opencontainers.image.source="https://github.com/dacrab/craftops"
