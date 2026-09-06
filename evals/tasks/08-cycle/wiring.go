package main

import "github.com/akim0v/ordo"

// BuildContainer is broken: the container it describes cannot be built.
func BuildContainer() (*ordo.Container, error) {
	return ordo.New(
		ordo.WithFactory(NewA),
		ordo.WithFactory(NewBWithA),
	)
}
