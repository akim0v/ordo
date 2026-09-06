// Example: factory_error
//
// A constructor may return (T, error). The graph here is valid, so the
// container builds; the failure happens later, when the service is resolved and
// the failing constructor actually runs. The error reaches the caller wrapped in
// a *di.DependencyError naming both types involved.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/akim0v/ordo/di"
)

func main() {
	// The graph is sound: UserService needs a UserRepository, and one is
	// registered. Construction succeeds.
	c, err := di.NewContainer(
		di.WithService[UserRepository](NewUserRepositoryImpl),
		di.WithFactory(NewUserService),
	)
	if err != nil {
		log.Fatalf("could not create the container: %s", err)
	}

	// Resolution runs NewUserRepositoryImpl, which fails.
	service, err := c.GetService[*UserService]()

	fmt.Println(service == nil) // true
	fmt.Println(err)
	// di: failed to create dependency "main.UserRepository" for service
	// "*main.UserService": database is unreachable

	// The cause is preserved, so the original sentinel is still classifiable.
	fmt.Println(errors.Is(err, ErrNoDatabase)) // true

	// The wrapper names which dependency failed and who asked for it.
	var depErr *di.DependencyError
	if errors.As(err, &depErr) {
		fmt.Printf("%s could not be built for %s\n", depErr.DependencyType, depErr.RequestingType)
	}
}
