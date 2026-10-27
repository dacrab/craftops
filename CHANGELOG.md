# Changelog

Notable changes to CraftOps. Older releases are available on the
[GitHub releases page](https://github.com/dacrab/craftops/releases).

## v2.8.2 - 2026-10-27

A maintenance release focused on reliability, security, and clearer
documentation. No configuration changes are required.

### Fixed

- Restart warnings now appear at the intended times.
- Mod downloads retry correctly and clean up temporary files.
- Server start and status checks are more dependable.
- Backups handle their file paths consistently.
- Terminal output stays aligned with long or non-English titles.
- File sizes are displayed the same way across commands.
- Config files with unknown settings now fail with a clear error instead of
  being silently accepted.
- The installer now gives a helpful message when GitHub cannot be reached.

### Security

- Mod downloads now accept only safe filenames from trusted HTTPS sources.
- The build now includes the latest Go standard-library security fixes.

### Improved

- Config files are saved safely, so an interrupted write cannot leave a
  damaged file.
- Docker, README, and configuration docs now match the actual defaults.
- Go patch releases are updated automatically.
