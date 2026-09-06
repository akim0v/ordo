package main

import "github.com/akim0v/ordo"

func BuildContainer() (*ordo.Container, error) {
	return ordo.New(
		ordo.WithService[Greeter](NewEnglishGreeter),
	)
}
