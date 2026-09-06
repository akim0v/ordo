## 1. Registration error types

- [x] 1.1 Add `di/registration.go` with the six cause sentinels (`ErrNilRegistration`, `ErrFactoryNotFunction`, `ErrFactoryNoReturn`, `ErrFactoryTooManyReturns`, `ErrFactorySecondReturnNotErr`, `ErrNotAssignable`) and verify `go build ./...` succeeds
- [x] 1.2 Add `InvalidRegistrationError` (`Index`, `ServiceType`, `ValueType`, `Site`, `Err`) with `Error()` and `Unwrap() error`, and verify a unit test asserts the message names the index, both types and the `file:line`, and that `errors.Is` reaches the wrapped sentinel
- [x] 1.3 Add `RegistrationError` (`Faults []error`) with `Error()` and `Unwrap() []error`, and verify a unit test asserts the `di: container registration failed:` header, one `  - ` line per fault, and that `errors.Is`/`errors.As` traverse every fault
- [x] 1.4 Add the `callSite` type and its `runtime.Caller` helper, and verify a unit test calling the helper from a known line reports that file's base name and a non-zero line

## 2. Factory validation without panics

- [x] 2.1 Change `newServiceFactory` to return `(*serviceFactory, error)` with the four shape checks returning their sentinels instead of `log.Panicf`, and verify `di/service_factory_test.go` cases rewritten from `suite.Panics` to `errors.Is` assertions pass
- [x] 2.2 Update every internal caller of `newServiceFactory` to propagate the error, and verify `go build ./...` and `go vet ./...` are clean

## 3. Option application returns errors

- [x] 3.1 Change the `Option` interface method to `apply(*Container) error` and update `serviceFactoryOption`, `serviceInstanceOption` and `factoryOption` to return a fault instead of calling `log.Panicf`, returning on the first fault per option, and verify `go build ./...` succeeds
- [x] 3.2 Thread a `callSite` captured in each exported `With*` constructor (`WithService`, `WithKeyedService`, `WithValue`, `WithKeyedValue`, `WithFactory`, `WithKeyedFactory`) down into the option structs, and verify a test registering a bad service from a helper function reports the helper's line, not `NewContainer`'s
- [x] 3.3 Handle a nil factory-or-instance argument in `withServiceKey` before the `reflect.TypeOf(...).Kind()` branch so it yields `ErrNilRegistration`, and verify a test asserts `di.WithService[Repo](nil)` returns that fault from `NewContainer` rather than panicking

## 4. Construction wiring

- [x] 4.1 Update `newContainer` to apply options in order, tag each fault with the caller-relative option index, collect all faults, and keep its internal `WithValue(c)` registration out of the index numbering; verify a test with faults at option 0 and option 2 reports exactly those indexes
- [x] 4.2 Update `NewContainer` to return `(nil, *RegistrationError)` when any registration fault exists and to run `verify()` only when there are none, and verify a test with one malformed registration plus a dependent service reports only the registration fault and no missing-dependency fault
- [x] 4.3 Verify with a test that a container of sound registrations still constructs, verifies, and resolves exactly as before, and that a sound-registration/faulty-graph container still returns `*VerificationError`

## 5. Resolution panic fixes

- [x] 5.1 Add `ErrServiceTypeMismatch` and rewrite `getServiceKey` to return `var zero T` with the error on the failure path and to use a comma-ok assertion on the success path, and verify tests cover an unregistered interface type, a keyed interface miss, and an interface-typed service whose factory errors — all returning `(nil, err)` without panicking
- [x] 5.2 Replace `log.Panicf` in `MustGetService` and `MustGetKeyedService` with `panic(err)`, and verify a test recovers the panic, asserts the recovered value is the error the equivalent `GetService` call returns, and asserts nothing was written to the standard logger
- [x] 5.3 Verify `grep -rn --include='*.go' -E 'log\.Panicf|panic\(' di/` reports only the two `Must*` panic sites and no `log` import outside them

## 6. Test suite migration and docs

- [x] 6.1 Rewrite the registration `suite.Panics` cases in `di/container_test.go` as `NewContainer` error assertions covering every fault in the container-registration spec (non-function factory, no return, too many returns, non-error second return, factory return not assignable, instance not assignable, nil argument), and verify `go test ./...` passes
- [x] 6.2 Add a test asserting registration validation invokes no factory, using factories that record invocation, and verify it passes
- [x] 6.3 Add a test asserting two malformed registrations of the same service type produce faults distinguishable by index and site, and verify it passes
- [x] 6.4 Update `README.md` and the `examples/` programs where they document or rely on panicking registration, and verify `go vet ./...` and `go run ./examples/...` for each example still behave as documented
- [x] 6.5 Run `go test ./... -race` and `go vet ./...` and verify both are clean
