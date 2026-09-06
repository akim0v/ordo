#!/usr/bin/env bash
#
# Runs an AI coding agent against every eval task and reports, per task,
# whether it passed on the first attempt and how many repair rounds it needed.
#
#   ./evals/run.sh                 # every task
#   ./evals/run.sh 04 07           # only tasks whose name matches
#
# The agent under test is configurable:
#   ORDO_EVAL_AGENT   command receiving the prompt on stdin, run in the workspace
#   ORDO_EVAL_MODEL   model passed to the default agent      (default: sonnet)
#   ORDO_EVAL_ROUNDS  repair rounds after the first attempt  (default: 3)
#   ORDO_EVAL_TIMEOUT seconds before an agent call is killed    (default: 300)

set -uo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
MODEL="${ORDO_EVAL_MODEL:-sonnet}"
ROUNDS="${ORDO_EVAL_ROUNDS:-3}"
TIMEOUT="${ORDO_EVAL_TIMEOUT:-300}"
RESULTS="$(mktemp -d)/results.tsv"
KEPT="$(dirname "$RESULTS")/kept-workspaces.txt"

# The library the tasks compile against is a snapshot with evals/ removed, so the
# reference solutions are unreachable while the documentation is not. It is
# copied into each workspace rather than kept beside it: a sandboxed agent can
# only read paths under its working directory, and an agent that cannot read the
# package documentation is not being measured on the documentation.
SNAPSHOT="$(mktemp -d)/ordo"
mkdir -p "$SNAPSHOT"
tar -C "$REPO" --exclude=evals --exclude=.git --exclude=.idea --exclude=.claude -cf - . \
  | tar -C "$SNAPSHOT" -xf -

# run_with_timeout runs a command in the background and kills it after N
# seconds. macOS ships no timeout(1), so this is done by hand.
run_with_timeout() {
  local seconds="$1" stdin_file="$2"; shift 2

  # Backgrounding a job in a non-interactive shell points its stdin at
  # /dev/null, so the prompt is restored from a file rather than inherited.
  "$@" < "$stdin_file" &
  local pid=$!

  ( sleep "$seconds"; kill -TERM "$pid" 2>/dev/null ) &
  local watchdog=$!

  wait "$pid" 2>/dev/null
  local status=$?
  kill -TERM "$watchdog" 2>/dev/null
  return "$status"
}

# agent runs the model under test. The prompt arrives both as an argument and
# on stdin, so a replacement command may take it either way.
agent() {
  local prompt="$1"

  if [[ -n "${ORDO_EVAL_AGENT:-}" ]]; then
    $ORDO_EVAL_AGENT
  else
    claude -p --permission-mode bypassPermissions --model "$MODEL" "$prompt"
  fi
}

# grade restores the files the agent was told not to touch, then builds, vets
# and tests. Restoring means an agent that edited the domain or the test to make
# things pass does not score for it.
grade() {
  local work="$1" task="$2"
  cp "$task/domain.go" "$work/domain.go" 2>/dev/null
  cp "$task/verify_test.go" "$work/verify_test.go"

  # . rather than ./... so the vendored library copy under .ordo is not graded
  (cd "$work" && go build . 2>&1 && go vet . 2>&1 && go test . 2>&1)
}

printf 'task\tfirst_attempt\titerations\tfinal\tdetail\n' > "$RESULTS"

for task in "$REPO"/evals/tasks/*/; do
  slug="$(basename "$task")"

  if [[ $# -gt 0 ]]; then
    match=0
    for pat in "$@"; do [[ "$slug" == *"$pat"* ]] && match=1; done
    [[ $match -eq 1 ]] || continue
  fi

  echo "=== $slug"
  work="$(mktemp -d)"
  cp "$task"/* "$work"/

  # The library lives inside the workspace so the agent can read its docs.
  mkdir -p "$work/.ordo"
  cp -R "$SNAPSHOT"/. "$work/.ordo"/

  # Written rather than edited in place, so the script does not depend on a
  # particular sed dialect.
  cat > "$work/go.mod" <<GOMOD
module ordoeval/$slug

go 1.27.0

require github.com/akim0v/ordo v0.0.0

replace github.com/akim0v/ordo => ./.ordo
GOMOD

  # The verification test is hidden while the agent works: the prompt states the
  # contract, and the question is whether the docs are enough to meet it.
  rm -f "$work/verify_test.go"

  prompt="$(cat "$task/PROMPT.md")"
  first="no"; iterations=0; final="fail"; detail=""

  for (( round=0; round<=ROUNDS; round++ )); do
    # The agent's output is kept rather than discarded: when a task fails it is
    # the only evidence of whether the agent stalled, errored or ran out of turns.
    printf '%s\n' "$prompt" > "$work/.prompt"
    ( cd "$work" && run_with_timeout "$TIMEOUT" "$work/.prompt" agent "$prompt" ) \
      > "$work/agent-round$round.log" 2>&1
    agent_status=$?

    if [[ $agent_status -ne 0 ]]; then
      printf '    agent exited %s on round %s\n' "$agent_status" "$round"
    fi

    if output="$(grade "$work" "$task" 2>&1)"; then
      final="pass"
      [[ $round -eq 0 ]] && first="yes"
      iterations=$round
      break
    fi

    iterations=$round
    detail="$(printf '%s' "$output" | grep -m1 -E 'ordo:|undefined:|cannot|expected|FAIL|\.go:[0-9]+' | cut -c1-160)"
    rm -f "$work/verify_test.go"

    prompt="The previous attempt did not pass. Fix the code in this directory.

go build ./... && go vet ./... && go test ./... reported:

$output"
  done

  printf '%s\t%s\t%s\t%s\t%s\n' "$slug" "$first" "$iterations" "$final" "$detail" >> "$RESULTS"
  printf '    %s after %s iteration(s)\n' "$final" "$iterations"

  # A failed workspace is kept: the code the agent actually wrote says more
  # about why it stalled than the last line of compiler output does.
  if [[ "$final" == "pass" ]]; then
    rm -rf "$work"
  else
    printf '    workspace kept at %s\n' "$work"
    echo "$work" >> "$KEPT"
  fi
done

echo
echo "| task | first attempt | iterations | final | stall cause |"
echo "| --- | --- | --- | --- | --- |"
tail -n +2 "$RESULTS" | while IFS=$'\t' read -r t f i r d; do
  echo "| \`$t\` | $f | $i | $r | ${d:-—} |"
done

passed=$(tail -n +2 "$RESULTS" | awk -F'\t' '$2=="yes"' | wc -l | tr -d ' ')
total=$(tail -n +2 "$RESULTS" | wc -l | tr -d ' ')
echo
echo "first-attempt pass rate: $passed/$total"
