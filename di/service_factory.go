package di

import (
	"reflect"
)

// serviceFactory is a service factory function description
type serviceFactory struct {
	// Type is a factory type
	Type reflect.Type

	// Value is factory value
	Value reflect.Value

	// DepsCount is a number of the factory dependencies
	DepsCount int

	// ReturnType is the return type of the factory function
	ReturnType reflect.Type

	// HasErr is true if the factory returns an error as the second return argument
	HasErr bool
}

// newServiceFactory creates a new serviceFactory for the provided factory
// function, or returns the registration cause sentinel describing why the
// provided value cannot serve as a factory.
//
// The returned error is a bare cause: the registration applying the factory
// wraps it in an *InvalidRegistrationError, since only the registration knows
// the declared service type and the call site
func newServiceFactory(factory any) (*serviceFactory, error) {
	if factory == nil {
		return nil, ErrNilRegistration
	}

	val := reflect.ValueOf(factory)
	typ := val.Type()

	if typ.Kind() != reflect.Func {
		return nil, ErrFactoryNotFunction
	}

	numOut := typ.NumOut()
	switch numOut {
	case 0:
		return nil, ErrFactoryNoReturn
	case 1:
	case 2:
		if typ.Out(1) != reflect.TypeFor[error]() {
			return nil, ErrFactorySecondReturnNotErr
		}
	default:
		return nil, ErrFactoryTooManyReturns
	}

	return &serviceFactory{
		Type:       typ,
		Value:      val,
		DepsCount:  typ.NumIn(),
		ReturnType: typ.Out(0),
		HasErr:     numOut == 2,
	}, nil
}

// Call calls the factory function with the provided dependencies
func (factory *serviceFactory) Call(deps ...reflect.Value) (reflect.Value, error) {
	values := factory.Value.Call(deps)
	if factory.HasErr && !values[1].IsNil() {
		return reflect.Zero(factory.ReturnType), values[1].Interface().(error)
	}

	return values[0], nil
}
