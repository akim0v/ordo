package ordo

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrServiceNotFound is the error returned when a requested service is not found is service provider
	ErrServiceNotFound = errors.New("ordo: requested service not found")

	// ErrServiceTypeMismatch is the error returned when a resolved service is
	// not assignable to the requested type.
	//
	// Registration rejects a service that is not assignable to its service
	// type, so this error reports a Container bug rather than a caller mistake
	ErrServiceTypeMismatch = errors.New("ordo: resolved service is not assignable to the requested type")
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
		"ordo: failed to create dependency %q for service %q: %v",
		e.DependencyType,
		e.RequestingType,
		e.Err,
	)
}

// Unwrap returns the underlying error
func (e *DependencyError) Unwrap() error {
	return e.Err
}
