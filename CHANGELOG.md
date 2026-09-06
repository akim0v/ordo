# Changelog

All notable changes to this project are documented in this file.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Tests covering `ErrValueIsFunction`, including that a function which cannot be a factory
  at all keeps the cause describing why, and that an explicit service type still reports
  `ErrNotAssignable`.

## [0.1.0]

First tagged release.

### Added

- `New` builds a container from `Option` values, validating every registration and then
  verifying the whole dependency graph. It never panics, and returns either an error or a
  container that is fully resolvable.
- Registration options `WithService`, `WithFactory`, `WithValue` and their keyed
  counterparts `WithKeyedService`, `WithKeyedFactory` and `WithKeyedValue`.
- Resolution through the generic methods `GetService`, `MustGetService`, `GetKeyedService`
  and `MustGetKeyedService`. Services are lazy and constructed once.
- Registering one service type more than once is supported: a `[]T` dependency or request
  receives every registration in registration order, and a plain request receives the last.
- Typed errors for every failure — `RegistrationError` and `InvalidRegistrationError` for
  malformed options, `VerificationError` with `MissingDependencyError` and
  `CircularDependencyError` for a broken graph, and `DependencyError` for a constructor that
  fails at resolution — each aggregating every fault of a call and classifiable with
  `errors.Is` and `errors.AsType`.
- Registration faults carry the `file:line` of the `With*` call that created them, and
  missing-dependency faults carry the source location of the registration that declared the
  dependency.
- The container registers itself, so `*Container` is always resolvable.

### Requires

- Go 1.27 or newer. Resolution uses generic methods, which are not valid Go before 1.27.

[Unreleased]: https://github.com/akim0v/ordo/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/akim0v/ordo/releases/tag/v0.1.0
