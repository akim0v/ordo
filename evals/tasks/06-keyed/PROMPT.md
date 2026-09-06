# Task

`domain.go` declares a `Cache` interface with two implementations.

Write `wiring.go` in this package. It must declare:

```go
func BuildContainer() (*ordo.Container, error)
func Get(c *ordo.Container, key string) (Cache, error)
```

Both implementations must be registered as `Cache` under the names `"redis"`
and `"memory"`, and `Get` must return the one matching the key it is given.

Do not edit `domain.go` or `verify_test.go`.
