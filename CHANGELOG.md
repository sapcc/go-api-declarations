# Changelog

All notable API changes will be documented in this file.

The sections should follow the order `Added`, `Changed`, `Fixed`, `Removed`, and
`Deprecated`.

## [Unreleased](https://github.com/sapcc/go-api-declarations/compare/v1.1.0...HEAD)

### Changed

- cadf: make API more stable by using `string` type for `Attachment.Content` field.
- cadf: ensure consistent format across all event by using `time.Time` type for
  `Event.EventTime` field and custom JSON marshalling behavior.

## v1.1.0 (2022-04-22)

### Added

- `cadf` package, imported from Hermes.

## v1.0.0 (2022-04-21)

### Added

- `limes` package, imported from Limes.
