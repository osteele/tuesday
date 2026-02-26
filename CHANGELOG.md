# Changelog

## [v1.0.4] - 2026-02-26

### Fixed

- `%Q` now returns milliseconds since epoch, matching Ruby's `DateTime#strftime`.
  Previously it returned microseconds.

### Removed

- Removed unused `%f` padding entry. Ruby does not implement `%f` as a strftime
  directive.

### Changed

- Minimum supported Go version is now 1.25 (following Go's two-release support
  policy).
- CI migrated from Travis CI and AppVeyor to GitHub Actions, testing on Linux,
  macOS, and Windows across Go 1.25.x and 1.26.x.

## [v1.0.3] - 2017-08-14

### Added

- Implement `%:::z` (minimal timezone offset format).

### Fixed

- VMS date (`%v`) now correctly uppercases the month abbreviation.

## [v1.0.2] - 2017-08-12

### Fixed

- Ruby week calculations (`%U`, `%W`) now differ correctly from ISO weeks.

## [v1.0.1] - 2017-08-10

### Fixed

- Fix `_` flag (space padding).
- Fix precedence of flag and width specifiers.

## [v1.0.0] - 2017-08-10

Initial release. Ruby-compatible `strftime` for Go, supporting:

- All standard `strftime` conversions
- Padding flags (`-`, `_`, `0`), case flags (`^`, `#`), and field widths
- Ruby-specific conversions: `%s`, `%N`, `%L`, `%:z`, `%::z`

[v1.0.4]: https://github.com/osteele/tuesday/compare/v1.0.3...v1.0.4
[v1.0.3]: https://github.com/osteele/tuesday/compare/v1.0.2...v1.0.3
[v1.0.2]: https://github.com/osteele/tuesday/compare/v1.0.1...v1.0.2
[v1.0.1]: https://github.com/osteele/tuesday/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/osteele/tuesday/releases/tag/v1.0.0
