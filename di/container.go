package di

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrServiceNotFound is the error returned when a requested service is not found is service provider
	ErrServiceNotFound = errors.New("di: requested service not found")

	// ErrServiceTypeMismatch is the error returned when a resolved service is
	// not assignable to the requested type.
	//
	// Registration rejects a service that is not assignable to its service
	// type, so this error reports a Container bug rather than a caller mistake
	ErrServiceTypeMismatch = errors.New("di: resolved service is not assignable to the requested type")
)

// DependencyError is a custom error type for dependency injection failures.
type DependencyError struct {
	// DependencyType is the type of the dependency that failed to be created
	DependencyType reflect.Type

	// RequestingType is the type of the service requesting the dependency
	RequestingType reflect.Type

	// Err is the underlying error that occurred
	Err error
}

// Error implements the error interface for DependencyError.
func (e *DependencyError) Error() string {
	return fmt.Sprintf(
		"di: failed to create dependency %q for service %q: %v",
		e.DependencyType,
		e.RequestingType,
		e.Err,
	)
}

// Unwrap returns the underlying error
func (e *DependencyError) Unwrap() error {
	return e.Err
}

// Option is a struct adding a new service to the Container
type Option interface {
	// apply registers the option service in the Container, or returns the
	// registration fault that keeps it from being registered
	apply(*Container) error
}

// invalidOption is an Option rejected before it could describe a registration,
// so it registers nothing and only reports its fault
type invalidOption struct {
	// serviceType is the declared service type of the rejected registration
	serviceType reflect.Type

	// site is the source location of the call that created the registration
	site callSite

	// err is the cause of the fault
	err error
}

// apply applies the Option
func (opt *invalidOption) apply(*Container) error {
	return &InvalidRegistrationError{
		ServiceType: opt.serviceType,
		Site:        opt.site,
		Err:         opt.err,
	}
}

// serviceFactoryOption adds a new keyed service with a factory to the Container
type serviceFactoryOption struct {
	typ     reflect.Type
	key     *string
	factory any
	site    callSite
}

// apply applies the Option
func (opt *serviceFactoryOption) apply(c *Container) error {
	f, err := newServiceFactory(opt.factory)
	if err != nil {
		return &InvalidRegistrationError{
			ServiceType: opt.typ,
			ValueType:   reflect.TypeOf(opt.factory),
			Site:        opt.site,
			Err:         err,
		}
	}

	if !f.ReturnType.AssignableTo(opt.typ) {
		return &InvalidRegistrationError{
			ServiceType: opt.typ,
			ValueType:   f.ReturnType,
			Site:        opt.site,
			Err:         ErrNotAssignable,
		}
	}

	id := newServiceIdentifier(opt.typ, opt.key)
	accessor := newServiceAccessor(id, c, f, nil)
	c.appendAccessor(id, accessor)

	return nil
}

// withServiceFactory returns a new instance of serviceFactoryOption
func withServiceFactory[T any](key *string, factory any, site callSite) Option {
	return &serviceFactoryOption{
		typ:     reflect.TypeFor[T](),
		key:     key,
		factory: factory,
		site:    site,
	}
}

// serviceInstanceOption adds a new keyed service with an instance to the Container
type serviceInstanceOption struct {
	typ      reflect.Type
	key      *string
	instance any
	site     callSite
}

// apply applies the Option
func (opt *serviceInstanceOption) apply(c *Container) error {
	instVal := reflect.ValueOf(opt.instance)
	if !instVal.IsValid() {
		return &InvalidRegistrationError{
			ServiceType: opt.typ,
			Site:        opt.site,
			Err:         ErrNilRegistration,
		}
	}

	instTyp := instVal.Type()
	if !instTyp.AssignableTo(opt.typ) {
		return &InvalidRegistrationError{
			ServiceType: opt.typ,
			ValueType:   instTyp,
			Site:        opt.site,
			Err:         ErrNotAssignable,
		}
	}

	id := newServiceIdentifier(opt.typ, opt.key)
	accessor := newServiceAccessor(id, c, nil, &instVal)
	c.appendAccessor(id, accessor)

	return nil
}

// withServiceInstance adds a new keyed service with an instance to the Container
func withServiceInstance[T any](key *string, instance any, site callSite) Option {
	return &serviceInstanceOption{
		typ:      reflect.TypeFor[T](),
		key:      key,
		instance: instance,
		site:     site,
	}
}

// withServiceKey adds service to the Container with the provided key
func withServiceKey[T any](key *string, factoryOrInstance any, site callSite) Option {
	// The kind of nil argument cannot be read, so it is rejected before the
	// factory and instance forms are told apart
	if factoryOrInstance == nil {
		return &invalidOption{
			serviceType: reflect.TypeFor[T](),
			site:        site,
			err:         ErrNilRegistration,
		}
	}

	if reflect.TypeOf(factoryOrInstance).Kind() == reflect.Func {
		return withServiceFactory[T](key, factoryOrInstance, site)
	} else {
		return withServiceInstance[T](key, factoryOrInstance, site)
	}
}

// WithService adds a new service to the Container with the provided factory or instance
func WithService[T any](factoryOrInstance any) Option {
	return withServiceKey[T](nil, factoryOrInstance, newCallSite(1))
}

// WithKeyedService adds a new keyed service to the Container with the provided factory or instance
func WithKeyedService[T any](key string, factoryOrInstance any) Option {
	return withServiceKey[T](&key, factoryOrInstance, newCallSite(1))
}

// WithValue adds a new value to the Container with the provided value
// Same as the WithService[T](value), but typed
func WithValue[T any](value T) Option {
	return withServiceKey[T](nil, value, newCallSite(1))
}

// WithKeyedValue adds a new keyed value to the Container with the provided value
// Same as the WithKeyedService[T](key, value), but typed
func WithKeyedValue[T any](key string, value T) Option {
	return withServiceKey[T](&key, value, newCallSite(1))
}

// factoryOption adds a new service factory to the Container
type factoryOption struct {
	factory any
	key     *string
	site    callSite
}

// apply applies the Option
func (opt *factoryOption) apply(c *Container) error {
	f, err := newServiceFactory(opt.factory)
	if err != nil {
		// The service type is inferred from the factory return type, so a
		// rejected factory has no service type to report
		return &InvalidRegistrationError{
			ValueType: reflect.TypeOf(opt.factory),
			Site:      opt.site,
			Err:       err,
		}
	}

	id := newServiceIdentifier(f.ReturnType, opt.key)
	accessor := newServiceAccessor(id, c, f, nil)
	c.appendAccessor(id, accessor)

	return nil
}

// WithFactory adds a new service factory to the Container
func WithFactory(factory any) Option {
	return &factoryOption{
		factory: factory,
		site:    newCallSite(1),
	}
}

// WithKeyedFactory adds a new keyed service factory to the Container
func WithKeyedFactory(key string, factory any) Option {
	return &factoryOption{
		factory: factory,
		key:     &key,
		site:    newCallSite(1),
	}
}

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
// NewContainer is the verified entry point
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

// NewContainer creates a new Container with the provided options, validating
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
func NewContainer(opts ...Option) (*Container, error) {
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
