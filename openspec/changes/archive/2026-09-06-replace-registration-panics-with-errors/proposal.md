## Why

Registering a malformed service today kills the process: `WithService`, `WithFactory` and friends validate their arguments inside `Option.apply` and call `log.Panicf` when validation fails, so a typo in a library's registration list crashes the host application instead of returning an error the caller can handle. This is inconsistent with `NewContainer`, which already returns a `*VerificationError` describing every graph fault it finds — a caller must therefore handle errors *and* recover from panics to construct a container safely.

Two further panics are latent bugs rather than deliberate contracts: `GetService[T]` panics with `interface conversion: interface is nil` whenever `T` is an interface type and resolution fails, and `WithService[T](nil)` panics with a nil dereference before any validation runs.

## What Changes

- **BREAKING** Option validation no longer panics. Every registration fault (factory is not a function, factory returns nothing / too many values / a non-error second value, factory return type or instance type not assignable to `T`, nil factory-or-instance) is collected and returned from `NewContainer` as an error.
- New exported error types: `RegistrationError` (the aggregate returned by `NewContainer`, mirroring `VerificationError`: `Faults []error`, `Unwrap() []error`, multi-line `Error()`) and a per-fault type carrying the offending option's index, the registered service type, the offending value's type, and the source location of the `With*` call.
- Registration faults short-circuit verification: when any option fails, `NewContainer` returns the `*RegistrationError` and never runs graph verification, so the caller is not shown cascading missing-dependency faults caused by a registration that was dropped.
- `With*` constructors capture their caller's `file:line` (`runtime.Caller`) so a fault points at the registration site, not at `NewContainer`.
- **Fix** `GetService[T]` / `GetKeyedService[T]` return `(zero T, err)` instead of panicking when `T` is an interface type and resolution fails.
- `MustGetService` / `MustGetKeyedService` keep panicking — a documented `Must*` contract — but panic with the underlying `error` value instead of `log.Panicf`, so `recover()` yields something `errors.As` can inspect and no logger side effect occurs.
- After this change the `di` package contains no `log.Panicf` call and no panic outside the two `Must*` methods.

## Capabilities

### New Capabilities

- `container-registration`: validation of service registrations at container construction, the faults that validation reports, and how those faults are aggregated, located, and ordered relative to graph verification.
- `service-resolution`: the error contract of service resolution — what resolution returns for every service type including interfaces, and the panic contract of the `Must*` accessors.

### Modified Capabilities

<!-- None. `container-verification` requirements are unchanged: verification still
     reports every graph fault it finds, still instantiates nothing, and its error
     type and semantics are untouched. This change only adds a phase that runs
     before it. -->

## Impact

- `di/container.go` — `Option.apply` signature, `serviceFactoryOption`, `serviceInstanceOption`, `factoryOption`, `withServiceKey`, `newContainer`, `NewContainer`, `getServiceKey`, `MustGetService`, `MustGetKeyedService`.
- `di/service_factory.go` — `newServiceFactory` returns an error instead of panicking.
- New file for the registration error types, alongside the existing `di/verification.go`.
- `Option` is an interface with an unexported method, so no code outside the module can implement it; changing `apply` to return an error is source-compatible for all callers.
- Behavior change for existing callers: code that today relies on a panic (or on `recover`) for an invalid registration now receives an error from `NewContainer`. Tests asserting `suite.Panics` on registration and factory validation (`di/container_test.go`, `di/service_factory_test.go`) must be rewritten as error assertions.
- `README.md` and the `examples/` programs may need updating if they document the panic behavior.
