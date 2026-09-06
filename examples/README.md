# Examples

Each example is a self-contained `package main` you can copy wholesale and run:

```bash
go run ./examples/simple
```

They are ordered below roughly as they are worth reading.

| Example | What it teaches | API it exercises |
| --- | --- | --- |
| [`simple`](./simple) | The smallest working container — bind an interface, register a constructor, resolve. | `WithService`, `WithFactory`, `MustGetService` |
| [`layered_app`](./layered_app) | A realistic config → storage → usecase → transport wiring, with a ready value and a slice of controllers. | `WithValue`, `WithService`, `WithFactory`, `MustGetService[[]T]` |
| [`multiple_implementations`](./multiple_implementations) | One service type registered several times; a slice dependency receives all of them, in registration order. | `WithService`, `MustGetService[[]T]` |
| [`verification`](./verification) | The core guarantee — a missing dependency and a cycle, both caught when the container is built rather than when it is used. | `VerificationError`, `MissingDependencyError`, `CircularDependencyError` |
| [`registration_errors`](./registration_errors) | Malformed options, aggregated into one error, each naming the `file:line` of the call responsible. | `RegistrationError`, `InvalidRegistrationError`, the `Err*` sentinels |
| [`keyed_services`](./keyed_services) | Telling several registrations of one type apart by name — and why a key never satisfies a constructor. | `WithKeyedService`, `WithKeyedFactory`, `WithKeyedValue`, `GetKeyedService` |
| [`factory_error`](./factory_error) | A constructor returning `(T, error)`; the failure reaches the caller wrapped in a `*di.DependencyError`. | `DependencyError` |

## The two failure phases

`verification` and `registration_errors` are the pair worth reading together. `NewContainer` checks options first and the graph second, and it never runs the second phase if the first one failed — a rejected registration is absent from the graph, so verifying it would report dependencies missing only because of the rejection.

Everything both phases report is an ordinary Go error. `NewContainer` does not panic.
