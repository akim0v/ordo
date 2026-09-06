// Example: multiple_implementations
//
// Registering one service type more than once is not a conflict. The container
// keeps every registration, and a dependency declared as a slice receives all
// of them in registration order.
package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/akim0v/ordo"
)

func main() {
	c, err := ordo.New(
		// Two implementations of the same interface.
		ordo.WithService[UserRepository](NewCacheRepositoryImpl),
		ordo.WithService[UserRepository](NewDBRepositoryImpl),

		// NewUserService takes []UserRepository and receives both, in the order
		// they were registered above.
		ordo.WithFactory(NewUserService),
	)
	if err != nil {
		log.Fatalf("could not create the container: %s", err)
	}

	service := c.MustGetService[*UserService]()

	fmt.Println("\n=== Creating user ===")
	userID := service.CreateUser("Akim")

	fmt.Println("\n=== Fetching user (cache first, then DB) ===")
	user := service.GetUser(userID)

	fmt.Println("\n=== Fetching non-existent user ===")
	missing := service.GetUser(rand.Int())

	fmt.Println("\n=== Results ===")
	fmt.Println(user.Name == "Akim") // true
	fmt.Println(missing == nil)      // true

	// Asking for the slice directly returns every registration; asking for the
	// bare type returns the last one.
	fmt.Println(len(c.MustGetService[[]UserRepository]())) // 2
}
