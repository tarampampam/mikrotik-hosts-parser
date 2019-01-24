# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog][keepachangelog] and this project adheres to [Semantic Versioning][semver].

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
