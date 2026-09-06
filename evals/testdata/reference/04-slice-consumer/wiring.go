package main

import "github.com/akim0v/ordo"

func BuildContainer() (*ordo.Container, error) {
	return ordo.New(
		ordo.WithService[Notifier](NewEmailNotifier),
		ordo.WithService[Notifier](NewSMSNotifier),
		ordo.WithFactory(NewDispatcher),
	)
}
