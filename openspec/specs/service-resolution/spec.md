## Purpose

Defines the failure contract of resolving a service from a constructed container: which resolution forms return errors, which are documented to panic, and what a caller receives in each case regardless of whether the requested service type is a concrete type or an interface.

## Requirements

### Requirement: Error-returning resolution never panics

Resolution that returns an error alongside the service SHALL report every failure through that error and SHALL NOT panic. This holds for every requested service type, including interface types, pointer types, slice types, and value types, and for both keyed and unkeyed resolution.

On failure, resolution SHALL return the zero value of the requested type together with a non-nil error.

#### Scenario: Unregistered interface type

- **WHEN** a caller resolves an interface type that is not registered
- **THEN** resolution returns the zero value of that interface type and a non-nil error
- **AND** the calling goroutine does not panic

#### Scenario: Interface-typed service whose factory fails

- **WHEN** a caller resolves an interface type whose registered factory returns an error
- **THEN** resolution returns the zero value of that interface type and a non-nil error identifying the factory failure
- **AND** the calling goroutine does not panic

#### Scenario: Keyed interface resolution failure

- **WHEN** a caller resolves an interface type under a key for which no registration exists
- **THEN** resolution returns the zero value of that interface type and a non-nil error
- **AND** the calling goroutine does not panic

#### Scenario: Successful resolution is unaffected

- **WHEN** a caller resolves a registered service of any type, including an interface type
- **THEN** resolution returns the service instance and a nil error, as before this change

### Requirement: Must-style resolution panics with the underlying error

Resolution accessors documented to panic SHALL panic with the underlying error value itself, so that a recovering caller can inspect the failure with the standard error-unwrapping helpers. Such a panic SHALL NOT write to the standard logger.

These accessors SHALL panic only when the equivalent error-returning resolution would return a non-nil error, and SHALL return the resolved service otherwise.

#### Scenario: Panic value is the resolution error

- **WHEN** a caller uses a must-style accessor for a service that cannot be resolved, and recovers from the resulting panic
- **THEN** the recovered value is the error the equivalent error-returning resolution would have returned
- **AND** the standard error-unwrapping helpers can classify it

#### Scenario: Panic writes no log output

- **WHEN** a must-style accessor panics
- **THEN** no message is written to the standard logger

#### Scenario: Successful must-style resolution does not panic

- **WHEN** a caller uses a must-style accessor for a registered service
- **THEN** the accessor returns the service instance and does not panic
