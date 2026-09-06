# Task

`domain.go` declares a `Store` interface, a real implementation, and a
`*Service` that depends on the interface.

Write `wiring.go` in this package. It must declare:

```go
func BuildContainer() (*ordo.Container, error)
func BuildContainerWith(store Store) (*ordo.Container, error)
```

`BuildContainer` wires the real store. `BuildContainerWith` wires the `Store`
it is given instead — this is how a test substitutes a mock — and in both cases
`*Service` must resolve.

Do not edit `domain.go` or `verify_test.go`.
