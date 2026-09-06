# Task

`domain.go` declares a `*Config` value type, a `Store` interface with an
implementation, and a `Service` whose constructor takes **both**.

Write `wiring.go` in this package. It must declare:

```go
func BuildContainer() (*ordo.Container, error)
```

The container must supply a `*Config` whose `Prefix` field is `"app"`, and
resolve `*Service`. Do not edit `domain.go` or `verify_test.go`.
