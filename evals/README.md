# Evals

These tasks measure whether an AI coding agent can use `ordo` **correctly on the first
attempt**. They exist because the library's registration API is `any`-typed by necessity, so
the compiler cannot guide a caller — documentation and error text are the only channels into
an agent's build-fix loop, and this is how we find out whether they work.

## Layout

```
tasks/<name>/PROMPT.md        what the agent is asked to do
tasks/<name>/domain.go        types and constructors, written by hand
tasks/<name>/verify_test.go   the grading test, written by hand
testdata/reference/<name>/wiring.go
                              a known-good solution, kept under testdata
                              so the root module ignores it
```

Every task is its own Go module with a `replace` pointing at the repository, so the suite
always runs against the working tree rather than a published version.

## Running

```bash
./evals/run.sh              # every task
./evals/run.sh 04 07        # only matching tasks
```

| variable | meaning | default |
| --- | --- | --- |
| `ORDO_EVAL_MODEL` | model passed to the default agent | `sonnet` |
| `ORDO_EVAL_ROUNDS` | repair rounds allowed after the first attempt | `3` |
| `ORDO_EVAL_AGENT` | command run in the workspace, prompt on stdin | `claude -p …` |

`ORDO_EVAL_AGENT` is what makes the suite model-agnostic: any command that reads a prompt and
edits files in its working directory can be measured.

## How a task is graded

1. The task is copied to a scratch workspace. **`verify_test.go` is removed** while the agent
   works — the prompt states the contract, and the question is whether the documentation is
   enough to meet it, not whether the agent can read the assertions.
2. The agent runs with the prompt.
3. `domain.go` and `verify_test.go` are restored from the task, then
   `go build ./... && go vet ./... && go test ./...` runs. Restoring means an agent that
   edited the domain or the test to make things pass scores nothing for it.
4. On failure the compiler and test output are fed back and the agent tries again, up to
   `ORDO_EVAL_ROUNDS` times.

Reported per task: whether it passed on the **first** attempt, how many repair rounds it
needed, and — when it failed — the error text it stalled on. That last column is the point
of the exercise: it names the message or the missing document that has to improve.

The agent compiles against a snapshot of the library with `evals/` removed, so it cannot
read the reference solutions through the `replace` directive.

## Results

Run on 2026-09-06 against `claude-sonnet-5`, at commit `d4d3a5b`, with
`ORDO_EVAL_ROUNDS=1` and `ORDO_EVAL_TIMEOUT=300`.

| task | first attempt | iterations | final | stall cause |
| --- | --- | --- | --- | --- |
| `01-interface-binding` | yes | 0 | pass | — |
| `02-config-and-dependency` | yes | 0 | pass | — |
| `03-constructor-error` | yes | 0 | pass | — |
| `04-slice-consumer` | yes | 0 | pass | — |
| `05-resolve-all-and-last` | yes | 0 | pass | — |
| `06-keyed` | yes | 0 | pass | — |
| `07-missing-dependency` | yes | 0 | pass | — |
| `08-cycle` | yes | 0 | pass | — |
| `09-classify-failure` | yes | 0 | pass | — |
| `10-mock-in-test` | yes | 0 | pass | — |

**first-attempt pass rate: 10/10**

The generic-method API was expected to be the obstacle, since a method could not
declare its own type parameters before Go 1.27 and no model was trained on code
that does. It was not. Nothing in this run stalled on `c.GetService[T]()`.

What did matter was whether the agent could read the package at all. In earlier runs
the library snapshot sat outside the agent's sandbox, so it could reach neither the
source nor `go doc`; it guessed the API as `Provide` and `Bind` and produced nothing
that compiled. Vendoring the snapshot into the workspace changed the result from
unusable to 10/10 with no compiler round-trips. The documentation carries the API.

### How much to read into this

- One model, one run. Every cell is a single sample.
- The vendored snapshot includes `examples/`, which contains wiring close to several
  tasks. A published module carries `examples/` into the module cache too, so this
  matches what a real consumer has — but it means a task may be answered from an
  example rather than from `go doc`. Excluding `examples/` would isolate the package
  documentation specifically, and is the sharper experiment.
- Passing workspaces are deleted, so the code behind a pass is not retained for
  inspection. Only failures are kept.

## Keeping the suite honest

`./evals/verify-reference.sh` builds and tests every task against its reference solution. It
runs in CI. If a task fails there, the fixture has rotted against the library and any agent
result it produces is noise rather than a finding.
