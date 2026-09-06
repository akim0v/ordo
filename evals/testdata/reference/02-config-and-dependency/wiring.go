package main

import "github.com/akim0v/ordo"

func BuildContainer() (*ordo.Container, error) {
	return ordo.New(
		ordo.WithValue(&Config{Prefix: "app"}),
		ordo.WithService[Store](NewMemStore),
		ordo.WithFactory(NewService),
	)
}
