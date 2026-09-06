// Example: layered_app
//
// A realistic wiring: config, storage, usecase, and transport, each layer
// depending only on the boundary below it. This is the shape most applications
// end up with, and the one to copy when starting a project.
//
// It also shows the two registration forms side by side — a ready value handed
// to the container, and constructors the container calls itself — plus a slice
// dependency collecting every controller.
package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/akim0v/ordo/di"
)

func main() {
	c, err := di.NewContainer(
		// Config is already built, so it is registered as a value under its own
		// type, *Config. Anything asking for a *Config receives this one.
		di.WithValue(LoadConfig()),

		// The implementation is bound to the interface the usecase layer
		// depends on, so nothing above storage names InMemoryUserRepository.
		di.WithService[UserRepository](NewInMemoryUserRepository),

		// Registered under its return type, *UserService.
		di.WithFactory(NewUserService),

		// Both controllers are bound to Controller, so both land in the
		// []Controller dependency below.
		di.WithService[Controller](NewUserController),
		di.WithService[Controller](NewHealthController),
	)
	if err != nil {
		// The graph was verified before this point. Reaching here means the
		// wiring is wrong, and the error says exactly how.
		log.Fatalf("could not create the container: %s", err)
	}

	mux := http.NewServeMux()

	// Every Controller registration, in registration order.
	for _, controller := range c.MustGetService[[]Controller]() {
		controller.Register(mux)
	}

	// Seed one user through the usecase layer.
	service := c.MustGetService[*UserService]()
	user := service.Create("Akim")

	fmt.Println()
	fmt.Println("GET /health   ->", request(mux, "/health"))
	fmt.Printf("GET /users/%d  -> %s\n", user.ID, request(mux, fmt.Sprintf("/users/%d", user.ID)))
	fmt.Println("GET /users/99 ->", request(mux, "/users/99"))
}

// request drives the mux in process, so the example needs no port.
func request(mux *http.ServeMux, target string) string {
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	return fmt.Sprintf("%d %s", rec.Code, strings.TrimSpace(rec.Body.String()))
}
