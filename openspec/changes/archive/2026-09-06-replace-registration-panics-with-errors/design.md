## Context

See `proposal.md` — Why. Current state relevant to the approach:

- `Option` is an exported interface with a single unexported method `apply(*Container)`. No code outside the module can implement it, so its signature can change without breaking any caller.
- Validation lives in two places: `newServiceFactory` (`di/service_factory.go`) checks the factory's shape, and each option's `apply` (`di/container.go`) checks assignability to the declared service type. All eight checks call `log.Panicf`.
- `withServiceKey` branches on `reflect.TypeOf(factoryOrInstance).Kind() == reflect.Func` *before* any validation, so a nil argument dereferences a nil `reflect.Type`.
- `NewContainer` already has the shape this change needs: build, then validate, then return `(nil, err)`. `verify()` returns `*VerificationError` aggregating `Faults []error` with `Unwrap() []error`. The new registration phase mirrors that shape rather than inventing a second style.
- `getServiceKey` performs a single-value type assertion `service.Interface().(T)` on the value returned by `getService`, including on the error path where that value is `reflect.Zero(id.Type)`. For an interface `T` this is a nil interface and the assertion panics.

## Goals / Non-Goals

**Goals:**

- One failure channel for container construction: an error return, with no panic reachable through `NewContainer` for any input.
- Registration faults that a developer can act on without a debugger: which registration, which types, which source line, what was wrong.
- Programmatic classification of a fault cause with `errors.Is`, not string matching.
- Symmetry with the existing verification phase, so the two construction phases read and behave alike.

**Non-Goals:**

- Changing verification behavior, its error type, or its message format.
- Changing resolution semantics beyond replacing a panic with the error that path already computes.
- Adding a `MustNewContainer` convenience constructor. It is additive, unrelated to removing panics, and can land separately once the error shape is settled.
- Detecting duplicate registrations, ambiguous keys, or any other new class of fault. This change relocates existing checks and fixes two latent panics; it does not widen validation.

## Decisions

### D1: `apply` returns an error

`apply(c *Container) error`. Each option validates its own arguments and returns the first fault it finds, or nil after registering its accessor.

Alternative considered: validate eagerly inside each `With*` constructor and store the fault on the option struct for `apply` to hand back. Rejected — it splits validation across two files with no benefit, and an option value that is never passed to a container has no one to report to. Validation stays where the reflection work already happens.

An option that fails validation registers nothing. Its service is simply absent from the graph, which is why registration faults must short-circuit verification (D6).

### D2: One fault per option, not per check

`apply` returns on its first fault rather than accumulating within one option. A registration with a malformed factory has one root cause; reporting "not a function" and "return type not assignable" for the same argument is noise. Aggregation happens *across* options (D5).

### D3: Sentinel causes plus one structured fault type

```go
var (
    ErrNilRegistration            = errors.New("di: registration value is nil")
    ErrFactoryNotFunction         = errors.New("di: service factory must be a function")
    ErrFactoryNoReturn            = errors.New("di: service factory must return at least one value")
    ErrFactoryTooManyReturns      = errors.New("di: service factory returns too many values")
    ErrFactorySecondReturnNotErr  = errors.New("di: second service factory return value must be an error")
    ErrNotAssignable              = errors.New("di: value is not assignable to the service type")
)

type InvalidRegistrationError struct {
    Index       int          // position among the options passed to NewContainer
    ServiceType reflect.Type // declared service type; nil when inferred from the factory
    ValueType   reflect.Type // type of the factory or instance; nil when the argument was nil
    Site        callSite     // file and line of the With* call
    Err         error        // one of the sentinels above
}
func (e *InvalidRegistrationError) Unwrap() error { return e.Err }
```

A sentinel per cause plus one fault type gives `errors.Is(err, di.ErrFactoryNotFunction)` for classification and `errors.As(err, &invalid)` for the details, with one type to document instead of six.

Alternative considered: a distinct error type per cause, mirroring `MissingDependencyError` / `CircularDependencyError` in verification. Rejected — those two carry genuinely different payloads (a pair of types vs. an ordered cycle); the six registration causes all carry the same fields and differ only in message.

### D4: `newServiceFactory` returns `(*serviceFactory, error)`

The four shape checks move from `log.Panicf` to returned sentinel errors. The factory-shape checks and the assignability check then compose in `apply`: build the factory, return its error if any, then check assignability. The error the constructor returns is a bare sentinel; `apply` wraps it in `InvalidRegistrationError` because only `apply` knows the index and the declared service type.

### D5: `RegistrationError` aggregate, mirroring `VerificationError`

```go
type RegistrationError struct { Faults []error }
func (e *RegistrationError) Error() string   // "di: container registration failed:" + one "  - " line per fault
func (e *RegistrationError) Unwrap() []error
```

Identical shape to `VerificationError`, so a caller learns one pattern for both phases and `errors.Is` / `errors.As` traverse every fault. Two aggregates rather than one shared type, because the spec-level contract of the two phases differs and a caller distinguishing "my registrations are malformed" from "my graph does not close" does it with a single `errors.As` on the aggregate.

Alternative considered: `errors.Join`. Rejected — no typed handle for the aggregate, and the joined message has no header, so it would not match the verification output the package already produces.

### D6: Registration short-circuits verification

`NewContainer` applies every option, collecting faults; if any fault exists it returns `(nil, &RegistrationError{...})` without calling `verify()`. A dropped registration would otherwise make `verify()` report missing dependencies that exist only because of it — the developer would chase a phantom fault whose cause is three lines up in the same error.

The cost is a second round trip when a container has both a malformed registration and a genuinely unsatisfiable dependency. That is the right trade: the second run's faults are then trustworthy.

`newContainer` appends its own `WithValue(c)` registration after the caller's options. It cannot fail, and index numbering counts only the caller's options, so the internal registration never appears in a fault or shifts an index.

### D7: Capture the call site in the exported constructors

```go
type callSite struct { File string; Line int }
func newCallSite() callSite // wraps runtime.Caller(2)
```

Every exported `With*` constructor captures its immediate caller. `runtime.Caller` at registration time is a single stack lookup at startup; the frame is stored as file and line, formatted only when a fault renders.

The skip depth must be fixed per call path, so each exported constructor calls the helper directly and threads the resulting `callSite` down to the option struct, rather than letting `withServiceKey` or `withServiceFactory` guess how deep they were called from.

Alternative considered: `runtime.Callers` + `CallersFrames` to also record the enclosing function name. Rejected — more allocation for information `file:line` already locates.

Index alone is not enough: options are commonly built in helper functions or appended to a slice, where the position in the final slice says nothing about where the registration was written. Site alone is not enough either: a loop registering many services reports the same line for each. Both together identify the registration in either style.

### D8: Fix `getServiceKey` with an error-path return and a comma-ok assertion

```go
service, err := c.getService(id)
if err != nil {
    var zero T
    return zero, err
}
instance, ok := service.Interface().(T)
if !ok {
    var zero T
    return zero, fmt.Errorf("%w: resolved %v for service type %v", ErrServiceTypeMismatch, service.Type(), reflect.TypeFor[T]())
}
return instance, nil
```

Returning `var zero T` on the error path is what makes the interface case correct: the zero value of the interface *type parameter* is a typed nil, whereas asserting on `reflect.Zero(ifaceType).Interface()` asserts on an untyped nil and panics. The comma-ok form on the success path covers the remaining case defensively — a registration that passed assignability checks should always assert cleanly, so `ErrServiceTypeMismatch` reports a container bug rather than crashing the caller.

### D9: `Must*` panic with the error value

`panic(err)` replaces `log.Panicf(...)`. `log.Panicf` writes to the standard logger — a side effect a library should not impose — and panics with a formatted string, so `recover()` yields a `string` that `errors.As` cannot inspect. Panicking with the error preserves the documented `Must*` contract while making the recovered value useful.

This changes the recovered value's type for anyone recovering today. It is called out as breaking in the proposal.

## Risks / Trade-offs

- **Callers relying on a panic for invalid registration silently change behavior** → The `NewContainer` signature already returns an error and every example already checks it; the release notes and README call the change out, and the examples stay the canonical usage.
- **Recovered `Must*` panic value changes from `string` to `error`** → Documented as breaking. The new value is strictly more useful, and `fmt.Sprint(recovered)` still yields the message.
- **`runtime.Caller` per registration** → One stack lookup per option at construction only, never on the resolution path. Negligible against the reflection each option already performs.
- **Tests asserting a source line become position-sensitive** → Assert on the file's base name and a non-zero line, not on an exact line number, so editing the test file above the assertion does not break it.
- **Short-circuiting hides real verification faults behind a registration typo** → Accepted deliberately (D6); the alternative is reporting faults that are artifacts of the typo.
- **Two aggregate error types to keep in sync** → Their `Error()` formats must stay parallel; a test asserting both headers share the `di: container ... failed:` shape keeps them honest.

## Migration Plan

Single release, no runtime migration:

1. Land the error-returning registration path with tests rewritten from `suite.Panics` to error assertions.
2. Update `README.md` and any example that documents or relies on panicking registration.
3. Note in the release notes: invalid registrations now return `*di.RegistrationError` from `NewContainer`; `Must*` accessors panic with an `error` instead of a logged string.

Rollback is a revert — no persisted state, no wire format, no data migration.
