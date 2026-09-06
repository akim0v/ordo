# Task

`domain.go` declares a `Notifier` interface with **two** implementations, and a
`Dispatcher` whose constructor takes `[]Notifier`.

Write `wiring.go` in this package. It must declare:

```go
func BuildContainer() (*ordo.Container, error)
```

`*Dispatcher` must resolve, having received both notifiers, with the email
notifier first. Do not edit `domain.go` or `verify_test.go`.
