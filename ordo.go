// Package ordo is a dependency injection container for Go.
//
// Registration is explicit and generic, dependencies are read from constructor
// parameter types, and the whole dependency graph is verified before New
// returns. A container either fails to build, or is fully resolvable.
//
//	c, err := ordo.New(
//		ordo.WithService[UserRepository](storage.NewUserRepository),
//		ordo.WithFactory(usecase.NewUserService),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	service := c.MustGetService[*usecase.UserService]()
//
// New never panics. A malformed option is returned as a *RegistrationError and
// a broken graph as a *VerificationError, both aggregating every fault of the
// call.
package ordo

import (
	"errors"
	"fmt"
	"reflect"
)

// Container is a service container
type Container struct {
	// accessors is a map for service identifiers of service descriptors lists
	accessors serviceAccessors
}

// newContainer creates a new Container with the provided options applied and
// returns every registration fault the options reported.
//
// An option reporting a fault registers nothing, and the registration graph is
// not verified, so the returned Container may be structurally broken.
// New is the verified entry point
func newContainer(opts ...Option) (*Container, []error) {
	c := &Container{
		accessors: make(serviceAccessors),
	}

	var faults []error

	for i, opt := range opts {
		err := opt.apply(c)
		if err == nil {
			continue
		}

		// Only this loop knows the position of the option among the ones the
		// caller passed, so the index is tagged here
		if invalid, ok := errors.AsType[*InvalidRegistrationError](err); ok {
			invalid.Index = i
		}

		faults = append(faults, err)
	}

	// The Container registers itself. This registration is built here from a
	// non-nil *Container under its own type, so it cannot fail validation and
	// never takes part in the caller-relative fault indexing above
	_ = WithValue(c).apply(c)

	return c, faults
}

// New creates a new Container with the provided options, validating
// every registration and then verifying the registration graph.
//
// Registration runs first and reports every malformed option it finds as a
// *RegistrationError. A malformed option registers nothing, so verification is
// skipped when registration fails, rather than reporting dependencies missing
// only because of a dropped registration.
//
// Verification runs once every option is applied, and reports every missing
// dependency and every dependency cycle it finds as a *VerificationError. It
// calls no factory and creates no service instance, so services are still
// created lazily on first resolution.
//
// Returns a nil Container if either phase fails.
func New(opts ...Option) (*Container, error) {
	c, faults := newContainer(opts...)
	if len(faults) > 0 {
		return nil, &RegistrationError{
			Faults: faults,
		}
	}

	if err := c.verify(); err != nil {
		return nil, err
	}

	return c, nil
}

// appendAccessor appends a service accessor to the container to the provided id
// creates a new serviceAccessorsList if the id does not exist
func (c *Container) appendAccessor(id serviceIdentifier, accessor *serviceAccessor) {
	if l, ok := c.accessors[id]; ok {
		l.Append(accessor)
	} else {
		c.accessors[id] = newServiceAccessorsList(accessor)
	}
}

// resolveFactoryDeps returns a slice of service dependency for the provided factory
func (c *Container) resolveFactoryDeps(factory *serviceFactory) ([]reflect.Value, error) {
	serviceDeps := make([]reflect.Value, factory.DepsCount)

	for i := 0; i < factory.DepsCount; i++ {
		depType := factory.Type.In(i)
		depID := newDependencyIdentifier(depType)

		dep, err := c.getService(depID)
		if err != nil {
			return nil, &DependencyError{
				RequestingType: factory.ReturnType,
				DependencyType: depType,
				Err:            err,
			}
		}

		serviceDeps[i] = dep
	}

	return serviceDeps, nil
}

// getService gets a service instance for the provided service identifier
func (c *Container) getService(id serviceIdentifier) (reflect.Value, error) {
	id, isSlice := id.resolveLookup()

	accessors, ok := c.accessors[id]
	if !ok {
		if isSlice {
			return reflect.Zero(reflect.SliceOf(id.Type)), ErrServiceNotFound
		}
		return reflect.Zero(id.Type), ErrServiceNotFound
	}

	if !isSlice {
		accessor := accessors.Last()
		return accessor.Instance()
	}

	slTyp := reflect.SliceOf(id.Type)
	res := reflect.MakeSlice(slTyp, accessors.Len(), accessors.Len())
	for i, accessor := range accessors.Iter() {
		instance, err := accessor.Instance()
		if err != nil {
			return reflect.Zero(slTyp), err
		}

		res.Index(i).Set(instance)
	}

	return res, nil
}

// getServiceKey returns an asserted service instance
// for the provided type and key
func (c *Container) getServiceKey[T any](key *string) (T, error) {
	var zero T

	id := newServiceIdentifier(reflect.TypeFor[T](), key)
	service, err := c.getService(id)
	if err != nil {
		// The failed resolution holds the zero reflect.Value of the service
		// type, which is an untyped nil for an interface service type and
		// cannot be asserted. The zero value of T is returned instead
		return zero, err
	}

	instance, ok := service.Interface().(T)
	if !ok {
		return zero, fmt.Errorf("%w: resolved %v for the service type %v",
			ErrServiceTypeMismatch, service.Type(), reflect.TypeFor[T]())
	}

	return instance, nil
}

// GetService returns the asserted service instance for the provided type
func (c *Container) GetService[T any]() (T, error) {
	return c.getServiceKey[T](nil)
}

// MustGetService returns the asserted service instance for the provided type
//
// Panics with the resolution error if no service is found or any error occurred
// while creating the instance
func (c *Container) MustGetService[T any]() T {
	service, err := c.GetService[T]()
	if err != nil {
		panic(err)
	}

	return service
}

// GetKeyedService returns the asserted service instance for the provided type and key
func (c *Container) GetKeyedService[T any](key string) (T, error) {
	return c.getServiceKey[T](&key)
}

// MustGetKeyedService returns the asserted service instance for the provided type and key
//
// Panics with the resolution error if no service is found or any error occurred
// while creating the instance
func (c *Container) MustGetKeyedService[T any](key string) T {
	service, err := c.GetKeyedService[T](key)
	if err != nil {
		panic(err)
	}

	return service
}
