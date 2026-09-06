package main

import "github.com/akim0v/ordo"

func BuildContainer(fail bool) (*ordo.Container, error) {
	return ordo.New(
		ordo.WithValue(&Config{Fail: fail}),
		ordo.WithFactory(NewFlaky),
	)
}
