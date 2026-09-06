# Task

`wiring.go` in this package does not build a working container: `BuildContainer`
returns an error.

Diagnose the failure and fix `wiring.go` so the container builds and `*A`
resolves. `domain.go` offers more than one constructor for `*B` — read it before
deciding what to change. Do not edit `domain.go` or `verify_test.go`, and do not
change the signature of `BuildContainer`.
