package main

import "github.com/akim0v/ordo"

func BuildContainer() (*ordo.Container, error) {
	return BuildContainerWith(NewRealStore())
}

func BuildContainerWith(store Store) (*ordo.Container, error) {
	return ordo.New(
		ordo.WithService[Store](store),
		ordo.WithFactory(NewService),
	)
}
