package ordo_test

import (
	"errors"
	"fmt"

	"github.com/akim0v/ordo"
)

// ExampleContainer_GetService resolves a service and reports a failure as an
// ordinary error.
func ExampleContainer_GetService() {
	c, _ := ordo.New(
		ordo.WithService[UserRepository](NewCacheRepository),
	)

	repository, err := c.GetService[UserRepository]()
	fmt.Println(repository.Name(), err)

	// A type that was never registered is not a panic.
	_, err = c.GetService[*Config]()
	fmt.Println(errors.Is(err, ordo.ErrServiceNotFound))

	// Output:
	// cache <nil>
	// true
}

// ExampleContainer_GetService_multipleImplementations shows that registering one
// service type more than once is not a conflict. Asking for the slice returns
// every registration in registration order; asking for the bare type returns the
// last one.
func ExampleContainer_GetService_multipleImplementations() {
	c, _ := ordo.New(
		ordo.WithValue(&Config{DSN: "localhost"}),
		ordo.WithService[UserRepository](NewCacheRepository),
		ordo.WithService[UserRepository](NewPostgresRepository),
		ordo.WithFactory(NewReportService),
	)

	all := c.MustGetService[[]UserRepository]()
	for _, repository := range all {
		fmt.Println(repository.Name())
	}

	fmt.Println(c.MustGetService[UserRepository]().Name())
	fmt.Println(c.MustGetService[*ReportService]().Count())

	// Output:
	// cache
	// postgres(localhost)
	// postgres(localhost)
	// 2
}

// ExampleContainer_MustGetService resolves a service directly. It panics with
// the exact error GetService would have returned, so a recovering caller can
// still classify it.
func ExampleContainer_MustGetService() {
	c, _ := ordo.New(
		ordo.WithService[UserRepository](NewCacheRepository),
	)

	fmt.Println(c.MustGetService[UserRepository]().Name())

	defer func() {
		err, ok := recover().(error)
		fmt.Println(ok, errors.Is(err, ordo.ErrServiceNotFound))
	}()

	c.MustGetService[*Config]()

	// Output:
	// cache
	// true true
}

// ExampleContainer_GetKeyedService resolves a registration by its key.
func ExampleContainer_GetKeyedService() {
	c, _ := ordo.New(
		ordo.WithKeyedService[UserRepository]("cache", NewCacheRepository),
	)

	repository, err := c.GetKeyedService[UserRepository]("cache")
	fmt.Println(repository.Name(), err)

	// A key that was never registered, and the same type without its key, are
	// both plain lookup misses.
	_, err = c.GetKeyedService[UserRepository]("postgres")
	fmt.Println(errors.Is(err, ordo.ErrServiceNotFound))

	_, err = c.GetService[UserRepository]()
	fmt.Println(errors.Is(err, ordo.ErrServiceNotFound))

	// Output:
	// cache <nil>
	// true
	// true
}

// ExampleContainer_MustGetKeyedService resolves a keyed registration directly.
func ExampleContainer_MustGetKeyedService() {
	c, _ := ordo.New(
		ordo.WithKeyedValue("replica", &Config{DSN: "replica"}),
	)

	fmt.Println(c.MustGetKeyedService[*Config]("replica").DSN)

	// Output:
	// replica
}

// ExampleDependencyError shows a constructor that fails. The graph is sound, so
// the container builds; the failure happens when the service is resolved and the
// constructor actually runs.
func ExampleDependencyError() {
	unreachable := errors.New("database is unreachable")

	c, _ := ordo.New(
		ordo.WithService[UserRepository](func() (*CacheRepository, error) {
			return nil, unreachable
		}),
		ordo.WithFactory(NewUserService),
	)

	_, err := c.GetService[*UserService]()
	fmt.Println(err)

	// The cause is preserved, and the wrapper names both types involved.
	fmt.Println(errors.Is(err, unreachable))

	var dependency *ordo.DependencyError
	if errors.As(err, &dependency) {
		fmt.Println(dependency.DependencyType, dependency.RequestingType)
	}

	// Output:
	// ordo: failed to create dependency "ordo_test.UserRepository" for service "*ordo_test.UserService": database is unreachable
	// true
	// ordo_test.UserRepository *ordo_test.UserService
}

// ExampleVerificationError shows the aggregate returned when the dependency
// graph is broken. Every fault of the pass is reported, not just the first.
func ExampleVerificationError() {
	type Orphan struct{}

	_, err := ordo.New(
		// Needs a UserRepository that is never registered.
		ordo.WithFactory(NewUserService),
		// Needs an *Orphan that is never registered.
		ordo.WithFactory(func(*Orphan) *Config { return nil }),
	)

	var verification *ordo.VerificationError
	if errors.As(err, &verification) {
		fmt.Println(len(verification.Faults))
		for _, fault := range verification.Faults {
			fmt.Println(fault)
		}
	}

	// Output:
	// 2
	// service "*ordo_test.Config" requires "*ordo_test.Orphan", which is not registered
	// service "*ordo_test.UserService" requires "ordo_test.UserRepository", which is not registered
}

// ExampleMissingDependencyError names the service that asked and the dependency
// that was absent. It wraps ErrServiceNotFound, the sentinel a failed runtime
// resolution of the same dependency would return.
func ExampleMissingDependencyError() {
	_, err := ordo.New(
		ordo.WithFactory(NewUserService),
	)

	var missing *ordo.MissingDependencyError
	if errors.As(err, &missing) {
		fmt.Println(missing.RequestingType)
		fmt.Println(missing.DependencyType)
	}

	fmt.Println(errors.Is(err, ordo.ErrServiceNotFound))

	// Output:
	// *ordo_test.UserService
	// ordo_test.UserRepository
	// true
}

// ExampleCircularDependencyError carries the ordered cycle, so the offending
// edge is readable rather than guessed at.
func ExampleCircularDependencyError() {
	type Billing struct{}
	type Accounts struct{}

	_, err := ordo.New(
		ordo.WithFactory(func(*Accounts) *Billing { return nil }),
		ordo.WithFactory(func(*Billing) *Accounts { return nil }),
	)

	var cycle *ordo.CircularDependencyError
	if errors.As(err, &cycle) {
		for i, typ := range cycle.Cycle {
			fmt.Println(i, typ)
		}
	}

	// Output:
	// 0 *ordo_test.Accounts
	// 1 *ordo_test.Billing
}

// ExampleRegistrationError shows the first construction phase. Every malformed
// option of the call is aggregated, and verification never runs: a rejected
// registration is absent from the graph, and verifying it would report
// dependencies missing only because of the rejection.
func ExampleRegistrationError() {
	_, err := ordo.New(
		ordo.WithFactory("not a function"),
		ordo.WithFactory(func() {}),
		ordo.WithService[UserRepository](42),
	)

	var registration *ordo.RegistrationError
	if errors.As(err, &registration) {
		fmt.Println(len(registration.Faults))
	}

	fmt.Println(errors.Is(err, ordo.ErrFactoryNotFunction))
	fmt.Println(errors.Is(err, ordo.ErrFactoryNoReturn))
	fmt.Println(errors.Is(err, ordo.ErrNotAssignable))

	// Verification is skipped entirely.
	var verification *ordo.VerificationError
	fmt.Println(errors.As(err, &verification))

	// Output:
	// 3
	// true
	// true
	// true
	// false
}

// ExampleInvalidRegistrationError is a single registration fault. It names the
// option index, the types involved, and the source location of the ordo.With*
// call that created it.
func ExampleInvalidRegistrationError() {
	_, err := ordo.New(
		ordo.WithService[UserRepository](NewCacheRepository),
		ordo.WithService[UserRepository](42),
	)

	var invalid *ordo.InvalidRegistrationError
	if errors.As(err, &invalid) {
		fmt.Println(invalid.Index)
		fmt.Println(invalid.ServiceType)
		fmt.Println(invalid.ValueType)
		fmt.Println(errors.Is(invalid, ordo.ErrNotAssignable))
	}

	// Output:
	// 1
	// ordo_test.UserRepository
	// int
	// true
}
