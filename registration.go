package ordo

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

var (
	// ErrNilRegistration is the cause of a registration fault reporting a nil
	// factory or instance argument
	ErrNilRegistration = errors.New("ordo: registration value is nil")

	// ErrFactoryNotFunction is the cause of a registration fault reporting a
	// factory argument that is not a function
	ErrFactoryNotFunction = errors.New("ordo: service factory must be a function")

	// ErrFactoryNoReturn is the cause of a registration fault reporting a
	// factory returning no value
	ErrFactoryNoReturn = errors.New("ordo: service factory must return at least one value")

	// ErrFactoryTooManyReturns is the cause of a registration fault reporting a
	// factory returning more than two values
	ErrFactoryTooManyReturns = errors.New("ordo: service factory returns too many values")

	// ErrFactorySecondReturnNotErr is the cause of a registration fault
	// reporting a factory whose second return value is not an error
	ErrFactorySecondReturnNotErr = errors.New("ordo: second service factory return value must be an error")

	// ErrNotAssignable is the cause of a registration fault reporting a factory
	// return type or an instance type that is not assignable to the service type
	ErrNotAssignable = errors.New("ordo: value is not assignable to the service type")
)

// callSite is the source location of the call that created a registration
type callSite struct {
	// File is the absolute path of the source file
	File string

	// Line is the line within the file
	Line int
}

// newCallSite captures a call site from the current stack.
//
// skip is the number of frames to ascend above newCallSite, so skip 0 reports
// the caller of newCallSite and skip 1 reports that caller's caller.
func newCallSite(skip int) callSite {
	_, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return callSite{}
	}

	return callSite{
		File: file,
		Line: line,
	}
}

// IsZero reports whether the call site was not captured
func (site callSite) IsZero() bool {
	return site.File == ""
}

// String renders the call site as the source file base name and its line
func (site callSite) String() string {
	if site.IsZero() {
		return "unknown source"
	}

	return fmt.Sprintf("%s:%d", filepath.Base(site.File), site.Line)
}

// InvalidRegistrationError is a container registration fault reporting a
// registration whose arguments cannot produce a usable service.
//
// It wraps one of the registration cause sentinels, so errors.Is reports the
// cause of a fault, and errors.As reads the registration it came from.
type InvalidRegistrationError struct {
	// Index is the position of the registration among the options passed to
	// New
	Index int

	// ServiceType is the declared service type of the registration.
	// Nil if the service type is inferred from the factory return type
	ServiceType reflect.Type

	// ValueType is the type of the offending factory or instance.
	// Nil if the registration argument was nil
	ValueType reflect.Type

	// Site is the source location of the call that created the registration
	Site callSite

	// Err is the cause of the fault
	Err error
}

// Error implements the error interface for InvalidRegistrationError.
func (e *InvalidRegistrationError) Error() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "option %d", e.Index)

	if !e.Site.IsZero() {
		fmt.Fprintf(&sb, " at %s", e.Site)
	}

	if e.ServiceType != nil {
		fmt.Fprintf(&sb, ", service %q", e.ServiceType)
	}

	// The cause is already prefixed with the package name by the aggregate
	// header, so it is trimmed here to keep the fault line readable
	fmt.Fprintf(&sb, ": %s", strings.TrimPrefix(e.Err.Error(), "ordo: "))

	if e.ValueType != nil {
		fmt.Fprintf(&sb, ", got %q", e.ValueType)
	}

	return sb.String()
}

// Unwrap returns the cause of the fault
func (e *InvalidRegistrationError) Unwrap() error {
	return e.Err
}

// RegistrationError is the error returned by New when one or more
// registrations are malformed.
//
// It aggregates every fault found while applying the options, the same way
// VerificationError aggregates the faults of the registration graph. Faults are
// reachable with errors.As and errors.Is, which traverse all of them.
//
// A container reporting registration faults is never verified, since a
// malformed registration is dropped from the graph and verifying it would
// report dependencies missing only because of the dropped registration.
type RegistrationError struct {
	// Faults is every fault found while applying the Container options.
	// Each entry is an *InvalidRegistrationError
	Faults []error
}

// Error implements the error interface for RegistrationError.
func (e *RegistrationError) Error() string {
	var sb strings.Builder
	sb.WriteString("ordo: container registration failed:")

	for _, fault := range e.Faults {
		sb.WriteString("\n  - ")
		sb.WriteString(fault.Error())
	}

	return sb.String()
}

// Unwrap returns every aggregated fault, so errors.As and errors.Is inspect all
// of them.
func (e *RegistrationError) Unwrap() []error {
	return e.Faults
}
