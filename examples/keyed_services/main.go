// Example: keyed_services
//
// A key tells apart several registrations of one service type at the call site.
// Every registration option has a keyed counterpart, and keyed services are read
// back with GetKeyedService / MustGetKeyedService.
//
// The rule worth learning here: keys exist only at the call site. Factory
// parameters are always resolved without a key, so a keyed registration never
// satisfies a constructor.
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/akim0v/ordo/di"
)

func main() {
	selectingByKey()
	fmt.Println()
	keysDoNotInject()
	fmt.Println()
	bothAtOnce()
}

func selectingByKey() {
	fmt.Println("=== selecting by key ===")

	c, err := di.NewContainer(
		// Constructor, bound to the interface, under a key.
		di.WithKeyedService[Cache]("redis", NewRedisCache),

		// Constructor registered under its own return type, under a key.
		di.WithKeyedFactory("memory", NewMemoryCache),

		// An already-built value, under a key.
		di.WithKeyedValue[Cache]("null", &NullCache{}),
	)
	if err != nil {
		log.Fatalf("could not create the container: %s", err)
	}

	fmt.Println(c.MustGetKeyedService[Cache]("redis").Name())         // redis
	fmt.Println(c.MustGetKeyedService[*MemoryCache]("memory").Name()) // memory
	fmt.Println(c.MustGetKeyedService[Cache]("null").Name())          // null

	// A key that was never registered is a plain ErrServiceNotFound.
	_, err = c.GetKeyedService[Cache]("postgres")
	fmt.Println(errors.Is(err, di.ErrServiceNotFound)) // true

	// So is the same type without its key.
	_, err = c.GetService[Cache]()
	fmt.Println(errors.Is(err, di.ErrServiceNotFound)) // true
}

func keysDoNotInject() {
	fmt.Println("=== keys do not inject ===")

	// NewSessionStore takes a Cache. The only Cache here is keyed, and a keyed
	// registration does not satisfy a factory parameter — so this is a missing
	// dependency, caught at construction.
	_, err := di.NewContainer(
		di.WithKeyedService[Cache]("redis", NewRedisCache),
		di.WithFactory(NewSessionStore),
	)

	fmt.Println(err)
	fmt.Println(errors.Is(err, di.ErrServiceNotFound)) // true
}

func bothAtOnce() {
	fmt.Println("=== registering both ways ===")

	// Register unkeyed for injection, and keyed for explicit selection.
	c, err := di.NewContainer(
		di.WithService[Cache](NewRedisCache),
		di.WithKeyedService[Cache]("memory", NewMemoryCache),
		di.WithFactory(NewSessionStore),
	)
	if err != nil {
		log.Fatalf("could not create the container: %s", err)
	}

	// The unkeyed registration is what the constructor received.
	fmt.Println(c.MustGetService[*SessionStore]().Describe()) // ... backed by redis

	// The keyed one is still reachable by name.
	fmt.Println(c.MustGetKeyedService[Cache]("memory").Name()) // memory
}
