# Task

Write `wiring.go` in this package. It must declare:

```go
func BadContainer() error
func IsNotAFunction(err error) bool
func FaultIndex(err error) (int, bool)
```

`BadContainer` calls `ordo.New` in a way that fails registration because a
factory argument is not a function, and returns the resulting error.

`IsNotAFunction` reports whether that error was caused by a non-function
factory, classified against the package's exported sentinel.

`FaultIndex` returns the index of the offending option, read off the concrete
registration fault, and false if the error carries no such fault.

Do not edit `verify_test.go`.
