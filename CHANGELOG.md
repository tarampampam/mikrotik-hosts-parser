# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog] and this project adheres to [Semantic Versioning][semver].

## UNRELEASED

### Changed

- Go version updated from `1.16.3` up to `1.16.6`

## v4.3.0

### Changed

- Go version updated from `1.16.2` up to `1.16.3`

### Added

- HTTP route `/metrics` with metrics in [prometheus](https://github.com/prometheus) format
- Comment in generated script with "script generation" time

## v4.2.0

### Changed

- Go version updated from `1.16.0` up to `1.16.2`

### Added

- Support for `linux/arm64`, `linux/arm/v6` and `linux/arm/v7` platforms for docker image

## v4.1.0

### Changed

- Go version updated from `1.15.7` up to `1.16.0`

## v4.0.1

### Fixed

- Mistake inside HTTP script generation handler (it caused handler panic when "excludes list" less than "sources list")

## v4.0.0

### Changed

- GitHub actions updated
- Docker image based on `scratch` (instead `alpine` image)
- Go version updated from `1.13` up to `1.15`
- Package name changed from `mikrotik-hosts-parser` to `github.com/tarampampam/mikrotik-hosts-parser/v4`
- Directory `public` renamed to `web`
- Config file location now is `./configs/config.yml` (instead `./serve.yml`)
- App packages refactored
- Docker image now contains only one layer
- More strict linter settings
- Config file now contains only sources list and script generator options
- Default values for the next `serve` sub-command flags:
  - For `--config` now is `%binary_file_dir%/configs/config.yml` (instead nothing)
  - For `--resources-dir` now is `%binary_file_dir%/web` (instead nothing)
  - For `--listen` flag now is `8080` (instead nothing)
- For static files serving disabling you can set `--resources-dir` empty value (`""`)
- Large performance improvements
- HTTP requests log records contains request processing duration
- Panics inside HTTP handlers now will be logged and JSON-formatted string will be returned (instead empty response)
- Frontend dependencies updated
- Docker image (for release) now supports `linux/amd64` + `linux/386` platforms

### Added

- Docker healthcheck
- Healthcheck sub-command (hidden in CLI help) that makes a simple HTTP request (with user-agent `HealthChecker/internal`) to the `http://127.0.0.1:8080/live` endpoint. Port number can be changed using `--port`, `-p` flag or `LISTEN_PORT` environment variable
- Two caching engines (memory and redis) instead file-based cache
- `serve` sub-command flags:
  - `--cache-ttl` for cache entries lifetime setting (examples: `50s`, `1h30m`); `30m` by default; environment variable: `CACHE_TTL`
  - `--caching-engine` for caching engine changing (`memory|redis`); `memory` by default; environment variable: `CACHING_ENGINE`
  - `--redis-dsn` for redis server URL setting; `redis://127.0.0.1:6379/0` by default; environment variable: `REDIS_DSN`. This flag is required only if `redis` caching engine is set
- Global (available for all sub-commands) flags:
  - `--log-json` for logging using JSON format (`stderr`)
  - `--debug` for debug information for logging messages
  - `--verbose` for verbose output
- Graceful shutdown support for `serve` sub-command
- HTTP endpoints:
  - `/live` for liveness probe
  - `/ready` for readiness probe
- E2E tests (using [postman](https://www.postman.com/))

### Removed

- File-based cache support
- HTTP `/api/routes` handler

### Fixed

- Wrong HTTP `Content-Type` header value for docker environment

## v3.0.3

### Fixed

- Dead link in config file replaced with mirror

### Added

- Log all `HTTP` requests to `stdout` [#39]
- `redirect_to` parameter validation [#37]

[#37]:https://github.com/tarampampam/mikrotik-hosts-parser/issues/37
[#39]:https://github.com/tarampampam/mikrotik-hosts-parser/pull/39

## v3.0.2

### Changed

- Docker image supports argument with application version value
- Docker image builds using GitHub Actions (not hub.docker.com)

## v3.0.1

### Fixed

- Version value extraction using GitHub Actions

## v3.0.0

### Changed

- Application re-wrote on GoLang _(previous HTTP endpoint still working)_
- Settings now defined in special configuration file
- Performance improvements

## v2.3.1

### Fixed

- Composer installation in dockerfile

## v2.3.0

### Changed

- Basic sources URIs
- (docker) Now docker image based on [PHPPM][phppm]

[phppm]:https://github.com/php-pm/php-pm

## v2.2.1

### Added

- Environment value `FORCE_HTTPS` for forcing `https` scheme usage

## v2.2.0

### Changed

- Dockerfile now based on `alpine`
- Bower-installed components removed (use `cdnjs.com` now)
- Make repository clear
- Added `delay 3` after `/tool fetch ...` [#11]

[#11]: https://github.com/tarampampam/mikrotik-hosts-parser/issues/11

## v2.1.2

### Fixed

- `WindowsSpyBlocker` hosts file URI [#10]

[#10]: https://github.com/tarampampam/mikrotik-hosts-parser/issues/10

[keepachangelog]:https://keepachangelog.com/en/1.0.0/
[semver]:https://semver.org/spec/v2.0.0.html
