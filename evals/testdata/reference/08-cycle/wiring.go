package main

import "github.com/akim0v/ordo"

func BuildContainer() (*ordo.Container, error) {
	return ordo.New(
		ordo.WithFactory(NewA),
		ordo.WithFactory(NewB),
	)
}
