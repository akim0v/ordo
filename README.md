# Ordo

**Ordo** is an application framework for Go. Its first component is `di`, a dependency
injection container that keeps your wiring explicit and verifies it before it hands you
anything.

## Why Ordo

Dependency injection in Go usually asks you to accept some magic. Struct tags a reflection
engine reads at runtime. Generated files you do not own rewriting your wiring. String names
where a typo becomes a failure at startup. Each of them moves knowledge out of the type
system and into a convention you have to remember.

Ordo takes the opposite position. Registration is a plain generic call:

```go
di.WithService[UserRepository](storage.NewPostgresUserRepository)
```

The compiler already knows the service type and the constructor. No tags, no code
generation, no build step, and no string names unless you deliberately ask for one.
Dependencies are read from the constructor's own parameter types.

Because the container knows the entire graph before it resolves a single service, it checks
that graph up front. `di.NewContainer` validates every registration, then verifies every
edge, and returns what it found — a malformed registration, a missing dependency, a cycle —
at the call site, with the `file:line` of the registration responsible.

That is the guarantee the name stands for. **A container either fails to build, or is fully
resolvable.** Nothing is deferred to the first request in production.

In practice:

- No code generation and no build step.
- No struct tags and no reflection-driven field injection.
- `NewContainer` never panics. Failures are ordinary Go errors, classified with `errors.Is`
  and `errors.As`.
- The dependency graph is verified once, at construction.

## Installation

Ordo requires **Go 1.27 or newer**. Add it to an existing module:

```bash
go get -u github.com/akim0v/ordo
```

## Quickstart

```go
package main

import (
	"log"

	"github.com/akim0v/ordo/di"

	"github.com/you/app/config"
	"github.com/you/app/rest"
	"github.com/you/app/storage"
	"github.com/you/app/usecase"
)

func main() {
	c, err := di.NewContainer(
		// Register a constructor against the interface it satisfies.
		// usecase.UserRepo is the interface usecase.NewUserService depends on;
		// storage.NewUserRepo is the constructor that implements it.
		di.WithService[usecase.UserRepo](storage.NewUserRepo),

		// Register a ready-made value under its own type.
		// Equivalent to di.WithService[*config.Config](config.New()).
		di.WithValue(config.New()),

		// Register a constructor under its return type.
		// Equivalent to di.WithService[*usecase.UserService](usecase.NewUserService).
		di.WithFactory(usecase.NewUserService),

		// rest.NewUserController takes a *usecase.UserService; the container supplies it.
		di.WithService[rest.Controller](rest.NewUserController),
	)
	if err != nil {
		log.Fatalf("could not create the container: %s", err)
	}

	// Every service registered as a rest.Controller, in registration order.
	for _, controller := range c.MustGetService[[]rest.Controller]() {
		controller.Init(...)
	}
}
```

Runnable versions of this and other setups live in [`examples/`](./examples).

## Registering services

| Option | Service type | Source |
| --- | --- | --- |
| `di.WithService[T](factoryOrInstance)` | `T`, given explicitly | a constructor or a ready value |
| `di.WithFactory(factory)` | the factory's return type | a constructor |
| `di.WithValue(value)` | the value's own type | a ready value |

`WithService` accepts either form and tells them apart by kind, so it is the option to reach
for when the service type differs from the concrete type — registering an implementation
against an interface, most often.

Each option has a keyed counterpart — `WithKeyedService`, `WithKeyedFactory`,
`WithKeyedValue` — for the case where one type has several registrations that callers need
to tell apart by name.

A constructor may return `(T, error)`. The container propagates a construction failure to
the caller resolving that service.

### Multiple implementations

Registering the same service type more than once is not a conflict. Ordo collects the
registrations, and a dependency declared as a slice receives all of them, in registration
order:

```go
c, err := di.NewContainer(
	di.WithService[UserRepository](NewCacheRepository),
	di.WithService[UserRepository](NewDBRepository),

	// NewUserService takes []UserRepository and receives both.
	di.WithFactory(NewUserService),
)
```

The same works at the call site: `c.MustGetService[[]UserRepository]()` returns every
registration, while `c.MustGetService[UserRepository]()` returns the last one.

## Resolving services

```go
svc, err := c.GetService[*usecase.UserService]()          // returns an error
svc := c.MustGetService[*usecase.UserService]()           // panics on failure

svc, err := c.GetKeyedService[Cache]("redis")
svc := c.MustGetKeyedService[Cache]("redis")
```

Resolution failures are the sentinels `di.ErrServiceNotFound` and
`di.ErrServiceTypeMismatch`. The `MustGet*` accessors panic — as their name says — with the
exact error the matching `Get*` method returns, so a recovering caller can inspect it with
`errors.Is` and `errors.As`.

## Container construction errors

`di.NewContainer` never panics. It reports in two phases and returns the first that fails.

**Registration.** A malformed option — a factory that is not a function, a factory returning
no value or a non-error second value, a return type or an instance not assignable to the
service type, a nil value — is returned as a `*di.RegistrationError`. It aggregates every
malformed registration in the call, and each fault names the option index, the types
involved, and the `file:line` of the `di.With*` call that created it. Each fault wraps one of
the exported sentinels:

```go
var invalid *di.InvalidRegistrationError

if errors.Is(err, di.ErrFactoryNotFunction) { /* classify */ }
if errors.As(err, &invalid) { /* read Site, ServiceType, ValueType */ }
```

| Sentinel | Cause |
| --- | --- |
| `ErrNilRegistration` | the factory or instance argument was nil |
| `ErrFactoryNotFunction` | the factory argument is not a function |
| `ErrFactoryNoReturn` | the factory returns no value |
| `ErrFactoryTooManyReturns` | the factory returns more than two values |
| `ErrFactorySecondReturnNotErr` | the factory's second return value is not an `error` |
| `ErrNotAssignable` | the return type or instance is not assignable to the service type |

**Verification.** Once every registration is sound, the dependency graph is checked, and
every missing dependency and every cycle is returned as a `*di.VerificationError`
aggregating `*di.MissingDependencyError` and `*di.CircularDependencyError` faults.

Verification is skipped when registration fails: a rejected registration is absent from the
graph and would report dependencies missing only because of it.

## Status

`di` is the foundation the rest of the framework will be built on. Routing, application
lifecycle, and modules are next. The container's API is settling but not yet frozen — pin a
version.

## License

[MIT](./LICENSE)
