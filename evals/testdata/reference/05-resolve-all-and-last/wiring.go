package main

import "github.com/akim0v/ordo"

func BuildContainer() (*ordo.Container, error) {
	return ordo.New(
		ordo.WithService[Notifier](NewEmailNotifier),
		ordo.WithService[Notifier](NewSMSNotifier),
	)
}

func All(c *ordo.Container) ([]Notifier, error) {
	return c.GetService[[]Notifier]()
}

func Last(c *ordo.Container) (Notifier, error) {
	return c.GetService[Notifier]()
}
