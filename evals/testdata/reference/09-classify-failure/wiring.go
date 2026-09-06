package main

import (
	"errors"

	"github.com/akim0v/ordo"
)

func BadContainer() error {
	_, err := ordo.New(
		ordo.WithFactory("not a function"),
	)
	return err
}

func IsNotAFunction(err error) bool {
	return errors.Is(err, ordo.ErrFactoryNotFunction)
}

func FaultIndex(err error) (int, bool) {
	if invalid, ok := errors.AsType[*ordo.InvalidRegistrationError](err); ok {
		return invalid.Index, true
	}

	return 0, false
}
