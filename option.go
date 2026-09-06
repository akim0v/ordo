package ordo

import (
	"reflect"
)

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
		// The service type being the registered function's own type means it
		// was inferred from that function, which is what WithValue does. The
		// generic ErrNotAssignable message reads backwards here — it names the
		// func type as the service and the return type as the offending value —
		// so the cause names the option to use instead.
		if opt.typ.Kind() == reflect.Func && reflect.TypeOf(opt.factory) == opt.typ {
			return &InvalidRegistrationError{
				ServiceType: opt.typ,
				Site:        opt.site,
				Err:         ErrValueIsFunction,
			}
		}

		return &InvalidRegistrationError{
			ServiceType: opt.typ,
			ValueType:   f.ReturnType,
			Site:        opt.site,
			Err:         ErrNotAssignable,
		}
	}

	id := newServiceIdentifier(opt.typ, opt.key)
	accessor := newServiceAccessor(id, c, f, nil, opt.site)
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
	accessor := newServiceAccessor(id, c, nil, &instVal, opt.site)
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
	accessor := newServiceAccessor(id, c, f, nil, opt.site)
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
