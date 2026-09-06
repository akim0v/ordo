## Purpose

Validates the arguments of every service registration at container construction time and reports malformed registrations as returned errors, so that a bad registration is a handleable failure of the calling code rather than a panic that terminates the host process.

## ADDED Requirements

### Requirement: Malformed registrations are reported as errors

Container construction SHALL NOT panic when a registration is malformed. A registration whose arguments cannot produce a usable service SHALL be reported through the error value returned by container construction, and construction SHALL return no container in that case.

This applies to every registration form the container accepts: registrations of a factory or an instance under a declared service type, keyed registrations, value registrations, and registrations whose service type is inferred from the factory's return type.

#### Scenario: Malformed registration returns an error

- **WHEN** a container is constructed with a registration whose arguments are malformed
- **THEN** construction returns a nil container and a non-nil error
- **AND** the calling goroutine does not panic

#### Scenario: Sound registrations construct successfully

- **WHEN** a container is constructed with registrations whose arguments are all well formed
- **THEN** registration reports no fault
- **AND** construction proceeds to graph verification and behaves exactly as before this change

#### Scenario: Registration reports no fault through a logger

- **WHEN** a registration is malformed
- **THEN** no message is written to the standard logger as part of reporting it

### Requirement: Registration fault coverage

Registration SHALL report each of the following as a fault:

- A factory argument that is not a function.
- A factory that returns no values.
- A factory that returns more than two values.
- A factory whose second return value is not an error.
- A factory whose first return type is not assignable to the declared service type.
- An instance or value whose type is not assignable to the declared service type.
- A nil factory-or-instance argument.

Each fault SHALL describe which of these conditions was violated.

#### Scenario: Factory is not a function

- **WHEN** a service is registered with a non-function, non-instance argument that cannot serve as the declared service type
- **THEN** construction returns an error reporting that the factory is not a function or that the value is not assignable to the service type

#### Scenario: Factory returns no value

- **WHEN** a service is registered with a function that returns nothing
- **THEN** construction returns an error reporting that a factory must return at least one value

#### Scenario: Factory returns too many values

- **WHEN** a service is registered with a function that returns three or more values
- **THEN** construction returns an error reporting that the factory returns too many values

#### Scenario: Second return value is not an error

- **WHEN** a service is registered with a function returning two values whose second value is not an error
- **THEN** construction returns an error reporting that the second return value must be an error

#### Scenario: Factory return type is not assignable to the service type

- **WHEN** a service is registered under a declared type with a factory whose return type is not assignable to that type
- **THEN** construction returns an error reporting the declared service type and the factory's return type

#### Scenario: Instance is not assignable to the service type

- **WHEN** a service is registered under a declared type with an instance whose type is not assignable to that type
- **THEN** construction returns an error reporting the declared service type and the instance type

#### Scenario: Nil registration argument

- **WHEN** a service is registered with a nil factory-or-instance argument
- **THEN** construction returns an error reporting the registration as invalid
- **AND** the calling goroutine does not panic

### Requirement: Aggregated registration fault reporting

A single construction attempt SHALL validate every registration and report every registration fault found, not only the first. The returned error SHALL expose the individual faults so a caller can inspect them programmatically, and SHALL render a message listing each fault when formatted as text.

Registration faults SHALL be distinguishable from graph verification faults by type, so a caller can tell a malformed registration from a structurally unsatisfiable one.

#### Scenario: Multiple malformed registrations

- **WHEN** a container is constructed with three separate malformed registrations
- **THEN** the single returned error reports all three faults

#### Scenario: Faults are programmatically inspectable

- **WHEN** a caller receives a construction error and inspects it with the standard error-unwrapping helpers
- **THEN** the caller can reach each individual registration fault and read the types involved in it

#### Scenario: Registration faults are distinguishable from verification faults

- **WHEN** a caller receives a construction error
- **THEN** the caller can determine whether it reports malformed registrations or graph verification faults

### Requirement: Registration faults identify their registration site

Every registration fault SHALL identify the registration that produced it by:

- the position of that registration among the registrations passed to construction,
- the declared service type of the registration, and
- the type of the offending factory or instance argument.

Every registration fault SHALL also carry the source location of the call that created the registration, so a fault points at the registration site rather than at the construction call. This location SHALL appear in the fault's text form.

#### Scenario: Fault names the registration position

- **WHEN** the third of five registrations is malformed
- **THEN** the reported fault identifies that registration by its position among the registrations passed to construction

#### Scenario: Fault names the source location of the registration

- **WHEN** a registration created in one source file is passed to construction in another
- **THEN** the reported fault names the file and line of the call that created the registration

#### Scenario: Two faulty registrations of the same service type are distinguishable

- **WHEN** the same service type is registered twice and both registrations are malformed
- **THEN** the two reported faults differ in position and source location

### Requirement: Registration faults short-circuit graph verification

When at least one registration fault is found, construction SHALL return the registration faults and SHALL NOT perform graph verification. A malformed registration is dropped from the graph, so verifying it would report missing dependencies that exist only because of the dropped registration.

When no registration fault is found, construction SHALL perform graph verification and report its result as before this change.

#### Scenario: No verification faults are reported alongside registration faults

- **WHEN** a container is constructed with one malformed registration and another registration that depends on the service the malformed registration would have provided
- **THEN** the returned error reports only the registration fault
- **AND** it reports no missing-dependency fault

#### Scenario: Verification still runs when registrations are sound

- **WHEN** a container is constructed with well-formed registrations whose dependency graph contains a fault
- **THEN** the returned error reports the graph verification faults as before this change

### Requirement: Registration does not instantiate services

Registration validation SHALL be a static analysis of the registration arguments. It SHALL NOT invoke any registered factory and SHALL NOT create any service instance, preserving the lazy singleton semantics of the container.

#### Scenario: Factories are not called during registration validation

- **WHEN** a container is constructed with factories that record whether they were invoked, together with a malformed registration
- **THEN** no factory has been invoked once construction returns
