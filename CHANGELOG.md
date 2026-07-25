# Changelog

All notable changes to squirrel are documented in this file.
Format: [Keep a Changelog](https://keepachangelog.com/en/1.1.0/); versions follow [SemVer](https://semver.org).
Release with `make release VERSION=x.y.z` — rotates this file, commits, tags `vx.y.z`.

## [Unreleased]

### Added
- `docker-compose.yml`: `mem_limit`/`cpus` caps and `json-file` log rotation (max-size 10m, max-file 3) on the `api` service.
- `logrotate.conf`: host-level rotation for the bind-mounted `./log/*.log` files (daily, 7 rotations, copytruncate).
- `docs/LOGGING.md`: logging setup and rotation reference.
- `docs/DEPLOYMENT_USING_DOCKER.md`: Docker-based production setup guide.

### Changed
- Stats cache in `cmd/api/main.go` now uses vendored `keeper/pkg/cache` instead of `squirrel/pkg/cache`. Behavior unchanged.
- `DEPLOYMENT.md` moved to `docs/DEPLOYMENT_WITHOUT_DOCKER.md` (bare-binary + systemd path), alongside the new Docker path doc.

### Removed
- `pkg/cache` (duplicate TTL cache, superseded by vendored `keeper/pkg/cache`).

## [0.0.2] - 2026-07-11

### Added
- Version in `GET /health` response, read from CHANGELOG.md.
- `make version` target; version shown in `make info`.

## [0.0.1] - 2026-07-11

### Added
- Changelog and `make release` versioning workflow.
