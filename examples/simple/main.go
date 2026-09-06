// Example: simple
//
// The smallest working container. One interface bound to its implementation,
// one constructor registered under its own return type, one resolution.
package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/akim0v/ordo/di"
)

func main() {
	c, err := di.NewContainer(
		// UserRepository is the interface; NewUserRepositoryImpl is the
		// constructor that satisfies it.
		di.WithService[UserRepository](NewUserRepositoryImpl),

		// Registered under its return type, *UserService. Its UserRepository
		// parameter is supplied from the registration above.
		di.WithFactory(NewUserService),
	)
	if err != nil {
		log.Fatalf("could not create the container: %s", err)
	}

	service := c.MustGetService[*UserService]()

	userID := service.CreateUser("Akim")
	user := service.GetUser(userID)
	missing := service.GetUser(rand.Int())

	fmt.Println(user.Name == "Akim") // true
	fmt.Println(missing == nil)      // true
}
