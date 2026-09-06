// Example: registration_errors
//
// The first phase of container construction. Before the dependency graph is
// walked, every option is checked for being a usable registration at all. Faults
// are aggregated — one call reports every malformed option, not just the first —
// and each names the option index, the types involved, and the file:line of the
// di.With* call responsible.
//
// NewContainer never panics. All of this arrives as an ordinary error value.
package main

import (
	"errors"
	"fmt"

	"github.com/akim0v/ordo/di"
)

type Greeter interface {
	Greet() string
}

func main() {
	_, err := di.NewContainer(
		// A factory that is not a function.
		di.WithFactory("not a function"),

		// A function returning nothing cannot produce a service.
		di.WithFactory(func() {}),

		// A second return value is allowed only if it is an error.
		di.WithFactory(func() (int, string) { return 0, "" }),

		// An int does not implement Greeter.
		di.WithService[Greeter](42),

		// Nothing to register.
		di.WithService[Greeter](nil),
	)

	// Every fault of the call, in one error.
	fmt.Println(err)
	fmt.Println()

	// Each cause is an exported sentinel, so faults classify with errors.Is
	// without matching on strings.
	fmt.Println("not a function: ", errors.Is(err, di.ErrFactoryNotFunction))
	fmt.Println("no return:      ", errors.Is(err, di.ErrFactoryNoReturn))
	fmt.Println("bad 2nd return: ", errors.Is(err, di.ErrFactorySecondReturnNotErr))
	fmt.Println("not assignable: ", errors.Is(err, di.ErrNotAssignable))
	fmt.Println("nil value:      ", errors.Is(err, di.ErrNilRegistration))
	fmt.Println()

	// The aggregate holds every fault.
	var registration *di.RegistrationError
	if errors.As(err, &registration) {
		fmt.Printf("faults: %d\n", len(registration.Faults))

		for _, fault := range registration.Faults {
			var invalid *di.InvalidRegistrationError
			if errors.As(fault, &invalid) {
				// Site is the di.With* call that created the registration, so a
				// fault points at your code rather than at the container.
				fmt.Printf("  option %d at %s: %s\n", invalid.Index, invalid.Site, invalid.Err)
			}
		}
	}

	// Verification is skipped entirely when registration fails: a rejected
	// registration is absent from the graph, and verifying it would report
	// dependencies missing only because of the rejection.
	var verification *di.VerificationError
	fmt.Println("\nverification ran:", errors.As(err, &verification)) // false
}
