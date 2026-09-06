# Task

`domain.go` declares a `Notifier` interface with two implementations.

Write `wiring.go` in this package. It must declare:

```go
func BuildContainer() (*ordo.Container, error)
func All(c *ordo.Container) ([]Notifier, error)
func Last(c *ordo.Container) (Notifier, error)
```

`BuildContainer` registers the email notifier and then the SMS notifier, both
as `Notifier`. `All` returns every registration in registration order. `Last`
returns the one that would satisfy a plain, non-slice request.

Do not edit `domain.go` or `verify_test.go`.
