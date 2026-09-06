# Ordo

**Ordo** is a dependency injection container for Go. Wiring is explicit, dependencies are
read from your constructors, and the whole graph is verified before the container hands you
anything.

## Why Ordo

Dependency injection in Go usually asks you to accept some magic. Struct tags a reflection
engine reads at runtime. Generated files you do not own rewriting your wiring. String names
where a typo becomes a failure at startup. Each of them moves knowledge out of the type
system and into a convention you have to remember.

Ordo takes the opposite position. Registration is a plain generic call:

```go
ordo.WithService[UserRepository](NewPostgresUserRepository)
```

The compiler already knows the service type and the constructor. No tags, no code
generation, no build step, and no string names unless you deliberately ask for one.
Dependencies are read from the constructor's own parameter types.

Because the container knows the entire graph before it resolves a single service, it checks
that graph up front. `ordo.New` validates every registration, then verifies every edge, and
returns what it found — a malformed registration, a missing dependency, a cycle — at the call
site, with the `file:line` of the registration responsible.

That is the guarantee the name stands for. **A container either fails to build, or is fully
resolvable.** Nothing is deferred to the first request in production.

In practice:

- No code generation and no build step.
- No struct tags and no reflection-driven field injection.
- `ordo.New` never panics. Failures are ordinary Go errors, classified with `errors.Is` and
  `errors.As`.
- The dependency graph is verified once, at construction.

## Installation

Ordo requires **Go 1.27 or newer**.

```bash
go get -u github.com/akim0v/ordo
```

Services are resolved through generic methods — `c.GetService[T]()` — and a method could
not declare its own type parameters until Go 1.27. On an older toolchain that call fails to
compile at *your* call site, usually as `syntax error: method must have no type
parameters`, which reads like a broken library rather than a version mismatch. If you see
it, check `go version`. There is no build-tag fallback: the generic-method form is the API.

## Quickstart

This program compiles and runs as written.

```go
package main

import (
	"fmt"
	"log"

	"github.com/akim0v/ordo"
)

// Config is built by the application and handed to the container as a value.
type Config struct {
	DSN string
}

// UserRepository is the interface the service layer depends on.
type UserRepository interface {
	Name() string
}

// PostgresUserRepository implements UserRepository and needs a *Config.
type PostgresUserRepository struct {
	dsn string
}

func NewPostgresUserRepository(cfg *Config) *PostgresUserRepository {
	return &PostgresUserRepository{dsn: cfg.DSN}
}

func (r *PostgresUserRepository) Name() string { return "postgres(" + r.dsn + ")" }

// UserService depends on the interface, never on the implementation.
type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository: repository}
}

func (s *UserService) RepositoryName() string { return s.repository.Name() }

func main() {
	c, err := ordo.New(
		// A ready value, registered under its own type, *Config.
		ordo.WithValue(&Config{DSN: "localhost"}),

		// A constructor bound to the interface it satisfies. Its *Config
		// parameter is supplied from the registration above.
		ordo.WithService[UserRepository](NewPostgresUserRepository),

		// A constructor registered under its own return type, *UserService.
		ordo.WithFactory(NewUserService),
	)
	if err != nil {
		// The graph was verified before this point, so this error names the
		// exact registration to change.
		log.Fatal(err)
	}

	service := c.MustGetService[*UserService]()
	fmt.Println(service.RepositoryName())
	// Output: postgres(localhost)
}
```

Every Go block in this README is compiled by `docscheck_test.go` in this repository, so
none of them can rot.

Runnable versions of this and other setups live in [`examples/`](./examples).

## Registering services

| Option | Service type | Source |
| --- | --- | --- |
| `ordo.WithService[T](factoryOrInstance)` | `T`, given explicitly | a constructor or a ready value |
| `ordo.WithFactory(factory)` | the factory's return type | a constructor |
| `ordo.WithValue(value)` | the value's own type | a ready value |

`WithService` accepts either form and tells them apart by kind, so it is the option to reach
for when the service type differs from the concrete type — registering an implementation
against an interface, most often.

Each option has a keyed counterpart — `WithKeyedService`, `WithKeyedFactory`,
`WithKeyedValue` — for the case where one type has several registrations that callers need
to tell apart by name. Keys are a call-site facility: a factory parameter is always resolved
without a key, so a keyed registration never satisfies a constructor.

A constructor may return `(T, error)`. The container propagates a construction failure to
the caller resolving that service.

### Multiple implementations

Registering the same service type more than once is not a conflict. Ordo collects the
registrations, and a dependency declared as a slice receives all of them, in registration
order:

```go
// Given a second implementation, NewCacheUserRepository, and a consumer
// NewReport whose parameter is []UserRepository:
c, err := ordo.New(
	ordo.WithValue(&Config{DSN: "localhost"}),
	ordo.WithService[UserRepository](NewPostgresUserRepository),
	ordo.WithService[UserRepository](NewCacheUserRepository),

	// NewReport receives both implementations, in registration order.
	ordo.WithFactory(NewReport),
)
if err != nil {
	log.Fatal(err)
}

fmt.Println(len(c.MustGetService[*Report]().Repositories()))
```

The same works at the call site: `c.MustGetService[[]UserRepository]()` returns every
registration, while `c.MustGetService[UserRepository]()` returns the last one.

## Resolving services

```go
service, err := c.GetService[*UserService]() // returns an error
fmt.Println(service.RepositoryName(), err)

service = c.MustGetService[*UserService]() // panics on failure
fmt.Println(service.RepositoryName())

// Every registration of a type, in registration order.
fmt.Println(len(c.MustGetService[[]UserRepository]()))

// A keyed registration is resolved by name.
cached, err := c.GetKeyedService[UserRepository]("cache")
fmt.Println(cached, err)
```

Services are resolved lazily and each registration is constructed once.

Resolution failures are the sentinels `ordo.ErrServiceNotFound` and
`ordo.ErrServiceTypeMismatch`. The `MustGet*` accessors panic — as their name says — with the
exact error the matching `Get*` method returns, so a recovering caller can inspect it with
`errors.Is` and `errors.As`.

## Container construction errors

`ordo.New` never panics. It reports in two phases and returns the first that fails.

**Registration.** A malformed option — a factory that is not a function, a factory returning
no value or a non-error second value, a return type or an instance not assignable to the
service type, a nil value — is returned as a `*ordo.RegistrationError`. It aggregates every
malformed registration in the call, and each fault names the option index, the types
involved, and the `file:line` of the `ordo.With*` call that created it. Each fault wraps one
of the exported sentinels:

```go
// Classify the cause with a sentinel.
if errors.Is(err, ordo.ErrFactoryNotFunction) {
	// a non-function was passed where a constructor was expected
}

// Read the fault itself. errors.AsType is the Go 1.26+ form and is what this
// package uses internally; it needs no out-parameter.
if invalid, ok := errors.AsType[*ordo.InvalidRegistrationError](err); ok {
	fmt.Println(invalid.Index, invalid.Site, invalid.ServiceType, invalid.ValueType)
}
```

| Sentinel | Cause |
| --- | --- |
| `ErrNilRegistration` | the factory or instance argument was nil |
| `ErrFactoryNotFunction` | the factory argument is not a function |
| `ErrFactoryNoReturn` | the factory returns no value |
| `ErrFactoryTooManyReturns` | the factory returns more than two values |
| `ErrFactorySecondReturnNotErr` | the factory's second return value is not an `error` |
| `ErrNotAssignable` | the return type or instance is not assignable to the service type |
| `ErrValueIsFunction` | a constructor was registered as a value, making the service type the function type |

**Verification.** Once every registration is sound, the dependency graph is checked, and
every missing dependency and every cycle is returned as a `*ordo.VerificationError`
aggregating `*ordo.MissingDependencyError` and `*ordo.CircularDependencyError` faults.

Verification is skipped when registration fails: a rejected registration is absent from the
graph and would report dependencies missing only because of it.

## Status

Ordo does one thing — dependency injection — and is not growing into a web framework. The
API is settling but not yet frozen, so pin a version.

## License

[MIT](./LICENSE)
