# CraftOps Deployment

Concise steps for building, releasing, and running CraftOps.

## Build and package

- Local build: `make build` (binary in `build/`)
- Install locally: `make install` or `make install-system`
- Cross‑package: `make package` (artifacts in `dist/`)

## Docker

Build (uses Go 1.23 multi‑stage and VERSION arg):

```bash
docker build --build-arg VERSION=$(git describe --tags --always 2>/dev/null || echo dev) -t craftops:local .
```

Run:

```bash
docker run --rm \
  -v /host/server:/minecraft/server \
  -v /host/backups:/minecraft/backups \
  -v /host/config:/config \
  ghcr.io/dacrab/craftops:latest health-check
```

## Releases (SemVer)

1) Tag: `git tag -a vX.Y.Z -m "release vX.Y.Z" && git push origin vX.Y.Z`
2) Build/upload binaries and container images
3) Publish brief notes and checksums

## CI/CD (outline)

- CI: `go mod tidy`, `gofmt -s -w`, `go vet`, `go test ./...`
- Release: matrix build for {linux, darwin} x {amd64, arm64}; attach assets; push GHCR image
