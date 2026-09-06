# Task

`domain.go` declares a `Greeter` interface and an implementation with a
constructor `NewEnglishGreeter`.

Write `wiring.go` in this package. It must declare:

```go
func BuildContainer() (*ordo.Container, error)
```

It returns a container in which the `Greeter` interface resolves to the
implementation. Do not edit `domain.go` or `verify_test.go`.
