# Contributing

## Development

```bash
make build        # build to build/craftops
make test         # run tests with race detector
make lint         # run golangci-lint
make fmt          # gofmt all packages
make package      # cross-compile for linux/darwin × amd64/arm64
```

## Releasing

Add a section for the new version to `CHANGELOG.md`, then use the helper script
to bump the semver tag and trigger the GitHub Actions release pipeline:

```bash
./scripts/release.sh patch                      # 2.3.0 → 2.3.1
./scripts/release.sh minor "New mod features"   # 2.3.1 → 2.4.0
./scripts/release.sh major                      # 2.4.0 → 3.0.0
```

The script uses the matching `CHANGELOG.md` section, prompts for confirmation,
then creates an annotated git tag and pushes it. GitHub Actions builds the
binaries and publishes the release notes automatically.
