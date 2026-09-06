# Task

`domain.go` declares a constructor that can fail: `NewFlaky` returns
`(*Flaky, error)` and returns `ErrFlaky` when its `*Config` has `Fail` set.

Write `wiring.go` in this package. It must declare:

```go
func BuildContainer(fail bool) (*ordo.Container, error)
```

It registers a `*Config` whose `Fail` field is the argument, plus `NewFlaky`.
Building the container must succeed in **both** cases — the constructor's
failure is a resolution-time failure, not a wiring error.

Do not edit `domain.go` or `verify_test.go`.
